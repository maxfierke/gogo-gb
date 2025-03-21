package mbc

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	MBC0_ROM_BANK = mem.MemRegion{Start: 0x0000, End: 0x7FFF}
	MBC0_RAM_BANK = mem.MemRegion{Start: 0xA000, End: 0xBFFF}
)

type MBC0 struct {
	rom []byte
}

var _ MBC = (*MBC0)(nil)

func NewMBC0(rom []byte) *MBC0 {
	return &MBC0{rom: rom}
}

func (m *MBC0) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr <= MBC0_ROM_BANK.End {
		return mem.ReadReplace(m.rom[addr])
	}

	return mem.ReadPassthrough()
}

func (m *MBC0) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr <= MBC0_ROM_BANK.End {
		// Put the Read-Only in ROM
		return mem.WriteBlock()
	}

	if MBC0_RAM_BANK.Contains(addr, false) {
		// RAM is RAM and this is a fake cartridge, so...
		return mem.WritePassthrough()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC0", value, addr))
}

func (m *MBC0) DebugPrint(w io.Writer) {
}

func (m *MBC0) Save(w io.Writer) error {
	return nil
}

func (m *MBC0) LoadSave(r io.Reader) error {
	return nil
}
