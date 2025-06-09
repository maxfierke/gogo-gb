package devices

import (
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/bits"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_BOOTROM_KEY0 = 0xFF4C
	REG_BOOTROM_EN   = 0xFF50

	REG_BOOTROM_KEY0_CPU_MODE_BIT = 2

	BOOTROM_SIZE_DMG = 0x100
	BOOTROM_SIZE_CGB = BOOTROM_SIZE_DMG + 0x800
)

type BootROM interface {
	mem.MemHandler
	AttachMemHandlers(mmu *mem.MMU)
	LoadROM(r io.Reader) error
}

type DMGBootROM struct {
	enabled        bool
	rom            [BOOTROM_SIZE_DMG]byte
	dmgModeEnabled bool
}

var _ BootROM = (*DMGBootROM)(nil)

func NewDMGBootROM() *DMGBootROM {
	return &DMGBootROM{}
}

func (br *DMGBootROM) AttachMemHandlers(mmu *mem.MMU) {
	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x00FF}, br) // DMG BootROM
	mmu.AddHandler(mem.MemRegion{Start: 0xFF50, End: 0xFF50}, br) // BootROM enable register
}

func (br *DMGBootROM) LoadROM(r io.Reader) error {
	if _, err := r.Read(br.rom[:]); err != nil {
		return fmt.Errorf("unable to load boot ROM: %w", err)
	}

	br.enabled = true

	return nil
}

func (br *DMGBootROM) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if br.enabled {
		return mem.ReadReplace(br.rom[addr])
	} else {
		return mem.ReadPassthrough()
	}
}

func (br *DMGBootROM) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_BOOTROM_EN && br.enabled {
		br.enabled = value == 0x00
		return mem.WriteBlock()
	} else if br.enabled {
		panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is not allowed for boot ROM", value, addr))
	} else {
		return mem.WritePassthrough()
	}
}

type CGBBootROM struct {
	enabled        bool
	rom            [BOOTROM_SIZE_CGB]byte
	dmgModeEnabled bool
}

func NewCGBBootROM() *CGBBootROM {
	return &CGBBootROM{}
}

func (br *CGBBootROM) AttachMemHandlers(mmu *mem.MMU) {
	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x00FF}, br) // BootROM (Part 1)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF50, End: 0xFF50}, br) // BootROM enable register
	mmu.AddHandler(mem.MemRegion{Start: 0xFF4C, End: 0xFF4C}, br) // KEY0 CPU Mode Select
	mmu.AddHandler(mem.MemRegion{Start: 0x0200, End: 0x08FF}, br) // BootROM (Part 2)
}

func (br *CGBBootROM) LoadROM(r io.Reader) error {
	if _, err := r.Read(br.rom[:]); err != nil {
		return fmt.Errorf("unable to load boot ROM: %w", err)
	}

	br.enabled = true

	return nil
}

func (br *CGBBootROM) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_BOOTROM_KEY0 {
		var value byte

		if br.dmgModeEnabled {
			value |= 1 << REG_BOOTROM_KEY0_CPU_MODE_BIT
		}

		return mem.ReadReplace(value)
	} else if br.enabled {
		return mem.ReadReplace(br.rom[addr])
	} else {
		return mem.ReadPassthrough()
	}
}

func (br *CGBBootROM) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_BOOTROM_EN && br.enabled {
		br.enabled = value == 0x00
		return mem.WriteBlock()
	} else if addr == REG_BOOTROM_KEY0 && br.enabled {
		br.dmgModeEnabled = bits.Read(value, REG_BOOTROM_KEY0_CPU_MODE_BIT) == 1
		return mem.WriteBlock()
	} else if br.enabled {
		panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is not allowed for boot ROM", value, addr))
	} else {
		return mem.WritePassthrough()
	}
}
