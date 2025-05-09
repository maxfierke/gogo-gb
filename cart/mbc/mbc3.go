package mbc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/maxfierke/gogo-gb/bits"
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

	ErrMBC3BadClockBattery = errors.New("unable to load saved RTC registers")
)

type mbc3RTCReg byte

func (rtcReg mbc3RTCReg) String() string {
	switch rtcReg {
	case MBC3_RTC_REG_NONE:
		return "None"
	case MBC3_RTC_REG_SECONDS:
		return "Seconds"
	case MBC3_RTC_REG_MINUTES:
		return "Minutes"
	case MBC3_RTC_REG_HOURS:
		return "Hours"
	case MBC3_RTC_REG_DAY_LOW:
		return "Day (Low)"
	case MBC3_RTC_REG_DAY_HIGH:
		return "Day (High) / Halt / Carry"
	default:
		return "Unknown"
	}
}

const (
	MBC3_RTC_REG_NONE     mbc3RTCReg = 0x00
	MBC3_RTC_REG_SECONDS  mbc3RTCReg = 0x08
	MBC3_RTC_REG_MINUTES  mbc3RTCReg = 0x09
	MBC3_RTC_REG_HOURS    mbc3RTCReg = 0x0A
	MBC3_RTC_REG_DAY_LOW  mbc3RTCReg = 0x0B
	MBC3_RTC_REG_DAY_HIGH mbc3RTCReg = 0x0C

	MBC3_RTC_REG_DAY_HIGH_BIT_DAY_MSB = 0
	MBC3_RTC_REG_DAY_HIGH_BIT_HALT    = 6
	MBC3_RTC_REG_DAY_HIGH_BIT_CARRY   = 7
)

type mbc3RTCRegs struct {
	Seconds      uint8
	Minutes      uint8
	Hours        uint8
	Days         uint32
	Halt         bool
	DaysOverflow bool
	Timestamp    time.Time
}

func (regs *mbc3RTCRegs) readReg(reg mbc3RTCReg) byte {
	switch reg {
	case MBC3_RTC_REG_SECONDS:
		return regs.Seconds & 0x3F
	case MBC3_RTC_REG_MINUTES:
		return regs.Minutes & 0x3F
	case MBC3_RTC_REG_HOURS:
		return regs.Hours & 0x1F
	case MBC3_RTC_REG_DAY_LOW:
		return byte(regs.Days & 0xFF)
	case MBC3_RTC_REG_DAY_HIGH:
		dayCounterMsb := byte(((regs.Days & 0x100) >> 8) & 0b1)
		overflow := byte(0x0)
		if regs.DaysOverflow {
			overflow = 1 << MBC3_RTC_REG_DAY_HIGH_BIT_CARRY
		}

		halt := byte(0x0)
		if regs.Halt {
			halt = 1 << MBC3_RTC_REG_DAY_HIGH_BIT_HALT
		}

		return (overflow | halt | dayCounterMsb)
	default:
		panic(fmt.Sprintf("Attempting to read RTC reg 0x%02X, which is out-of-bounds for MBC3", reg))
	}
}

func (regs *mbc3RTCRegs) writeReg(reg mbc3RTCReg, value byte) {
	switch reg {
	case MBC3_RTC_REG_SECONDS:
		regs.Seconds = value & 0x3F
	case MBC3_RTC_REG_MINUTES:
		regs.Minutes = value & 0x3F
	case MBC3_RTC_REG_HOURS:
		regs.Hours = value & 0x1F
	case MBC3_RTC_REG_DAY_LOW:
		regs.Days = (regs.Days & 0x100) | uint32(value)
	case MBC3_RTC_REG_DAY_HIGH:
		regs.Days = (regs.Days & 0xFF) | (uint32(value&0b1) << 8)
		regs.Halt = bits.Read(value, MBC3_RTC_REG_DAY_HIGH_BIT_HALT) == 1
		regs.DaysOverflow = bits.Read(value, MBC3_RTC_REG_DAY_HIGH_BIT_CARRY) == 1
	default:
		panic(fmt.Sprintf("Attempting to write RTC reg 0x%02X with 0x%02X, which is out-of-bounds for MBC3", reg, value))
	}
}

func (regs *mbc3RTCRegs) advanceTime(now time.Time) {
	rtcDiff := now.Sub(regs.Timestamp).Truncate(time.Second)

	if rtcDiff.Seconds() > 0.0 && !regs.Halt {
		deltaDays := uint32(rtcDiff.Hours()) / 24
		if deltaDays > 0 {
			rtcDiff = rtcDiff - (24 * time.Hour * time.Duration(deltaDays))
		}

		deltaHours := uint(rtcDiff.Hours())
		rtcDiff = rtcDiff - (time.Hour * time.Duration(deltaHours))

		deltaMinutes := uint(rtcDiff.Minutes())
		rtcDiff = rtcDiff - (time.Minute * time.Duration(deltaMinutes))

		deltaSeconds := uint(rtcDiff.Seconds())

		newSeconds := uint(regs.Seconds) + deltaSeconds
		if newSeconds >= 60 {
			deltaMinutes += newSeconds / 60
			newSeconds = newSeconds % 60
		}

		newMinutes := uint(regs.Minutes) + deltaMinutes
		if newMinutes >= 60 {
			deltaHours += newMinutes / 60
			newMinutes = newMinutes % 60
		}

		newHours := uint(regs.Hours) + deltaHours
		if newHours >= 24 {
			deltaDays += uint32(newHours / 24)
			newHours = newHours % 24
		}

		newDays := regs.Days + deltaDays
		if newDays >= 512 {
			newDays = newDays % 512
			regs.DaysOverflow = true
		}

		regs.Days = newDays
		regs.Hours = uint8(newHours)
		regs.Minutes = uint8(newMinutes)
		regs.Seconds = uint8(newSeconds)
		regs.Timestamp = now
	}
}

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
	rtcRegSelected    mbc3RTCReg
	rtc               mbc3RTCRegs
	latchedRTC        mbc3RTCRegs
	rtcClock          uint
}

type mbc3SaveRTC struct {
	CurrentSeconds              uint32
	CurrentMinutes              uint32
	CurrentHours                uint32
	CurrentDays                 uint32
	CurrentDaysHighOverflowHalt uint32
	LatchedSeconds              uint32
	LatchedMinutes              uint32
	LatchedHours                uint32
	LatchedDays                 uint32
	LatchedDaysHighOverflowHalt uint32
	UnixTimestamp               int64
}

var _ MBC = (*MBC3)(nil)

func NewMBC3(rom []byte, ram []byte, rtcAvailable bool) *MBC3 {
	return &MBC3{
		ram:          ram,
		rom:          rom,
		rtcAvailable: rtcAvailable,
	}
}

// TODO(GBC): Use GBC clock rate here
const cyclesPerRTCSecond = 4194304

func (m *MBC3) Step(cycles uint8) {
	if m.rtcAvailable {
		m.rtcClock += uint(cycles)

		now := time.Now()
		if m.rtc.Timestamp.IsZero() || m.rtc.Timestamp.After(now) {
			m.rtc.Timestamp = now
		}

		if m.rtc.Halt {
			m.rtcClock -= uint(cycles)
		}

		if m.rtcClock >= cyclesPerRTCSecond {
			m.rtcClock -= cyclesPerRTCSecond
			m.rtc.advanceTime(m.rtc.Timestamp.Add(time.Second))
		}
	}
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
			value := m.rtc.readReg(m.rtcRegSelected)
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
			m.rtcRegSelected = mbc3RTCReg(value)
		}
		return mem.WriteBlock()
	} else if MBC3_REG_RTC_LATCH_DATA.Contains(addr, false) {
		if value == 0x00 && !m.rtcLatchRequested {
			m.rtcLatchRequested = true
		} else if value == 0x01 && m.rtcLatchRequested {
			m.latchedRTC = m.rtc
			m.rtcLatchRequested = false
		} else {
			m.rtcLatchRequested = false
		}

		return mem.WriteBlock()
	} else if MBC3_RAM_BANKS.Contains(addr, false) {
		if m.ramEnabled && m.ramSelected && len(m.ram) > 0 {
			writeBankAddr(
				m.ram,
				MBC3_RAM_BANKS,
				RAM_BANK_SIZE,
				uint16(m.curRamBank),
				addr,
				value,
			)
		} else if m.rtcEnabled && m.rtcRegSelected != MBC3_RTC_REG_NONE {
			m.rtc.writeReg(m.rtcRegSelected, value)
		}
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for MBC3", value, addr))
}

func (m *MBC3) DebugPrint(w io.Writer) {
	fmt.Fprintf(w, "== MBC3/MBC30 ==\n\n")

	fmt.Fprintf(w, "Current ROM bank: %d\n", m.curRomBank)
	fmt.Fprintf(w, "Current RAM bank: %d\n", m.curRamBank)
	fmt.Fprintf(w, "RAM enabled: %t\n", m.ramEnabled)
	fmt.Fprintf(w, "RAM selected: %t\n", m.ramSelected)

	fmt.Fprintf(w, "RTC available: %t\n", m.rtcAvailable)
	if m.rtcAvailable {
		fmt.Fprintf(w, "RTC enabled: %t\n", m.rtcEnabled)
		fmt.Fprintf(w, "RTC register selected: %s (0x%02X)\n", m.rtcRegSelected.String(), byte(m.rtcRegSelected))
		fmt.Fprintf(w, "RTC latch requested: %t\n", m.rtcLatchRequested)
		fmt.Fprintf(w, "RTC halt: %t\n", m.rtc.Halt)
		fmt.Fprintf(w, "RTC days overflow: %t\n", m.rtc.DaysOverflow)

		fmt.Fprintf(w, "Current time: %02d:%02d:%02d\t\tDay: %d\n", m.rtc.Hours, m.rtc.Minutes, m.rtc.Seconds, m.rtc.Days)
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

	if m.rtcAvailable {
		err := m.saveRTCRegsToSave(w)
		if err != nil {
			return fmt.Errorf("mbc3: saving RTC registers: %w", err)
		}
	}

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

	if m.rtcAvailable {
		err = m.loadRTCRegsFromSave(r)
		if err != nil {
			return fmt.Errorf("mbc3: %w: %w", ErrMBC3BadClockBattery, err)
		}
	}

	return nil
}

func (m *MBC3) saveRTCRegsToSave(w io.Writer) error {
	rtc := &mbc3SaveRTC{
		CurrentSeconds:              uint32(m.rtc.readReg(MBC3_RTC_REG_SECONDS)),
		CurrentMinutes:              uint32(m.rtc.readReg(MBC3_RTC_REG_MINUTES)),
		CurrentHours:                uint32(m.rtc.readReg(MBC3_RTC_REG_HOURS)),
		CurrentDays:                 uint32(m.rtc.readReg(MBC3_RTC_REG_DAY_LOW)),
		CurrentDaysHighOverflowHalt: uint32(m.rtc.readReg(MBC3_RTC_REG_DAY_HIGH)),
		LatchedSeconds:              uint32(m.latchedRTC.readReg(MBC3_RTC_REG_SECONDS)),
		LatchedMinutes:              uint32(m.latchedRTC.readReg(MBC3_RTC_REG_MINUTES)),
		LatchedHours:                uint32(m.latchedRTC.readReg(MBC3_RTC_REG_HOURS)),
		LatchedDays:                 uint32(m.latchedRTC.readReg(MBC3_RTC_REG_DAY_LOW)),
		LatchedDaysHighOverflowHalt: uint32(m.latchedRTC.readReg(MBC3_RTC_REG_DAY_HIGH)),
		UnixTimestamp:               time.Now().Unix(),
	}

	err := binary.Write(w, binary.LittleEndian, rtc)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}

	return nil
}

func (m *MBC3) loadRTCRegsFromSave(r io.Reader) error {
	savedRTC := mbc3SaveRTC{}
	err := binary.Read(r, binary.LittleEndian, &savedRTC)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("decoding: %w", err)
	}

	rtc := mbc3RTCRegs{
		Seconds:   uint8(savedRTC.CurrentSeconds),
		Minutes:   uint8(savedRTC.CurrentMinutes),
		Hours:     uint8(savedRTC.CurrentHours),
		Days:      uint32(savedRTC.CurrentDays & 0xFF),
		Timestamp: time.Unix(savedRTC.UnixTimestamp, 0),
	}
	rtc.writeReg(MBC3_RTC_REG_DAY_HIGH, byte(savedRTC.CurrentDaysHighOverflowHalt&0xFF))

	latchedRTC := mbc3RTCRegs{
		Seconds:   uint8(savedRTC.LatchedSeconds),
		Minutes:   uint8(savedRTC.LatchedMinutes),
		Hours:     uint8(savedRTC.LatchedHours),
		Days:      uint32(savedRTC.LatchedDays & 0xFF),
		Timestamp: time.Unix(savedRTC.UnixTimestamp, 0),
	}
	latchedRTC.writeReg(MBC3_RTC_REG_DAY_HIGH, byte(savedRTC.LatchedDaysHighOverflowHalt&0xFF))

	m.rtc = rtc
	m.rtc.advanceTime(time.Now())

	m.latchedRTC = latchedRTC

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

func (m *MBC30) DebugPrint(w io.Writer) {
	m.MBC3.DebugPrint(w)
}

func (m *MBC30) Save(w io.Writer) error {
	return m.MBC3.Save(w)
}

func (m *MBC30) LoadSave(r io.Reader) error {
	return m.MBC3.LoadSave(r)
}
