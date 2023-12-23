package devices

import "github.com/maxfierke/gogo-gb/mem"

const (
	REG_LCD_LY = 0xFF44
)

type LCD struct {
}

func NewLCD() *LCD {
	return &LCD{}
}

func (lcd *LCD) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return mem.ReadPassthrough()
}

func (lcd *LCD) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WriteBlock()
}
