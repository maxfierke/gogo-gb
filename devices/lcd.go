package devices

import (
	"fmt"
	"image"
	"image/color"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_LCD_LCDC = 0xFF40
	REG_LCD_STAT = 0xFF41

	VBLANK_PERIOD_BEGIN = 144
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

const (
	LCD_STAT_BIT_PPU_MODE = iota
	LCD_STAT_BIT_LYC_EQ_LY
	LCD_STAT_BIT_MODE_0_INT_SEL
	LCD_STAT_BIT_MODE_1_INT_SEL
	LCD_STAT_BIT_MODE_2_INT_SEL
	LCD_STAT_BIT_LYC_INT_SEL
)

type lcdStatus struct {
	mode0IntSel bool
	mode1IntSel bool
	mode2IntSel bool
	lycIntSel   bool

	shouldInterrupt bool
}

func (stat *lcdStatus) Read(lcd *LCD, ppu *PPU) uint8 {
	var (
		lycEqLy     uint8
		mode0IntSel uint8
		mode1IntSel uint8
		mode2IntSel uint8
		lycIntSel   uint8
	)

	if stat.mode0IntSel {
		lycIntSel = 1 << LCD_STAT_BIT_MODE_0_INT_SEL
	}

	if stat.mode1IntSel {
		lycIntSel = 1 << LCD_STAT_BIT_MODE_1_INT_SEL
	}

	if stat.mode2IntSel {
		lycIntSel = 1 << LCD_STAT_BIT_MODE_2_INT_SEL
	}

	if stat.lycIntSel {
		lycIntSel = 1 << LCD_STAT_BIT_LYC_INT_SEL
	}

	if ppu.IsCurrentLineEqualToCompare() {
		lycEqLy = 1 << LCD_STAT_BIT_LYC_EQ_LY
	}

	return (lycIntSel |
		mode2IntSel |
		mode1IntSel |
		mode0IntSel |
		lycEqLy |
		uint8(ppu.Mode))
}

func (stat *lcdStatus) ShouldInterrupt() bool {
	return stat.shouldInterrupt
}

func (stat *lcdStatus) statIntLine() bool {
	return stat.mode0IntSel || stat.mode1IntSel || stat.mode2IntSel || stat.lycIntSel
}

func (stat *lcdStatus) Write(value uint8) {
	prevStatIntLine := stat.statIntLine()

	stat.mode0IntSel = readBit(value, LCD_STAT_BIT_MODE_0_INT_SEL) == 1
	stat.mode1IntSel = readBit(value, LCD_STAT_BIT_MODE_1_INT_SEL) == 1
	stat.mode2IntSel = readBit(value, LCD_STAT_BIT_MODE_2_INT_SEL) == 1
	stat.lycIntSel = readBit(value, LCD_STAT_BIT_LYC_INT_SEL) == 1

	nextStatIntLine := stat.statIntLine()
	stat.shouldInterrupt = !prevStatIntLine && nextStatIntLine
}

var grayShades = color.Palette{
	color.White,
	color.Gray{Y: 192},
	color.Gray{Y: 96},
	color.Black,
}

type LCD struct {
	ctrl   lcdCtrl
	status lcdStatus

	ic  *InterruptController
	ppu *PPU
}

func NewLCD(ic *InterruptController, ppu *PPU) *LCD {
	return &LCD{
		ic:  ic,
		ppu: ppu,
	}
}

func (lcd *LCD) Draw() image.Image {
	image := image.NewPaletted(
		image.Rect(0, 0, 160, 144),
		grayShades,
	)
	image.Set(80, 77, grayShades[2])

	return image
}

func (lcd *LCD) Step(cycles uint8) {
	if lcd.ppu.CurrentLine() == VBLANK_PERIOD_BEGIN {
		lcd.ic.RequestVBlank()
	}
}

func (lcd *LCD) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_LCD_LCDC {
		return mem.ReadReplace(lcd.ctrl.Read())
	}

	if addr == REG_LCD_STAT {
		return mem.ReadReplace(lcd.status.Read(lcd, lcd.ppu))
	}

	return mem.ReadPassthrough()
}

func (lcd *LCD) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_LCD_LCDC {
		lcd.ctrl.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_LCD_STAT {
		lcd.status.Write(value)

		if lcd.status.ShouldInterrupt() {
			lcd.ic.RequestLCD()
		}

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for LCD", value, addr))
}

func readBit(value byte, bit int) uint8 {
	return ((value >> bit) & 0b1)
}
