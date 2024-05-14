package hardware

import (
	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/debug"
)

type Console interface {
	AttachDebugger(debugger debug.Debugger)
	DetachDebugger()
	LoadCartridge(r *cart.Reader) error
	Step() error
	Run() error
}
