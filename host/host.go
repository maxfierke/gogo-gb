package host

import (
	"log"

	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type Host interface {
	devices.HostInterface

	AttachSerialCable(serialCable devices.SerialCable)
	SetLogger(logger *log.Logger)
	Run(console hardware.Console) error
}
