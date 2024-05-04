package host

import (
	"log"

	"github.com/maxfierke/gogo-gb/devices"
)

type CLIHost struct {
	logger      *log.Logger
	serialCable devices.SerialCable
}

func NewCLIHost() *CLIHost {
	return &CLIHost{
		logger:      log.Default(),
		serialCable: &devices.NullSerialCable{},
	}
}

func (h *CLIHost) Log(msg string, args ...any) {
	h.logger.Printf(msg+"\n", args...)
}

func (h *CLIHost) LogErr(msg string, args ...any) {
	h.Log("ERROR: "+msg, args...)
}

func (h *CLIHost) LogWarn(msg string, args ...any) {
	h.Log("WARN: "+msg, args...)
}

func (h *CLIHost) SetLogger(logger *log.Logger) {
	h.logger = logger
}

func (h *CLIHost) SerialCable() devices.SerialCable {
	return h.serialCable
}

func (h *CLIHost) AttachSerialCable(serialCable devices.SerialCable) {
	h.serialCable = serialCable
}
