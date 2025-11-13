package mbc

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRom(banks int) []byte {
	rom := make([]byte, ROM_BANK_SIZE*banks)

	for bankNum := range banks {
		for bankSlot := range ROM_BANK_SIZE {
			rom[(ROM_BANK_SIZE*bankNum)+bankSlot] = byte(bankNum)
		}
	}

	return rom
}

func makeRam(banks int) []byte {
	ram := make([]byte, RAM_BANK_SIZE*banks)

	for bankNum := range banks {
		for bankSlot := range RAM_BANK_SIZE {
			ram[(RAM_BANK_SIZE*bankNum)+bankSlot] = byte(bankNum)
		}
	}

	return ram
}

func TestMBC3RTCRegs_writeReg_readReg(t *testing.T) {
	testCases := map[string]struct {
		initialRegs          mbc3RTCRegs
		reg                  mbc3RTCReg
		writeValue           uint8
		expectedValue        uint8
		expectedDays         uint32
		expectedHalt         bool
		expectedDaysOverflow bool
	}{
		"seconds - in-range value": {
			reg:           MBC3_RTC_REG_SECONDS,
			writeValue:    59,
			expectedValue: 59,
		},
		"seconds - out-of-range value": {
			reg:           MBC3_RTC_REG_SECONDS,
			writeValue:    64,
			expectedValue: 0,
		},
		"minutes - in-range value": {
			reg:           MBC3_RTC_REG_MINUTES,
			writeValue:    59,
			expectedValue: 59,
		},
		"minutes - out-of-range value": {
			reg:           MBC3_RTC_REG_MINUTES,
			writeValue:    64,
			expectedValue: 0,
		},
		"hours - in-range value": {
			reg:           MBC3_RTC_REG_HOURS,
			writeValue:    23,
			expectedValue: 23,
		},
		"hours - out-of-range value": {
			reg:           MBC3_RTC_REG_HOURS,
			writeValue:    30,
			expectedValue: 30,
		},
		"days low - days is not using high bit": {
			reg:           MBC3_RTC_REG_DAY_LOW,
			writeValue:    255,
			expectedValue: 255,
		},
		"days low - days is using high bit": {
			initialRegs: mbc3RTCRegs{
				Days: 256,
			},
			reg:           MBC3_RTC_REG_DAY_LOW,
			writeValue:    255,
			expectedValue: 255,
			expectedDays:  511,
		},
		"days high - days is not using high bit, no halt bit, no overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 255,
			},
			reg:          MBC3_RTC_REG_DAY_HIGH,
			expectedDays: 255,
		},
		"days high - days is not using high bit, halt bit, no overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 255,
			},
			reg:           MBC3_RTC_REG_DAY_HIGH,
			expectedValue: 0b01000000,
			writeValue:    1 << MBC3_RTC_REG_DAY_HIGH_BIT_HALT,
			expectedDays:  255,
			expectedHalt:  true,
		},
		"days high - days is not using high bit, no halt bit, overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 255,
			},
			reg:                  MBC3_RTC_REG_DAY_HIGH,
			writeValue:           1 << MBC3_RTC_REG_DAY_HIGH_BIT_CARRY,
			expectedValue:        0b10000000,
			expectedDays:         255,
			expectedDaysOverflow: true,
		},
		"days high - days is not using high bit, halt bit, overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 255,
			},
			reg:                  MBC3_RTC_REG_DAY_HIGH,
			writeValue:           (1 << MBC3_RTC_REG_DAY_HIGH_BIT_HALT) | (1 << MBC3_RTC_REG_DAY_HIGH_BIT_CARRY),
			expectedValue:        0b11000000,
			expectedDays:         255,
			expectedDaysOverflow: true,
			expectedHalt:         true,
		},
		"days high - days is using high bit, no halt bit, no overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 257,
			},
			reg:          MBC3_RTC_REG_DAY_HIGH,
			expectedDays: 1,
		},
		"days high - days is using high bit, halt bit, no overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 257,
			},
			reg:           MBC3_RTC_REG_DAY_HIGH,
			writeValue:    (1 << MBC3_RTC_REG_DAY_HIGH_BIT_HALT) | 1,
			expectedValue: 0b01000001,
			expectedDays:  257,
			expectedHalt:  true,
		},
		"days high - days is using high bit, no halt bit, overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 257,
			},
			reg:                  MBC3_RTC_REG_DAY_HIGH,
			writeValue:           (1 << MBC3_RTC_REG_DAY_HIGH_BIT_CARRY) | 1,
			expectedValue:        0b10000001,
			expectedDays:         257,
			expectedDaysOverflow: true,
		},
		"days high - days is using high bit, halt bit, overflow": {
			initialRegs: mbc3RTCRegs{
				Days: 257,
			},
			reg:                  MBC3_RTC_REG_DAY_HIGH,
			writeValue:           (1 << MBC3_RTC_REG_DAY_HIGH_BIT_CARRY) | (1 << MBC3_RTC_REG_DAY_HIGH_BIT_HALT) | 1,
			expectedValue:        0b11000001,
			expectedDays:         257,
			expectedDaysOverflow: true,
			expectedHalt:         true,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			regs := testCase.initialRegs

			regs.writeReg(testCase.reg, testCase.writeValue)

			assert.Equal(testCase.expectedValue, regs.readReg(testCase.reg))
			assert.Equal(testCase.expectedHalt, regs.Halt)
			assert.Equal(testCase.expectedDaysOverflow, regs.DaysOverflow)
		})
	}
}

func TestMBC3RTCRegs_advanceTime(t *testing.T) {
	assert := assert.New(t)

	startTimestamp := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)

	regs := mbc3RTCRegs{Timestamp: startTimestamp}
	originalRegs := regs

	// No change in time
	regs.advanceTime(startTimestamp)
	assert.Equal(originalRegs, regs)

	// Below second threshold
	regs.advanceTime(startTimestamp.Add(time.Microsecond))
	assert.Equal(originalRegs, regs)

	regs.advanceTime(startTimestamp.Add(time.Millisecond))
	assert.Equal(originalRegs, regs)

	// Halt
	regs.Halt = true

	// +1 second (halted)
	regs.advanceTime(startTimestamp.Add(time.Second))
	assert.EqualValues(0, regs.Days)
	assert.EqualValues(0, regs.Hours)
	assert.EqualValues(0, regs.Minutes)
	assert.EqualValues(0, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// Resume
	regs.Halt = false

	// +1 second
	regs.advanceTime(startTimestamp.Add(time.Second))
	assert.EqualValues(0, regs.Days)
	assert.EqualValues(0, regs.Hours)
	assert.EqualValues(0, regs.Minutes)
	assert.EqualValues(1, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +1 minute
	regs.advanceTime(regs.Timestamp.Add(time.Minute))
	assert.EqualValues(0, regs.Days)
	assert.EqualValues(0, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(1, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +1 hour
	regs.advanceTime(regs.Timestamp.Add(time.Hour))
	assert.EqualValues(0, regs.Days)
	assert.EqualValues(1, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(1, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +23 Hours
	regs.advanceTime(regs.Timestamp.Add(23 * time.Hour))
	assert.EqualValues(1, regs.Days)
	assert.EqualValues(0, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(1, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +59 Minutes
	regs.advanceTime(regs.Timestamp.Add(59 * time.Minute))
	assert.EqualValues(1, regs.Days)
	assert.EqualValues(1, regs.Hours)
	assert.EqualValues(0, regs.Minutes)
	assert.EqualValues(1, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +59 seconds
	regs.advanceTime(regs.Timestamp.Add(59 * time.Second))
	assert.EqualValues(1, regs.Days)
	assert.EqualValues(1, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(0, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +30 hours
	regs.advanceTime(regs.Timestamp.Add(30 * time.Hour))
	assert.EqualValues(2, regs.Days)
	assert.EqualValues(7, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(0, regs.Seconds)
	assert.False(regs.DaysOverflow)

	// +510 Days
	regs.advanceTime(regs.Timestamp.Add(12240 * time.Hour))
	assert.EqualValues(0, regs.Days)
	assert.EqualValues(7, regs.Hours)
	assert.EqualValues(1, regs.Minutes)
	assert.EqualValues(0, regs.Seconds)
	assert.True(regs.DaysOverflow)
}

func TestMBC3_Save_LoadSave_rtc(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	mbc3 := NewMBC3(makeRom(8), makeRam(4), true)

	currentRtcTimestamp := time.Now().Add(-time.Hour)
	latchedRtcTimestamp := currentRtcTimestamp.Add(-5 * time.Second)

	currentRTC := mbc3RTCRegs{
		Seconds:   15,
		Minutes:   30,
		Hours:     12,
		Days:      36,
		Timestamp: currentRtcTimestamp,
	}
	latchedRTC := mbc3RTCRegs{
		Seconds:      10,
		Minutes:      30,
		Hours:        12,
		Days:         36,
		Timestamp:    latchedRtcTimestamp,
		DaysOverflow: true,
		Halt:         true,
	}

	mbc3.rtc = currentRTC
	mbc3.latchedRTC = latchedRTC

	var saveFile bytes.Buffer

	err := mbc3.Save(&saveFile)
	require.NoError(err)

	mbc3 = NewMBC3(makeRom(8), makeRam(4), true)
	err = mbc3.LoadSave(&saveFile)
	require.NoError(err)

	rtcDiff := time.Since(currentRtcTimestamp)

	assert.Equal(currentRTC.Days, mbc3.rtc.Days)
	assert.Equal(currentRTC.DaysOverflow, mbc3.rtc.DaysOverflow)
	assert.Equal(currentRTC.Halt, mbc3.rtc.Halt)
	assert.Equal(currentRTC.Hours+1, mbc3.rtc.Hours)
	assert.Equal(currentRTC.Minutes, mbc3.rtc.Minutes)
	assert.Equal(currentRTC.Seconds+(uint8(rtcDiff.Seconds())/60), mbc3.rtc.Seconds)
	assert.WithinDuration(currentRTC.Timestamp.Add(rtcDiff), mbc3.rtc.Timestamp, time.Second)

	assert.Equal(latchedRTC.Days, mbc3.latchedRTC.Days)
	assert.Equal(latchedRTC.DaysOverflow, mbc3.latchedRTC.DaysOverflow)
	assert.Equal(latchedRTC.Halt, mbc3.latchedRTC.Halt)
	assert.Equal(latchedRTC.Hours, mbc3.latchedRTC.Hours)
	assert.Equal(latchedRTC.Minutes, mbc3.latchedRTC.Minutes)
	assert.Equal(latchedRTC.Seconds, mbc3.latchedRTC.Seconds)
	assert.WithinDuration(currentRTC.Timestamp.Add(rtcDiff), mbc3.rtc.Timestamp, time.Second)
}
