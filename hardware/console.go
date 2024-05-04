package hardware

import "github.com/maxfierke/gogo-gb/cart"

type Console interface {
	LoadCartridge(r *cart.Reader) error
	Step() error
	Run() error
}
