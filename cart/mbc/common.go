package mbc

import "github.com/maxfierke/gogo-gb/mem"

const (
	RAM_BANK_SIZE = 0x2000
	ROM_BANK_SIZE = 0x4000
)

func readBankAddr(memory []byte, banks_region mem.MemRegion, bank_size uint16, current_bank uint16, addr uint16) byte {
	bank_base_addr := current_bank * bank_size
	bank_slot_addr := addr - banks_region.Start
	return memory[bank_base_addr+bank_slot_addr]
}

func writeBankAddr(memory []byte, banks_region mem.MemRegion, bank_size uint16, current_bank uint16, addr uint16, value byte) {
	bank_base_addr := current_bank * bank_size
	bank_slot_addr := addr - banks_region.Start
	memory[bank_base_addr+bank_slot_addr] = value
}
