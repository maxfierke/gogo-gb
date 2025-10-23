package rendering

import (
	"image"
	"image/color"

	gfx "github.com/maxfierke/gogo-gb/ppu"
)

type fetcherStep int

const (
	FETCHER_STEP_GET_TILE_INDEX fetcherStep = iota
	FETCHER_STEP_GET_TILE_LOW
	FETCHER_STEP_GET_TILE_HIGH
	FETCHER_STEP_PUSH
	FETCHER_STEP_SLEEP
)

type pixelFIFO struct {
	pixels []RenderedPixel
}

func (pf *pixelFIFO) IsFull() bool {
	return len(pf.pixels) == 16
}

func (pf *pixelFIFO) CanPushRow() bool {
	return len(pf.pixels) <= 8
}

func (pf *pixelFIFO) Clear() {
	pf.pixels = nil
}

func (pf *pixelFIFO) Push(pixel RenderedPixel) {
	if pf.IsFull() {
		panic("fifo is full")
	}
	pf.pixels = append(pf.pixels, pixel)
}

func (pf *pixelFIFO) Pop() (RenderedPixel, bool) {
	if len(pf.pixels) == 0 {
		return RenderedPixel{}, false
	}

	pixel := pf.pixels[0]
	pf.pixels = pf.pixels[1:]
	return pixel, true
}

type bgFetcher struct {
	currentX       uint8
	fifo           pixelFIFO
	fetcherStep    fetcherStep
	tileIndex      uint8
	tileRow        [8]gfx.PPUPixel
	tileAttributes gfx.BGAttributes
	tilePixelY     uint8
	tileX          uint8
	tileY          uint8
	renderedWindow bool
}

func (b *bgFetcher) Reset() {
	b.renderedWindow = false
	b.currentX = 0
	b.tileIndex = 0
	clear(b.tileRow[:])
	b.tileAttributes = gfx.BGAttributes{}
	b.tilePixelY = 0
	b.tileX = 0
	b.tileY = 0
	b.fifo.Clear()
}

func (b *bgFetcher) Step(ppu *gfx.PPU, vram *gfx.VRAM) {
	switch b.fetcherStep {
	case FETCHER_STEP_GET_TILE_INDEX:
		currentScanLine := ppu.CurrentScanline()
		scrollBackgroundX := ppu.ScrollBackgroundX()
		scrollBackgroundY := ppu.ScrollBackgroundY()
		windowX := ppu.WindowX()
		windowY := ppu.WindowY()
		isWithinWindow := currentScanLine >= windowY &&
			uint16(b.currentX+7) >= uint16(windowX)

		tileMap := ppu.GetBGTilemap()
		b.tileY = (currentScanLine + scrollBackgroundY) / 8
		b.tilePixelY = (currentScanLine + scrollBackgroundY) % 8
		scrollAdjustedLineX := (uint16(b.currentX) + uint16(scrollBackgroundX)) % 256
		b.tileX = uint8(scrollAdjustedLineX / 8)

		if isWithinWindow {
			// TODO: do this in PPU instead
			if currentScanLine == windowY {
				ppu.ResetWindow()
			}

			currentWindowLine := ppu.CurrentWindowLine()
			tileMap = ppu.GetWindowTilemap()

			b.tileY = currentWindowLine / 8
			b.tilePixelY = currentWindowLine % 8
			windowAdjustedLineX := (uint16(b.currentX+7) - uint16(windowX))
			b.tileX = uint8(windowAdjustedLineX / 8)
			b.renderedWindow = true
		}

		b.tileAttributes = vram.GetBGTileAttributes(
			tileMap,
			b.tileX,
			b.tileY,
		)

		b.tileIndex = vram.GetBGTileIndex(
			tileMap,
			uint8(b.tileX),
			b.tileY,
		)
		b.fetcherStep = FETCHER_STEP_GET_TILE_LOW
	case FETCHER_STEP_GET_TILE_LOW:
		var tileVRAMBank uint8
		if ppu.IsColorEnabled() {
			tileVRAMBank = b.tileAttributes.VRAMBank
		}

		tile := vram.GetBGTile(
			tileVRAMBank,
			ppu.GetBGWindowTileset(),
			b.tileIndex,
		)

		b.tileRow = tile[b.tilePixelY]
		if b.tileAttributes.FlipY {
			b.tileRow = tile[7-b.tilePixelY]
		}
		b.fetcherStep = FETCHER_STEP_GET_TILE_HIGH
	case FETCHER_STEP_GET_TILE_HIGH:
		// We already have the whole tile, so until this really matters, just skip
		// over it.
		b.fetcherStep = FETCHER_STEP_SLEEP
	case FETCHER_STEP_SLEEP:
		b.fetcherStep = FETCHER_STEP_PUSH
	case FETCHER_STEP_PUSH:
		if b.fifo.CanPushRow() {
			for tilePixelX := range 8 {
				if b.tileAttributes.FlipX {
					tilePixelX = 7 - tilePixelX
				}

				tilePixelValue := b.tileRow[tilePixelX]
				pixelColorID := gfx.ColorID(tilePixelValue)
				pixelLayer := PIXEL_LAYER_BG
				if b.tileAttributes.Priority && ppu.IsColorEnabled() {
					pixelLayer = PIXEL_LAYER_BGP
				}

				b.fifo.Push(RenderedPixel{
					Layer:     pixelLayer,
					ColorID:   pixelColorID,
					PaletteID: b.tileAttributes.PaletteID,
				})
				b.currentX++
			}
			b.fetcherStep = FETCHER_STEP_GET_TILE_INDEX
		}
	}
}

type FIFORenderer struct {
	ppu  *gfx.PPU
	oam  *gfx.OAM
	vram *gfx.VRAM

	framebuf [FB_HEIGHT][FB_WIDTH]RenderedPixel

	currentX        uint8
	renderedObjects uint8

	bg *bgFetcher
}

var _ gfx.Renderer = (*FIFORenderer)(nil)

func FIFO(ppu *gfx.PPU, oam *gfx.OAM, vram *gfx.VRAM) gfx.Renderer {
	return &FIFORenderer{
		bg:   &bgFetcher{},
		ppu:  ppu,
		oam:  oam,
		vram: vram,
	}
}

func (r *FIFORenderer) DrawImage() image.Image {
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

func (r *FIFORenderer) Reset() {
	r.currentX = 0
	r.renderedObjects = 0
	r.bg.Reset()
}

func (r *FIFORenderer) writePixel(x, y uint8, colorID gfx.ColorID, color color.Color, layer PixelLayer) {
	r.framebuf[y][x].Color = color
	r.framebuf[y][x].ColorID = colorID
	r.framebuf[y][x].Layer = layer
}

func (r *FIFORenderer) Step(cycles uint8) uint8 {
	if !r.ppu.IsLCDEnabled() || r.ppu.CurrentScanline() >= FB_HEIGHT {
		return 0
	}

	pixelsRendered := uint8(0)

	for cycle := range cycles {
		if cycle%2 == 0 {
			r.bg.Step(r.ppu, r.vram)
		}

		if pixel, ok := r.bg.fifo.Pop(); ok {
			color := r.ppu.GetBGPaletteColor(pixel.ColorID, pixel.PaletteID)
			r.writePixel(r.currentX, r.ppu.CurrentScanline(), pixel.ColorID, color, pixel.Layer)
			pixelsRendered++
			r.currentX++
		}

		if r.currentX >= FB_WIDTH {
			if r.bg.renderedWindow {
				// TODO: Do this in PPU
				r.ppu.IncrementWindowLine()
			}
			break
		}
	}

	return pixelsRendered
}
