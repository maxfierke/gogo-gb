package host

import (
	"image"
	"log"
	"time"

	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type CLIHost struct {
	fbChan      chan image.Image
	frameChan   chan struct{}
	inputChan   chan devices.JoypadInputs
	logger      *log.Logger
	serialCable devices.SerialCable
}

var _ Host = (*CLIHost)(nil)

func NewCLIHost() *CLIHost {
	return &CLIHost{
		fbChan:      make(chan image.Image),
		frameChan:   make(chan struct{}),
		inputChan:   make(chan devices.JoypadInputs),
		logger:      log.Default(),
		serialCable: &devices.NullSerialCable{},
	}
}

func (h *CLIHost) Framebuffer() chan<- image.Image {
	return h.fbChan
}

func (h *CLIHost) JoypadInput() <-chan devices.JoypadInputs {
	return h.inputChan
}

func (h *CLIHost) RequestFrame() <-chan struct{} {
	return h.frameChan
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

func (h *CLIHost) Run(console hardware.Console) error {
	done := make(chan error)
	defer close(h.inputChan)
	defer close(h.frameChan)

	// "Renderer"
	go func() {
		ticker := time.NewTicker(time.Second / 60)

		for range ticker.C {
			h.frameChan <- struct{}{}

			// Consume frame
			<-h.fbChan
		}
	}()

	go func() {
		if err := hardware.Run(console, h); err != nil {
			h.LogErr("unexpected error occurred during runtime: %w", err)
			done <- err
			return
		}

		done <- nil
	}()

	return <-done
}
