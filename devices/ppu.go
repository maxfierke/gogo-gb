package devices

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_PPU_SCY  = 0xFF42
	REG_PPU_SCX  = 0xFF43
	REG_PPU_LY   = 0xFF44
	REG_PPU_LYC  = 0xFF45
	REG_PPU_DMA  = 0xFF46
	REG_PPU_BGP  = 0xFF47
	REG_PPU_OBP0 = 0xFF48
	REG_PPU_OBP1 = 0xFF49
	REG_PPU_WY   = 0xFF4A
	REG_PPU_WX   = 0xFF4B
)

const (
	VRAM_START           = 0x8000
	VRAM_TILESET_1_START = VRAM_START
	VRAM_TILESET_2_START = 0x8800
	VRAM_TILESET_1_END   = 0x8FFF
	VRAM_TILESET_2_END   = 0x97FF
	VRAM_TILESET_SIZE    = 384
	VRAM_TILE_ROW_MASK   = 0xFFFE
	VRAM_END             = 0x9FFF
	VRAM_SIZE            = VRAM_END - VRAM_START + 1 // 8K
)

const (
	VRAM_TILE_PIXEL_ZERO TilePixel = iota
	VRAM_TILE_PIXEL_ONE
	VRAM_TILE_PIXEL_TWO
	VRAM_TILE_PIXEL_THREE
)

type ShadeID uint8

const (
	COLOR_WHITE ShadeID = iota
	COLOR_LIGHT_GRAY
	COLOR_DARK_GRAY
	COLOR_BLACK
)

type bgPalette [4]ShadeID

func (pal *bgPalette) Read() uint8 {
	return (uint8(pal[3])<<6 |
		uint8(pal[2])<<4 |
		uint8(pal[1])<<2 |
		uint8(pal[0]))
}

func (pal *bgPalette) Write(value uint8) {
	pal[0] = ShadeID(value & 0b0000_0011)
	pal[1] = ShadeID(value & 0b0000_1100)
	pal[2] = ShadeID(value & 0b0011_0000)
	pal[3] = ShadeID(value & 0b1100_0000)
}

type objPalette [3]ShadeID

func (pal *objPalette) Read() uint8 {
	return (uint8(pal[2])<<6 |
		uint8(pal[1])<<4 |
		uint8(pal[0])<<2)
}

func (pal *objPalette) Write(value uint8) {
	pal[0] = ShadeID(value & 0b0000_1100)
	pal[1] = ShadeID(value & 0b0011_0000)
	pal[2] = ShadeID(value & 0b1100_0000)
}

type PPUMode uint8

const (
	PPU_MODE_HBLANK PPUMode = iota
	PPU_MODE_VBLANK
	PPU_MODE_OAM
	PPU_MODE_VRAM
)

type PPU struct {
	Mode PPUMode

	curScanLine       uint8 // LY
	cmpScanLine       uint8 // LYC
	scrollBackgroundX uint8 // SCX
	scrollBackgroundY uint8 // SCY
	windowX           uint8 // WX
	windowY           uint8 // WY

	bgPalette   bgPalette
	objPalette0 objPalette
	objPalette1 objPalette

	vram    [VRAM_SIZE]byte
	tileset [VRAM_TILESET_SIZE]Tile
}

func NewPPU() *PPU {
	return &PPU{}
}

func (ppu *PPU) CurrentLine() uint8 {
	return ppu.curScanLine
}

func (ppu *PPU) IsCurrentLineEqualToCompare() bool {
	return ppu.curScanLine == ppu.cmpScanLine
}

func (ppu *PPU) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_PPU_SCX {
		return mem.ReadReplace(ppu.scrollBackgroundX)
	}

	if addr == REG_PPU_SCY {
		return mem.ReadReplace(ppu.scrollBackgroundY)
	}

	if addr == REG_PPU_LY {
		return mem.ReadReplace(ppu.curScanLine)
	}

	if addr == REG_PPU_LYC {
		return mem.ReadReplace(ppu.cmpScanLine)
	}

	if addr == REG_PPU_BGP {
		return mem.ReadReplace(ppu.bgPalette.Read())
	}

	if addr == REG_PPU_OBP0 {
		return mem.ReadReplace(ppu.objPalette0.Read())
	}

	if addr == REG_PPU_OBP1 {
		return mem.ReadReplace(ppu.objPalette1.Read())
	}

	if addr == REG_PPU_WY {
		return mem.ReadReplace(ppu.windowY)
	}

	if addr == REG_PPU_WX {
		return mem.ReadReplace(ppu.windowX)
	}

	vramAddr := addr - VRAM_START

	if vramAddr < VRAM_SIZE {
		if ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.vram[vramAddr])
	}

	panic(fmt.Sprintf("Attempting to read @ 0x%04X, which is out-of-bounds for PPU", addr))
}

func (ppu *PPU) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_PPU_SCX {
		ppu.scrollBackgroundX = value
		return mem.WriteBlock()
	}

	if addr == REG_PPU_SCY {
		ppu.scrollBackgroundY = value
		return mem.WriteBlock()
	}

	if addr == REG_PPU_LY {
		// Ignore. LY is read-only
		return mem.WriteBlock()
	}

	if addr == REG_PPU_LYC {
		ppu.cmpScanLine = value
		return mem.WriteBlock()
	}

	if addr == REG_PPU_DMA {
		// TODO: Implement DMA
		return mem.WriteBlock()
	}

	if addr == REG_PPU_BGP {
		ppu.bgPalette.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_OBP0 {
		ppu.objPalette0.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_OBP1 {
		ppu.objPalette1.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_WY {
		ppu.windowY = value
		return mem.WriteBlock()
	}

	if addr == REG_PPU_WX {
		ppu.windowX = value
		return mem.WriteBlock()
	}

	vramAddr := addr - VRAM_START

	if vramAddr < VRAM_SIZE {
		if ppu.Mode == PPU_MODE_VRAM {
			return mem.WriteBlock()
		}

		ppu.vram[vramAddr] = value

		if addr <= VRAM_TILESET_2_END {
			ppu.writeTile(vramAddr)
		}

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for PPU", value, addr))
}

func (ppu *PPU) writeTile(vramAddr uint16) {
	// https://rylev.github.io/DMG-01/public/book/graphics/tile_ram.html
	rowAddr := vramAddr & VRAM_TILE_ROW_MASK

	tileRowTop := ppu.vram[rowAddr]
	tileRowBottom := ppu.vram[rowAddr+1]

	tileIdx := vramAddr / 16
	rowIdx := (vramAddr % 16) / 2
	row := ppu.tileset[tileIdx][rowIdx]

	for pixelIdx := range row {
		pixelMask := byte(1 << (7 - pixelIdx))
		lsb := tileRowTop & pixelMask
		msb := tileRowBottom & pixelMask

		if lsb == 0 && msb == 0 {
			row[pixelIdx] = VRAM_TILE_PIXEL_ZERO
		} else if lsb != 0 && msb == 0 {
			row[pixelIdx] = VRAM_TILE_PIXEL_ONE
		} else if lsb == 0 && msb != 0 {
			row[pixelIdx] = VRAM_TILE_PIXEL_TWO
		} else {
			row[pixelIdx] = VRAM_TILE_PIXEL_THREE
		}
	}
}

type TilePixel uint8
type Tile [8][8]TilePixel

func NewTile() Tile {
	return Tile{}
}
