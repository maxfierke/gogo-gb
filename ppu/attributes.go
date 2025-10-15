package ppu

import "github.com/maxfierke/gogo-gb/bits"

const (
	BG_ATTR_BIT_VRAM_BANK   = 3
	BG_ATTR_BIT_X_FLIP      = 5
	BG_ATTR_BIT_Y_FLIP      = 6
	BG_ATTR_BIT_BG_PRIORITY = 7

	BG_ATTR_MASK_PALETTE_ID = 0x7
)

type bgAttributes struct {
	priority  bool
	flipY     bool
	flipX     bool
	vramBank  uint8
	paletteID uint8
}

func (attrs *bgAttributes) Read() uint8 {
	var (
		priority  uint8
		flipY     uint8
		flipX     uint8
		vramBank  uint8
		paletteID uint8
	)

	if attrs.priority {
		priority = 1 << BG_ATTR_BIT_BG_PRIORITY
	}

	if attrs.flipY {
		flipY = 1 << BG_ATTR_BIT_Y_FLIP
	}

	if attrs.flipX {
		flipX = 1 << BG_ATTR_BIT_X_FLIP
	}

	vramBank = attrs.vramBank << BG_ATTR_BIT_VRAM_BANK
	paletteID = attrs.paletteID & BG_ATTR_MASK_PALETTE_ID

	return priority | flipY | flipX | vramBank | paletteID
}

func (attrs *bgAttributes) Write(value uint8) {
	attrs.priority = bits.Read(value, BG_ATTR_BIT_BG_PRIORITY) == 1
	attrs.flipY = bits.Read(value, BG_ATTR_BIT_Y_FLIP) == 1
	attrs.flipX = bits.Read(value, BG_ATTR_BIT_X_FLIP) == 1
	attrs.vramBank = bits.Read(value, BG_ATTR_BIT_VRAM_BANK)
	attrs.paletteID = value & BG_ATTR_MASK_PALETTE_ID
}

type objectData struct {
	posY       uint8
	posX       uint8
	tileIndex  uint8
	attributes objectAttributes
}

const (
	OAM_ATTR_BIT_VRAM_BANK      = 3
	OAM_ATTR_BIT_DMG_PALETTE_ID = 4
	OAM_ATTR_BIT_X_FLIP         = 5
	OAM_ATTR_BIT_Y_FLIP         = 6
	OAM_ATTR_BIT_BG_PRIORITY    = 7

	OAM_ATTR_MASK_CGB_PALETTE_ID = 0x7
)

type objectAttributes struct {
	bgPriority   bool
	flipY        bool
	flipX        bool
	dmgPaletteID uint8
	vramBank     uint8
	cgbPaletteID uint8
}

func (attrs *objectAttributes) Read() uint8 {
	var (
		bgPriority   uint8
		flipY        uint8
		flipX        uint8
		dmgPaletteID uint8
		vramBank     uint8
		cgbPaletteID uint8
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

	dmgPaletteID = attrs.dmgPaletteID << OAM_ATTR_BIT_DMG_PALETTE_ID
	vramBank = attrs.vramBank << OAM_ATTR_BIT_VRAM_BANK
	cgbPaletteID = attrs.cgbPaletteID & OAM_ATTR_MASK_CGB_PALETTE_ID

	return bgPriority | flipY | flipX | dmgPaletteID | vramBank | cgbPaletteID
}

func (attrs *objectAttributes) Write(value uint8) {
	attrs.bgPriority = bits.Read(value, OAM_ATTR_BIT_BG_PRIORITY) == 1
	attrs.flipY = bits.Read(value, OAM_ATTR_BIT_Y_FLIP) == 1
	attrs.flipX = bits.Read(value, OAM_ATTR_BIT_X_FLIP) == 1
	attrs.dmgPaletteID = bits.Read(value, OAM_ATTR_BIT_DMG_PALETTE_ID)
	attrs.vramBank = bits.Read(value, OAM_ATTR_BIT_VRAM_BANK)
	attrs.cgbPaletteID = value & OAM_ATTR_MASK_CGB_PALETTE_ID
}

const (
	OAM_START                    uint16 = 0xFE00
	OAM_END                      uint16 = 0xFE9F
	OAM_MAX_OBJECTS_PER_SCANLINE        = 10
	OAM_MAX_OBJECT_COUNT                = 40
	OAM_SIZE                            = OAM_END - OAM_START + 1
)

type oam struct {
	raw        [OAM_SIZE]byte
	objectData [OAM_MAX_OBJECT_COUNT]objectData
}

func (o *oam) Objects() [OAM_MAX_OBJECT_COUNT]objectData {
	return o.objectData
}

func (o *oam) Read(oamAddr uint8) byte {
	return o.raw[oamAddr]
}

func (o *oam) Write(oamAddr uint8, value byte) {
	o.raw[oamAddr] = value
	o.writeObj(oamAddr, value)
}

func (o *oam) writeObj(oamAddr uint8, value byte) {
	objIndex := oamAddr / 4
	if objIndex > OAM_MAX_OBJECT_COUNT {
		return
	}

	byteIndex := oamAddr % 4

	switch byteIndex {
	case 0:
		o.objectData[objIndex].posY = value - 16
	case 1:
		o.objectData[objIndex].posX = value - 8
	case 2:
		o.objectData[objIndex].tileIndex = value
	default:
		o.objectData[objIndex].attributes.Write(value)
	}
}
