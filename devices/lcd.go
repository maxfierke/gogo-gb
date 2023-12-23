package devices

import "github.com/maxfierke/gogo-gb/mem"

const (
	lcd_reg_ly = 0xFF44
)

type LCD struct {
}

func NewLCD() *LCD {
	return &LCD{}
}

func (lcd *LCD) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == lcd_reg_ly {
		return mem.ReadReplace(0x90) // gameboy-doctor
	}

	return mem.ReadPassthrough()
}

func (lcd *LCD) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WriteBlock()
}
