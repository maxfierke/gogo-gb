package cart

import (
	"encoding/binary"
	"fmt"
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
)

type Header struct {
	Title           string
	Cgb             byte
	NewLicenseeCode uint16
	Sgb             byte
	CartType        byte
	RomSize         byte
	RamSize         byte
	DestinationCode byte
	OldLicenseeCode byte
	MaskROMVersion  byte
	HeaderChecksum  byte
	GlobalChecksum  uint16
}

func NewHeader(bytes []byte) Header {
	return Header{
		Title:           string(bytes[titleOffset:cgbOffset]),
		Cgb:             bytes[cgbOffset],
		NewLicenseeCode: be.Uint16(bytes[newLicenseeOffset:sgbOffset]),
		Sgb:             bytes[sgbOffset],
		CartType:        bytes[cartTypeOffset],
		RomSize:         bytes[romSizeOffset],
		RamSize:         bytes[ramSizeOffset],
		DestinationCode: bytes[destCodeOffset],
		OldLicenseeCode: bytes[oldLicenseeOffset],
		MaskROMVersion:  bytes[maskRomVerOffset],
		HeaderChecksum:  bytes[headerChkOffset],
		GlobalChecksum:  be.Uint16(bytes[globalChkOffset : HEADER_END+1]),
	}
}

func (hdr *Header) DebugPrint() {
	fmt.Printf("== Cartridge Info ==\n\n")

	fmt.Printf("Title:			%s\n", hdr.Title)
	fmt.Printf("CGB flag:		0x%x\n", hdr.Cgb)
	fmt.Printf("New Licensee Code:	0x%x\n", hdr.NewLicenseeCode)
	fmt.Printf("SGB flag:		0x%x\n", hdr.Sgb)
	fmt.Printf("Cart Type:		0x%x\n", hdr.CartType)
	fmt.Printf("ROM Size:		0x%x\n", hdr.RomSize)
	fmt.Printf("RAM Size:		0x%x\n", hdr.RamSize)
	fmt.Printf("Destination Code:	0x%x\n", hdr.DestinationCode)
	fmt.Printf("Old License Code:	0x%x\n", hdr.OldLicenseeCode)
	fmt.Printf("Mask ROM Version:	0x%x\n", hdr.MaskROMVersion)
	fmt.Printf("Header Checksum:	0x%x\n", hdr.HeaderChecksum)
	fmt.Printf("Global Checksum:	0x%x\n", hdr.GlobalChecksum)
}
