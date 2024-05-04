package host

import (
	"log"

	"github.com/maxfierke/gogo-gb/devices"
)

type Host interface {
	devices.HostInterface

	AttachSerialCable(serialCable devices.SerialCable)
	SetLogger(logger *log.Logger)
}
