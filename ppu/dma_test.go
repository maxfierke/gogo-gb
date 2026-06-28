package ppu

import (
	"testing"

	"github.com/maxfierke/gogo-gb/mem"
	"github.com/stretchr/testify/require"
)

const DMG_RAM_SIZE = 0x10000

func populateRam(mmu *mem.MMU, start uint16, len int) {
	for i := range len {
		mmu.Write8(start+uint16(i), byte(i+1))
	}
}

func TestDMAOnRead_ReadsLastWrittenValue(t *testing.T) {
	require := require.New(t)

	dma := NewDMA()
	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	mmu.AddHandler(mem.MemRegion{Start: REG_DMA_OAM, End: REG_DMA_OAM}, dma)

	mmu.Write8(REG_DMA_OAM, 0x80)

	require.Equal(byte(0x80), mmu.Read8(REG_DMA_OAM))
}

func TestDMAOnWrite_StartsTransfer(t *testing.T) {
	require := require.New(t)

	dma := NewDMA()
	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)

	populateRam(mmu, 0x8000, OAM_TRANSFER_LENGTH)

	dma.OnWrite(mmu, REG_DMA_OAM, 0x80)

	require.True(dma.enabled)
	require.Equal(dma.pendingDMA[0].addr, uint16(0xFE00))
	require.Equal(dma.pendingDMA[0].value, byte(0x01))
}

func TestDMAOnWrite_IgnoredWhileActive(t *testing.T) {
	require := require.New(t)

	dma := NewDMA()
	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)

	for i := range OAM_TRANSFER_LENGTH {
		mmu.Write8(VRAM_START+uint16(i), 0xAA)
		mmu.Write8(0x8A00+uint16(i), 0xBB)
	}

	// Start a transfer
	dma.OnWrite(mmu, REG_DMA_OAM, 0x80)

	dma.Step(mmu, 255)

	// Attempt to restart from a different region while active. should be ignored.
	dma.OnWrite(mmu, REG_DMA_OAM, 0x8A)

	dma.Step(mmu, 255)
	dma.Step(mmu, 130)

	for i := range OAM_TRANSFER_LENGTH {
		require.Equal(byte(0xAA), mmu.Read8(OAM_START+uint16(i)))
	}
}

func TestDMAStep_NoOpWhenInactive(t *testing.T) {
	require := require.New(t)

	dma := NewDMA()
	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	populateRam(mmu, 0x8000, OAM_TRANSFER_LENGTH)

	// Step a whole DMA transfer length
	dma.Step(mmu, 255)
	dma.Step(mmu, 255)
	dma.Step(mmu, 130)

	for i := range OAM_TRANSFER_LENGTH {
		require.Equal(byte(0), mmu.Read8(OAM_START+uint16(i)))
	}
}

func TestDMATransfer(t *testing.T) {
	require := require.New(t)

	dma := NewDMA()
	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	populateRam(mmu, 0x8000, OAM_TRANSFER_LENGTH)

	dma.OnWrite(mmu, REG_DMA_OAM, 0x80)

	// Step 639 cycles
	dma.Step(mmu, 255)
	dma.Step(mmu, 255)
	dma.Step(mmu, 129)

	require.True(dma.enabled)
	require.Len(dma.pendingDMA, OAM_TRANSFER_LENGTH)

	// Verify OAM not written yet
	require.Equal(byte(0), mmu.Read8(OAM_START))

	dma.Step(mmu, 1)

	for i := range OAM_TRANSFER_LENGTH {
		dest := OAM_START + uint16(i)
		require.Equal(byte(i+1), mmu.Read8(dest))
	}

	// Verify only OAM was written
	require.Equal(byte(0), mmu.Read8(OAM_END+1))
	require.False(dma.enabled)
	require.Empty(dma.pendingDMA)
}
