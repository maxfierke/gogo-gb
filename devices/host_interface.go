package devices

import "image"

type HostInterface interface {
	Exited() <-chan bool
	Framebuffer() chan<- image.Image
	JoypadInput() <-chan JoypadInputs
	Log(msg string, args ...any)
	LogErr(msg string, args ...any)
	LogWarn(msg string, args ...any)
	SerialCable() SerialCable
}
