package mbc

import (
	"fmt"

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

func NewMBC5(rom []byte, ram []byte) *MBC5 {
	return &MBC5{
		curRamBank: 0,
		curRomBank: 0,
		ram:        ram,
		ramEnabled: false,
		rom:        rom,
	}
}

func (m *MBC5) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC5_ROM_BANK_00.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if MBC5_ROM_BANKS.Contains(addr, false) {
		bankByte := readBankAddr(
			m.rom,
			MBC5_ROM_BANKS,
			ROM_BANK_SIZE,
			m.curRomBank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if MBC5_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled {
			bankByte := readBankAddr(
				m.ram,
				MBC5_RAM_BANKS,
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

func (m *MBC5) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC5_REG_RAM_ENABLE.Contains(addr, false) {
		if value&MBC5_REG_RAM_ENABLE_MASK == MBC5_REG_RAM_ENABLED {
			m.ramEnabled = true
		} else if value&MBC5_REG_RAM_ENABLE_MASK == MBC5_REG_RAM_DISABLED {
			m.ramEnabled = false
		}

		// TODO: Log something / panic if unexpected value?

		return mem.WriteBlock()
	} else if MBC5_REG_LSB_ROM_BANK.Contains(addr, false) {
		m.curRomBank = (m.curRomBank & MBC5_REG_LSB_ROM_BANK_SEL_MASK) | uint16(value)
		return mem.WriteBlock()
	} else if MBC5_REG_MSB_ROM_BANK.Contains(addr, false) {
		msb := uint16(value) & 0x1 << 8
		m.curRomBank = (m.curRomBank & MBC5_REG_MSB_ROM_BANK_SEL_MASK) | msb
		return mem.WriteBlock()
	} else if MBC5_REG_RAM_BANK.Contains(addr, false) {
		m.curRamBank = value & MBC5_REG_RAM_BANK_SEL_MASK
		return mem.WriteBlock()
	} else if MBC5_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled {
			writeBankAddr(
				m.ram,
				MBC5_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
				value,
			)
		}
		return mem.WriteBlock()
	} else if MBC5_ROM_BANKS.Contains(addr, false) {
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC5", value, addr))
}
