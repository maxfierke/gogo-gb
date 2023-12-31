package main

import (
	"flag"
	"log"
	"os"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/hardware"
)

type CLIOptions struct {
	cartPath   string
	debugger   string
	debugPrint string
}

func main() {
	options := CLIOptions{}

	flag.StringVar(&options.cartPath, "cart", "", "Path to cartridge file (.gb, .gbc)")
	flag.StringVar(&options.debugger, "debugger", "none", "Specify debugger to use (\"none\", \"gameboy-doctor\")")
	flag.StringVar(&options.debugPrint, "debug-print", "", "Print out something for debugging purposes. (\"cart-header\", \"opcodes\")")
	flag.Parse()

	if options.debugPrint != "" {
		debugPrint(&options)
	} else {
		runCart(&options)
	}
}

func debugPrint(options *CLIOptions) {
	switch options.debugPrint {
	case "cart-header":
		debugPrintCartHeader(options)
	case "opcodes":
		debugPrintOpcodes(options)
	default:
		log.Fatalf("unrecognized \"debug-print\" option: %v\n", options.debugPrint)
	}
}

func debugPrintCartHeader(options *CLIOptions) {
	cartFile, err := os.Open(options.cartPath)
	if options.cartPath == "" || err != nil {
		log.Fatalf("Unable to load cartridge. Please ensure it's inserted correctly (exists): %v\n", err)
	}
	defer cartFile.Close()

	cartReader, err := cart.NewReader(cartFile)
	if err == cart.ErrHeader {
		log.Printf("Warning: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		log.Fatalf("Unable to load cartridge. Please ensure it's inserted correctly or trying blowing on it: %v\n", err)
	}

	cartReader.Header.DebugPrint()
}

func debugPrintOpcodes(options *CLIOptions) {
	opcodes, err := isa.LoadOpcodes()
	if err != nil {
		log.Fatalf("Unable to load opcodes: %v\n", err)
	}

	opcodes.DebugPrint()
}

func initDMG(options *CLIOptions) *hardware.DMG {
	debugger, err := debug.NewDebugger(options.debugger)
	if err != nil {
		log.Fatalf("Unable to initialize Debugger: %v\n", err)
	}

	dmg, err := hardware.NewDMGDebug(debugger)
	if err != nil {
		log.Fatalf("Unable to initialize DMG: %v\n", err)
	}
	return dmg
}

func loadCart(dmg *hardware.DMG, options *CLIOptions) {
	if options.cartPath == "" {
		return
	}

	cartFile, err := os.Open(options.cartPath)
	if options.cartPath == "" || err != nil {
		log.Fatalf("Unable to load cartridge. Please ensure it's inserted correctly (e.g. file exists): %v\n", err)
	}
	defer cartFile.Close()

	cartReader, err := cart.NewReader(cartFile)
	if err == cart.ErrHeader {
		log.Printf("Warning: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		log.Fatalf("Unable to load cartridge. Please ensure it's inserted correctly (e.g. file exists): %v\n", err)
	}

	err = dmg.LoadCartridge(cartReader)
	if err == cart.ErrHeader {
		log.Printf("Warning: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
	} else if err != nil {
		log.Fatalf("Unable to load cartridge: %v\n", err)
	}
}

func runCart(options *CLIOptions) {
	dmg := initDMG(options)
	loadCart(dmg, options)
	dmg.Run()
}
