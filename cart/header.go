package cart

import (
	"encoding/binary"
	"log"
)

var be = binary.BigEndian

const (
	entrypointOffset  = 0x100 // 0x100 - 0x103
	logoOffset        = 0x104 // 0x104 - 0x133
	titleOffset       = 0x134 // 0x134 - 0x142
	cgbOffset         = 0x143
	newLicenseeOffset = 0x144 // 0x144-0x145
	sgbOffset         = 0x146
	cartTypeOffset    = 0x147
	romSizeOffset     = 0x148
	ramSizeOffset     = 0x149
	destCodeOffset    = 0x14A
	oldLicenseeOffset = 0x14B
	maskRomVerOffset  = 0x14C
	headerChkOffset   = 0x14D
	globalChkOffset   = 0x14E // 0x14E-0x14F
	HEADER_START      = 0x100
	HEADER_END        = 0x14F
	HEADER_SIZE       = HEADER_END + 1
)

const (
	CGB_COLOR_NONE     = "No"
	CGB_COLOR_ENHANCED = "Color-enhanced"
	CGB_COLOR_ONLY     = "Color-only"
	CGB_UNKNOWN        = "Unknown"
)

type cartType byte

const (
	CART_TYPE_MBC0                       cartType = 0x00
	CART_TYPE_MBC1                       cartType = 0x01
	CART_TYPE_MBC1_RAM                   cartType = 0x02
	CART_TYPE_MBC1_RAM_BAT               cartType = 0x03
	CART_TYPE_MBC2                       cartType = 0x05
	CART_TYPE_MBC2_BAT                   cartType = 0x06
	CART_TYPE_UNK_ROM_RAM                cartType = 0x08
	CART_TYPE_UNK_ROM_RAM_BAT            cartType = 0x09
	CART_TYPE_MMM01                      cartType = 0x0B
	CART_TYPE_MMM01_RAM                  cartType = 0x0C
	CART_TYPE_MMM01_RAM_BAT              cartType = 0x0D
	CART_TYPE_MBC3_RTC_BAT               cartType = 0x0F
	CART_TYPE_MBC3_RTC_RAM_BAT           cartType = 0x10
	CART_TYPE_MBC3                       cartType = 0x11
	CART_TYPE_MBC3_RAM                   cartType = 0x12
	CART_TYPE_MBC3_RAM_BAT               cartType = 0x13
	CART_TYPE_MBC5                       cartType = 0x19
	CART_TYPE_MBC5_RAM                   cartType = 0x1A
	CART_TYPE_MBC5_RAM_BAT               cartType = 0x1B
	CART_TYPE_MBC5_RUMBLE                cartType = 0x1C
	CART_TYPE_MBC5_RUMBLE_RAM            cartType = 0x1D
	CART_TYPE_MBC5_RUMBLE_RAM_BAT        cartType = 0x1E
	CART_TYPE_MBC6                       cartType = 0x20
	CART_TYPE_MBC7_SENSOR_RUMBLE_RAM_BAT cartType = 0x22
	CART_TYPE_POCKET_CAM                 cartType = 0xFC
	CART_TYPE_BANDAI_TAMA5               cartType = 0xFD
	CART_TYPE_HUC3                       cartType = 0xFE
	CART_TYPE_HUC1_RAM_BAT               cartType = 0xFF
)

type Header struct {
	Title           string
	cgb             byte
	newLicenseeCode string
	sgb             byte
	CartType        cartType
	romSize         byte
	ramSize         byte
	destinationCode byte
	oldLicenseeCode byte
	maskROMVersion  byte
	HeaderChecksum  byte
	GlobalChecksum  uint16
}

func NewHeader(bytes []byte) Header {
	return Header{
		Title:           string(bytes[titleOffset:cgbOffset]),
		cgb:             bytes[cgbOffset],
		newLicenseeCode: string(bytes[newLicenseeOffset:sgbOffset]),
		sgb:             bytes[sgbOffset],
		CartType:        cartType(bytes[cartTypeOffset]),
		romSize:         bytes[romSizeOffset],
		ramSize:         bytes[ramSizeOffset],
		destinationCode: bytes[destCodeOffset],
		oldLicenseeCode: bytes[oldLicenseeOffset],
		maskROMVersion:  bytes[maskRomVerOffset],
		HeaderChecksum:  bytes[headerChkOffset],
		GlobalChecksum:  be.Uint16(bytes[globalChkOffset : HEADER_END+1]),
	}
}

func (hdr *Header) CartTypeName() string {
	switch hdr.CartType {
	case CART_TYPE_MBC0:
		return "ROM-only / MBC0"
	case CART_TYPE_MBC1:
		return "MBC1"
	case CART_TYPE_MBC1_RAM:
		return "MBC1+RAM"
	case CART_TYPE_MBC1_RAM_BAT:
		return "MBC1+RAM+BATTERY"
	case CART_TYPE_MBC2:
		return "MBC2"
	case CART_TYPE_MBC2_BAT:
		return "MBC2+BATTERY"
	case CART_TYPE_UNK_ROM_RAM:
		return "ROM+RAM"
	case CART_TYPE_UNK_ROM_RAM_BAT:
		return "ROM+RAM+BATTERY"
	case CART_TYPE_MMM01:
		return "MMM01"
	case CART_TYPE_MMM01_RAM:
		return "MMM01+RAM"
	case CART_TYPE_MMM01_RAM_BAT:
		return "MMM01+RAM+BATTERY"
	case CART_TYPE_MBC3_RTC_BAT:
		return "MBC3+TIMER+BATTERY"
	case CART_TYPE_MBC3_RTC_RAM_BAT:
		if hdr.IsMBC30() {
			return "MBC30+TIMER+RAM+BATTERY"
		}
		return "MBC3+TIMER+RAM+BATTERY"
	case CART_TYPE_MBC3:
		if hdr.IsMBC30() {
			return "MBC30"
		}
		return "MBC3"
	case CART_TYPE_MBC3_RAM:
		if hdr.IsMBC30() {
			return "MBC30+RAM"
		}
		return "MBC3+RAM"
	case CART_TYPE_MBC3_RAM_BAT:
		if hdr.IsMBC30() {
			return "MBC30+RAM+BATTERY"
		}
		return "MBC3+RAM+BATTERY"
	case CART_TYPE_MBC5:
		return "MBC5"
	case CART_TYPE_MBC5_RAM:
		return "MBC5+RAM"
	case CART_TYPE_MBC5_RAM_BAT:
		return "MBC5+RAM+BATTERY"
	case CART_TYPE_MBC5_RUMBLE:
		return "MBC5+RUMBLE"
	case CART_TYPE_MBC5_RUMBLE_RAM:
		return "MBC5+RUMBLE+RAM"
	case CART_TYPE_MBC5_RUMBLE_RAM_BAT:
		return "MBC5+RUMBLE+RAM+BATTERY"
	case CART_TYPE_MBC6:
		return "MBC6"
	case CART_TYPE_MBC7_SENSOR_RUMBLE_RAM_BAT:
		return "MBC7+SENSOR+RUMBLE+RAM+BATTERY"
	case CART_TYPE_POCKET_CAM:
		return "POCKET CAMERA"
	case CART_TYPE_BANDAI_TAMA5:
		return "BANDAI TAMA5"
	case CART_TYPE_HUC3:
		return "HuC3"
	case CART_TYPE_HUC1_RAM_BAT:
		return "HuC1+RAM+BATTERY"
	default:
		return "Unknown"
	}
}

func (hdr *Header) Cgb() string {
	if hdr.cgb == 0x00 {
		return CGB_COLOR_NONE
	} else if hdr.cgb == 0x80 {
		return CGB_COLOR_ENHANCED
	} else if hdr.cgb == 0xC0 {
		return CGB_COLOR_ONLY
	} else {
		return CGB_UNKNOWN
	}
}

func (hdr *Header) Destination() string {
	if hdr.destinationCode == 0x00 {
		return "JPN"
	} else {
		return "Non-JPN"
	}
}

func (hdr *Header) IsMBC30() bool {
	switch hdr.CartType {
	case CART_TYPE_MBC3,
		CART_TYPE_MBC3_RAM,
		CART_TYPE_MBC3_RAM_BAT,
		CART_TYPE_MBC3_RTC_BAT,
		CART_TYPE_MBC3_RTC_RAM_BAT:
		return hdr.ramSize == 0x05 || hdr.romSize == 0x07
	default:
		return false
	}
}

func (hdr *Header) Sgb() bool {
	return hdr.sgb == 0x03
}

func (hdr *Header) SgbMode() string {
	if hdr.Sgb() {
		return "Yes"
	} else {
		return "No"
	}
}

func (hdr *Header) RomSizeBytes() uint {
	return 32768 * (1 << hdr.romSize)
}

func (hdr *Header) RamSizeBytes() uint {
	switch hdr.ramSize {
	case 0x02:
		return 8192
	case 0x03:
		return 32768
	case 0x04:
		return 131072
	case 0x05:
		return 65536
	default:
		return 0
	}
}

func (hdr *Header) DebugPrint(logger *log.Logger) {
	logger.Printf("== Cartridge Info ==\n")
	logger.Printf("\n")
	logger.Printf("Title:		%s\n", hdr.Title)
	logger.Printf("Licensee:		%s\n", hdr.Licensee())
	logger.Printf("Color:		%s (0x%x)\n", hdr.Cgb(), hdr.cgb)
	logger.Printf("TV-Ready:		%s (0x%x)\n", hdr.SgbMode(), hdr.sgb)
	logger.Printf("Cart Type:		%s (0x%x)\n", hdr.CartTypeName(), hdr.CartType)
	logger.Printf("ROM Size:		%d KiB (0x%x)\n", hdr.RomSizeBytes()/1024, hdr.romSize)
	logger.Printf("RAM Size:		%d KiB (0x%x)\n", hdr.RamSizeBytes()/1024, hdr.ramSize)
	logger.Printf("Destination:	%s (0x%x)\n", hdr.Destination(), hdr.destinationCode)
	logger.Printf("Mask ROM Version:	0x%x\n", hdr.maskROMVersion)
	logger.Printf("Header Checksum:	0x%x\n", hdr.HeaderChecksum)
	logger.Printf("Global Checksum:	0x%x\n", hdr.GlobalChecksum)
}

func (hdr *Header) Licensee() string {
	if hdr.oldLicenseeCode != 0x33 {
		// https://gbdev.io/pandocs/The_Cartridge_Header.html#014b--old-licensee-code
		switch hdr.oldLicenseeCode {
		case 0x00:
			return "None"
		case 0x01:
			return "Nintendo"
		case 0x08:
			return "Capcom"
		case 0x09:
			return "Hot-B"
		case 0x0A:
			return "Jaleco"
		case 0x0B:
			return "Coconuts Japan"
		case 0x0C:
			return "Elite Systems"
		case 0x13:
			return "EA (Electronic Arts)"
		case 0x18:
			return "Hudsonsoft"
		case 0x19:
			return "ITC Entertainment"
		case 0x1A:
			return "Yanoman"
		case 0x1D:
			return "Japan Clary"
		case 0x1F:
			return "Virgin Interactive"
		case 0x24:
			return "PCM Complete"
		case 0x25:
			return "San-X"
		case 0x28:
			return "Kotobuki Systems"
		case 0x29:
			return "Seta"
		case 0x30:
			return "Infogrames"
		case 0x31:
			return "Nintendo"
		case 0x32:
			return "Bandai"
		case 0x34:
			return "Konami"
		case 0x35:
			return "HectorSoft"
		case 0x38:
			return "Capcom"
		case 0x39:
			return "Banpresto"
		case 0x3C:
			return ".Entertainment i"
		case 0x3E:
			return "Gremlin"
		case 0x41:
			return "Ubisoft"
		case 0x42:
			return "Atlus"
		case 0x44:
			return "Malibu"
		case 0x46:
			return "Angel"
		case 0x47:
			return "Spectrum Holoby"
		case 0x49:
			return "Irem"
		case 0x4A:
			return "Virgin Interactive"
		case 0x4D:
			return "Malibu"
		case 0x4F:
			return "U.S. Gold"
		case 0x50:
			return "Absolute"
		case 0x51:
			return "Acclaim"
		case 0x52:
			return "Activision"
		case 0x53:
			return "American Sammy"
		case 0x54:
			return "GameTek"
		case 0x55:
			return "Park Place"
		case 0x56:
			return "LJN"
		case 0x57:
			return "Matchbox"
		case 0x59:
			return "Milton Bradley"
		case 0x5A:
			return "Mindscape"
		case 0x5B:
			return "Romstar"
		case 0x5C:
			return "Naxat Soft"
		case 0x5D:
			return "Tradewest"
		case 0x60:
			return "Titus"
		case 0x61:
			return "Virgin Interactive"
		case 0x67:
			return "Ocean Interactive"
		case 0x69:
			return "EA (Electronic Arts)"
		case 0x6E:
			return "Elite Systems"
		case 0x6F:
			return "Electro Brain"
		case 0x70:
			return "Infogrames"
		case 0x71:
			return "Interplay"
		case 0x72:
			return "Broderbund"
		case 0x73:
			return "Sculptered Soft"
		case 0x75:
			return "The Sales Curve"
		case 0x78:
			return "t.hq"
		case 0x79:
			return "Accolade"
		case 0x7A:
			return "Triffix Entertainment"
		case 0x7C:
			return "Microprose"
		case 0x7F:
			return "Kemco"
		case 0x80:
			return "Misawa Entertainment"
		case 0x83:
			return "Lozc"
		case 0x86:
			return "Tokuma Shoten Intermedia"
		case 0x8B:
			return "Bullet-Proof Software"
		case 0x8C:
			return "Vic Tokai"
		case 0x8E:
			return "Ape"
		case 0x8F:
			return "I’Max"
		case 0x91:
			return "Chunsoft Co."
		case 0x92:
			return "Video System"
		case 0x93:
			return "Tsubaraya Productions Co."
		case 0x95:
			return "Varie Corporation"
		case 0x96:
			return "Yonezawa/S’Pal"
		case 0x97:
			return "Kaneko"
		case 0x99:
			return "Arc"
		case 0x9A:
			return "Nihon Bussan"
		case 0x9B:
			return "Tecmo"
		case 0x9C:
			return "Imagineer"
		case 0x9D:
			return "Banpresto"
		case 0x9F:
			return "Nova"
		case 0xA1:
			return "Hori Electric"
		case 0xA2:
			return "Bandai"
		case 0xA4:
			return "Konami"
		case 0xA6:
			return "Kawada"
		case 0xA7:
			return "Takara"
		case 0xA9:
			return "Technos Japan"
		case 0xAA:
			return "Broderbund"
		case 0xAC:
			return "Toei Animation"
		case 0xAD:
			return "Toho"
		case 0xAF:
			return "Namco"
		case 0xB0:
			return "acclaim"
		case 0xB1:
			return "ASCII or Nexsoft"
		case 0xB2:
			return "Bandai"
		case 0xB4:
			return "Square Enix"
		case 0xB6:
			return "HAL Laboratory"
		case 0xB7:
			return "SNK"
		case 0xB9:
			return "Pony Canyon"
		case 0xBA:
			return "Culture Brain"
		case 0xBB:
			return "Sunsoft"
		case 0xBD:
			return "Sony Imagesoft"
		case 0xBF:
			return "Sammy"
		case 0xC0:
			return "Taito"
		case 0xC2:
			return "Kemco"
		case 0xC3:
			return "Squaresoft"
		case 0xC4:
			return "Tokuma Shoten Intermedia"
		case 0xC5:
			return "Data East"
		case 0xC6:
			return "Tonkinhouse"
		case 0xC8:
			return "Koei"
		case 0xC9:
			return "UFL"
		case 0xCA:
			return "Ultra"
		case 0xCB:
			return "Vap"
		case 0xCC:
			return "Use Corporation"
		case 0xCD:
			return "Meldac"
		case 0xCE:
			return ".Pony Canyon or"
		case 0xCF:
			return "Angel"
		case 0xD0:
			return "Taito"
		case 0xD1:
			return "Sofel"
		case 0xD2:
			return "Quest"
		case 0xD3:
			return "Sigma Enterprises"
		case 0xD4:
			return "ASK Kodansha Co."
		case 0xD6:
			return "Naxat Soft"
		case 0xD7:
			return "Copya System"
		case 0xD9:
			return "Banpresto"
		case 0xDA:
			return "Tomy"
		case 0xDB:
			return "LJN"
		case 0xDD:
			return "NCS"
		case 0xDE:
			return "Human"
		case 0xDF:
			return "Altron"
		case 0xE0:
			return "Jaleco"
		case 0xE1:
			return "Towa Chiki"
		case 0xE2:
			return "Yutaka"
		case 0xE3:
			return "Varie"
		case 0xE5:
			return "Epcoh"
		case 0xE7:
			return "Athena"
		case 0xE8:
			return "Asmik ACE Entertainment"
		case 0xE9:
			return "Natsume"
		case 0xEA:
			return "King Records"
		case 0xEB:
			return "Atlus"
		case 0xEC:
			return "Epic/Sony Records"
		case 0xEE:
			return "IGS"
		case 0xF0:
			return "A Wave"
		case 0xF3:
			return "Extreme Entertainment"
		case 0xFF:
			return "LJN"
		default:
			return "Unknown"
		}
	} else {
		// https://gbdev.io/pandocs/The_Cartridge_Header.html#01440145--new-licensee-code
		switch hdr.newLicenseeCode {
		case "00":
			return "None"
		case "01":
			return "Nintendo R&D1"
		case "08":
			return "Capcom"
		case "13":
			return "Electronic Arts"
		case "18":
			return "Hudson Soft"
		case "19":
			return "b-ai"
		case "20":
			return "kss"
		case "22":
			return "pow"
		case "24":
			return "PCM Complete"
		case "25":
			return "san-x"
		case "28":
			return "Kemco Japan"
		case "29":
			return "seta"
		case "30":
			return "Viacom"
		case "31":
			return "Nintendo"
		case "32":
			return "Bandai"
		case "33":
			return "Ocean/Acclaim"
		case "34":
			return "Konami"
		case "35":
			return "Hector"
		case "37":
			return "Taito"
		case "38":
			return "Hudson"
		case "39":
			return "Banpresto"
		case "41":
			return "Ubi Soft"
		case "42":
			return "Atlus"
		case "44":
			return "Malibu"
		case "46":
			return "angel"
		case "47":
			return "Bullet-Proof"
		case "49":
			return "irem"
		case "50":
			return "Absolute"
		case "51":
			return "Acclaim"
		case "52":
			return "Activision"
		case "53":
			return "American sammy"
		case "54":
			return "Konami"
		case "55":
			return "Hi tech entertainment"
		case "56":
			return "LJN"
		case "57":
			return "Matchbox"
		case "58":
			return "Mattel"
		case "59":
			return "Milton Bradley"
		case "60":
			return "Titus"
		case "61":
			return "Virgin"
		case "64":
			return "LucasArts"
		case "67":
			return "Ocean"
		case "69":
			return "Electronic Arts"
		case "70":
			return "Infogrames"
		case "71":
			return "Interplay"
		case "72":
			return "Broderbund"
		case "73":
			return "sculptured"
		case "75":
			return "sci"
		case "78":
			return "THQ"
		case "79":
			return "Accolade"
		case "80":
			return "misawa"
		case "83":
			return "lozc"
		case "86":
			return "Tokuma Shoten Intermedia"
		case "87":
			return "Tsukuda Original"
		case "91":
			return "Chunsoft"
		case "92":
			return "Video system"
		case "93":
			return "Ocean/Acclaim"
		case "95":
			return "Varie"
		case "96":
			return "Yonezawa/s’pal"
		case "97":
			return "Kaneko"
		case "99":
			return "Pack in soft"
		case "9H":
			return "Bottom Up"
		case "A4":
			return "Konami (Yu-Gi-Oh!)"
		default:
			return "Unknown"
		}
	}
}
