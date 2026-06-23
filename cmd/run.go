package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
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
	model        string
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

		ctx := context.Background()

		cartPath := args[0]
		runCmdOptions.cartPath = cartPath

		logger.Println("welcome to gogo-gb, the go-getting GB emulator")
		if err := runCart(ctx, logger, &runCmdOptions); err != nil {
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
	runCmd.Flags().StringVarP(&runCmdOptions.model, "model", "m", "auto", "Specify model to use (\"auto\", \"dmg\", \"cgb\")")
	runCmd.Flags().StringVarP(&runCmdOptions.serialPort, "serial-port", "p", "", "Path to serial port IO (unix domain socket)")
	runCmd.Flags().BoolVar(&runCmdOptions.skipBootRom, "skip-bootrom", false, "Skip loading a boot ROM")
	runCmd.Flags().BoolVar(&runCmdOptions.headless, "headless", false, "Launch without UI")
}

var DEFAULT_DMG_BOOT_ROM_PATHS = []string{
	"gb_bios.bin",
	"dmg_bios.bin",
	"mgb_bios.bin",
	"dmg0_bios.bin",
	"gb_boot.bin",
	"dmg_boot.bin",
	"mgb_boot.bin",
	"dmg0_boot.bin",
}

var DEFAULT_CGB_BOOT_ROM_PATHS = []string{
	"cgb_bios.bin",
	"cgb0_bios.bin",
	"gbc_bios.bin",
	"gbc_boot.bin",
	"cgb_boot.bin",
	"cgb0_boot.bin",
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

func initHost(ctx context.Context, logger *log.Logger, options *RunCmdOptions) (host.Host, error) {
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
			serialCable.SetWriter(ctx, os.Stdout)
		case "stderr", "/dev/stderr":
			serialCable.SetWriter(ctx, os.Stderr)
		default:
			if info, _ := os.Stat(options.serialPort); info != nil && info.Mode().Type() == os.ModeSocket {
				if err := os.Remove(options.serialPort); err != nil {
					return nil, fmt.Errorf("unable to remove existing unix socket '%s': %w", options.serialPort, err)
				}
			}

			listener, err := net.Listen("unix", options.serialPort)
			if err != nil {
				return nil, fmt.Errorf("unable to open unix socket '%s' as serial port: %w", options.serialPort, err)
			}

			go func() {
				defer os.RemoveAll(options.serialPort)

				for {
					if ctx.Err() != nil {
						return
					}

					conn, err := listener.Accept()
					if err != nil {
						logger.Printf("ERROR: unable to accept connections to serial port domain socket: %v", err)

						return
					}
					defer conn.Close()

					serialCable.SetReader(ctx, conn)
					serialCable.SetWriter(ctx, conn)
				}
			}()
		}

		hostDevice.AttachSerialCable(serialCable)
	}

	return hostDevice, nil
}

func initConsole(cartridge *cart.Cartridge, logger *log.Logger, options *RunCmdOptions) (hardware.Console, error) {
	var model hardware.ConsoleModel
	switch options.model {
	case "auto":
		{
			switch cartridge.Header.Cgb() {
			case cart.CGB_COLOR_ONLY, cart.CGB_COLOR_ENHANCED:
				model = hardware.ConsoleModelCGB
			case cart.CGB_COLOR_NONE:
				model = hardware.ConsoleModelDMG
			default:
				return nil, errors.New("unable to auto-detect model. Please specify with --model/-m")
			}

			logger.Printf("using model %s based on cartridge header\n", model)
		}
	case "dmg":
		model = hardware.ConsoleModelDMG
	case "cgb":
		model = hardware.ConsoleModelCGB
	default:
		return nil, fmt.Errorf("unrecognized model: %s", options.model)
	}

	debugger, err := debug.NewDebugger(options.debugger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize debugger: %w", err)
	}

	opts := []hardware.ConsoleOption{
		hardware.WithCartridge(cartridge),
		hardware.WithDebugger(debugger),
	}

	if options.skipBootRom {
		opts = append(opts, hardware.WithFakeBootROM())
	} else {
		bootRomFile, err := loadBootROM(model, logger, options)
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

	console, err := hardware.NewConsole(
		model,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return console, nil
}

func loadBootROM(model hardware.ConsoleModel, logger *log.Logger, options *RunCmdOptions) (*os.File, error) {
	bootRomPath := options.bootRomPath

	var bootRomFile *os.File
	var err error

	if bootRomPath == "" {
		lookupPaths := DEFAULT_DMG_BOOT_ROM_PATHS
		if model == hardware.ConsoleModelCGB {
			lookupPaths = DEFAULT_CGB_BOOT_ROM_PATHS
		}

		for _, romPath := range lookupPaths {
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

func loadCart(logger *log.Logger, options *RunCmdOptions) (*cart.Cartridge, error) {
	if options.cartPath == "" {
		return cart.NewCartridge(), nil
	}

	cartFile, err := os.Open(options.cartPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load cartridge. Please ensure it's inserted correctly (e.g. file exists): %w", err)
	}
	defer cartFile.Close()

	cartridge, err := cart.LoadCartridge(cartFile)
	if errors.Is(err, cart.ErrChecksum) {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		return nil, fmt.Errorf("unable to load cartridge: %w", err)
	}

	logger.Printf("loaded cartridge from %s\n", options.cartPath)

	return cartridge, nil
}

func loadCartSave(cartridge *cart.Cartridge, logger *log.Logger, options *RunCmdOptions) error {
	cartSaveFilePath := getCartSaveFilePath(options)

	cartSaveFile, err := os.Open(cartSaveFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to open or create cartridge save file: %w", err)
	}

	if cartSaveFile != nil {
		defer cartSaveFile.Close()

		err = cartridge.LoadSave(cartSaveFile)
		if err != nil {
			switch {
			case errors.Is(err, mbc.ErrMBC3BadClockBattery):
				logger.Printf("WARN: Unable to load RTC data from save. In-game clock may be incorrect")
			default:
				return fmt.Errorf("unable to load cartridge save: %w", err)
			}
		}

		logger.Printf("loaded cartridge save from %s\n", cartSaveFilePath)
	}

	return nil
}

func saveCart(cartridge *cart.Cartridge, logger *log.Logger, options *RunCmdOptions) error {
	cartSaveFilePath := getCartSaveFilePath(options)

	cartSaveFile, err := os.OpenFile(cartSaveFilePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open or create cartridge save file: %w", err)
	}
	defer cartSaveFile.Close()

	err = cartridge.Save(cartSaveFile)
	if err != nil {
		return fmt.Errorf("unable to write cartridge save file: %w", err)
	}

	logger.Printf("Saved cartridge save to %s\n", cartSaveFilePath)

	return nil
}

func runCart(ctx context.Context, logger *log.Logger, options *RunCmdOptions) error {
	cartridge, err := loadCart(logger, options)
	if err != nil {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	consoleHost, err := initHost(ctx, logger, options)
	if err != nil {
		return fmt.Errorf("unable to initialize host device: %w", err)
	}

	console, err := initConsole(cartridge, logger, options)
	if err != nil {
		return fmt.Errorf("initializing console: %w", err)
	}

	if cartridge.Header.SupportsSaving() {
		err := loadCartSave(cartridge, logger, options)
		if err != nil {
			return fmt.Errorf("loading cartridge save: %w", err)
		}

		defer func() {
			err := saveCart(cartridge, logger, options)
			if err != nil {
				logger.Printf("WARN: Error occurred while saving: %s", err.Error())
			}
		}()
	}

	err = consoleHost.Run(console)
	if err != nil {
		return fmt.Errorf("running emulation: %w", err)
	}

	return nil
}
