package ppu

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

	REG_BOOTROM_KEY0              = 0xFF4C
	REG_BOOTROM_KEY0_CPU_MODE_BIT = 2

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

	// CGB Registers
	REG_PPU_VBK       uint16 = 0xFF4F
	REG_PPU_BCPS_BGPI uint16 = 0xFF68
	REG_PPU_BCPD_BGPD uint16 = 0xFF69
	REG_PPU_OCPS_OBPI uint16 = 0xFF6A
	REG_PPU_OCPD_OBPD uint16 = 0xFF6B
	REG_PPU_OPRI      uint16 = 0xFF6C
)

type objectPriorityMode uint8

const (
	objectPriorityModeCGB objectPriorityMode = iota // objectPriorityModeCGB prioritizes by OAM location
	objectPriorityModeDMG                           // objectPriorityModeDMG prioritizes by x-coordinate
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

type pixelLayer uint8

const (
	PIXEL_LAYER_BG = iota
	PIXEL_LAYER_BGP
	PIXEL_LAYER_OBJ
)

type scanLine struct {
	layer   pixelLayer
	colorID ColorID
	color   color.Color
}

type PPUMode uint8

const (
	PPU_MODE_HBLANK PPUMode = iota
	PPU_MODE_VBLANK
	PPU_MODE_OAM
	PPU_MODE_VRAM
)

type InterruptRequester interface {
	RequestLCD()
	RequestVBlank()
}

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
	curWindowLine     uint8 // Internal counter for window rendering

	// Monochrome palettes (DMG)
	bgPalette   bgPalette
	objPalettes [2]objPalette

	// Color palettes (CGB)
	cgbBGPalettes  cgbPalettes
	cgbObjPalettes cgbPalettes

	oam            oam
	objectPriority objectPriorityMode

	vram *VRAM

	scanLines [FB_HEIGHT][FB_WIDTH]scanLine

	clock uint

	ic InterruptRequester

	color                   bool
	dmgCompatibilityEnabled bool
	hdma                    *HDMA
}

func NewPPU(ic InterruptRequester) *PPU {
	return &PPU{
		ic:             ic,
		objectPriority: objectPriorityModeDMG,
		vram:           NewVRAM(),
	}
}

var grayScales = []color.Color{
	color.White,
	color.GrayModel.Convert(color.RGBA{R: 170, G: 170, B: 170}),
	color.GrayModel.Convert(color.RGBA{R: 85, G: 85, B: 85}),
	color.Black,
}

func (ppu *PPU) Draw() image.Image {
	fbImage := image.NewRGBA(
		image.Rect(0, 0, FB_WIDTH, FB_HEIGHT),
	)

	for y := range FB_HEIGHT {
		for x, scanLine := range ppu.scanLines[y] {
			if scanLine.color != nil {
				fbImage.Set(x, y, scanLine.color)
			}
		}
	}

	return fbImage
}

func (ppu *PPU) ConnectHDMA(hdma *HDMA) {
	ppu.hdma = hdma
}

func (ppu *PPU) EnableColor() {
	ppu.color = true
	ppu.objectPriority = objectPriorityModeCGB
}

func (ppu *PPU) IsColorEnabled() bool {
	return ppu.color && !ppu.dmgCompatibilityEnabled
}

func (ppu *PPU) IsCurrentLineEqualToCompare() bool {
	return ppu.curScanLine == ppu.cmpScanLine
}

func (ppu *PPU) SetDMGCompatibilityEnabled(enabled bool) {
	ppu.dmgCompatibilityEnabled = enabled
}

func (ppu *PPU) Step(mmu *mem.MMU, cycles uint8) {
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
				ppu.curWindowLine = 0
				ppu.ic.RequestVBlank()
				ppu.requestLCD()
			} else {
				if ppu.hdma != nil {
					ppu.hdma.Step(mmu)
				}

				ppu.Mode = PPU_MODE_OAM
				ppu.requestLCD()
			}

			if ppu.IsCurrentLineEqualToCompare() {
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

			if ppu.IsCurrentLineEqualToCompare() {
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

	if addr == REG_PPU_VBK {
		return mem.ReadReplace(0xFE | ppu.vram.CurrentBank)
	}

	if addr == REG_PPU_BCPS_BGPI {
		return mem.ReadReplace(ppu.cgbBGPalettes.Read())
	}

	if addr == REG_PPU_BCPD_BGPD {
		if ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.cgbBGPalettes.ReadPalette())
	}

	if addr == REG_PPU_OCPS_OBPI {
		return mem.ReadReplace(ppu.cgbObjPalettes.Read())
	}

	if addr == REG_PPU_OCPD_OBPD {
		if ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.cgbObjPalettes.ReadPalette())
	}

	if addr == REG_PPU_OPRI {
		return mem.ReadReplace(byte(ppu.objectPriority))
	}

	if addr >= OAM_START && addr <= OAM_END {
		oamAddr := uint8(addr - OAM_START)

		if ppu.Mode == PPU_MODE_OAM || ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.oam.Read(oamAddr))
	}

	if addr >= VRAM_START && addr <= VRAM_END {
		vramAddr := addr - VRAM_START

		if ppu.Mode == PPU_MODE_VRAM {
			return mem.ReadReplace(0xFF)
		}

		return mem.ReadReplace(ppu.vram.Read(vramAddr))
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

	if addr == REG_PPU_BCPS_BGPI {
		ppu.cgbBGPalettes.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_BCPD_BGPD {
		if ppu.Mode == PPU_MODE_VRAM {
			if ppu.cgbBGPalettes.autoIncrement {
				ppu.cgbBGPalettes.addr = (ppu.cgbBGPalettes.addr + 1) % 64
			}
		} else {
			ppu.cgbBGPalettes.WritePalette(value)
		}

		return mem.WriteBlock()
	}

	if addr == REG_PPU_OCPS_OBPI {
		ppu.cgbObjPalettes.Write(value)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_OCPD_OBPD {
		if ppu.Mode == PPU_MODE_VRAM {
			if ppu.cgbObjPalettes.autoIncrement {
				ppu.cgbObjPalettes.addr = (ppu.cgbObjPalettes.addr + 1) % 64
			}
		} else {
			ppu.cgbObjPalettes.WritePalette(value)
		}

		return mem.WriteBlock()
	}

	if addr == REG_PPU_OPRI {
		ppu.objectPriority = objectPriorityMode(value & 0x1)
		return mem.WriteBlock()
	}

	if addr == REG_BOOTROM_KEY0 && !ppu.dmgCompatibilityEnabled {
		// TODO: This should only be set by the bootrom,
		// so this probably doesn't belong here
		ppu.SetDMGCompatibilityEnabled(bits.Read(value, REG_BOOTROM_KEY0_CPU_MODE_BIT) == 1)
		return mem.WriteBlock()
	}

	if addr == REG_PPU_VBK {
		ppu.vram.SetCurrentBank(value & 0b1)
		return mem.WriteBlock()
	}

	if addr >= OAM_START && addr <= OAM_END {
		oamAddr := uint8(addr - OAM_START)

		if ppu.Mode == PPU_MODE_OAM || ppu.Mode == PPU_MODE_VRAM {
			return mem.WriteBlock()
		}

		ppu.oam.Write(oamAddr, value)

		return mem.WriteBlock()
	}

	if addr >= VRAM_START && addr <= VRAM_END {
		vramAddr := addr - VRAM_START

		if ppu.Mode == PPU_MODE_VRAM {
			return mem.WriteBlock()
		}

		ppu.vram.Write(vramAddr, value)

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for PPU", value, addr))
}

func (ppu *PPU) drawScanline() {
	if !ppu.lcdCtrl.enabled || ppu.curScanLine >= FB_HEIGHT {
		return
	}

	if ppu.lcdCtrl.bgWindowEnabled || ppu.IsColorEnabled() { // TODO: Extract method
		ppu.drawBgScanline()
	}

	if (ppu.lcdCtrl.bgWindowEnabled || ppu.IsColorEnabled()) && ppu.lcdCtrl.windowEnabled { // TODO: Extract method
		ppu.drawWinScanline()
	}

	if ppu.lcdCtrl.objectEnabled {
		ppu.drawObjScanline()
	}
}

func (ppu *PPU) drawBgScanline() {
	tileY := (ppu.curScanLine + ppu.scrollBackgroundY) / 8
	tilePixelY := (ppu.curScanLine + ppu.scrollBackgroundY) % 8

	for lineX := uint16(0); lineX < FB_WIDTH; lineX++ {
		scrollAdjustedLineX := (lineX + uint16(ppu.scrollBackgroundX)) % 256
		tileX := uint8(scrollAdjustedLineX / 8)

		tileIndex := ppu.vram.GetBGTileIndex(
			ppu.lcdCtrl.bgTilemap,
			uint8(tileX),
			tileY,
		)

		bgAttributes := ppu.vram.GetBGTileAttributes(
			tileX,
			tileY,
		)

		var tileVRAMBank uint8
		if ppu.IsColorEnabled() {
			tileVRAMBank = bgAttributes.vramBank
		}

		tile := ppu.vram.GetBGTile(
			tileVRAMBank,
			ppu.lcdCtrl.bgWindowTileset,
			tileIndex,
		)

		tileRow := tile[tilePixelY]
		if bgAttributes.flipY {
			tileRow = tile[7-tilePixelY]
		}

		tilePixelX := scrollAdjustedLineX % 8
		if bgAttributes.flipX {
			tilePixelX = 7 - tilePixelX
		}

		tilePixelValue := tileRow[tilePixelX]

		if ppu.IsColorEnabled() {
			color := ppu.cgbBGPalettes.palettes[bgAttributes.paletteID][tilePixelValue]
			ppu.scanLines[ppu.curScanLine][lineX].color = color
		} else {
			color := ppu.bgPalette[tilePixelValue]
			ppu.scanLines[ppu.curScanLine][lineX].color = grayScales[color]
		}

		ppu.scanLines[ppu.curScanLine][lineX].colorID = ColorID(tilePixelValue)

		if bgAttributes.priority && ppu.IsColorEnabled() {
			ppu.scanLines[ppu.curScanLine][lineX].layer = PIXEL_LAYER_BGP
		} else {
			ppu.scanLines[ppu.curScanLine][lineX].layer = PIXEL_LAYER_BG
		}
	}
}

func (ppu *PPU) drawWinScanline() {
	if ppu.curScanLine >= ppu.windowY {
		if ppu.curScanLine == ppu.windowY {
			ppu.curWindowLine = 0
		}

		tileY := ppu.curWindowLine / 8
		tilePixelY := ppu.curWindowLine % 8

		rendered := false

		for lineX := uint16(0); lineX < FB_WIDTH; lineX++ {
			if (lineX + 7) < uint16(ppu.windowX) {
				continue
			}

			rendered = true

			windowAdjustedLineX := (lineX + 7 - uint16(ppu.windowX))
			tileX := uint8(windowAdjustedLineX / 8)

			tileIndex := ppu.vram.GetBGTileIndex(
				ppu.lcdCtrl.windowTilemap,
				uint8(tileX),
				tileY,
			)
			bgAttributes := ppu.vram.GetBGTileAttributes(
				tileX,
				tileY,
			)

			var tileVRAMBank uint8
			if ppu.IsColorEnabled() {
				tileVRAMBank = bgAttributes.vramBank
			}

			tile := ppu.vram.GetBGTile(
				tileVRAMBank,
				ppu.lcdCtrl.bgWindowTileset,
				tileIndex,
			)

			tileRow := tile[tilePixelY]
			if bgAttributes.flipY {
				tileRow = tile[7-tilePixelY]
			}

			tilePixelX := windowAdjustedLineX % 8
			if bgAttributes.flipX {
				tilePixelX = 7 - tilePixelX
			}

			tilePixelValue := tileRow[tilePixelX]

			if ppu.IsColorEnabled() {
				color := ppu.cgbBGPalettes.palettes[bgAttributes.paletteID][tilePixelValue]
				ppu.scanLines[ppu.curScanLine][lineX].color = color
			} else {
				color := ppu.bgPalette[tilePixelValue]
				ppu.scanLines[ppu.curScanLine][lineX].color = grayScales[color]
			}

			ppu.scanLines[ppu.curScanLine][lineX].colorID = ColorID(tilePixelValue)

			if bgAttributes.priority && ppu.IsColorEnabled() {
				ppu.scanLines[ppu.curScanLine][lineX].layer = PIXEL_LAYER_BGP
			} else {
				ppu.scanLines[ppu.curScanLine][lineX].layer = PIXEL_LAYER_BG
			}
		}

		if rendered {
			ppu.curWindowLine++
		}
	}
}

func (ppu *PPU) drawObjScanline() {
	objHeight := uint8(8)
	if ppu.lcdCtrl.objectSize == OBJ_SIZE_8x16 {
		objHeight = 16
	}

	renderedObjects := 0

	renderedObjectsX := map[uint8]uint8{}

	for _, object := range ppu.oam.Objects() {
		if renderedObjects == OAM_MAX_OBJECTS_PER_SCANLINE {
			break
		}

		if object.posY <= ppu.curScanLine && (object.posY+objHeight) > ppu.curScanLine {
			objPixelY := ppu.curScanLine - object.posY

			tile := ppu.vram.GetObjTile(
				object,
				ppu.lcdCtrl.objectSize,
				objPixelY,
				ppu.IsColorEnabled(),
			)

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

				renderedObjX, hasRenderedObj := renderedObjectsX[pixelX]

				if pixelX < FB_WIDTH &&
					// Skip transparent pixels
					tilePixelValue != VRAM_TILE_PIXEL_ZERO &&
					((ppu.objectPriority == objectPriorityModeCGB && !ppu.dmgCompatibilityEnabled && !hasRenderedObj) || // CGB mode: Earlier Object hasn't rendered at pixel
						// DMG mode: Object has higher priority x coordinate than currently rendered object
						(ppu.objectPriority == objectPriorityModeDMG &&
							(!hasRenderedObj || (hasRenderedObj && renderedObjX > object.posX)))) && // TODO: Extract method
					(ppu.scanLines[ppu.curScanLine][pixelX].colorID == COLOR_ID_WHITE || // BG is color 0
						// CGB: BG master priority isn't set
						(ppu.objectPriority == objectPriorityModeCGB && !ppu.dmgCompatibilityEnabled && !ppu.lcdCtrl.bgWindowEnabled) ||
						// BG doesn't have priority (CGB) AND OBJ has priority over BG
						(ppu.scanLines[ppu.curScanLine][pixelX].layer != PIXEL_LAYER_BGP && !object.attributes.bgPriority)) { // TODO: Extract method

					if ppu.IsColorEnabled() {
						color := ppu.cgbObjPalettes.palettes[object.attributes.cgbPaletteID][tilePixelValue]
						ppu.scanLines[ppu.curScanLine][pixelX].color = color
					} else {
						color := ppu.objPalettes[object.attributes.dmgPaletteID][tilePixelValue]
						ppu.scanLines[ppu.curScanLine][pixelX].color = grayScales[color]
					}

					ppu.scanLines[ppu.curScanLine][pixelX].colorID = ColorID(tilePixelValue)
					ppu.scanLines[ppu.curScanLine][pixelX].layer = PIXEL_LAYER_OBJ
					renderedObject = true
					renderedObjectsX[pixelX] = object.posX
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
