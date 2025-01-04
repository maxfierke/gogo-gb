package mbc

import (
	"testing"

	"github.com/maxfierke/gogo-gb/mem"
	"github.com/stretchr/testify/assert"
)

func makeRom(banks int) []byte {
	rom := make([]byte, ROM_BANK_SIZE*banks)

	for bankNum := 0; bankNum < banks; bankNum++ {
		for bankSlot := 0; bankSlot < ROM_BANK_SIZE; bankSlot++ {
			rom[(ROM_BANK_SIZE*bankNum)+bankSlot] = byte(bankNum)
		}
	}

	return rom
}

func TestReadBankAddr(t *testing.T) {
	assert := assert.New(t)
	rom := makeRom(8)

	banksRegion := mem.MemRegion{
		Start: 0x0000,
		End:   0x3FFF,
	}

	assert.Equal(
		byte(0x0),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 0, 0),
	)
	assert.Equal(
		byte(0x0),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 0, ROM_BANK_SIZE-1),
	)
	assert.Equal(
		byte(0x1),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 1, 0),
	)
	assert.Equal(
		byte(0x1),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 1, ROM_BANK_SIZE-1),
	)
	assert.Equal(
		byte(0x6),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 6, 0),
	)
	assert.Equal(
		byte(0x6),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 6, ROM_BANK_SIZE-1),
	)

	// Out-of-bounds reads should return from masked bank addr (so, 0x2 in this case)
	assert.Equal(
		byte(0x2),
		readBankAddr(rom, banksRegion, ROM_BANK_SIZE, 10, ROM_BANK_SIZE-1),
	)
}

func TestWriteBankAddr(t *testing.T) {
	assert := assert.New(t)
	rom := makeRom(8)

	banksRegion := mem.MemRegion{
		Start: 0x0000,
		End:   0x3FFF,
	}

	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 0, 0, 0xFF)
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 0, ROM_BANK_SIZE-1, 0xFF)
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 1, 0, 0xFE)
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 1, ROM_BANK_SIZE-1, 0xFE)
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 6, 0, 0xFD)
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 6, ROM_BANK_SIZE-1, 0xFD)

	assert.Equal(
		byte(0xFF),
		rom[0],
	)
	assert.Equal(
		byte(0xFF),
		rom[ROM_BANK_SIZE-1],
	)
	assert.Equal(
		byte(0xFE),
		rom[ROM_BANK_SIZE*1],
	)
	assert.Equal(
		byte(0xFE),
		rom[ROM_BANK_SIZE*1+(ROM_BANK_SIZE-1)],
	)
	assert.Equal(
		byte(0xFD),
		rom[ROM_BANK_SIZE*6],
	)
	assert.Equal(
		byte(0xFD),
		rom[ROM_BANK_SIZE*6+(ROM_BANK_SIZE-1)],
	)

	// Out-of-bounds writes should write to masked bank address
	writeBankAddr(rom, banksRegion, ROM_BANK_SIZE, 10, ROM_BANK_SIZE-1, 0x22)
	assert.Equal(
		byte(0x22),
		rom[ROM_BANK_SIZE*2+(ROM_BANK_SIZE-1)],
	)
}
