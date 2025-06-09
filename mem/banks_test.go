package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const bankSize16KB = 0x4000

func makeBankedMemory(bankSize, banks uint) []byte {
	memory := make([]byte, bankSize*banks)

	for bankNum := range banks {
		for bankSlot := range bankSize {
			memory[(bankSize*bankNum)+bankSlot] = byte(bankNum)
		}
	}

	return memory
}

func TestReadBankAddr(t *testing.T) {
	assert := assert.New(t)
	memory := makeBankedMemory(bankSize16KB, 8)

	banksRegion := MemRegion{
		Start: 0x0000,
		End:   0x3FFF,
	}

	assert.Equal(
		byte(0x0),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 0, 0),
	)
	assert.Equal(
		byte(0x0),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 0, bankSize16KB-1),
	)
	assert.Equal(
		byte(0x1),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 1, 0),
	)
	assert.Equal(
		byte(0x1),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 1, bankSize16KB-1),
	)
	assert.Equal(
		byte(0x6),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 6, 0),
	)
	assert.Equal(
		byte(0x6),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 6, bankSize16KB-1),
	)

	// Out-of-bounds reads should return from masked bank addr (so, 0x2 in this case)
	assert.Equal(
		byte(0x2),
		ReadBankAddr(memory, banksRegion, bankSize16KB, 10, bankSize16KB-1),
	)
}

func TestWriteBankAddr(t *testing.T) {
	assert := assert.New(t)
	memory := makeBankedMemory(bankSize16KB, 8)

	banksRegion := MemRegion{
		Start: 0x0000,
		End:   0x3FFF,
	}

	WriteBankAddr(memory, banksRegion, bankSize16KB, 0, 0, 0xFF)
	WriteBankAddr(memory, banksRegion, bankSize16KB, 0, bankSize16KB-1, 0xFF)
	WriteBankAddr(memory, banksRegion, bankSize16KB, 1, 0, 0xFE)
	WriteBankAddr(memory, banksRegion, bankSize16KB, 1, bankSize16KB-1, 0xFE)
	WriteBankAddr(memory, banksRegion, bankSize16KB, 6, 0, 0xFD)
	WriteBankAddr(memory, banksRegion, bankSize16KB, 6, bankSize16KB-1, 0xFD)

	assert.Equal(
		byte(0xFF),
		memory[0],
	)
	assert.Equal(
		byte(0xFF),
		memory[bankSize16KB-1],
	)
	assert.Equal(
		byte(0xFE),
		memory[bankSize16KB*1],
	)
	assert.Equal(
		byte(0xFE),
		memory[bankSize16KB*1+(bankSize16KB-1)],
	)
	assert.Equal(
		byte(0xFD),
		memory[bankSize16KB*6],
	)
	assert.Equal(
		byte(0xFD),
		memory[bankSize16KB*6+(bankSize16KB-1)],
	)

	// Out-of-bounds writes should write to masked bank address
	WriteBankAddr(memory, banksRegion, bankSize16KB, 10, bankSize16KB-1, 0x22)
	assert.Equal(
		byte(0x22),
		memory[bankSize16KB*2+(bankSize16KB-1)],
	)
}
