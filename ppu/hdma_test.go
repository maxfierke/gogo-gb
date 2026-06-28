package ppu

import (
	"testing"

	"github.com/maxfierke/gogo-gb/mem"
	"github.com/stretchr/testify/require"
)

func startHDMATransfer(hdma *HDMA, mmu *mem.MMU, srcAddr, destAddr uint16, blocks uint8, mode HDMAMode) {
	hdma.OnWrite(mmu, REG_HDMA_SRC_HIGH, byte(srcAddr>>8))
	hdma.OnWrite(mmu, REG_HDMA_SRC_LOW, byte(srcAddr&REG_HDMA_SRC_LOW_MASK))
	hdma.OnWrite(mmu, REG_HDMA_DST_HIGH, byte((destAddr>>8)&REG_HDMA_DST_HIGH_MASK))
	hdma.OnWrite(mmu, REG_HDMA_DST_LOW, byte(destAddr&REG_HDMA_DST_LOW_MASK))
	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, (uint8(mode)<<7)|((blocks-1)&REG_HDMA_LEN_MASK))
}

func TestHDMAOnWrite_SrcHigh(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_SRC_HIGH, 0x23)

	require.Equal(uint16(0x2300), hdma.srcAddr)
}

func TestHDMAOnWrite_SrcLow(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_SRC_LOW, 0xAF)

	require.Equal(uint16(0x00A0), hdma.srcAddr)
}

func TestHDMAOnWrite_DstHigh(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_DST_HIGH, 0xFF)
	require.Equal(uint16(0x9F00), hdma.destAddr)

	hdma.OnWrite(mmu, REG_HDMA_DST_HIGH, 0x00)
	require.Equal(VRAM_START, hdma.destAddr)
}

func TestHDMAOnWrite_DstLow(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_DST_HIGH, 0x00)
	hdma.OnWrite(mmu, REG_HDMA_DST_LOW, 0x0F)
	require.Equal(VRAM_START, hdma.destAddr)

	hdma.OnWrite(mmu, REG_HDMA_DST_LOW, 0xFF)
	require.Equal(uint16(0x80F0), hdma.destAddr)
}

func TestHDMAOnWrite_TransferLength_MinusOne(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, 0x00)

	require.Equal(uint8(1), hdma.length)
}

func TestHDMAOnWrite_TransferLength_MaxMasked(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, 0xFF)

	require.Equal(uint8(128), hdma.length)
}

func TestHDMAOnWrite_StartGeneralMode(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, 0x00)

	require.Equal(HDMAMode(HDMA_MODE_GENERAL), hdma.mode)
}

func TestHDMAOnWrite_StartHBlankMode(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, uint8(HDMA_MODE_HBLANK)<<7)

	require.Equal(HDMAMode(HDMA_MODE_HBLANK), hdma.mode)
}

func TestHDMAOnWrite_IgnoredWhileActive_SameMode(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 5, HDMA_MODE_HBLANK)
	require.True(hdma.active)
	require.Equal(uint8(5), hdma.length)

	// Attempt to start HBlank with different length. should be ignored.
	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, (uint8(HDMA_MODE_HBLANK)<<7)|0x9)

	require.True(hdma.active)
	require.Equal(uint8(5), hdma.length)
}

func TestHDMAOnWrite_CancelHBlank(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	for i := range 160 {
		mmu.Write8(0x2000+uint16(i), byte(i))
	}

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 10, HDMA_MODE_HBLANK)

	hdma.Step(mmu, false)
	hdma.Step(mmu, false)
	hdma.Step(mmu, false)
	require.True(hdma.active)
	require.Equal(uint8(7), hdma.length)

	hdma.OnWrite(mmu, REG_HDMA_LEN_MODE_START, 0x00)

	result := hdma.OnRead(mmu, REG_HDMA_LEN_MODE_START)
	require.Equal(mem.ReadReplace((uint8(HDMA_MODE_HBLANK)<<7)|0x6), result)
	require.False(hdma.active)
	require.Equal(HDMA_MODE_HBLANK, hdma.mode)
	require.Equal(uint8(7), hdma.length)
}

func TestHDMAOnRead_Inactive(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	result := hdma.OnRead(mmu, REG_HDMA_LEN_MODE_START)

	require.Equal(mem.ReadReplace(0xFF), result)
}

func TestHDMAOnRead_HBlankInProgress(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 5, HDMA_MODE_HBLANK)
	require.True(hdma.active)
	require.Equal(uint8(5), hdma.length)

	result := hdma.OnRead(mmu, REG_HDMA_LEN_MODE_START)

	require.Equal(mem.ReadReplace(hdma.length-1), result)
}

func TestHDMAOnRead_Inactive_GeneralCompleted(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 1, HDMA_MODE_GENERAL)
	hdma.Step(mmu, false)
	require.False(hdma.active)

	result := hdma.OnRead(mmu, REG_HDMA_LEN_MODE_START)

	require.Equal(mem.ReadReplace(0xFF), result)
}

func TestHDMAOnRead_Inactive_HBlankCompleted(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 1, HDMA_MODE_HBLANK)
	hdma.Step(mmu, false)
	require.False(hdma.active)

	result := hdma.OnRead(mmu, REG_HDMA_LEN_MODE_START)

	require.Equal(mem.ReadReplace(0xFF), result)
}

func TestHDMAIsActive_Inactive(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)

	require.False(hdma.IsActive(HDMA_MODE_GENERAL))
	require.False(hdma.IsActive(HDMA_MODE_HBLANK))
}

func TestHDMAIsActive_HBlankActive(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 5, HDMA_MODE_HBLANK)

	require.False(hdma.IsActive(HDMA_MODE_GENERAL))
	require.True(hdma.IsActive(HDMA_MODE_HBLANK))
}

func TestHDMAIsActive_GeneralActive(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 5, HDMA_MODE_GENERAL)

	require.True(hdma.IsActive(HDMA_MODE_GENERAL))
	require.False(hdma.IsActive(HDMA_MODE_HBLANK))
}

func TestHDMAStep_Cycles_Inactive(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	cycles := hdma.Step(mmu, false)

	require.Equal(uint8(0), cycles)
}

func TestHDMAStep_TransferBlock(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	for i := range 32 {
		mmu.Write8(0x2000+uint16(i), byte(i+1))
	}

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 2, HDMA_MODE_GENERAL)
	require.Equal(uint16(0x2000), hdma.srcAddr)
	require.Equal(uint16(VRAM_START), hdma.destAddr)
	require.Equal(uint8(2), hdma.length)
	require.True(hdma.active)
	require.Equal(HDMA_MODE_GENERAL, hdma.mode)

	cycles := hdma.Step(mmu, false)
	require.Equal(uint8(16), cycles)

	// First 16 bytes in VRAM should match source block
	for i := range 16 {
		require.Equal(byte(i+1), vram.Read(uint16(i)))
	}

	// Next block should not be written yet, but queued up
	require.Equal(byte(0), vram.Read(16))
	require.Equal(uint16(0x2010), hdma.srcAddr)
	require.Equal(uint16(0x8010), hdma.destAddr)
	require.Equal(uint8(1), hdma.length)
	require.True(hdma.active)
	require.Equal(HDMA_MODE_GENERAL, hdma.mode)
}

func TestHDMAStep_TransferBlock_DoubleSpeed(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	for i := range 32 {
		mmu.Write8(0x2000+uint16(i), byte(i+1))
	}

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, 1, HDMA_MODE_GENERAL)
	require.Equal(uint16(0x2000), hdma.srcAddr)
	require.Equal(uint16(VRAM_START), hdma.destAddr)
	require.Equal(uint8(1), hdma.length)
	require.True(hdma.active)
	require.Equal(HDMA_MODE_GENERAL, hdma.mode)

	cycles := hdma.Step(mmu, true)
	require.Equal(uint8(64), cycles)

	// First 16 bytes in VRAM should match source block
	for i := range 16 {
		require.Equal(byte(i+1), vram.Read(uint16(i)))
	}

	require.False(hdma.active)
	require.Equal(HDMA_MODE_GENERAL, hdma.mode)
}

func TestHDMAStep_FullTransfer(t *testing.T) {
	require := require.New(t)

	vram := NewVRAM()
	hdma := NewHDMA(vram)
	mmu := mem.NewMMU(make([]byte, DMG_RAM_SIZE))

	var numBlocks uint8 = 3

	for i := range hdmaBytesInBlock * numBlocks {
		mmu.Write8(0x2000+uint16(i), byte(i+1))
	}

	startHDMATransfer(hdma, mmu, 0x2000, VRAM_START, numBlocks, HDMA_MODE_GENERAL)

	for i := range numBlocks {
		hdma.Step(mmu, false)
		require.Equal(hdma.length != 0, hdma.active)
		require.Equal(numBlocks-(i+1), hdma.length)
	}

	for i := range hdmaBytesInBlock * numBlocks {
		require.Equal(byte(i+1), vram.Read(uint16(i)))
	}
	require.False(hdma.active)
	require.Empty(hdma.length)
}
