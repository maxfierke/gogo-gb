package cart

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/maxfierke/gogo-gb/cart/mbc"
	"github.com/maxfierke/gogo-gb/mem"
)

var (
	ErrCartridgeAlreadyLoaded = errors.New("cartridge already loaded")
	ErrUnsupportedMbc         = errors.New("unsupported or unknown MBC type")
)

type Cartridge struct {
	Header Header
	mbc    mbc.MBC
}

func NewCartridge() *Cartridge {
	return &Cartridge{}
}

func (c *Cartridge) DebugPrint(w io.Writer) {
	if c.mbc != nil {
		c.Header.DebugPrint(w)
	}
}

func (c *Cartridge) LoadCartridge(r *Reader) error {
	if c.mbc != nil {
		return ErrCartridgeAlreadyLoaded
	}

	c.Header = r.Header

	romSize := r.Header.RomSizeBytes()
	romBuffer := new(bytes.Buffer)
	romBuffer.Grow(int(romSize))
	headerBytes, err := romBuffer.Write(r.headerBuf[:])
	if err != nil {
		return fmt.Errorf("copying cartridge header: %w. read %d bytes", err, headerBytes)
	}

	n, err := romBuffer.ReadFrom(r)
	if err != nil {
		return fmt.Errorf("reading cartridge ROM: %w. read %d bytes", err, n)
	}

	rom := make([]byte, romSize)
	copy(rom, romBuffer.Bytes())

	ram := make([]byte, r.Header.RamSizeBytes())

	switch r.Header.CartType {
	case CART_TYPE_MBC0:
		c.mbc = mbc.NewMBC0(rom)
	case CART_TYPE_MBC1, CART_TYPE_MBC1_RAM, CART_TYPE_MBC1_RAM_BAT:
		c.mbc = mbc.NewMBC1(rom, ram)
	case CART_TYPE_MBC3, CART_TYPE_MBC3_RAM, CART_TYPE_MBC3_RAM_BAT:
		if r.Header.IsMBC30() {
			c.mbc = mbc.NewMBC30(rom, ram, false)
		} else {
			c.mbc = mbc.NewMBC3(rom, ram, false)
		}
	case CART_TYPE_MBC3_RTC_BAT, CART_TYPE_MBC3_RTC_RAM_BAT:
		if r.Header.IsMBC30() {
			c.mbc = mbc.NewMBC30(rom, ram, true)
		} else {
			c.mbc = mbc.NewMBC3(rom, ram, true)
		}
	case CART_TYPE_MBC5, CART_TYPE_MBC5_RAM, CART_TYPE_MBC5_RAM_BAT:
		c.mbc = mbc.NewMBC5(rom, ram)
	default:
		return fmt.Errorf("unsupported or unknown MBC type: %s", r.Header.CartTypeName())
	}

	return nil
}

func (c *Cartridge) Save(w io.Writer) error {
	return c.mbc.Save(w)
}

func (c *Cartridge) LoadSave(r io.Reader) error {
	return c.mbc.LoadSave(r)
}

func (c *Cartridge) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if c.mbc == nil {
		return mem.ReadPassthrough()
	}

	return c.mbc.OnRead(mmu, addr)
}

func (c *Cartridge) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if c.mbc == nil {
		return mem.WriteBlock()
	}

	return c.mbc.OnWrite(mmu, addr, value)
}
