package mbc

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

var (
	MBC3_ROM_BANK_00 = mem.MemRegion{Start: 0x0000, End: 0x3FFF}
	MBC3_ROM_BANKS   = mem.MemRegion{Start: 0x4000, End: 0x7FFF}
	MBC3_RAM_BANKS   = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	MBC3_REG_RTC = mem.MemRegion{Start: 0xA000, End: 0xBFFF}

	MBC3_REG_RAM_RTC_ENABLE      = mem.MemRegion{Start: 0x0000, End: 0x1FFF}
	MBC3_REG_RAM_RTC_ENABLE_MASK = byte(0xF)
	MBC3_REG_RAM_RTC_ENABLED     = byte(0xA)

	MBC3_REG_ROM_BANK          = mem.MemRegion{Start: 0x2000, End: 0x3FFF}
	MBC3_REG_ROM_BANK_SEL_MASK = uint16(0x7F)

	MBC3_REG_RAM_BANK_OR_RTC_REG_SEL = mem.MemRegion{Start: 0x4000, End: 0x5FFF}

	MBC3_REG_RTC_LATCH_DATA = mem.MemRegion{Start: 0x6000, End: 0x7FFF}
)

type mbc3RtcReg byte

const (
	MBC3_RTC_REG_NONE     mbc3RtcReg = 0x00
	MBC3_RTC_REG_SECONDS  mbc3RtcReg = 0x08
	MBC3_RTC_REG_MINUTES  mbc3RtcReg = 0x09
	MBC3_RTC_REG_HOURS    mbc3RtcReg = 0x0A
	MBC3_RTC_REG_DAY_LOW  mbc3RtcReg = 0x0B
	MBC3_RTC_REG_DAY_HIGH mbc3RtcReg = 0x0C

	MBC3_RTC_REG_DAY_HIGH_BIT_DAY_MSB = 0
	MBC3_RTC_REG_DAY_HIGH_BIT_HALT    = 1 << 6
	MBC3_RTC_REG_DAY_HIGH_BIT_CARRY   = 1 << 7
)

type MBC3 struct {
	curRamBank  uint8
	curRomBank  uint16
	ram         []byte
	ramEnabled  bool
	ramSelected bool
	rtcEnabled  bool
	rom         []byte

	rtcAvailable      bool
	rtcLatchRequested bool
	rtcRegSelected    mbc3RtcReg
	rtcSeconds        uint8
	rtcMinutes        uint8
	rtcHours          uint8
	rtcDays           uint
	rtcHalt           bool
	rtcDaysOverflow   bool
}

var _ MBC = (*MBC3)(nil)

func NewMBC3(rom []byte, ram []byte, rtcAvailable bool) *MBC3 {
	return &MBC3{
		ram:          ram,
		rom:          rom,
		rtcAvailable: rtcAvailable,
	}
}

func (m *MBC3) Step(cycles uint8) {
	// TODO: tick the RTC
}

func (m *MBC3) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC3_ROM_BANK_00.Contains(addr, false) {
		return mem.ReadReplace(m.rom[addr])
	} else if MBC3_ROM_BANKS.Contains(addr, false) {
		bankByte := readBankAddr(
			m.rom,
			MBC3_ROM_BANKS,
			ROM_BANK_SIZE,
			m.curRomBank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	} else if MBC3_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled && m.ramSelected {
			bankByte := readBankAddr(
				m.ram,
				MBC3_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
			)
			return mem.ReadReplace(bankByte)
		} else if m.rtcEnabled && m.rtcRegSelected != MBC3_RTC_REG_NONE {
			value := m.readRtcReg(m.rtcRegSelected)
			return mem.ReadReplace(value)
		} else {
			// Docs say this is usually 0xFF, but not guaranteed. Randomness needed?
			return mem.ReadReplace(0xFF)
		}
	}

	return mem.ReadPassthrough()
}

func (m *MBC3) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC3_REG_RAM_RTC_ENABLE.Contains(addr, false) {
		if value&MBC3_REG_RAM_RTC_ENABLE_MASK == MBC3_REG_RAM_RTC_ENABLED {
			m.ramEnabled = true
			m.rtcEnabled = m.rtcAvailable
		} else {
			m.ramEnabled = false
			m.rtcEnabled = false
		}

		return mem.WriteBlock()
	} else if MBC3_REG_ROM_BANK.Contains(addr, false) {
		m.curRomBank = uint16(value) & MBC3_REG_ROM_BANK_SEL_MASK
		if m.curRomBank == 0 {
			m.curRomBank = 1
		}
		return mem.WriteBlock()
	} else if MBC3_REG_RAM_BANK_OR_RTC_REG_SEL.Contains(addr, false) {
		if value <= 0x3 {
			m.curRamBank = value & 0x3
			m.ramSelected = true
			m.rtcRegSelected = MBC3_RTC_REG_NONE
		} else if value >= 0x08 && value <= 0x0C {
			m.ramSelected = false
			m.rtcRegSelected = mbc3RtcReg(value)
		}
		return mem.WriteBlock()
	} else if MBC3_REG_RTC_LATCH_DATA.Contains(addr, false) {
		if value == 0x00 && !m.rtcLatchRequested {
			m.rtcLatchRequested = true
		} else if value == 0x01 && m.rtcLatchRequested {
			// TODO: Latch current time into RTC registers (?)
			m.rtcLatchRequested = false
		} else {
			m.rtcLatchRequested = false
		}

		return mem.WriteBlock()
	} else if MBC3_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled && m.ramSelected {
			writeBankAddr(
				m.ram,
				MBC3_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
				value,
			)
		} else if m.rtcEnabled && m.rtcRegSelected != MBC3_RTC_REG_NONE {
			m.writeRtcReg(m.rtcRegSelected, value)
		}
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC3", value, addr))
}

func (m *MBC3) readRtcReg(reg mbc3RtcReg) byte {
	switch reg {
	case MBC3_RTC_REG_SECONDS:
		return m.rtcSeconds
	case MBC3_RTC_REG_MINUTES:
		return m.rtcMinutes
	case MBC3_RTC_REG_HOURS:
		return m.rtcHours
	case MBC3_RTC_REG_DAY_LOW:
		return byte(m.rtcDays & 0xFF)
	case MBC3_RTC_REG_DAY_HIGH:
		dayCounterMsb := byte((m.rtcDays >> 8) & 0b1)
		overflow := byte(0x0)
		if m.rtcDaysOverflow {
			overflow = 1 << 7
		}

		halt := byte(0x0)
		if m.rtcHalt {
			halt = 1 << 6
		}

		return (overflow | halt | dayCounterMsb)
	}

	panic(fmt.Sprintf("Attempting to read RTC reg 0x%02X, which is out-of-bounds for MBC3", reg))
}

func (m *MBC3) writeRtcReg(reg mbc3RtcReg, value byte) {
	switch reg {
	case MBC3_RTC_REG_SECONDS:
		m.rtcSeconds = value & 0x3F
	case MBC3_RTC_REG_MINUTES:
		m.rtcMinutes = value & 0x3F
	case MBC3_RTC_REG_HOURS:
		m.rtcHours = value & 0x1F
	case MBC3_RTC_REG_DAY_LOW:
		m.rtcDays = (m.rtcDays & 0x100) | uint(value)
	case MBC3_RTC_REG_DAY_HIGH:
		m.rtcDays = (m.rtcDays & 0xFF) | (uint(value&0b1) << 8)
		m.rtcHalt = ((value >> 6) & 0b1) == 1
		m.rtcDaysOverflow = ((value >> 7) & 0b1) == 1
	default:
		panic(fmt.Sprintf("Attempting to write RTC reg 0x%02X with 0x%02X, which is out-of-bounds for MBC3", reg, value))
	}
}

func (m *MBC3) Save(w io.Writer) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := w.Write(m.ram)
	if err != nil {
		return fmt.Errorf("mbc3: saving SRAM: %w. wrote %d bytes", err, n)
	}

	// TODO: Write RTC registers

	return nil
}

func (m *MBC3) LoadSave(r io.Reader) error {
	if len(m.ram) == 0 {
		return nil
	}

	n, err := io.ReadFull(r, m.ram)
	if err != nil {
		return fmt.Errorf("mbc3: loading save into SRAM: %w. read %d bytes", err, n)
	}

	// TODO: Read RTC registers

	return nil
}

var (
	MBC30_ROM_BANKS = mem.MemRegion{Start: 0x4000, End: 0x7FFF}

	MBC30_REG_ROM_BANK = mem.MemRegion{Start: 0x2000, End: 0x3FFF}

	MBC30_REG_RAM_BANK_OR_RTC_REG_SEL = mem.MemRegion{Start: 0x4000, End: 0x5FFF}
)

type MBC30 struct {
	MBC3
}

func NewMBC30(rom []byte, ram []byte, rtcAvailable bool) *MBC30 {
	return &MBC30{
		MBC3: MBC3{
			ram:          ram,
			rom:          rom,
			rtcAvailable: rtcAvailable,
		},
	}
}

func (m *MBC30) Step(cycles uint8) {
	m.MBC3.Step(cycles)
}

func (m *MBC30) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if MBC30_ROM_BANKS.Contains(addr, false) {
		bankByte := readBankAddr(
			m.rom,
			MBC30_ROM_BANKS,
			ROM_BANK_SIZE,
			m.curRomBank,
			addr,
		)
		return mem.ReadReplace(bankByte)
	}

	return m.MBC3.OnRead(mmu, addr)
}

func (m *MBC30) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if MBC30_REG_ROM_BANK.Contains(addr, false) {
		m.curRomBank = uint16(value)
		if m.curRomBank == 0 {
			m.curRomBank = 1
		}
		return mem.WriteBlock()
	}

	if MBC30_REG_RAM_BANK_OR_RTC_REG_SEL.Contains(addr, false) && value <= 0x7 {
		m.curRamBank = value & 0x7
		m.ramSelected = true
		m.rtcRegSelected = MBC3_RTC_REG_NONE
		return mem.WriteBlock()
	}

	return m.MBC3.OnWrite(mmu, addr, value)
}

func (m *MBC30) Save(w io.Writer) error {
	return m.MBC3.Save(w)
}

func (m *MBC30) LoadSave(r io.Reader) error {
	return m.MBC3.LoadSave(r)
}
