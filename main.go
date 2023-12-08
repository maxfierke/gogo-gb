package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/maxfierke/gogo-gb/cpu"
)

func main() {
	fmt.Println("welcome to gogo-gb")

	debugPrintPtr := flag.String("debug-print", "", "Print out something for debugging purposes. Currently just 'opcodes'")
	flag.Parse()

	opcodes, err := cpu.LoadOpcodes("./opcodes.json")
	if err != nil {
		log.Fatalf("Unable to load opcodes: %v\n", err)
	}

	if debugPrintPtr != nil {
		if *debugPrintPtr == "opcodes" {
			printOpcodes(opcodes)
		}
	}
}

func printOpcodes(opcodes *cpu.Opcodes) {
	fmt.Println("== Opcodes ==")

	fmt.Printf("=== Unprefixed: \n\n")
	for k := range opcodes.Unprefixed {
		fmt.Printf("%s %s\n", k, opcodes.Unprefixed[k].String())
	}

	fmt.Printf("\n=== Cbprefixed: \n\n")
	for k := range opcodes.CbPrefixed {
		fmt.Printf("%s %s\n", k, opcodes.CbPrefixed[k].String())
	}
}
