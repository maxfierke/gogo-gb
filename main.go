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

func main() {
	cartPathPtr := flag.String("cart", "", "Path to cartridge file (.gb, .gbc)")
	debuggerPtr := flag.String("debugger", "none", "Specify debugger to use (\"none\", \"gameboy-doctor\")")
	debugPrintPtr := flag.String("debug-print", "", "Print out something for debugging purposes. (\"cart-header\", \"opcodes\")")
	flag.Parse()

	if debugPrintPtr != nil && *debugPrintPtr != "" {
		if *debugPrintPtr == "cart-header" {
			cartFile, err := os.Open(*cartPathPtr)
			if *cartPathPtr == "" || err != nil {
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

		if *debugPrintPtr == "opcodes" {
			opcodes, err := isa.LoadOpcodes()
			if err != nil {
				log.Fatalf("Unable to load opcodes: %v\n", err)
			}

			opcodes.DebugPrint()
		}
	} else {
		debugger, err := debug.NewDebugger(*debuggerPtr)
		if err != nil {
			log.Fatalf("Unable to initialize Debugger: %v\n", err)
		}

		dmg, err := hardware.NewDMGDebug(debugger)
		if err != nil {
			log.Fatalf("Unable to initialize DMG: %v\n", err)
		}

		cartFile, err := os.Open(*cartPathPtr)
		if *cartPathPtr == "" || err != nil {
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

		dmg.Run()
	}
}
