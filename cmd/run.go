package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cart/mbc"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
	"github.com/maxfierke/gogo-gb/host"
	"github.com/spf13/cobra"
)

type RunCmdOptions struct {
	bootRomPath  string
	cartPath     string
	cartSavePath string
	debugger     string
	headless     bool
	serialPort   string
	skipBootRom  bool
}

var runCmdOptions = RunCmdOptions{}

var runCmd = &cobra.Command{
	Use:   "run [path to cartridge]",
	Short: "Run a cartridge",
	Long: `Run a cartridge under emulation.

Options can be specified to attach a debugger, control peripherals, and specify paths for saves and the boot ROM.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := getLogger(cmd)
		if err != nil {
			return fmt.Errorf("getting logger: %w", err)
		}

		cartPath := args[0]
		runCmdOptions.cartPath = cartPath

		logger.Println("welcome to gogo-gb, the go-getting GB emulator")
		if err := runCart(logger, &runCmdOptions); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&runCmdOptions.bootRomPath, "bootrom", "", "Path to boot ROM file (dmg_bios.bin, etc.). Defaults to a lookup on common boot ROM filenames in current directory")
	_ = runCmd.MarkFlagFilename("bootrom", ".bin", ".rom")

	runCmd.Flags().StringVarP(&runCmdOptions.cartSavePath, "save", "s", "", "Path to cartridge save file (.sav). Defaults to a .sav file with the same name as the cartridge file")
	_ = runCmd.MarkFlagFilename("save", ".sav")

	runCmd.Flags().StringVarP(&runCmdOptions.debugger, "debugger", "d", "", "Specify debugger to use (\"gameboy-doctor\", \"interactive\")")
	runCmd.Flags().StringVarP(&runCmdOptions.serialPort, "serial-port", "p", "", "Path to serial port IO (could be a file, UNIX socket, etc.)")
	runCmd.Flags().BoolVar(&runCmdOptions.skipBootRom, "skip-bootrom", false, "Skip loading a boot ROM")
	runCmd.Flags().BoolVar(&runCmdOptions.headless, "headless", false, "Launch without UI")
}

var DEFAULT_BOOT_ROM_PATHS = []string{
	"gb_bios.bin",
	"dmg_bios.bin",
	"mgb_bios.bin",
	"dmg0_bios.bin",
}

func getCartSaveFilePath(options *RunCmdOptions) string {
	cartSaveFilePath := options.cartSavePath

	if cartSaveFilePath == "" && options.cartPath != "" {
		cartSaveDir := filepath.Dir(options.cartPath)
		cartSaveFileName := strings.Replace(
			filepath.Base(options.cartPath),
			filepath.Ext(options.cartPath),
			".sav",
			1,
		)

		cartSaveFilePath = filepath.Join(cartSaveDir, cartSaveFileName)
	}

	return cartSaveFilePath
}

func initHost(logger *log.Logger, options *RunCmdOptions) (host.Host, error) {
	var hostDevice host.Host

	if options.headless {
		hostDevice = host.NewCLIHost()
	} else {
		hostDevice = host.NewUIHost()
	}

	hostDevice.SetLogger(logger)

	if options.serialPort != "" {
		serialCable := devices.NewHostSerialCable()

		switch options.serialPort {
		case "stdout", "/dev/stdout":
			serialCable.SetWriter(os.Stdout)
		case "stderr", "/dev/stderr":
			serialCable.SetWriter(os.Stderr)
		default:
			serialPort, err := os.Create(options.serialPort)
			if err != nil {
				return nil, fmt.Errorf("unable to open file '%s' as serial port: %w", options.serialPort, err)
			}

			serialCable.SetReader(serialPort)
			serialCable.SetWriter(serialPort)
		}

		hostDevice.AttachSerialCable(serialCable)
	}

	return hostDevice, nil
}

func initDMG(logger *log.Logger, options *RunCmdOptions) (*hardware.DMG, error) {
	debugger, err := debug.NewDebugger(options.debugger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Debugger: %w", err)
	}

	opts := []hardware.ConsoleOption{
		hardware.WithDebugger(debugger),
	}

	if options.skipBootRom {
		opts = append(opts, hardware.WithFakeBootROM())
	} else {
		bootRomFile, err := loadBootROM(logger, options)
		if err != nil {
			return nil, fmt.Errorf("unable to load boot ROM: %w", err)
		}
		if bootRomFile == nil {
			opts = append(opts, hardware.WithFakeBootROM())
		} else {
			defer bootRomFile.Close()
			opts = append(opts, hardware.WithBootROM(bootRomFile))
		}
	}

	dmg, err := hardware.NewDMG(opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize DMG: %w", err)
	}

	return dmg, nil
}

func loadBootROM(logger *log.Logger, options *RunCmdOptions) (*os.File, error) {
	bootRomPath := options.bootRomPath

	var bootRomFile *os.File
	var err error

	if bootRomPath == "" {
		for _, romPath := range DEFAULT_BOOT_ROM_PATHS {
			if bootRomFile, err = os.Open(romPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			} else if bootRomFile != nil {
				// yay! we found one!
				break
			}
		}
	} else if bootRomFile, err = os.Open(bootRomPath); err != nil {
		return nil, err
	}

	if bootRomFile == nil {
		// Bail out if no boot ROM loaded
		logger.Printf("WARN: No boot ROM provided. Some emulation functionality may be incorrect")
		return nil, nil
	}

	logger.Printf("loaded boot ROM: %s\n", bootRomFile.Name())

	return bootRomFile, nil
}

func loadCart(dmg *hardware.DMG, logger *log.Logger, options *RunCmdOptions) error {
	if options.cartPath == "" {
		return nil
	}

	cartFile, err := os.Open(options.cartPath)
	if options.cartPath == "" || err != nil {
		return fmt.Errorf("unable to load cartridge. Please ensure it's inserted correctly (e.g. file exists): %w", err)
	}
	defer cartFile.Close()

	err = dmg.LoadCartridge(cartFile)
	if errors.Is(err, cart.ErrChecksum) {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		return fmt.Errorf("unable to load cartridge: %w", err)
	}

	return nil
}

func loadCartSave(dmg *hardware.DMG, logger *log.Logger, options *RunCmdOptions) error {
	cartSaveFilePath := getCartSaveFilePath(options)

	cartSaveFile, err := os.Open(cartSaveFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to open or create cartridge save file: %w", err)
	}

	if cartSaveFile != nil {
		defer cartSaveFile.Close()

		err = dmg.LoadSave(cartSaveFile)
		if err != nil {
			switch {
			case errors.Is(err, mbc.ErrMBC3BadClockBattery):
				logger.Printf("WARN: Unable to load RTC data from save. In-game clock may be incorrect")
			default:
				return fmt.Errorf("unable to load cartridge save: %w", err)
			}
		}

		logger.Printf("Loaded cartridge save from %s\n", cartSaveFilePath)
	}

	return nil
}

func saveCart(dmg *hardware.DMG, logger *log.Logger, options *RunCmdOptions) error {
	cartSaveFilePath := getCartSaveFilePath(options)

	cartSaveFile, err := os.OpenFile(cartSaveFilePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open or create cartridge save file: %w", err)
	}
	defer cartSaveFile.Close()

	err = dmg.Save(cartSaveFile)
	if err != nil {
		return fmt.Errorf("unable to write cartridge save file: %w", err)
	}

	logger.Printf("Saved cartridge save to %s\n", cartSaveFilePath)

	return nil
}

func runCart(logger *log.Logger, options *RunCmdOptions) error {
	hostDevice, err := initHost(logger, options)
	if err != nil {
		return fmt.Errorf("unable to initialize host device: %w", err)
	}

	dmg, err := initDMG(logger, options)
	if err != nil {
		return fmt.Errorf("initializing DMG: %w", err)
	}

	err = loadCart(dmg, logger, options)
	if err != nil {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	if dmg.CartridgeHeader().SupportsSaving() {
		err := loadCartSave(dmg, logger, options)
		if err != nil {
			return fmt.Errorf("loading cartridge save: %w", err)
		}

		defer func() {
			err := saveCart(dmg, logger, options)
			if err != nil {
				logger.Printf("WARN: Error occurred while saving: %s", err.Error())
			}
		}()
	}

	err = hostDevice.Run(dmg)
	if err != nil {
		return fmt.Errorf("running emulation: %w", err)
	}

	return nil
}
