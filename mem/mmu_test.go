package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	handlerReadReplaceValue  byte = 0xAB
	handlerWriteReplaceValue byte = 0xEA
)

type testReplacementHandler struct{}

func (h *testReplacementHandler) OnRead(mmu *MMU, addr uint16) MemRead {
	return ReadReplace(handlerReadReplaceValue)
}

func (h *testReplacementHandler) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	return WriteReplace(handlerWriteReplaceValue)
}

type testPassthroughHandler struct{}

func (h *testPassthroughHandler) OnRead(mmu *MMU, addr uint16) MemRead {
	return ReadPassthrough()
}

func (h *testPassthroughHandler) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	return WritePassthrough()
}

type testWriteBlockHandler struct{}

func (h *testWriteBlockHandler) OnRead(mmu *MMU, addr uint16) MemRead {
	return ReadPassthrough()
}

func (h *testWriteBlockHandler) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	return WriteBlock()
}

func TestMmuBasicReads(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	ram[0x100] = 0xC0
	ram[0x101] = 0xEE
	ram[0x102] = 0xFF

	assert.EqualValues(mmu.Read8(0x100), byte(0xC0))
	assert.EqualValues(mmu.Read16(0x101), 0xFFEE)
}

func TestMmuBasicWrites(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.Write8(0x100, 0xC0)
	mmu.Write16(0x101, 0xFFEE)

	assert.Equal(ram[0x100], byte(0xC0))
	assert.Equal(ram[0x101], byte(0xEE))
	assert.Equal(ram[0x102], byte(0xFF))
}

func TestMmuReadHandlerReplacement(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assert.Equal(mmu.Read8(0x103), handlerReadReplaceValue)
	assert.Equal(mmu.Read8(0x201), byte(0x22))
}

func TestMmuReadHandlerPassthrough(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assert.Equal(mmu.Read8(0x103), byte(0x11))
	assert.Equal(mmu.Read8(0x201), byte(0x22))
}

func TestMmuWriteHandlerReplacement(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(ram[0x103], handlerWriteReplaceValue)
	assert.Equal(ram[0x201], byte(0x22))
}

func TestMmuWriteHandlerPassthrough(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(ram[0x103], byte(0x11))
	assert.Equal(ram[0x201], byte(0x22))
}

func TestMmuWriteHandlerBlock(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testWriteBlockHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(ram[0x103], byte(0x00))
	assert.Equal(ram[0x201], byte(0x22))
}
