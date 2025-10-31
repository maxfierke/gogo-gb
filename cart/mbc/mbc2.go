package mbc

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	MBC2_ROM_BANK_X0 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	MBC2_ROM_BANKS   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	MBC2_RAM_BANK    = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	MBC2_RAM_SIZE      uint16 = 0x200
	MBC2_RAM_ADDR_MASK uint16 = 0x1FF

	MBC2_REG_RAM_ENABLE_OR_ROM_BANK                  = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	MBC2_REG_RAM_ENABLE_OR_ROM_BANK_MODE_MASK        = uint16(0x100)
	MBC2_REG_RAM_ENABLED                             = byte(0xA)
	MBC2_REG_RAM_ENABLE_OR_ROM_BANK_RAM_SEL   uint16 = 0
)

type MBC2 struct {
	curRomBank uint8
	hasBattery bool
	ram        []byte
	ramEnabled bool
	rom        []byte
}

var _ MBC = (*MBC2)(nil)

func NewMBC2(rom []byte, hasBattery bool) *MBC2 {
	return &MBC2{
		curRomBank: 0,
		hasBattery: hasBattery,
		ram:        make([]byte, MBC2_RAM_SIZE),
		ramEnabled: false,
		rom:        rom,
	}
}

func (m *MBC2) Step(cycles uint8) {}

func (m *MBC2) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC2_ROM_BANK_X0.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if MBC2_ROM_BANKS.Contains(addr, false) {
		// see https://gbdev.io/pandocs/MBC1.html#00003fff--rom-bank-x0-read-only
		romBank := max(m.curRomBank, 1)

		bankByte := mem.ReadBankAddr(
			m.rom,
			MBC2_ROM_BANKS,
			ROM_BANK_SIZE,
			uint16(romBank),
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if MBC2_RAM_BANK.Contains(addr, false) {
		if m.ramEnabled {
			bankByte := mem.ReadBankAddr(
				m.ram,
				MBC2_RAM_BANK,
				MBC2_RAM_SIZE,
				0,
				(addr & MBC2_RAM_ADDR_MASK),
			)
			return mem.ReadReplace(bankByte & 0xF)
		} else {
			return mem.ReadReplace(0xFF)
		}
	}

	return mem.ReadPassthrough()
}

func (m *MBC2) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC2_REG_RAM_ENABLE_OR_ROM_BANK.Contains(addr, false) {
		mode := (addr & MBC2_REG_RAM_ENABLE_OR_ROM_BANK_MODE_MASK) >> 7

		if mode == MBC2_REG_RAM_ENABLE_OR_ROM_BANK_RAM_SEL {
			m.ramEnabled = (value & 0xF) == MBC2_REG_RAM_ENABLED
		} else {
			m.curRomBank = value & 0xF
		}

		return mem.WriteBlock()
	} else if MBC2_RAM_BANK.Contains(addr, false) {
		if m.ramEnabled {
			mem.WriteBankAddr(
				m.ram,
				MBC2_RAM_BANK,
				MBC2_RAM_SIZE,
				0,
				(addr & MBC2_RAM_ADDR_MASK),
				(value & 0xF),
			)
		}

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC2", value, addr))
}

func (m *MBC2) DebugPrint(w io.Writer) {
	fmt.Fprintf(w, "== MBC2 ==\n\n")
	fmt.Fprintf(w, "Current ROM bank: %d\n", m.curRomBank)
	fmt.Fprintf(w, "RAM enabled: %t\n", m.ramEnabled)
}

func (m *MBC2) Save(w io.Writer) error {
	if !m.hasBattery {
		return nil
	}

	n, err := w.Write(m.ram)
	if err != nil {
		return fmt.Errorf("MBC2: saving built-in RAM: %w. wrote %d bytes", err, n)
	}

	return nil
}

func (m *MBC2) LoadSave(r io.Reader) error {
	if !m.hasBattery {
		return nil
	}

	n, err := io.ReadFull(r, m.ram)
	if err != nil {
		return fmt.Errorf("MBC2: loading save into built-in RAM: %w. read %d bytes", err, n)
	}

	return nil
}
