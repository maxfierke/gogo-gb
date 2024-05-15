package devices

import "image"

type HostInterface interface {
	Framebuffer() chan<- image.Image
	Log(msg string, args ...any)
	LogErr(msg string, args ...any)
	LogWarn(msg string, args ...any)
	Running() <-chan bool
	SerialCable() SerialCable
}
