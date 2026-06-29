package rendering

import (
	"image"
	"image/color"

	"github.com/maxfierke/gogo-gb/ppu"
)

const (
	FB_WIDTH  = 160
	FB_HEIGHT = 144

	// CLK_MODE3_PERIOD_LEN is the dot length of Mode 3 (VRAM / drawing).
	// 172 dots is the floor, but 174 dots was chosen for compatibility with
	// orangeglo's LED Screen Timer test ROM. Mode 0 and Mode 3 just need to add
	// up to 376.
	// This is probably a bug of some kind, but not one I feel like fixing right now.
	SCANLINE_CLK_MODE3_PERIOD_LEN = 174
)

type RenderedPixel struct {
	Layer   PixelLayer
	ColorID ppu.ColorID
	Color   color.Color
}

type PixelLayer uint8

const (
	PIXEL_LAYER_BG  PixelLayer = iota // Background/window layer
	PIXEL_LAYER_BGP                   // Background/window layer w/ priority over objects
	PIXEL_LAYER_OBJ                   // Object layer
)

type ScanlineRenderer struct {
	ppu  *ppu.PPU
	oam  *ppu.OAM
	vram *ppu.VRAM

	framebuf [FB_HEIGHT][FB_WIDTH]RenderedPixel
}

var _ ppu.Renderer = (*ScanlineRenderer)(nil)

func Scanline(ppu *ppu.PPU, oam *ppu.OAM, vram *ppu.VRAM) ppu.Renderer {
	return &ScanlineRenderer{
		ppu:  ppu,
		oam:  oam,
		vram: vram,
	}
}

func (r *ScanlineRenderer) DrawImage() image.Image {
	fbImage := image.NewRGBA(
		image.Rect(0, 0, FB_WIDTH, FB_HEIGHT),
	)

	for y := range FB_HEIGHT {
		for x, pixel := range r.framebuf[y] {
			if pixel.Color != nil {
				fbImage.Set(x, y, pixel.Color)
			}
		}
	}

	return fbImage
}

func (r *ScanlineRenderer) Step(cycles uint8) uint8 {
	if !r.ppu.IsLCDEnabled() || r.ppu.CurrentScanline() >= FB_HEIGHT {
		return 0
	}

	if cycles < SCANLINE_CLK_MODE3_PERIOD_LEN {
		return 0
	}

	if r.ppu.IsBackgroundEnabled() {
		r.drawBgScanline()
	}

	if r.ppu.IsWindowEnabled() {
		r.drawWinScanline()
	}

	if r.ppu.IsObjectEnabled() {
		r.drawObjScanline()
	}

	return 160
}

func (r *ScanlineRenderer) drawBgScanline() {
	currentScanLine := r.ppu.CurrentScanline()
	scrollBackgroundX := r.ppu.ScrollBackgroundX()
	scrollBackgroundY := r.ppu.ScrollBackgroundY()

	tileY := (currentScanLine + scrollBackgroundY) / 8
	tilePixelY := (currentScanLine + scrollBackgroundY) % 8
	tileMap := r.ppu.GetBGTilemap()

	for lineX := range uint16(FB_WIDTH) {
		scrollAdjustedLineX := (lineX + uint16(scrollBackgroundX)) % 256
		tileX := uint8(scrollAdjustedLineX / 8)

		tileIndex := r.vram.GetBGTileIndex(
			tileMap,
			uint8(tileX),
			tileY,
		)

		bgAttributes := r.vram.GetBGTileAttributes(
			tileMap,
			tileX,
			tileY,
		)

		var tileVRAMBank uint8
		if r.ppu.IsColorEnabled() {
			tileVRAMBank = bgAttributes.VRAMBank
		}

		tile := r.vram.GetBGTile(
			tileVRAMBank,
			r.ppu.GetBGWindowTileset(),
			tileIndex,
		)

		tileRow := tile[tilePixelY]
		if bgAttributes.FlipY {
			tileRow = tile[7-tilePixelY]
		}

		tilePixelX := scrollAdjustedLineX % 8
		if bgAttributes.FlipX {
			tilePixelX = 7 - tilePixelX
		}

		tilePixelValue := tileRow[tilePixelX]
		pixelColorID := ppu.ColorID(tilePixelValue)
		color := r.ppu.GetBGPaletteColor(pixelColorID, bgAttributes.PaletteID)
		pixelLayer := PIXEL_LAYER_BG
		if bgAttributes.Priority && r.ppu.IsColorEnabled() {
			pixelLayer = PIXEL_LAYER_BGP
		}

		r.writePixel(uint8(lineX), currentScanLine, pixelColorID, color, pixelLayer)
	}
}

func (r *ScanlineRenderer) drawWinScanline() {
	currentScanLine := r.ppu.CurrentScanline()
	windowX := r.ppu.WindowX()
	windowY := r.ppu.WindowY()

	if currentScanLine >= windowY {
		// TODO: do this in PPU instead
		if currentScanLine == windowY {
			r.ppu.ResetWindow()
		}

		currentWindowLine := r.ppu.CurrentWindowLine()
		tileMap := r.ppu.GetWindowTilemap()
		tileY := currentWindowLine / 8
		tilePixelY := currentWindowLine % 8

		rendered := false

		for lineX := range uint16(FB_WIDTH) {
			if (lineX + 7) < uint16(windowX) {
				continue
			}

			rendered = true

			windowAdjustedLineX := (lineX + 7 - uint16(windowX))
			tileX := uint8(windowAdjustedLineX / 8)

			tileIndex := r.vram.GetBGTileIndex(
				tileMap,
				uint8(tileX),
				tileY,
			)
			bgAttributes := r.vram.GetBGTileAttributes(
				tileMap,
				tileX,
				tileY,
			)

			var tileVRAMBank uint8
			if r.ppu.IsColorEnabled() {
				tileVRAMBank = bgAttributes.VRAMBank
			}

			tile := r.vram.GetBGTile(
				tileVRAMBank,
				r.ppu.GetBGWindowTileset(),
				tileIndex,
			)

			tileRow := tile[tilePixelY]
			if bgAttributes.FlipY {
				tileRow = tile[7-tilePixelY]
			}

			tilePixelX := windowAdjustedLineX % 8
			if bgAttributes.FlipX {
				tilePixelX = 7 - tilePixelX
			}

			tilePixelValue := tileRow[tilePixelX]
			pixelColorID := ppu.ColorID(tilePixelValue)
			color := r.ppu.GetBGPaletteColor(pixelColorID, bgAttributes.PaletteID)
			pixelLayer := PIXEL_LAYER_BG
			if bgAttributes.Priority && r.ppu.IsColorEnabled() {
				pixelLayer = PIXEL_LAYER_BGP
			}

			r.writePixel(uint8(lineX), currentScanLine, pixelColorID, color, pixelLayer)
		}

		// TODO: Do this in PPU
		if rendered {
			r.ppu.IncrementWindowLine()
		}
	}
}

func (r *ScanlineRenderer) drawObjScanline() {
	objHeight := uint8(8)
	if r.ppu.ObjectSize() == ppu.OBJ_SIZE_8x16 {
		objHeight = 16
	}

	currentScanLine := r.ppu.CurrentScanline()

	renderedObjects := 0
	renderedObjectsX := map[uint8]uint8{}

	for _, object := range r.oam.Objects() {
		if object == nil || renderedObjects == ppu.OAM_MAX_OBJECTS_PER_SCANLINE {
			break
		}

		if object.PosY <= int16(currentScanLine) && (object.PosY+int16(objHeight)) > int16(currentScanLine) {
			objPixelY := currentScanLine - uint8(object.PosY)

			tile := r.vram.GetObjTile(
				*object,
				r.ppu.ObjectSize(),
				objPixelY,
				r.ppu.IsColorEnabled(),
			)

			tilePixelY := objPixelY % 8
			tileRow := tile[tilePixelY]
			if object.Attributes.FlipY {
				tileRow = tile[7-tilePixelY]
			}

			renderedObject := false
			for x := range uint8(8) {
				tilePixelX := x
				if object.Attributes.FlipX {
					tilePixelX = 7 - x
				}

				pixelX := uint8(object.PosX) + x

				if pixelX >= FB_WIDTH {
					// Skip pixels outside of rendering area
					continue
				}

				tilePixel := tileRow[tilePixelX]

				if r.isObjOverBackground(object, pixelX, tilePixel, renderedObjectsX) {
					pixelColorID := ppu.ColorID(tilePixel)
					color := r.ppu.GetObjPaletteColor(pixelColorID, object.Attributes)
					pixelLayer := PIXEL_LAYER_OBJ

					r.writePixel(pixelX, currentScanLine, pixelColorID, color, pixelLayer)

					renderedObject = true
					renderedObjectsX[pixelX] = uint8(object.PosX)
				}
			}

			if renderedObject {
				renderedObjects += 1
			}
		}
	}
}

func (r *ScanlineRenderer) isObjOverBackground(object *ppu.ObjectData, pixelX uint8, tilePixel ppu.PPUPixel, renderedObjectsX map[uint8]uint8) bool {
	if tilePixel == ppu.VRAM_TILE_PIXEL_ZERO { // Skip transparent pixels
		return false
	}

	objectPriorityMode := r.ppu.ObjectPriority()

	renderedObjX, hasRenderedObj := renderedObjectsX[pixelX]

	currentPixel := r.readPixel(pixelX, r.ppu.CurrentScanline())

	switch objectPriorityMode {
	case ppu.ObjectPriorityModeCGB:
		if !r.ppu.IsColorEnabled() { // TODO: Is this actually possible?
			return false
		}

		if hasRenderedObj { // Earlier object has already rendered at pixel
			return false
		}

		if currentPixel.ColorID == ppu.COLOR_ID_WHITE { // BG is color 0
			return true
		}

		// BG master priority isn't set
		if !r.ppu.IsMasterBGPriorityEnabled() {
			return true
		}

		// BG doesn't have priority (CGB) AND OBJ has priority over BG
		if currentPixel.Layer != PIXEL_LAYER_BGP && !object.Attributes.BGPriority {
			return true
		}
	case ppu.ObjectPriorityModeDMG:
		// DMG mode: Object has lower priority x coordinate than currently rendered object
		if hasRenderedObj && renderedObjX <= uint8(object.PosX) {
			return false
		}

		if currentPixel.ColorID == ppu.COLOR_ID_WHITE { // BG is color 0
			return true
		}

		// OBJ has priority over BG
		if !object.Attributes.BGPriority {
			return true
		}
	}

	return false
}

func (r *ScanlineRenderer) readPixel(x, y uint8) RenderedPixel {
	return r.framebuf[y][x]
}

func (r *ScanlineRenderer) writePixel(x, y uint8, colorID ppu.ColorID, color color.Color, layer PixelLayer) {
	r.framebuf[y][x].Color = color
	r.framebuf[y][x].ColorID = colorID
	r.framebuf[y][x].Layer = layer
}
