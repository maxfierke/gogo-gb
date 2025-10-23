package ppu

const (
	VRAM_BANKS                  = 2 // Bank 0/1. Bank 1 is CGB only
	VRAM_START           uint16 = 0x8000
	VRAM_TILESET_1_START uint16 = VRAM_START
	VRAM_TILESET_2_START uint16 = 0x8800
	VRAM_TILESET_1_END   uint16 = 0x8FFF
	VRAM_TILESET_2_END   uint16 = 0x97FF
	VRAM_TILEMAP_1_START uint16 = 0x9800
	VRAM_TILEMAP_1_END   uint16 = 0x9BFF
	VRAM_TILEMAP_2_START uint16 = 0x9C00
	VRAM_TILEMAP_2_END   uint16 = 0x9FFF
	VRAM_TILEMAP_SIZE           = VRAM_TILEMAP_2_END - VRAM_TILEMAP_1_START + 1 // 2K
	VRAM_BG_ATTR_SIZE           = VRAM_TILEMAP_2_END - VRAM_TILEMAP_1_START + 1 // 2K
	VRAM_TILESET_SIZE           = 384
	VRAM_TILE_ROW_MASK   uint16 = 0xFFFE
	VRAM_END             uint16 = 0x9FFF
	VRAM_SIZE                   = VRAM_END - VRAM_START + 1 // 8K
)

type (
	PPUPixel uint8
	Tile     [8][8]PPUPixel
)

const (
	VRAM_TILE_PIXEL_ZERO PPUPixel = iota
	VRAM_TILE_PIXEL_ONE
	VRAM_TILE_PIXEL_TWO
	VRAM_TILE_PIXEL_THREE
)

type VRAM struct {
	CurrentBank uint8

	cgbBGAttributes [VRAM_BG_ATTR_SIZE]BGAttributes
	vram            [VRAM_BANKS][VRAM_SIZE]byte
	tileset         [VRAM_BANKS][VRAM_TILESET_SIZE]Tile
}

func NewVRAM() *VRAM {
	return &VRAM{}
}

func (v *VRAM) Read(vramAddr uint16) byte {
	return v.vram[v.CurrentBank][vramAddr]
}

func (v *VRAM) Write(vramAddr uint16, value byte) {
	v.vram[v.CurrentBank][vramAddr] = value

	if vramAddr <= (VRAM_TILESET_2_END - VRAM_START) {
		v.writeTile(vramAddr)
	} else if v.CurrentBank == 1 {
		v.writeBGAttr(vramAddr, value)
	}
}

func (v *VRAM) SetCurrentBank(value uint8) {
	if value > 1 {
		panic("illegal vram bank")
	}

	v.CurrentBank = value
}

func (v *VRAM) GetBGTileIndex(tilemapArea tileMapArea, tileX, tileY uint8) uint8 {
	tileMapAddr := VRAM_TILEMAP_1_START
	if tilemapArea == TILEMAP_AREA2 {
		tileMapAddr = VRAM_TILEMAP_2_START
	}

	tileMapOffset := tileMapAddr - VRAM_START
	tileMapIndex := uint16(tileY)*32 + uint16(tileX)
	tileIndex := v.vram[0][tileMapOffset+tileMapIndex]

	return tileIndex
}

func (v *VRAM) GetBGTileAttributes(tilemapArea tileMapArea, tileX, tileY uint8) BGAttributes {
	var tileMapOffset uint16
	if tilemapArea == TILEMAP_AREA2 {
		tileMapOffset = VRAM_BG_ATTR_SIZE / 2
	}

	tileMapIndex := uint16(tileY)*32 + uint16(tileX)
	return v.cgbBGAttributes[tileMapOffset+tileMapIndex]
}

func (v *VRAM) GetBGTile(bank uint8, tilesetArea tileSetArea, tileIndex uint8) Tile {
	tile := v.tileset[bank][tileIndex]
	if tilesetArea == TILESET_1 && tileIndex < 128 {
		tile = v.tileset[bank][VRAM_TILESET_SIZE-128+uint16(tileIndex)]
	}

	return tile
}

func (v *VRAM) GetObjTile(object ObjectData, objSize objectSize, tileY uint8, colorEnabled bool) Tile {
	tileIndex := object.TileIndex
	if objSize == OBJ_SIZE_8x16 {
		// Ignore bit 0 for 8x16
		tileIndex &= 0xFE

		if (!object.Attributes.FlipY && tileY > 7) || (object.Attributes.FlipY && tileY <= 7) {
			tileIndex += 1
		}
	}

	var tileVRAMBank uint8
	if colorEnabled {
		tileVRAMBank = object.Attributes.VRAMBank
	}

	return v.tileset[tileVRAMBank][tileIndex]
}

func (v *VRAM) writeBGAttr(vramAddr uint16, value uint8) {
	attrIndex := vramAddr - (VRAM_TILEMAP_1_START - VRAM_START)
	if vramAddr >= (VRAM_TILEMAP_2_START - VRAM_START) {
		attrIndex = vramAddr - (VRAM_TILEMAP_2_START - VRAM_START) + (VRAM_BG_ATTR_SIZE / 2)
	}
	v.cgbBGAttributes[attrIndex].Write(value)
}

func (v *VRAM) writeTile(vramAddr uint16) {
	// https://rylev.github.io/DMG-01/public/book/graphics/tile_ram.html
	rowAddr := vramAddr & VRAM_TILE_ROW_MASK

	tileRowTop := v.vram[v.CurrentBank][rowAddr]
	tileRowBottom := v.vram[v.CurrentBank][rowAddr+1]

	tileIdx := vramAddr / 16
	rowIdx := (vramAddr % 16) / 2

	for pixelIdx := range v.tileset[v.CurrentBank][tileIdx][rowIdx] {
		pixelMask := byte(1 << (7 - pixelIdx))
		lsb := tileRowTop & pixelMask
		msb := tileRowBottom & pixelMask

		if lsb == 0 && msb == 0 {
			v.tileset[v.CurrentBank][tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_ZERO
		} else if lsb != 0 && msb == 0 {
			v.tileset[v.CurrentBank][tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_ONE
		} else if lsb == 0 && msb != 0 {
			v.tileset[v.CurrentBank][tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_TWO
		} else {
			v.tileset[v.CurrentBank][tileIdx][rowIdx][pixelIdx] = VRAM_TILE_PIXEL_THREE
		}
	}
}
