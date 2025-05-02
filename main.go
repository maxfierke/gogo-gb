package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cart/mbc"
	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
	"github.com/maxfierke/gogo-gb/host"
)

type CLIOptions struct {
	bootRomPath  string
	cartPath     string
	cartSavePath string
	debugger     string
	debugPrint   string
	logPath      string
	logger       *log.Logger
	serialPort   string
	skipBootRom  bool
	ui           bool
}

const LOG_PREFIX = ""

var DEFAULT_BOOT_ROM_PATHS = []string{
	"gb_bios.bin",
	"dmg_bios.bin",
	"mgb_bios.bin",
	"dmg0_bios.bin",
}

func main() {
	options := CLIOptions{}

	parseOptions(&options)

	if options.logPath == "" || options.logPath == "stdout" {
		options.logger = log.New(os.Stdout, LOG_PREFIX, log.LstdFlags)
	} else if options.logPath == "stderr" {
		options.logger = log.New(os.Stderr, LOG_PREFIX, log.LstdFlags)
	} else {
		logFile, err := os.Create(options.logPath)
		if err != nil {
			log.Fatalf("ERROR: Unable to open log file '%s' for writing: %v\n", options.logPath, err)
		}
		defer logFile.Close()

		options.logger = log.New(logFile, LOG_PREFIX, log.LstdFlags)
	}

	if options.debugPrint != "" {
		debugPrint(&options)
	} else {
		options.logger.Println("welcome to gogo-gb, the go-getting GB emulator")
		if err := runCart(&options); err != nil {
			options.logger.Fatalf("ERROR: %v\n", err)
		}
	}
}

func parseOptions(options *CLIOptions) {
	flag.StringVar(&options.bootRomPath, "bootrom", "", "Path to boot ROM file (dmg_bios.bin, mgb_bios.bin, etc.). Defaults to a lookup on common boot ROM filenames in current directory")
	flag.StringVar(&options.cartPath, "cart", "", "Path to cartridge file (.gb, .gbc)")
	flag.StringVar(&options.cartSavePath, "cart-save", "", "Path to cartridge save file (.sav). Defaults to a .sav file with the same name as the cartridge file")
	flag.StringVar(&options.debugger, "debugger", "none", "Specify debugger to use (\"none\", \"gameboy-doctor\", \"interactive\")")
	flag.StringVar(&options.debugPrint, "debug-print", "", "Print out something for debugging purposes (\"cart-header\", \"opcodes\")")
	flag.StringVar(&options.logPath, "log", "", "Path to log file. Default/empty implies stdout")
	flag.StringVar(&options.serialPort, "serial-port", "", "Path to serial port IO (could be a file, UNIX socket, etc.)")
	flag.BoolVar(&options.skipBootRom, "skip-bootrom", false, "Skip loading a boot ROM")
	flag.BoolVar(&options.ui, "ui", false, "Launch with UI")
	flag.Parse()
}

func debugPrint(options *CLIOptions) {
	switch options.debugPrint {
	case "cart-header":
		debugPrintCartHeader(options)
	case "opcodes":
		debugPrintOpcodes(options)
	default:
		options.logger.Fatalf("ERROR: unrecognized \"debug-print\" option: %v\n", options.debugPrint)
	}
}

func debugPrintCartHeader(options *CLIOptions) {
	logger := options.logger

	cartFile, err := os.Open(options.cartPath)
	if options.cartPath == "" || err != nil {
		logger.Fatalf("ERROR: Unable to load cartridge. Please ensure it's inserted correctly (exists): %v\n", err)
	}
	defer cartFile.Close()

	cartReader, err := cart.NewReader(cartFile)
	if errors.Is(err, cart.ErrChecksum) {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		logger.Fatalf("ERROR: Unable to load cartridge. Please ensure it's inserted correctly or trying blowing on it: %v\n", err)
	}

	cartReader.Header.DebugPrint(logger.Writer())
}

func debugPrintOpcodes(options *CLIOptions) {
	logger := options.logger

	opcodes, err := isa.LoadOpcodes()
	if err != nil {
		logger.Fatalf("ERROR: Unable to load opcodes: %v\n", err)
	}

	opcodes.DebugPrint(logger.Writer())
}

func getCartSaveFilePath(options *CLIOptions) string {
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

func initHost(options *CLIOptions) (host.Host, error) {
	var hostDevice host.Host

	if options.ui {
		hostDevice = host.NewUIHost()
	} else {
		hostDevice = host.NewCLIHost()
	}

	hostDevice.SetLogger(options.logger)

	if options.serialPort != "" {
		serialCable := devices.NewHostSerialCable()

		if options.serialPort == "stdout" || options.serialPort == "/dev/stdout" {
			serialCable.SetWriter(os.Stdout)
		} else if options.serialPort == "stderr" || options.serialPort == "/dev/stderr" {
			serialCable.SetWriter(os.Stderr)
		} else {
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

func initDMG(options *CLIOptions) (*hardware.DMG, error) {
	debugger, err := debug.NewDebugger(options.debugger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Debugger: %w", err)
	}

	opts := []hardware.DMGOption{
		hardware.WithDebugger(debugger),
	}

	if options.skipBootRom {
		opts = append(opts, hardware.WithFakeBootROM())
	} else {
		bootRomFile, err := loadBootROM(options)
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

func loadBootROM(options *CLIOptions) (*os.File, error) {
	logger := options.logger
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

	logger.Printf("Loaded boot ROM: %s\n", bootRomFile.Name())

	return bootRomFile, nil
}

func loadCart(dmg *hardware.DMG, options *CLIOptions) error {
	if options.cartPath == "" {
		return nil
	}

	logger := options.logger

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

func loadCartSave(dmg *hardware.DMG, options *CLIOptions) error {
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
				options.logger.Printf("WARN: Unable to load RTC data from save. In-game clock may be incorrect")
			default:
				return fmt.Errorf("unable to load cartridge save: %w", err)
			}
		}

		options.logger.Printf("Loaded cartridge save from %s\n", cartSaveFilePath)
	}

	return nil
}

func saveCart(dmg *hardware.DMG, options *CLIOptions) error {
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

	options.logger.Printf("Saved cartridge save to %s\n", cartSaveFilePath)

	return nil
}

func runCart(options *CLIOptions) error {
	hostDevice, err := initHost(options)
	if err != nil {
		return fmt.Errorf("unable to initialize host device: %w", err)
	}

	dmg, err := initDMG(options)
	if err != nil {
		return fmt.Errorf("initializing DMG: %w", err)
	}

	err = loadCart(dmg, options)
	if err != nil {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	if dmg.CartridgeHeader().SupportsSaving() {
		err := loadCartSave(dmg, options)
		if err != nil {
			return fmt.Errorf("loading cartridge save: %w", err)
		}

		defer func() {
			err := saveCart(dmg, options)
			if err != nil {
				options.logger.Printf("WARN: Error occurred while saving: %s", err.Error())
			}
		}()
	}

	err = hostDevice.Run(dmg)
	if err != nil {
		return fmt.Errorf("running emulation: %w", err)
	}

	return nil
}
