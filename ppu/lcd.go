package ppu

import "github.com/maxfierke/gogo-gb/bits"

const (
	REG_LCD_LCDC = 0xFF40
	REG_LCD_STAT = 0xFF41
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
	lcdc.enabled = bits.Read(value, LCDC_BIT_LCD_ENABLE) == 1
	lcdc.windowTilemap = tileMapArea(bits.Read(value, LCDC_BIT_WINDOW_TILEMAP))
	lcdc.windowEnabled = bits.Read(value, LCDC_BIT_WINDOW_ENABLE) == 1
	lcdc.bgWindowTileset = tileSetArea(bits.Read(value, LCDC_BIT_BG_WINDOW_TILESET))
	lcdc.bgTilemap = tileMapArea(bits.Read(value, LCDC_BIT_BG_TILEMAP))
	lcdc.objectSize = objectSize(bits.Read(value, LCDC_BIT_OBJ_SIZE))
	lcdc.objectEnabled = bits.Read(value, LCDC_BIT_OBJ_ENABLE) == 1
	lcdc.bgWindowEnabled = bits.Read(value, LCDC_BIT_BG_WINDOW_ENABLE) == 1
}

const (
	LCD_STAT_BIT_LYC_EQ_LY      = 2
	LCD_STAT_BIT_MODE_0_INT_SEL = 3
	LCD_STAT_BIT_MODE_1_INT_SEL = 4
	LCD_STAT_BIT_MODE_2_INT_SEL = 5
	LCD_STAT_BIT_LYC_INT_SEL    = 6
)

type lcdStatus struct {
	mode0IntSel bool
	mode1IntSel bool
	mode2IntSel bool
	lycIntSel   bool
}

func (stat *lcdStatus) Read(ppu *PPU) uint8 {
	var (
		lycEqLy     uint8
		mode0IntSel uint8
		mode1IntSel uint8
		mode2IntSel uint8
		lycIntSel   uint8
	)

	if stat.mode0IntSel {
		mode0IntSel = 1 << LCD_STAT_BIT_MODE_0_INT_SEL
	}

	if stat.mode1IntSel {
		mode1IntSel = 1 << LCD_STAT_BIT_MODE_1_INT_SEL
	}

	if stat.mode2IntSel {
		mode2IntSel = 1 << LCD_STAT_BIT_MODE_2_INT_SEL
	}

	if stat.lycIntSel {
		lycIntSel = 1 << LCD_STAT_BIT_LYC_INT_SEL
	}

	if ppu.IsCurrentLineEqualToCompare() {
		lycEqLy = 1 << LCD_STAT_BIT_LYC_EQ_LY
	}

	return (1<<7 | // Always set, but supposedly unused
		lycIntSel |
		mode2IntSel |
		mode1IntSel |
		mode0IntSel |
		lycEqLy |
		uint8(ppu.Mode))
}

func (stat *lcdStatus) InterruptEnabled(ppu *PPU) bool {
	if stat.lycIntSel && ppu.IsCurrentLineEqualToCompare() {
		return true
	}

	switch ppu.Mode {
	case PPU_MODE_HBLANK:
		return stat.mode0IntSel
	case PPU_MODE_VBLANK:
		return stat.mode1IntSel
	case PPU_MODE_OAM:
		return stat.mode2IntSel
	default:
		return false
	}
}

func (stat *lcdStatus) Write(ppu *PPU, value uint8) {
	stat.mode0IntSel = bits.Read(value, LCD_STAT_BIT_MODE_0_INT_SEL) == 1
	stat.mode1IntSel = bits.Read(value, LCD_STAT_BIT_MODE_1_INT_SEL) == 1
	stat.mode2IntSel = bits.Read(value, LCD_STAT_BIT_MODE_2_INT_SEL) == 1
	stat.lycIntSel = bits.Read(value, LCD_STAT_BIT_LYC_INT_SEL) == 1
}
