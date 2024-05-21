package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
	"github.com/maxfierke/gogo-gb/host"
)

type CLIOptions struct {
	cartPath   string
	debugger   string
	debugPrint string
	logPath    string
	logger     *log.Logger
	serialPort string
	ui         bool
}

const LOG_PREFIX = ""

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
	flag.StringVar(&options.cartPath, "cart", "", "Path to cartridge file (.gb, .gbc)")
	flag.StringVar(&options.serialPort, "serial-port", "", "Path to serial port IO (could be a file, UNIX socket, etc.)")
	flag.StringVar(&options.debugger, "debugger", "none", "Specify debugger to use (\"none\", \"gameboy-doctor\")")
	flag.StringVar(&options.debugPrint, "debug-print", "", "Print out something for debugging purposes (\"cart-header\", \"opcodes\")")
	flag.StringVar(&options.logPath, "log", "", "Path to log file. Default/empty implies stdout")
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
	if err == cart.ErrHeader {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		logger.Fatalf("ERROR: Unable to load cartridge. Please ensure it's inserted correctly or trying blowing on it: %v\n", err)
	}

	cartReader.Header.DebugPrint(logger)
}

func debugPrintOpcodes(options *CLIOptions) {
	logger := options.logger

	opcodes, err := isa.LoadOpcodes()
	if err != nil {
		logger.Fatalf("ERROR: Unable to load opcodes: %v\n", err)
	}

	opcodes.DebugPrint(logger)
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

	dmg, err := hardware.NewDMG(
		hardware.WithDebugger(debugger),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize DMG: %w", err)
	}

	return dmg, nil
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

	cartReader, err := cart.NewReader(cartFile)
	if err == cart.ErrHeader {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		return fmt.Errorf("unable to load cartridge. Please ensure it's inserted correctly (e.g. file exists): %w", err)
	}

	err = dmg.LoadCartridge(cartReader)
	if err == cart.ErrHeader {
		logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		return fmt.Errorf("unable to load cartridge: %w", err)
	}

	return nil
}

func runCart(options *CLIOptions) error {
	hostDevice, err := initHost(options)
	if err != nil {
		return fmt.Errorf("unable to initialize host device: %w", err)
	}

	dmg, err := initDMG(options)
	if err != nil {
		return err
	}

	if err := loadCart(dmg, options); err != nil {
		return err
	}

	err = hostDevice.Run(dmg)
	if err != nil {
		return err
	}

	return nil
}
