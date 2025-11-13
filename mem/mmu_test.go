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

	assert.Equal(byte(0xC0), mmu.Read8(0x100))
	assert.Equal(uint16(0xFFEE), mmu.Read16(0x101))
}

func TestMmuBasicWrites(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.Write8(0x100, 0xC0)
	mmu.Write16(0x101, 0xFFEE)

	assert.Equal(byte(0xC0), ram[0x100])
	assert.Equal(byte(0xEE), ram[0x101])
	assert.Equal(byte(0xFF), ram[0x102])
}

func TestMmuReadHandlerReplacement(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assert.Equal(handlerReadReplaceValue, mmu.Read8(0x103))
	assert.Equal(byte(0x22), mmu.Read8(0x201))
}

func TestMmuReadHandlerPassthrough(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assert.Equal(byte(0x11), mmu.Read8(0x103))
	assert.Equal(byte(0x22), mmu.Read8(0x201))
}

func TestMmuWriteHandlerReplacement(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(handlerWriteReplaceValue, ram[0x103])
	assert.Equal(byte(0x22), ram[0x201])
}

func TestMmuWriteHandlerPassthrough(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(byte(0x11), ram[0x103])
	assert.Equal(byte(0x22), ram[0x201])
}

func TestMmuWriteHandlerBlock(t *testing.T) {
	assert := assert.New(t)

	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testWriteBlockHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assert.Equal(byte(0x00), ram[0x103])
	assert.Equal(byte(0x22), ram[0x201])
}
