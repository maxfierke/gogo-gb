package host

import (
	"log"

	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type CLIHost struct {
	console     hardware.Console
	logger      *log.Logger
	serialCable devices.SerialCable
}

var _ Host = (*CLIHost)(nil)

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

func (h *CLIHost) SetConsole(console hardware.Console) {
	h.console = console
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

func (h *CLIHost) Run() error {
	return h.console.Run()
}
