package mbc

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	MBC1_ROM_BANK_X0 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	MBC1_ROM_BANKS   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	MBC1_RAM_BANKS   = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	MBC1_REG_RAM_ENABLE      = mem.MemRegion{Start: 0x0000, End: 0x1FFF}
	MBC1_REG_RAM_ENABLE_MASK = byte(0xF)
	MBC1_REG_RAM_ENABLED     = byte(0xA)

	MBC1_REG_ROM_BANK          = mem.MemRegion{Start: 0x2000, End: 0x3FFF}
	MBC1_REG_ROM_BANK_SEL_MASK = uint16(0x1F)

	MBC1_REG_RAM_BANK_OR_MSB_ROM_BANK = mem.MemRegion{Start: 0x4000, End: 0x5FFF}
	MBC1_REG_MSB_ROM_BANK_SEL_MASK    = ^uint16(0x60)

	MBC1_REG_BANK_MODE_SEL = mem.MemRegion{Start: 0x6000, End: 0x7FFF}
)

// Struct for MBC1 support (minus MBC1M, currently)
type MBC1 struct {
	curRamBank  uint8
	curRomBank  uint16
	ram         []byte
	ramEnabled  bool
	ramSelected bool
	rom         []byte
}

var _ MBC = (*MBC1)(nil)

func NewMBC1(rom []byte, ram []byte) *MBC1 {
	return &MBC1{
		curRamBank:  0,
		curRomBank:  0,
		ram:         ram,
		ramEnabled:  false,
		ramSelected: false,
		rom:         rom,
	}
}

func (m *MBC1) Step(cycles uint8) {}

func (m *MBC1) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC1_ROM_BANK_X0.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if MBC1_ROM_BANKS.Contains(addr, false) {
		// see https://gbdev.io/pandocs/MBC1.html#00003fff--rom-bank-x0-read-only
		romBank := max(m.curRomBank, 1)
		if romBank == 0x20 || romBank == 0x40 || romBank == 0x60 {
			romBank += 1
		}

		bankByte := readBankAddr(
			m.rom,
			MBC1_ROM_BANKS,
			ROM_BANK_SIZE,
			romBank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if MBC1_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled {
			bankByte := readBankAddr(
				m.ram,
				MBC1_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
			)
			return mem.ReadReplace(bankByte)
		} else {
			// Docs say this is usually 0xFF, but not guaranteed. Randomness needed?
			return mem.ReadReplace(0xFF)
		}
	}

	return mem.ReadPassthrough()
}

func (m *MBC1) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC1_REG_RAM_ENABLE.Contains(addr, false) {
		if value&MBC1_REG_RAM_ENABLE_MASK == MBC1_REG_RAM_ENABLED {
			m.ramEnabled = true
		} else {
			m.ramEnabled = false
		}

		return mem.WriteBlock()
	} else if MBC1_REG_ROM_BANK.Contains(addr, false) {
		m.curRomBank = (m.curRomBank & ^MBC1_REG_ROM_BANK_SEL_MASK) |
			(uint16(value) & MBC1_REG_ROM_BANK_SEL_MASK)
		return mem.WriteBlock()
	} else if MBC1_REG_RAM_BANK_OR_MSB_ROM_BANK.Contains(addr, false) {
		if m.ramSelected {
			m.curRamBank = value & 0x3
		} else {
			msb := (uint16(value) & 0x3) << 5
			m.curRomBank = (m.curRomBank & MBC1_REG_MSB_ROM_BANK_SEL_MASK) | msb
		}
		return mem.WriteBlock()
	} else if MBC1_REG_BANK_MODE_SEL.Contains(addr, false) {
		switch value {
		case 0x00:
			m.ramSelected = false
		case 0x01:
			m.ramSelected = true
		}

		return mem.WriteBlock()
	} else if MBC1_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled && len(m.ram) > 0 {
			writeBankAddr(
				m.ram,
				MBC1_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
				value,
			)
		}

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC1", value, addr))
}

func (m *MBC1) DebugPrint(w io.Writer) {
	fmt.Fprintf(w, "== MBC1 ==\n\n")

	bankMode := 0
	if m.ramSelected {
		bankMode = 1
	}

	fmt.Fprintf(w, "Current ROM bank: %d\n", m.curRomBank)
	fmt.Fprintf(w, "Current RAM bank: %d\n", m.curRamBank)
	fmt.Fprintf(w, "RAM enabled: %t\n", m.ramEnabled)
	fmt.Fprintf(w, "Bank mode: %d\n", bankMode)
}

func (m *MBC1) Save(w io.Writer) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := w.Write(m.ram)
	if err != nil {
		return fmt.Errorf("mbc1: saving SRAM: %w. wrote %d bytes", err, n)
	}

	return nil
}

func (m *MBC1) LoadSave(r io.Reader) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := io.ReadFull(r, m.ram)
	if err != nil {
		return fmt.Errorf("mbc1: loading save into SRAM: %w. read %d bytes", err, n)
	}

	return nil
}
