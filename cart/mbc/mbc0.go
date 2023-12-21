package mbc

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	mbc0_rom_bank = mem.MemRegion{Start: 0x0000, End: 0x7FFF}
	mbc0_ram_bank = mem.MemRegion{Start: 0xA000, End: 0xBFFF}
)

type MBC0 struct {
	rom []byte
}

func NewMBC0(rom []byte) *MBC0 {
	return &MBC0{rom: rom}
}

func (m *MBC0) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr <= mbc0_rom_bank.End {
		return mem.ReadReplace(m.rom[addr])
	}

	return mem.ReadPassthrough()
}

func (m *MBC0) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr <= mbc0_rom_bank.End {
		// Put the Read-Only in ROM
		return mem.WriteBlock()
	}

	if mbc0_ram_bank.Contains(addr, false) {
		// RAM is RAM and this is a fake cartridge, so...
		return mem.WritePassthrough()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC0", value, addr))
}
