package devices

import "image"

type HostInterface interface {
	Framebuffer() chan<- image.Image
	JoypadInput() <-chan JoypadInputs
	RequestFrame() <-chan struct{}
	Log(msg string, args ...any)
	LogErr(msg string, args ...any)
	LogWarn(msg string, args ...any)
	SerialCable() SerialCable
}
