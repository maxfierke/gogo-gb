package ppu

import (
	"image/color"

	"github.com/maxfierke/gogo-gb/bits"
)

type ColorID uint8

const (
	COLOR_ID_WHITE ColorID = iota
	COLOR_ID_LIGHT_GRAY
	COLOR_ID_DARK_GRAY
	COLOR_ID_BLACK
	COLOR_ID_TRANSPARENT = COLOR_ID_WHITE
)

type objPalette [4]uint8

func (pal *objPalette) Read() uint8 {
	return (uint8(pal[3])<<6 |
		uint8(pal[2])<<4 |
		uint8(pal[1])<<2)
}

func (pal *objPalette) Write(value uint8) {
	pal[0] = 0x0 // This should be skipped over/treated as transparent
	pal[1] = (value & 0b0000_1100) >> 2
	pal[2] = (value & 0b0011_0000) >> 4
	pal[3] = (value & 0b1100_0000) >> 6
}

type bgPalette [4]uint8

func (pal *bgPalette) Read() uint8 {
	return (uint8(pal[3])<<6 |
		uint8(pal[2])<<4 |
		uint8(pal[1])<<2 |
		uint8(pal[0]))
}

func (pal *bgPalette) Write(value uint8) {
	pal[0] = value & 0b0000_0011
	pal[1] = (value & 0b0000_1100) >> 2
	pal[2] = (value & 0b0011_0000) >> 4
	pal[3] = (value & 0b1100_0000) >> 6
}

type cgbPalette [4]rgb555

type rgb555 struct {
	R uint8
	G uint8
	B uint8
}

func NewRGB555(r, g, b uint8) rgb555 {
	return rgb555{
		R: r & 0x1F,
		G: g & 0x1F,
		B: b & 0x1F,
	}
}

func (c rgb555) RGBA() (r, g, b, a uint32) {
	color := color.RGBA{
		R: c.R << 3,
		G: c.G << 3,
		B: c.B << 3,
		A: 255,
	}
	color.R |= color.R >> 2
	color.G |= color.G >> 2
	color.B |= color.B >> 2

	return color.RGBA()
}

const (
	REG_BCPS_OCPS_BIT_AUTO_INCREMENT = 7

	REG_BCPS_OCPS_ADDR_MASK = 0x3F
)

type cgbPalettes struct {
	palettes   [8]cgbPalette
	paletteRAM [64]byte

	autoIncrement bool
	addr          uint8
}

func (cgbp *cgbPalettes) Read() uint8 {
	var (
		autoIncrement uint8
		addr          uint8
	)

	if cgbp.autoIncrement {
		autoIncrement = 1 << REG_BCPS_OCPS_BIT_AUTO_INCREMENT
	}

	addr = cgbp.addr & REG_BCPS_OCPS_ADDR_MASK

	return (autoIncrement | addr)
}

func (cgbp *cgbPalettes) Write(value byte) {
	cgbp.autoIncrement = bits.Read(value, REG_BCPS_OCPS_BIT_AUTO_INCREMENT) == 1
	cgbp.addr = value & REG_BCPS_OCPS_ADDR_MASK
}

func (cgbp *cgbPalettes) ReadPalette() byte {
	return cgbp.paletteRAM[cgbp.addr]
}

func (cgbp *cgbPalettes) WritePalette(value byte) {
	cgbp.paletteRAM[cgbp.addr] = value

	index := cgbp.addr / 8
	colorIndex := cgbp.addr % 8
	currentColor := cgbp.palettes[index][colorIndex/2]

	if colorIndex%2 == 0 {
		cgbp.palettes[index][colorIndex/2] = NewRGB555(
			value,
			(currentColor.G&0b11000)|(value>>5),
			currentColor.B,
		)
	} else {
		cgbp.palettes[index][colorIndex/2] = NewRGB555(
			currentColor.R,
			(currentColor.G&0b111)|((value&0x3)<<3),
			(value >> 2),
		)
	}

	if cgbp.autoIncrement {
		cgbp.addr = (cgbp.addr + 1) % 64
	}
}
