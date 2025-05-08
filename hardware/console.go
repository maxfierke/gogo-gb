package hardware

import (
	"io"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
)

type Console interface {
	AttachDebugger(debugger debug.Debugger)
	DetachDebugger()
	CartridgeHeader() cart.Header
	LoadCartridge(r io.Reader) error
	Save(w io.Writer) error
	LoadSave(r io.Reader) error
	Step() (uint8, error)
	Run(host devices.HostInterface) error
}
