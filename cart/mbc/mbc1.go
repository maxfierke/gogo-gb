package mbc

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	mbc1_rom_bank_x0 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	mbc1_rom_banks   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	mbc1_ram_banks   = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	mbc1_reg_ram_enable      = mem.MemRegion{Start: 0x0000, End: 0x1FFF}
	mbc1_reg_ram_enable_mask = byte(0xF)
	mbc1_reg_ram_enabled     = byte(0xA)

	mbc1_reg_rom_bank          = mem.MemRegion{Start: 0x2000, End: 0x3FFF}
	mbc1_reg_rom_bank_sel_mask = uint16(0x1F)

	mbc1_reg_ram_bank_or_msb_rom_bank = mem.MemRegion{Start: 0x4000, End: 0x5FFF}
	mbc1_reg_msb_rom_bank_sel_mask    = ^uint16(0x60)

	mbc1_reg_bank_mode_sel = mem.MemRegion{Start: 0x6000, End: 0x7FFF}
)

// Struct for MBC1 support (minus MBC1M, currently)
type MBC1 struct {
	cur_ram_bank uint8
	cur_rom_bank uint16
	ram          []byte
	ram_enabled  bool
	ram_selected bool
	rom          []byte
}

func NewMBC1(rom []byte, ram []byte) *MBC1 {
	return &MBC1{
		cur_ram_bank: 0,
		cur_rom_bank: 0,
		ram:          ram,
		ram_enabled:  false,
		ram_selected: false,
		rom:          rom,
	}
}

func (m *MBC1) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if mbc1_rom_bank_x0.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if mbc1_rom_banks.Contains(addr, false) {
		// see https://gbdev.io/pandocs/MBC1.html#00003fff--rom-bank-x0-read-only
		romBank := max(m.cur_rom_bank, 1)
		if romBank == 0x20 || romBank == 0x40 || romBank == 0x60 {
			romBank += 1
		}

		bankByte := readBankAddr(
			m.rom,
			mbc1_rom_banks,
			ROM_BANK_SIZE,
			romBank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if mbc1_ram_banks.Contains(addr, false) {
		if m.ram_enabled {
			bankByte := readBankAddr(
				m.ram,
				mbc1_ram_banks,
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

func (m *MBC1) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if mbc1_reg_ram_enable.Contains(addr, false) {
		if value&mbc1_reg_ram_enable_mask == mbc1_reg_ram_enabled {
			m.ram_enabled = true
		} else {
			m.ram_enabled = false
		}

		return mem.WriteBlock()
	} else if mbc1_reg_rom_bank.Contains(addr, false) {
		m.cur_rom_bank =
			(m.cur_rom_bank & ^mbc1_reg_rom_bank_sel_mask) |
				(uint16(value) & mbc1_reg_rom_bank_sel_mask)
		return mem.WriteBlock()
	} else if mbc1_reg_ram_bank_or_msb_rom_bank.Contains(addr, false) {
		if m.ram_selected {
			m.cur_ram_bank = value & 0x3
		} else {
			msb := uint16(value) & 0x3 << 5
			m.cur_rom_bank = (m.cur_rom_bank & mbc1_reg_msb_rom_bank_sel_mask) | msb
		}
		return mem.WriteBlock()
	} else if mbc1_reg_bank_mode_sel.Contains(addr, false) {
		if value == 0x00 {
			m.ram_selected = false
		} else if value == 0x01 {
			m.ram_selected = true
		}

		// TODO: Log something / panic if unexpected value?

		return mem.WriteBlock()
	} else if mbc1_ram_banks.Contains(addr, false) {
		writeBankAddr(
			m.ram,
			mbc1_ram_banks,
			RAM_BANK_SIZE,
			uint16(m.cur_ram_bank),
			addr,
			value,
		)
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%x @ 0x%x, which is out-of-bounds for MBC1", value, addr))
}
