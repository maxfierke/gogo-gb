package hardware

import (
	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
)

type Console interface {
	AttachDebugger(debugger debug.Debugger)
	DetachDebugger()
	LoadCartridge(r *cart.Reader) error
	Step() error
	Run(host devices.HostInterface) error
}
