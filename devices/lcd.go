package devices

import (
	"image"
	"image/color"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_LCD_LCDC = 0xFF40
	REG_LCD_LY   = 0xFF44
)

type objectSize uint8

const (
	OBJ_SIZE_8x8  objectSize = 0
	OBJ_SIZE_8x16 objectSize = 1
)

type tileMapArea uint8

const (
	TILEMAP_AREA1 tileMapArea = 0 // 0x9800–0x9BFF
	TILEMAP_AREA2 tileMapArea = 1 // 0x9C00–0x9FFF
)

type tileSetArea uint8

const (
	TILESET_1 tileSetArea = 0 // 0x8800–0x97FF
	TILESET_2 tileSetArea = 1 // 0x8000–0x8FFF
)

const (
	LCDC_BIT_BG_WINDOW_ENABLE = iota
	LCDC_BIT_OBJ_ENABLE
	LCDC_BIT_OBJ_SIZE
	LCDC_BIT_BG_TILEMAP
	LCDC_BIT_BG_WINDOW_TILESET
	LCDC_BIT_WINDOW_ENABLE
	LCDC_BIT_WINDOW_TILEMAP
	LCDC_BIT_LCD_ENABLE
)

type lcdCtrl struct {
	enabled         bool
	bgWindowEnabled bool
	windowEnabled   bool
	objectEnabled   bool
	bgTilemap       tileMapArea
	bgWindowTileset tileSetArea
	objectSize      objectSize
	windowTilemap   tileMapArea
}

func (lcdc *lcdCtrl) Read() uint8 {
	var (
		enabled         uint8
		windowTilemap   uint8
		windowEnabled   uint8
		bgWindowTileset uint8
		bgTilemap       uint8
		objectSize      uint8
		objectEnabled   uint8
		bgWindowEnabled uint8
	)

	if lcdc.enabled {
		enabled = 1 << LCDC_BIT_LCD_ENABLE
	}

	windowTilemap = uint8(lcdc.windowTilemap) << LCDC_BIT_WINDOW_TILEMAP

	if lcdc.windowEnabled {
		windowEnabled = 1 << LCDC_BIT_WINDOW_ENABLE
	}

	bgWindowTileset = uint8(lcdc.bgWindowTileset) << LCDC_BIT_BG_WINDOW_TILESET
	bgTilemap = uint8(lcdc.bgTilemap) << LCDC_BIT_BG_TILEMAP
	objectSize = uint8(lcdc.objectSize) << LCDC_BIT_OBJ_SIZE

	if lcdc.objectEnabled {
		objectEnabled = 1 << LCDC_BIT_OBJ_ENABLE
	}

	if lcdc.bgWindowEnabled {
		bgWindowEnabled = 1 << LCDC_BIT_BG_WINDOW_ENABLE
	}

	return (enabled |
		windowTilemap |
		windowEnabled |
		bgWindowTileset |
		bgTilemap |
		objectSize |
		objectEnabled |
		bgWindowEnabled)
}

func (lcdc *lcdCtrl) Write(value uint8) {
	lcdc.enabled = readBit(value, LCDC_BIT_WINDOW_ENABLE) == 1
	lcdc.windowTilemap = tileMapArea(readBit(value, LCDC_BIT_WINDOW_TILEMAP))
	lcdc.windowEnabled = readBit(value, LCDC_BIT_WINDOW_ENABLE) == 1
	lcdc.bgWindowTileset = tileSetArea(readBit(value, LCDC_BIT_BG_WINDOW_TILESET))
	lcdc.bgTilemap = tileMapArea(readBit(value, LCDC_BIT_BG_TILEMAP))
	lcdc.objectSize = objectSize(readBit(value, LCDC_BIT_OBJ_SIZE))
	lcdc.objectEnabled = readBit(value, LCDC_BIT_OBJ_ENABLE) == 1
	lcdc.bgWindowEnabled = readBit(value, LCDC_BIT_BG_WINDOW_ENABLE) == 1
}

type LCD struct {
	Ctrl lcdCtrl
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
	if addr == REG_LCD_LCDC {
		return mem.ReadReplace(lcd.Ctrl.Read())
	}

	return mem.ReadPassthrough()
}

func (lcd *LCD) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_LCD_LCDC {
		lcd.Ctrl.Write(value)
		return mem.WriteBlock()
	}

	return mem.WriteBlock()
}

func readBit(value byte, bit int) uint8 {
	return ((value >> bit) & 0b1)
}
