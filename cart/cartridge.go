package cart

import (
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
	mbc    mem.MemHandler
}

func NewCartridge() *Cartridge {
	return &Cartridge{}
}

func (c *Cartridge) DebugPrint() {
	if c.mbc != nil {
		c.Header.DebugPrint()
	}
}

func (c *Cartridge) LoadCartridge(r *Reader) error {
	if c.mbc != nil {
		return ErrCartridgeAlreadyLoaded
	}

	c.Header = r.Header

	rom := make([]byte, r.Header.RomSizeBytes())
	copy(rom, r.headerBuf[:])
	_, err := io.ReadFull(r, rom[HEADER_END+1:])
	if err != nil {
		return err
	}

	ram := make([]byte, r.Header.RamSizeBytes())

	switch r.Header.CartType {
	case CART_TYPE_MBC0:
		c.mbc = mbc.NewMBC0(rom)
	case CART_TYPE_MBC1, CART_TYPE_MBC1_RAM, CART_TYPE_MBC1_RAM_BAT:
		c.mbc = mbc.NewMBC1(rom, ram)
	case CART_TYPE_MBC5, CART_TYPE_MBC5_RAM, CART_TYPE_MBC5_RAM_BAT:
		c.mbc = mbc.NewMBC5(rom, ram)
	default:
		return fmt.Errorf("unsupported or unknown MBC type: %s", r.Header.CartTypeName())
	}

	return nil
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
