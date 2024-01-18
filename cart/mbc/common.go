package mbc

import "github.com/maxfierke/gogo-gb/mem"

const (
	RAM_BANK_SIZE = 0x2000
	ROM_BANK_SIZE = 0x4000
)

func readBankAddr(memory []byte, banksRegion mem.MemRegion, bankSize uint16, currentBank uint16, addr uint16) byte {
	bankBaseAddr := currentBank * bankSize
	bankSlotAddr := addr - banksRegion.Start
	return memory[bankBaseAddr+bankSlotAddr]
}

func writeBankAddr(memory []byte, banksRegion mem.MemRegion, bankSize uint16, currentBank uint16, addr uint16, value byte) {
	bankBaseAddr := currentBank * bankSize
	bankSlotAddr := addr - banksRegion.Start
	memory[bankBaseAddr+bankSlotAddr] = value
}
