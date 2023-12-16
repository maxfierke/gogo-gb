package mbc

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	mbc0_rom_bank_start = 0x0000
	mbc0_rom_bank_end   = 0x7FFF
	mbc0_ram_bank_start = 0xA000
	mbc0_ram_bank_end   = 0xBFFF
)

type MBC0 struct {
	rom []byte
}

func NewMBC0(rom []byte) *MBC0 {
	return &MBC0{rom: rom}
}

func (m *MBC0) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr <= mbc0_rom_bank_end {
		return mem.ReadReplace(m.rom[addr])
	}

	return mem.ReadPassthrough()
}

func (m *MBC0) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr <= mbc0_rom_bank_end {
		// Put the Read-Only in ROM
		return mem.WriteBlock()
	}

	if addr >= mbc0_ram_bank_start && addr <= mbc0_ram_bank_end {
		// RAM is RAM and this is a fake cartridge, so...
		return mem.WritePassthrough()
	}

	panic(fmt.Sprintf("Attempting to write 0x%x @ 0x%x, which is out-of-bounds for MBC0", value, addr))
}
