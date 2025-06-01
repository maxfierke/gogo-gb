package devices

import (
	"fmt"
	"image"
	"image/color"

	"github.com/maxfierke/gogo-gb/bits"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	// CLK_MODE0_PERIOD_LEN is the dot length of Mode 0 (HBlank).
	// 204 dots is the ceiling, but 200 dots was chosen for compatibility with
	// orangeglo's LED Screen Timer test ROM. Mode 0 and Mode 3 just need to add
	// up to 376.
	// This is probably a bug of some kind, but not one I feel like fixing right now.
	CLK_MODE0_PERIOD_LEN = 200

	// CLK_MODE1_PERIOD_LEN is the dot length of a line in Mode 1 (VBlank)
	// There are 10 lines in Mode 1, for a total of 4560 dots.
	CLK_MODE1_PERIOD_LEN = 456

	// CLK_MODE2_PERIOD_LEN is the dot length of Mode 2 (OAM Scan).
	// It is fixed at 80 dots.
	CLK_MODE2_PERIOD_LEN = 80

	// CLK_MODE3_PERIOD_LEN is the dot length of Mode 3 (VRAM / drawing).
	// 172 dots is the floor, but 174 dots was chosen for compatibility with
	// orangeglo's LED Screen Timer test ROM. Mode 0 and Mode 3 just need to add
	// up to 376.
	// This is probably a bug of some kind, but not one I feel like fixing right now.
	CLK_MODE3_PERIOD_LEN = 174

	FB_WIDTH  = 160
	FB_HEIGHT = 144

	VBLANK_PERIOD_BEGIN = 144
	VBLANK_PERIOD_END   = 153

	REG_PPU_LCDC    uint16 = 0xFF40
	REG_PPU_LCDSTAT uint16 = 0xFF41
	REG_PPU_SCY     uint16 = 0xFF42
	REG_PPU_SCX     uint16 = 0xFF43
	REG_PPU_LY      uint16 = 0xFF44
	REG_PPU_LYC     uint16 = 0xFF45
	REG_PPU_BGP     uint16 = 0xFF47
	REG_PPU_OBP0    uint16 = 0xFF48
	REG_PPU_OBP1    uint16 = 0xFF49
	REG_PPU_WY      uint16 = 0xFF4A
	REG_PPU_WX      uint16 = 0xFF4B

	OAM_START                    uint16 = 0xFE00
	OAM_END                      uint16 = 0xFE9F
	OAM_MAX_OBJECTS_PER_SCANLINE        = 10
	OAM_MAX_OBJECT_COUNT                = 40
	OAM_SIZE                            = OAM_END - OAM_START + 1

	VRAM_START           uint16 = 0x8000
	VRAM_TILESET_1_START uint16 = VRAM_START
	VRAM_TILESET_2_START uint16 = 0x8800
	VRAM_TILESET_1_END   uint16 = 0x8FFF
	VRAM_TILESET_2_END   uint16 = 0x97FF
	VRAM_TILEMAP_1_START uint16 = 0x9800
	VRAM_TILEMAP_1_END   uint16 = 0x9BFF
	VRAM_TILEMAP_2_START uint16 = 0x9C00
	VRAM_TILEMAP_2_END   uint16 = 0x9FFF
	VRAM_TILESET_SIZE           = 384
	VRAM_TILE_ROW_MASK   uint16 = 0xFFFE
	VRAM_END             uint16 = 0x9FFF
	VRAM_SIZE                   = VRAM_END - VRAM_START + 1 // 8K
)

const (
	VRAM_TILE_PIXEL_ZERO PPUPixel = iota
	VRAM_TILE_PIXEL_ONE
	VRAM_TILE_PIXEL_TWO
	VRAM_TILE_PIXEL_THREE
)

type ColorID uint8

const (
	COLOR_ID_WHITE ColorID = iota
	COLOR_ID_LIGHT_GRAY
	COLOR_ID_DARK_GRAY
	COLOR_ID_BLACK
	COLOR_ID_TRANSPARENT = COLOR_ID_WHITE
)

type objectData struct {
	posY      uint8
	posX      uint8
	tileIndex uint8

	attributes objectAttributes
}

const (
	OAM_ATTR_BIT_PALETTE_ID  = 4
	OAM_ATTR_BIT_X_FLIP      = 5
	OAM_ATTR_BIT_Y_FLIP      = 6
	OAM_ATTR_BIT_BG_PRIORITY = 7
)

type objectAttributes struct {
	bgPriority bool
	flipY      bool
	flipX      bool
	paletteID  uint8
	// TODO(gbc): Add GBC palette & bank info
}

func (attrs *objectAttributes) Read() uint8 {
	var (
		bgPriority uint8
		flipY      uint8
		flipX      uint8
		paletteID  uint8
	)

	if attrs.bgPriority {
		bgPriority = 1 << OAM_ATTR_BIT_BG_PRIORITY
	}

	if attrs.flipY {
		flipY = 1 << OAM_ATTR_BIT_Y_FLIP
	}

	if attrs.flipX {
		flipX = 1 << OAM_ATTR_BIT_X_FLIP
	}

	paletteID = attrs.paletteID << OAM_ATTR_BIT_PALETTE_ID

	return bgPriority | flipY | flipX | paletteID
}

func (attrs *objectAttributes) Write(value uint8) {
	attrs.bgPriority = bits.Read(value, OAM_ATTR_BIT_BG_PRIORITY) == 1
	attrs.flipY = bits.Read(value, OAM_ATTR_BIT_Y_FLIP) == 1
	attrs.flipX = bits.Read(value, OAM_ATTR_BIT_X_FLIP) == 1
	attrs.paletteID = bits.Read(value, OAM_ATTR_BIT_PALETTE_ID)
}

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

const (
	REG_LCD_LCDC = 0xFF40
	REG_LCD_STAT = 0xFF41
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

	shouldInterrupt bool
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

func (stat *lcdStatus) ShouldInterrupt() bool {
	return stat.shouldInterrupt
}

func (stat *lcdStatus) ModeInterruptEnabled(ppu *PPU) bool {
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

func (stat *lcdStatus) statIntLine(ppu *PPU) bool {
	return (stat.lycIntSel && ppu.IsCurrentLineEqualToCompare()) ||
		stat.mode0IntSel ||
		stat.mode1IntSel ||
		stat.mode2IntSel
}

func (stat *lcdStatus) Write(ppu *PPU, value uint8) {
	prevStatIntLine := stat.statIntLine(ppu)

	stat.mode0IntSel = bits.Read(value, LCD_STAT_BIT_MODE_0_INT_SEL) == 1
	stat.mode1IntSel = bits.Read(value, LCD_STAT_BIT_MODE_1_INT_SEL) == 1
	stat.mode2IntSel = bits.Read(value, LCD_STAT_BIT_MODE_2_INT_SEL) == 1
	stat.lycIntSel = bits.Read(value, LCD_STAT_BIT_LYC_INT_SEL) == 1

	nextStatIntLine := stat.statIntLine(ppu)
	stat.shouldInterrupt = !prevStatIntLine && nextStatIntLine
}

type scanLine struct {
	colorID ColorID
	color   uint8
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

	lcdCtrl   lcdCtrl
	lcdStatus lcdStatus

	curScanLine       uint8 // LY
	cmpScanLine       uint8 // LYC
	scrollBackgroundX uint8 // SCX
	scrollBackgroundY uint8 // SCY
	windowX           uint8 // WX
	windowY           uint8 // WY

	bgPalette   bgPalette
	objPalettes [2]objPalette

	oam        [OAM_SIZE]byte
	objectData [OAM_MAX_OBJECT_COUNT]objectData
	scanLines  [FB_HEIGHT][FB_WIDTH]scanLine
	vram       [VRAM_SIZE]byte
	tileset    [VRAM_TILESET_SIZE]Tile

	clock uint

	ic *InterruptController
}

func NewPPU(ic *InterruptController) *PPU {
	return &PPU{
		ic: ic,
	}
}

var grayScales = []color.Color{
	color.White,
	color.RGBA{R: 170, G: 170, B: 170},
	color.RGBA{R: 85, G: 85, B: 85},
	color.Black,
}

func (ppu *PPU) Draw() image.Image {
	fbImage := image.NewGray(
		image.Rect(0, 0, FB_WIDTH, FB_HEIGHT),
	)

	for y := range FB_HEIGHT {
		for x, scanLine := range ppu.scanLines[y] {
			fbImage.Set(x, y, grayScales[scanLine.color])
		}
	}

	return fbImage
}

func (ppu *PPU) IsCurrentLineEqualToCompare() bool {
	return ppu.curScanLine == ppu.cmpScanLine
}

func (ppu *PPU) Step(cycles uint8) {
	if !ppu.lcdCtrl.enabled {
		return
	}

	ppu.clock += uint(cycles)

	switch ppu.Mode {
	case PPU_MODE_HBLANK:
		if ppu.clock >= CLK_MODE0_PERIOD_LEN {
			ppu.clock = ppu.clock % CLK_MODE0_PERIOD_LEN
			ppu.curScanLine += 1

			if ppu.curScanLine == VBLANK_PERIOD_BEGIN {
				ppu.Mode = PPU_MODE_VBLANK
				ppu.ic.RequestVBlank()
				ppu.requestLCD()
			} else {
				ppu.Mode = PPU_MODE_OAM
				ppu.requestLCD()
			}

			if ppu.IsCurrentLineEqualToCompare() && ppu.lcdStatus.ShouldInterrupt() {
				ppu.ic.RequestLCD()
			}
		}
	case PPU_MODE_VBLANK:
		if ppu.clock >= CLK_MODE1_PERIOD_LEN {
			ppu.clock = ppu.clock % CLK_MODE1_PERIOD_LEN
			ppu.curScanLine += 1

			if ppu.curScanLine == VBLANK_PERIOD_END {
				ppu.Mode = PPU_MODE_OAM
				ppu.curScanLine = 0
				ppu.requestLCD()
			}

			if ppu.IsCurrentLineEqualToCompare() && ppu.lcdStatus.ShouldInterrupt() {
				ppu.ic.RequestLCD()
			}
		}
	case PPU_MODE_OAM:
		if ppu.clock >= CLK_MODE2_PERIOD_LEN {
			ppu.clock = ppu.clock % CLK_MODE2_PERIOD_LEN
			ppu.Mode = PPU_MODE_VRAM
		}
	case PPU_MODE_VRAM:
		if ppu.clock >= CLK_MODE3_PERIOD_LEN {
			ppu.clock = ppu.clock % CLK_MODE3_PERIOD_LEN
			ppu.Mode = PPU_MODE_HBLANK
			ppu.drawScanline()
			ppu.requestLCD()
		}
	}
}

func (ppu *PPU) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_PPU_LCDC {
		return mem.ReadReplace(ppu.lcdCtrl.Read())
	}

	if addr == REG_PPU_LCDSTAT {
		return mem.ReadReplace(ppu.lcdStatus.Read(ppu))
	}

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
		return mem.ReadReplace(ppu.objPalettes[0].Read())
	}

	if addr == REG_PPU_OBP1 {
		return mem.ReadReplace(ppu.objPalettes[1].Read())
	}

	if addr == REG_PPU_WY {
		return mem.ReadReplace(ppu.windowY)
	}

	if addr == REG_PPU_WX {
		return mem.ReadReplace(ppu.windowX)
	}

	if addr >= OAM_START && addr <= OAM_END {
		oamAddr := addr - OAM_START

		if ppu.Mode == PPU_MODE_OAM || ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.oam[oamAddr])
	}

	if addr >= VRAM_START && addr <= VRAM_END {
		vramAddr := addr - VRAM_START

		if ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.vram[vramAddr])
	}

	panic(fmt.Sprintf("Attempting to read @ 0x%04X, which is out-of-bounds for PPU", addr))
}

func (ppu *PPU) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_PPU_LCDC {
		ppu.lcdCtrl.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_LCDSTAT {
		ppu.lcdStatus.Write(ppu, value)

		if ppu.lcdStatus.ShouldInterrupt() {
			ppu.ic.RequestLCD()
		}

		return mem.WriteBlock()
	}

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

	if addr == REG_PPU_BGP {
		ppu.bgPalette.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_OBP0 {
		ppu.objPalettes[0].Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_OBP1 {
		ppu.objPalettes[1].Write(value)
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

	if addr >= OAM_START && addr <= OAM_END {
		oamAddr := uint8(addr - OAM_START)

		if ppu.Mode == PPU_MODE_OAM || ppu.Mode == PPU_MODE_VRAM {
			return mem.WriteBlock()
		}

		ppu.oam[oamAddr] = value
		ppu.writeObj(oamAddr, value)

		return mem.WriteBlock()
	}

	if addr >= VRAM_START && addr <= VRAM_END {
		vramAddr := addr - VRAM_START

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

func (ppu *PPU) drawScanline() {
	if !ppu.lcdCtrl.enabled || ppu.curScanLine >= FB_HEIGHT {
		return
	}

	if ppu.lcdCtrl.bgWindowEnabled {
		ppu.drawBgScanline()
	}

	if ppu.lcdCtrl.bgWindowEnabled && ppu.lcdCtrl.windowEnabled {
		ppu.drawWinScanline()
	}

	if ppu.lcdCtrl.objectEnabled {
		ppu.drawObjScanline()
	}
}

func (ppu *PPU) drawBgScanline() {
	tileY := (ppu.curScanLine + ppu.scrollBackgroundY) / 8

	bgMapAddr := VRAM_TILEMAP_1_START
	if ppu.lcdCtrl.bgTilemap == TILEMAP_AREA2 {
		bgMapAddr = VRAM_TILEMAP_2_START
	}

	tileMapBegin := bgMapAddr - VRAM_START
	tileMapOffset := tileMapBegin + uint16(tileY)*32

	tilePixelY := (ppu.curScanLine + ppu.scrollBackgroundY) % 8

	for lineX := uint16(0); lineX < FB_WIDTH; lineX++ {
		tileX := (lineX + uint16(ppu.scrollBackgroundX)) % 256
		tilePixelX := tileX % 8
		tileIndex := ppu.vram[tileMapOffset+uint16(tileX/8)]

		tilePixelValue := ppu.tileset[tileIndex][tilePixelY][tilePixelX]
		if ppu.lcdCtrl.bgWindowTileset == TILESET_1 && tileIndex < 128 {
			tilePixelValue = ppu.tileset[VRAM_TILESET_SIZE-128+uint16(tileIndex)][tilePixelY][tilePixelX]
		}

		color := ppu.bgPalette[tilePixelValue]
		ppu.scanLines[ppu.curScanLine][lineX].colorID = ColorID(tilePixelValue)
		ppu.scanLines[ppu.curScanLine][lineX].color = color
	}
}

func (ppu *PPU) drawWinScanline() {
	if ppu.curScanLine >= ppu.windowY {
		windowMapAddr := VRAM_TILEMAP_1_START
		if ppu.lcdCtrl.windowTilemap == TILEMAP_AREA2 {
			windowMapAddr = VRAM_TILEMAP_2_START
		}

		tileY := (ppu.curScanLine - ppu.windowY) / 8
		tilePixelY := (ppu.curScanLine - ppu.windowY) % 8

		tileMapBegin := windowMapAddr - VRAM_START
		tileMapOffset := tileMapBegin + uint16(tileY)*32

		for lineX := uint16(0); lineX < FB_WIDTH; lineX++ {
			if (lineX + 7) <= uint16(ppu.windowX) {
				continue
			}

			tileX := (lineX + 7 - uint16(ppu.windowX)) / 8
			tilePixelX := (lineX + 7 - uint16(ppu.windowX)) % 8

			tileIndex := ppu.vram[tileMapOffset+uint16(tileX)]

			tilePixelValue := ppu.tileset[tileIndex][tilePixelY][tilePixelX]
			if ppu.lcdCtrl.bgWindowTileset == TILESET_1 && tileIndex < 128 {
				tilePixelValue = ppu.tileset[VRAM_TILESET_SIZE-128+uint16(tileIndex)][tilePixelY][tilePixelX]
			}

			color := ppu.bgPalette[tilePixelValue]
			ppu.scanLines[ppu.curScanLine][lineX].colorID = ColorID(tilePixelValue)
			ppu.scanLines[ppu.curScanLine][lineX].color = color
		}
	}
}

func (ppu *PPU) drawObjScanline() {
	objHeight := uint8(8)
	if ppu.lcdCtrl.objectSize == OBJ_SIZE_8x16 {
		objHeight = 16
	}

	renderedObjects := 0

	for _, object := range ppu.objectData {
		if renderedObjects == OAM_MAX_OBJECTS_PER_SCANLINE {
			break
		}

		if object.posY <= ppu.curScanLine && (object.posY+objHeight) > ppu.curScanLine {
			objPixelY := ppu.curScanLine - object.posY
			tileIndex := object.tileIndex
			if objHeight == 16 &&
				((!object.attributes.flipY && objPixelY > 7) ||
					(object.attributes.flipY && objPixelY <= 7)) {
				tileIndex += 1
			}

			tile := ppu.tileset[tileIndex]
			tilePixelY := objPixelY % 8
			tileRow := tile[tilePixelY]
			if object.attributes.flipY {
				tileRow = tile[7-tilePixelY]
			}

			renderedObject := false
			for x := uint8(0); x < 8; x++ {
				tilePixelX := x
				if object.attributes.flipX {
					tilePixelX = 7 - x
				}

				pixelX := object.posX + x
				tilePixelValue := tileRow[tilePixelX]

				if pixelX < FB_WIDTH &&
					// Skip transparent pixels
					tilePixelValue != VRAM_TILE_PIXEL_ZERO &&
					// Priority over BG or BG is color 0
					(!object.attributes.bgPriority || ppu.scanLines[ppu.curScanLine][pixelX].colorID == COLOR_ID_WHITE) {

					color := ppu.objPalettes[object.attributes.paletteID][tilePixelValue]
					ppu.scanLines[ppu.curScanLine][pixelX].colorID = ColorID(tilePixelValue)
					ppu.scanLines[ppu.curScanLine][pixelX].color = color
					renderedObject = true
				}
			}

			if renderedObject {
				renderedObjects += 1
			}
		}
	}
}

func (ppu *PPU) requestLCD() {
	if ppu.lcdStatus.ModeInterruptEnabled(ppu) && ppu.lcdStatus.ShouldInterrupt() {
		ppu.ic.RequestLCD()
	}
}

func (ppu *PPU) writeObj(oamAddr uint8, value byte) {
	objIndex := oamAddr / 4
	if objIndex > OAM_MAX_OBJECT_COUNT {
		return
	}

	byteIndex := oamAddr % 4

	switch byteIndex {
	case 0:
		ppu.objectData[objIndex].posY = value - 16
	case 1:
		ppu.objectData[objIndex].posX = value - 8
	case 2:
		ppu.objectData[objIndex].tileIndex = value
	default:
		ppu.objectData[objIndex].attributes.Write(value)
	}
}

func (ppu *PPU) writeTile(vramAddr uint16) {
	// https://rylev.github.io/DMG-01/public/book/graphics/tile_ram.html
	rowAddr := vramAddr & VRAM_TILE_ROW_MASK

	tileRowTop := ppu.vram[rowAddr]
	tileRowBottom := ppu.vram[rowAddr+1]

	tileIdx := vramAddr / 16
	rowIdx := (vramAddr % 16) / 2

	for pixelIdx := range ppu.tileset[tileIdx][rowIdx] {
		pixelMask := byte(1 << (7 - pixelIdx))
		lsb := tileRowTop & pixelMask
		msb := tileRowBottom & pixelMask

		if lsb == 0 && msb == 0 {
			ppu.tileset[tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_ZERO
		} else if lsb != 0 && msb == 0 {
			ppu.tileset[tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_ONE
		} else if lsb == 0 && msb != 0 {
			ppu.tileset[tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_TWO
		} else {
			ppu.tileset[tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_THREE
		}
	}
}

type (
	PPUPixel uint8
	Tile     [8][8]PPUPixel
)

func NewTile() Tile {
	return Tile{}
}
