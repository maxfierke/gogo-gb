package devices

import (
	"log"
)

type HostContext interface {
	Logger() *log.Logger
	Log(msg string, args ...any)
	LogErr(msg string, args ...any)
	LogWarn(msg string, args ...any)
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

func (h *Host) Log(msg string, args ...any) {
	h.logger.Printf(msg+"\n", args...)
}

func (h *Host) LogErr(msg string, args ...any) {
	h.Log("ERROR: "+msg, args...)
}

func (h *Host) LogWarn(msg string, args ...any) {
	h.Log("WARN: "+msg, args...)
}

func (h *Host) SetLogger(logger *log.Logger) {
	h.logger = logger
}

func (h *Host) SerialCable() SerialCable {
	return h.serialCable
}

func (h *Host) AttachSerialCable(serialCable SerialCable) {
	h.serialCable = serialCable
}
