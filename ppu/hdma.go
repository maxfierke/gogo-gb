package ppu

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/bits"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_HDMA_SRC_HIGH uint16 = 0xFF51
	REG_HDMA_SRC_LOW  uint16 = 0xFF52

	REG_HDMA_DST_HIGH uint16 = 0xFF53
	REG_HDMA_DST_LOW  uint16 = 0xFF54

	REG_HDMA_LEN_MODE_START uint16 = 0xFF55

	REG_HDMA_SRC_LOW_MASK  = 0xF0
	REG_HDMA_DST_HIGH_MASK = 0x1F
	REG_HDMA_DST_LOW_MASK  = 0xF0
	REG_HDMA_LEN_MASK      = 0x7F
)

type HDMAMode uint8

const (
	HDMA_MODE_GENERAL HDMAMode = iota
	HDMA_MODE_HBLANK
)

type HDMA struct {
	active bool

	mode     HDMAMode
	srcAddr  uint16
	destAddr uint16
	length   uint8

	vram *VRAM
}

const hdmaBytesInBlock = 16

var _ mem.MemHandler = (*HDMA)(nil)

func NewHDMA(vram *VRAM) *HDMA {
	return &HDMA{
		vram: vram,
	}
}

func (d *HDMA) IsActive(dmaMode HDMAMode) bool {
	return d.active && d.mode == dmaMode
}

func (d *HDMA) Step(mmu *mem.MMU, doubleSpeed bool) uint8 {
	if !d.active {
		return 0
	}

	for i := range uint16(hdmaBytesInBlock) {
		destAddr := d.destAddr + i
		srcAddr := d.srcAddr + i
		value := mmu.Read8(srcAddr)

		if destAddr < VRAM_START || destAddr > VRAM_END {
			panic(fmt.Sprintf("illegal HDMA write to 0x%04X", destAddr))
		}
		d.vram.Write(destAddr-VRAM_START, value)
	}
	d.srcAddr += hdmaBytesInBlock
	d.destAddr += hdmaBytesInBlock
	d.length--

	if d.length == 0 {
		d.active = false
	}

	// https://gbdev.io/pandocs/CGB_Registers.html#transfer-timings
	cycles := 16
	if doubleSpeed {
		cycles = 64
	}

	return uint8(cycles)
}

func (d *HDMA) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_HDMA_LEN_MODE_START {
		if d.active {
			return mem.ReadReplace((d.length - 1) & REG_HDMA_LEN_MASK)
		} else if d.mode == HDMA_MODE_HBLANK {
			return mem.ReadReplace((uint8(HDMA_MODE_HBLANK) << 7) | ((d.length - 1) & REG_HDMA_LEN_MASK))
		}

		return mem.ReadReplace(0xFF)
	}

	panic(fmt.Sprintf("Attempting to read @ 0x%04X, which is out-of-bounds for HDMA", addr))
}

func (d *HDMA) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_HDMA_LEN_MODE_START {
		if !d.active {
			d.active = true
			d.mode = HDMAMode(bits.Read(value, 7))
			d.length = (value & REG_HDMA_LEN_MASK) + 1
		} else if d.active && d.mode != HDMAMode(bits.Read(value, 7)) {
			d.active = false
		}

		return mem.WriteBlock()
	}

	if addr == REG_HDMA_SRC_HIGH {
		d.srcAddr = (d.srcAddr & 0xFF) | uint16(value)<<8

		return mem.WriteBlock()
	}

	if addr == REG_HDMA_SRC_LOW {
		d.srcAddr = (d.srcAddr & 0xFF00) | uint16(value&REG_HDMA_SRC_LOW_MASK)

		return mem.WriteBlock()
	}

	if addr == REG_HDMA_DST_HIGH {
		d.destAddr = (d.destAddr & 0xFF) | uint16(value&REG_HDMA_DST_HIGH_MASK)<<8
		d.destAddr |= VRAM_START

		return mem.WriteBlock()
	}

	if addr == REG_HDMA_DST_LOW {
		d.destAddr = (d.destAddr & 0xFF00) | uint16(value&REG_HDMA_DST_LOW_MASK)
		d.destAddr |= VRAM_START

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for HDMA", value, addr))
}
