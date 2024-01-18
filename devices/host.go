package devices

import (
	"log"
)

type HostContext interface {
	Logger() *log.Logger
	SerialCable() SerialCable
}

type Host struct {
	logger      *log.Logger
	serialCable SerialCable
}

func NewHost() *Host {
	return &Host{
		logger:      log.Default(),
		serialCable: &NullSerialCable{},
	}
}

func (h *Host) Logger() *log.Logger {
	return h.logger
}

func (h *Host) SetLogger(logger *log.Logger) {
	h.logger = logger
}

func (h *Host) SerialCable() SerialCable {
	return h.serialCable
}

func (h *Host) SetSerialCable(serialCable SerialCable) {
	h.serialCable = serialCable
}
