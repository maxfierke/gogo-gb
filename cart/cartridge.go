package cart

import (
	"errors"
	"fmt"

	"github.com/maxfierke/gogo-gb/cart/mbc"
	"github.com/maxfierke/gogo-gb/mem"
)

var (
	ErrUnsupportedMbc = errors.New("unsupported or unknown MBC type")
)

type Cartridge struct {
	Header Header
	mbc    mem.MemHandler
}

func NewCartridge(r *Reader) (*Cartridge, error) {
	cartridge := new(Cartridge)
	cartridge.Header = r.Header

	rom := make([]byte, r.Header.RomSizeBytes())
	copy(rom, r.headerBuf[:])

	switch r.Header.CartType {
	case CART_TYPE_MBC0:
		cartridge.mbc = mbc.NewMBC0(rom)
	default:
		return nil, fmt.Errorf("unsupported or unknown MBC type: %s", r.Header.CartTypeName())
	}

	return cartridge, nil
}

func (c *Cartridge) DebugPrint() {
	c.Header.DebugPrint()
}

func (c *Cartridge) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return c.mbc.OnRead(mmu, addr)
}

func (c *Cartridge) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return c.mbc.OnWrite(mmu, addr, value)
}
