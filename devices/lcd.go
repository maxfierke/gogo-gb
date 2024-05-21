package devices

import (
	"image"
	"image/color"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_LCD_LY = 0xFF44
)

type LCD struct {
}

func NewLCD() *LCD {
	return &LCD{}
}

func (lcd *LCD) Draw() image.Image {
	image := image.NewPaletted(
		image.Rect(0, 0, 160, 144),
		color.Palette{color.Black, color.Gray{Y: 96}, color.Gray16{Y: 128}, color.White},
	)
	image.Set(80, 77, color.Gray{Y: 96})

	return image
}

func (lcd *LCD) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return mem.ReadPassthrough()
}

func (lcd *LCD) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WriteBlock()
}
