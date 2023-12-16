package mem

import (
	"testing"
)

const (
	handlerReadReplaceValue  = 0xAB
	handlerWriteReplaceValue = 0xEA
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

func assertAddrEquals[T uint8 | uint16](t *testing.T, actual T, expected T) {
	if actual != expected {
		t.Errorf("Expected 0x%X, but got 0x%x", expected, actual)
	}
}

func TestMmuBasicReads(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	ram[0x100] = 0xC0
	ram[0x101] = 0xEE
	ram[0x102] = 0xFF

	assertAddrEquals(t, mmu.Read8(0x100), 0xC0)
	assertAddrEquals(t, mmu.Read16(0x101), 0xFFEE)
}

func TestMmuBasicWrites(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.Write8(0x100, 0xC0)
	mmu.Write16(0x101, 0xFFEE)

	assertAddrEquals(t, ram[0x100], 0xC0)
	assertAddrEquals(t, ram[0x101], 0xEE)
	assertAddrEquals(t, ram[0x102], 0xFF)
}

func TestMmuReadHandlerReplacement(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assertAddrEquals(t, mmu.Read8(0x103), handlerReadReplaceValue)
	assertAddrEquals(t, mmu.Read8(0x201), 0x22)
}

func TestMmuReadHandlerPassthrough(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	ram[0x103] = 0x11
	ram[0x201] = 0x22

	assertAddrEquals(t, mmu.Read8(0x103), 0x11)
	assertAddrEquals(t, mmu.Read8(0x201), 0x22)
}

func TestMmuWriteHandlerReplacement(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testReplacementHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assertAddrEquals(t, ram[0x103], handlerWriteReplaceValue)
	assertAddrEquals(t, ram[0x201], 0x22)
}

func TestMmuWriteHandlerPassthrough(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testPassthroughHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assertAddrEquals(t, ram[0x103], 0x11)
	assertAddrEquals(t, ram[0x201], 0x22)
}

func TestMmuWriteHandlerBlock(t *testing.T) {
	ram := make([]byte, 0xFFFF)
	mmu := NewMMU(ram)

	mmu.AddHandler(MemRegion{Start: 0x100, End: 0x200}, &testWriteBlockHandler{})

	mmu.Write8(0x103, 0x11)
	mmu.Write8(0x201, 0x22)

	assertAddrEquals(t, ram[0x103], 0x00)
	assertAddrEquals(t, ram[0x201], 0x22)
}
