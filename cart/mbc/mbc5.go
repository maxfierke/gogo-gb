package mbc

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	MBC5_ROM_BANK_00 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	MBC5_ROM_BANKS   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	MBC5_RAM_BANKS   = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	MBC5_REG_RAM_ENABLE      = mem.MemRegion{Start: 0x0000, End: 0x1FFF}
	MBC5_REG_RAM_ENABLE_MASK = byte(0xF)
	MBC5_REG_RAM_ENABLED     = byte(0xA)
	MBC5_REG_RAM_DISABLED    = byte(0x00)

	MBC5_REG_LSB_ROM_BANK          = mem.MemRegion{Start: 0x2000, End: 0x2FFF}
	MBC5_REG_LSB_ROM_BANK_SEL_MASK = ^uint16(0xFF)

	MBC5_REG_MSB_ROM_BANK          = mem.MemRegion{Start: 0x3000, End: 0x3FFF}
	MBC5_REG_MSB_ROM_BANK_SEL_MASK = ^uint16(0x100)

	MBC5_REG_RAM_BANK          = mem.MemRegion{Start: 0x4000, End: 0x5FFF}
	MBC5_REG_RAM_BANK_SEL_MASK = byte(0xF)
)

type MBC5 struct {
	curRamBank uint8
	curRomBank uint16
	ram        []byte
	ramEnabled bool
	rom        []byte
}

var _ MBC = (*MBC5)(nil)

func NewMBC5(rom []byte, ram []byte) *MBC5 {
	return &MBC5{
		curRamBank: 0,
		curRomBank: 0,
		ram:        ram,
		ramEnabled: false,
		rom:        rom,
	}
}

func (m *MBC5) Step(cycles uint8) {}

func (m *MBC5) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC5_ROM_BANK_00.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	}

	if MBC5_ROM_BANKS.Contains(addr, false) {
		bankByte := mem.ReadBankAddr(
			m.rom,
			MBC5_ROM_BANKS,
			ROM_BANK_SIZE,
			m.curRomBank,
			addr,
		)

		return mem.ReadReplace(bankByte)
	}

	if MBC5_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled {
			bankByte := mem.ReadBankAddr(
				m.ram,
				MBC5_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
			)

			return mem.ReadReplace(bankByte)
		}

		// Docs say this is usually 0xFF, but not guaranteed. Randomness needed?
		return mem.ReadReplace(0xFF)
	}

	return mem.ReadPassthrough()
}

func (m *MBC5) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC5_REG_RAM_ENABLE.Contains(addr, false) {
		switch value & MBC5_REG_RAM_ENABLE_MASK {
		case MBC5_REG_RAM_ENABLED:
			m.ramEnabled = true
		case MBC5_REG_RAM_DISABLED:
			m.ramEnabled = false
		}

		return mem.WriteBlock()
	}

	if MBC5_REG_LSB_ROM_BANK.Contains(addr, false) {
		m.curRomBank = (m.curRomBank & MBC5_REG_LSB_ROM_BANK_SEL_MASK) | uint16(value)

		return mem.WriteBlock()
	}

	if MBC5_REG_MSB_ROM_BANK.Contains(addr, false) {
		msb := uint16(value) & 0x1 << 8
		m.curRomBank = (m.curRomBank & MBC5_REG_MSB_ROM_BANK_SEL_MASK) | msb

		return mem.WriteBlock()
	}

	if MBC5_REG_RAM_BANK.Contains(addr, false) {
		m.curRamBank = value & MBC5_REG_RAM_BANK_SEL_MASK

		return mem.WriteBlock()
	}

	if MBC5_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled && len(m.ram) > 0 {
			mem.WriteBankAddr(
				m.ram,
				MBC5_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
				value,
			)
		}

		return mem.WriteBlock()
	}

	if MBC5_ROM_BANKS.Contains(addr, false) {
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC5", value, addr))
}

func (m *MBC5) DebugPrint(w io.Writer) {
	fmt.Fprintf(w, "== MBC5 ==\n\n")

	fmt.Fprintf(w, "Current ROM bank: %d\n", m.curRomBank)
	fmt.Fprintf(w, "Current RAM bank: %d\n", m.curRamBank)
	fmt.Fprintf(w, "RAM enabled: %t\n", m.ramEnabled)
}

func (m *MBC5) Save(w io.Writer) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := w.Write(m.ram)
	if err != nil {
		return fmt.Errorf("mbc5: saving SRAM: %w. wrote %d bytes", err, n)
	}

	return nil
}

func (m *MBC5) LoadSave(r io.Reader) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := io.ReadFull(r, m.ram)
	if err != nil {
		return fmt.Errorf("mbc5: loading save into SRAM: %w. read %d bytes", err, n)
	}

	return nil
}
