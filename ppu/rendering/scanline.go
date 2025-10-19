package rendering

import "github.com/maxfierke/gogo-gb/ppu"

type ScanlineRenderer struct {
	ppu  *ppu.PPU
	oam  *ppu.OAM
	vram *ppu.VRAM
}

var _ ppu.Renderer = (*ScanlineRenderer)(nil)

func Scanline(ppu *ppu.PPU, oam *ppu.OAM, vram *ppu.VRAM) ppu.Renderer {
	return &ScanlineRenderer{
		ppu:  ppu,
		oam:  oam,
		vram: vram,
	}
}

func (r *ScanlineRenderer) DrawPixels() {
	if !r.ppu.IsLCDEnabled() || r.ppu.CurrentScanline() >= ppu.FB_HEIGHT {
		return
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
}

func (r *ScanlineRenderer) drawBgScanline() {
	currentScanLine := r.ppu.CurrentScanline()
	scrollBackgroundX := r.ppu.ScrollBackgroundX()
	scrollBackgroundY := r.ppu.ScrollBackgroundY()

	tileY := (currentScanLine + scrollBackgroundY) / 8
	tilePixelY := (currentScanLine + scrollBackgroundY) % 8

	for lineX := uint16(0); lineX < ppu.FB_WIDTH; lineX++ {
		scrollAdjustedLineX := (lineX + uint16(scrollBackgroundX)) % 256
		tileX := uint8(scrollAdjustedLineX / 8)

		tileIndex := r.vram.GetBGTileIndex(
			r.ppu.GetBGTilemap(),
			uint8(tileX),
			tileY,
		)

		bgAttributes := r.vram.GetBGTileAttributes(
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
		pixelLayer := ppu.PIXEL_LAYER_BG
		if bgAttributes.Priority && r.ppu.IsColorEnabled() {
			pixelLayer = ppu.PIXEL_LAYER_BGP
		}

		r.ppu.WritePixel(uint8(lineX), currentScanLine, pixelColorID, color, pixelLayer)
	}
}

func (r *ScanlineRenderer) drawWinScanline() {
	currentScanLine := r.ppu.CurrentScanline()
	windowX := r.ppu.WindowX()
	windowY := r.ppu.WindowY()

	if currentScanLine >= windowY {
		if currentScanLine == windowY {
			r.ppu.ResetWindow()
		}

		currentWindowLine := r.ppu.CurrentWindowLine()
		tileY := currentWindowLine / 8
		tilePixelY := currentWindowLine % 8

		rendered := false

		for lineX := uint16(0); lineX < ppu.FB_WIDTH; lineX++ {
			if (lineX + 7) < uint16(windowX) {
				continue
			}

			rendered = true

			windowAdjustedLineX := (lineX + 7 - uint16(windowX))
			tileX := uint8(windowAdjustedLineX / 8)

			tileIndex := r.vram.GetBGTileIndex(
				r.ppu.GetWindowTilemap(),
				uint8(tileX),
				tileY,
			)
			bgAttributes := r.vram.GetBGTileAttributes(
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
			pixelLayer := ppu.PIXEL_LAYER_BG
			if bgAttributes.Priority && r.ppu.IsColorEnabled() {
				pixelLayer = ppu.PIXEL_LAYER_BGP
			}

			r.ppu.WritePixel(uint8(lineX), currentScanLine, pixelColorID, color, pixelLayer)
		}

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
	objectPriorityMode := r.ppu.ObjectPriority()

	renderedObjects := 0
	renderedObjectsX := map[uint8]uint8{}

	for _, object := range r.oam.Objects() {
		if renderedObjects == ppu.OAM_MAX_OBJECTS_PER_SCANLINE {
			break
		}

		if object.PosY <= currentScanLine && (object.PosY+objHeight) > currentScanLine {
			objPixelY := currentScanLine - object.PosY

			tile := r.vram.GetObjTile(
				object,
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
			for x := uint8(0); x < 8; x++ {
				tilePixelX := x
				if object.Attributes.FlipX {
					tilePixelX = 7 - x
				}

				pixelX := object.PosX + x

				if pixelX >= ppu.FB_WIDTH {
					// Skip pixels outside of rendering area
					continue
				}

				tilePixelValue := tileRow[tilePixelX]

				renderedObjX, hasRenderedObj := renderedObjectsX[pixelX]

				currentPixel := r.ppu.ReadPixel(pixelX, currentScanLine)

				if tilePixelValue != ppu.VRAM_TILE_PIXEL_ZERO && // Skip transparent pixels
					((objectPriorityMode == ppu.ObjectPriorityModeCGB && r.ppu.IsColorEnabled() && !hasRenderedObj) || // CGB mode: Earlier Object hasn't rendered at pixel
						// DMG mode: Object has higher priority x coordinate than currently rendered object
						(objectPriorityMode == ppu.ObjectPriorityModeDMG &&
							(!hasRenderedObj || (hasRenderedObj && renderedObjX > object.PosX)))) && // TODO: Extract method
					(currentPixel.ColorID == ppu.COLOR_ID_WHITE || // BG is color 0
						// CGB: BG master priority isn't set
						objectPriorityMode == ppu.ObjectPriorityModeCGB && !r.ppu.IsMasterBGPriorityEnabled() ||
						// BG doesn't have priority (CGB) AND OBJ has priority over BG
						(currentPixel.Layer != ppu.PIXEL_LAYER_BGP && !object.Attributes.BGPriority)) { // TODO: Extract method

					pixelColorID := ppu.ColorID(tilePixelValue)
					color := r.ppu.GetObjPaletteColor(pixelColorID, object.Attributes)
					pixelLayer := ppu.PIXEL_LAYER_OBJ

					r.ppu.WritePixel(pixelX, currentScanLine, pixelColorID, color, pixelLayer)

					renderedObject = true
					renderedObjectsX[pixelX] = object.PosX
				}
			}

			if renderedObject {
				renderedObjects += 1
			}
		}
	}
}
