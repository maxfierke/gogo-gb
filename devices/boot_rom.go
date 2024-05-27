package devices

import (
	"fmt"
	"log"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_BOOTROM_EN = 0xFF50
)

type BootROM struct {
	enabled bool
	rom     []byte
}

func NewBootROM() *BootROM {
	return &BootROM{enabled: false, rom: []byte{}}
}

func (br *BootROM) LoadROM(rom []byte) {
	br.enabled = true
	br.rom = rom
}

func (br *BootROM) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if br.enabled {
		return mem.ReadReplace(br.rom[addr])
	} else {
		return mem.ReadPassthrough()
	}
}

func (br *BootROM) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_BOOTROM_EN && br.enabled {
		br.enabled = value == 0x00
		log.Printf("Unloaded boot ROM")
		return mem.WriteBlock()
	} else if br.enabled {
		panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is not allowed for boot ROM", value, addr))
	} else {
		return mem.WritePassthrough()
	}
}
