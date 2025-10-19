package ppu

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_DMA_OAM uint16 = 0xFF46
)

type dmaRequest struct {
	addr  uint16
	value byte
}

type DMA struct {
	enabled    bool
	clock      uint
	pendingDMA []*dmaRequest
}

func (d *DMA) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_DMA_OAM {
		return mem.ReadPassthrough()
	}

	panic(fmt.Sprintf("Attempting to read @ 0x%04X, which is out-of-bounds for DMA", addr))
}

func (d *DMA) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_DMA_OAM {
		if !d.enabled {
			srcAddrStart := uint16(value) << 8

			for oamAddr := uint8(0); oamAddr < 160; oamAddr++ {
				srcAddr := srcAddrStart + uint16(oamAddr)
				copiedValue := mmu.Read8(srcAddr)

				d.pendingDMA = append(d.pendingDMA,
					&dmaRequest{
						addr:  OAM_START + uint16(oamAddr),
						value: copiedValue,
					},
				)
			}
			d.enabled = true
		}

		// Reads should return last written value, so we'll pass it through it
		return mem.WritePassthrough()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for DMA", value, addr))
}

var _ mem.MemHandler = (*DMA)(nil)

func NewDMA() *DMA {
	return &DMA{
		pendingDMA: make([]*dmaRequest, 0, 160),
	}
}

func (d *DMA) Step(mmu *mem.MMU, cycles uint8) {
	if !d.enabled {
		return
	}

	d.clock += uint(cycles)

	if d.clock >= 640 {
		d.clock = 0

		for _, request := range d.pendingDMA {
			mmu.Write8(request.addr, request.value)
		}

		d.pendingDMA = make([]*dmaRequest, 0, 160)
		d.enabled = false
	}
}
