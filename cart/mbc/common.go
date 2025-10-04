package mbc

import (
	"io"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	RAM_BANK_SIZE = 0x2000
	ROM_BANK_SIZE = 0x4000
)

type MBC interface {
	mem.MemHandler

	Step(cycles uint8)
	DebugPrint(w io.Writer)
	Save(w io.Writer) error
	LoadSave(r io.Reader) error
}
