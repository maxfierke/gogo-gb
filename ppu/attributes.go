package ppu

import "github.com/maxfierke/gogo-gb/bits"

const (
	BG_ATTR_BIT_VRAM_BANK   = 3
	BG_ATTR_BIT_X_FLIP      = 5
	BG_ATTR_BIT_Y_FLIP      = 6
	BG_ATTR_BIT_BG_PRIORITY = 7

	BG_ATTR_MASK_PALETTE_ID = 0x7
)

type BGAttributes struct {
	Priority  bool
	FlipY     bool
	FlipX     bool
	VRAMBank  uint8
	PaletteID uint8
}

func (attrs *BGAttributes) Read() uint8 {
	var (
		priority  uint8
		flipY     uint8
		flipX     uint8
		vramBank  uint8
		paletteID uint8
	)

	if attrs.Priority {
		priority = 1 << BG_ATTR_BIT_BG_PRIORITY
	}

	if attrs.FlipY {
		flipY = 1 << BG_ATTR_BIT_Y_FLIP
	}

	if attrs.FlipX {
		flipX = 1 << BG_ATTR_BIT_X_FLIP
	}

	vramBank = attrs.VRAMBank << BG_ATTR_BIT_VRAM_BANK
	paletteID = attrs.PaletteID & BG_ATTR_MASK_PALETTE_ID

	return priority | flipY | flipX | vramBank | paletteID
}

func (attrs *BGAttributes) Write(value uint8) {
	attrs.Priority = bits.Read(value, BG_ATTR_BIT_BG_PRIORITY) == 1
	attrs.FlipY = bits.Read(value, BG_ATTR_BIT_Y_FLIP) == 1
	attrs.FlipX = bits.Read(value, BG_ATTR_BIT_X_FLIP) == 1
	attrs.VRAMBank = bits.Read(value, BG_ATTR_BIT_VRAM_BANK)
	attrs.PaletteID = value & BG_ATTR_MASK_PALETTE_ID
}

type ObjectData struct {
	PosY       uint8
	PosX       uint8
	TileIndex  uint8
	Attributes ObjectAttributes
}

const (
	OAM_ATTR_BIT_VRAM_BANK      = 3
	OAM_ATTR_BIT_DMG_PALETTE_ID = 4
	OAM_ATTR_BIT_X_FLIP         = 5
	OAM_ATTR_BIT_Y_FLIP         = 6
	OAM_ATTR_BIT_BG_PRIORITY    = 7

	OAM_ATTR_MASK_CGB_PALETTE_ID = 0x7
)

type ObjectAttributes struct {
	BGPriority   bool
	FlipY        bool
	FlipX        bool
	DMGPaletteID uint8
	VRAMBank     uint8
	CGBPaletteID uint8
}

func (attrs *ObjectAttributes) Read() uint8 {
	var (
		bgPriority   uint8
		flipY        uint8
		flipX        uint8
		dmgPaletteID uint8
		vramBank     uint8
		cgbPaletteID uint8
	)

	if attrs.BGPriority {
		bgPriority = 1 << OAM_ATTR_BIT_BG_PRIORITY
	}

	if attrs.FlipY {
		flipY = 1 << OAM_ATTR_BIT_Y_FLIP
	}

	if attrs.FlipX {
		flipX = 1 << OAM_ATTR_BIT_X_FLIP
	}

	dmgPaletteID = attrs.DMGPaletteID << OAM_ATTR_BIT_DMG_PALETTE_ID
	vramBank = attrs.VRAMBank << OAM_ATTR_BIT_VRAM_BANK
	cgbPaletteID = attrs.CGBPaletteID & OAM_ATTR_MASK_CGB_PALETTE_ID

	return bgPriority | flipY | flipX | dmgPaletteID | vramBank | cgbPaletteID
}

func (attrs *ObjectAttributes) Write(value uint8) {
	attrs.BGPriority = bits.Read(value, OAM_ATTR_BIT_BG_PRIORITY) == 1
	attrs.FlipY = bits.Read(value, OAM_ATTR_BIT_Y_FLIP) == 1
	attrs.FlipX = bits.Read(value, OAM_ATTR_BIT_X_FLIP) == 1
	attrs.DMGPaletteID = bits.Read(value, OAM_ATTR_BIT_DMG_PALETTE_ID)
	attrs.VRAMBank = bits.Read(value, OAM_ATTR_BIT_VRAM_BANK)
	attrs.CGBPaletteID = value & OAM_ATTR_MASK_CGB_PALETTE_ID
}

const (
	OAM_START                    uint16 = 0xFE00
	OAM_END                      uint16 = 0xFE9F
	OAM_MAX_OBJECTS_PER_SCANLINE        = 10
	OAM_MAX_OBJECT_COUNT                = 40
	OAM_SIZE                            = OAM_END - OAM_START + 1
)

type OAM struct {
	raw        [OAM_SIZE]byte
	objectData [OAM_MAX_OBJECT_COUNT]ObjectData
}

func (o *OAM) Objects() []ObjectData {
	return o.objectData[:]
}

func (o *OAM) Read(oamAddr uint8) byte {
	return o.raw[oamAddr]
}

func (o *OAM) Write(oamAddr uint8, value byte) {
	o.raw[oamAddr] = value
	o.writeObj(oamAddr, value)
}

func (o *OAM) writeObj(oamAddr uint8, value byte) {
	objIndex := oamAddr / 4
	if objIndex > OAM_MAX_OBJECT_COUNT {
		return
	}

	byteIndex := oamAddr % 4

	switch byteIndex {
	case 0:
		o.objectData[objIndex].PosY = value - 16
	case 1:
		o.objectData[objIndex].PosX = value - 8
	case 2:
		o.objectData[objIndex].TileIndex = value
	default:
		o.objectData[objIndex].Attributes.Write(value)
	}
}
