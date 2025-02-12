package host

import (
	"image"
	"log"

	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type CLIHost struct {
	fbChan      chan image.Image
	inputChan   chan devices.JoypadInputs
	logger      *log.Logger
	exitedChan  chan bool
	serialCable devices.SerialCable
}

var _ Host = (*CLIHost)(nil)

func NewCLIHost() *CLIHost {
	return &CLIHost{
		fbChan:      make(chan image.Image, 3),
		inputChan:   make(chan devices.JoypadInputs),
		exitedChan:  make(chan bool),
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

func (h *CLIHost) Exited() <-chan bool {
	return h.exitedChan
}

func (h *CLIHost) SerialCable() devices.SerialCable {
	return h.serialCable
}

func (h *CLIHost) AttachSerialCable(serialCable devices.SerialCable) {
	h.serialCable = serialCable
}

func (h *CLIHost) Run(console hardware.Console) error {
	done := make(chan error)
	defer close(h.exitedChan)
	defer close(h.inputChan)

	// "Renderer"
	go func() {
		for range h.fbChan {
			// Consume frames
		}
	}()

	go func() {
		if err := console.Run(h); err != nil {
			h.LogErr("unexpected error occurred during runtime: %w", err)
			done <- err
			return
		}

		done <- nil
	}()

	return <-done
}
