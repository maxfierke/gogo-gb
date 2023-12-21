package mbc

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	mbc5_rom_bank_00 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	mbc5_rom_banks   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	mbc5_ram_banks   = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	mbc5_reg_ram_enable      = mem.MemRegion{Start: 0x0000, End: 0x1FFF}
	mbc5_reg_ram_enable_mask = byte(0xF)
	mbc5_reg_ram_enabled     = byte(0xA)
	mbc5_reg_ram_disabled    = byte(0x00)

	mbc5_reg_lsb_rom_bank          = mem.MemRegion{Start: 0x2000, End: 0x2FFF}
	mbc5_reg_lsb_rom_bank_sel_mask = ^uint16(0xFF)

	mbc5_reg_msb_rom_bank          = mem.MemRegion{Start: 0x3000, End: 0x3FFF}
	mbc5_reg_msb_rom_bank_sel_mask = ^uint16(0x100)

	mbc5_reg_ram_bank          = mem.MemRegion{Start: 0x4000, End: 0x5FFF}
	mbc5_reg_ram_bank_sel_mask = byte(0xF)
)

type MBC5 struct {
	cur_ram_bank uint8
	cur_rom_bank uint16
	ram          []byte
	ram_enabled  bool
	rom          []byte
}

func NewMBC5(rom []byte, ram []byte) *MBC5 {
	return &MBC5{
		cur_ram_bank: 0,
		cur_rom_bank: 0,
		ram:          ram,
		ram_enabled:  false,
		rom:          rom,
	}
}

func (m *MBC5) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if mbc5_rom_bank_00.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if mbc5_rom_banks.Contains(addr, false) {
		bankByte := readBankAddr(
			m.rom,
			mbc5_rom_banks,
			ROM_BANK_SIZE,
			m.cur_rom_bank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if mbc5_ram_banks.Contains(addr, false) {
		if m.ram_enabled {
			bankByte := readBankAddr(
				m.ram,
				mbc5_ram_banks,
				RAM_BANK_SIZE,
				uint16(m.cur_ram_bank),
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
	if mbc5_reg_ram_enable.Contains(addr, false) {
		if value&mbc5_reg_ram_enable_mask == mbc5_reg_ram_enabled {
			m.ram_enabled = true
		} else if value&mbc5_reg_ram_enable_mask == mbc5_reg_ram_disabled {
			m.ram_enabled = false
		}

		// TODO: Log something / panic if unexpected value?

		return mem.WriteBlock()
	} else if mbc5_reg_lsb_rom_bank.Contains(addr, false) {
		m.cur_rom_bank = (m.cur_rom_bank & mbc5_reg_lsb_rom_bank_sel_mask) | uint16(value)
		return mem.WriteBlock()
	} else if mbc5_reg_msb_rom_bank.Contains(addr, false) {
		msb := uint16(value) & 0x1 << 8
		m.cur_rom_bank = (m.cur_rom_bank & mbc5_reg_msb_rom_bank_sel_mask) | msb
		return mem.WriteBlock()
	} else if mbc5_reg_ram_bank.Contains(addr, false) {
		m.cur_ram_bank = value & mbc5_reg_ram_bank_sel_mask
		return mem.WriteBlock()
	} else if mbc5_ram_banks.Contains(addr, false) {
		writeBankAddr(
			m.ram,
			mbc5_ram_banks,
			RAM_BANK_SIZE,
			uint16(m.cur_ram_bank),
			addr,
			value,
		)
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC5", value, addr))
}
