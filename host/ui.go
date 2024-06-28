package host

import (
	"errors"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type UI struct {
	fbChan      chan image.Image
	logger      *log.Logger
	exitedChan  chan bool
	serialCable devices.SerialCable

	framebufferImage *ebiten.Image
}

var _ Host = (*UI)(nil)

func NewUIHost() *UI {
	return &UI{
		fbChan:      make(chan image.Image, 3),
		exitedChan:  make(chan bool),
		logger:      log.Default(),
		serialCable: &devices.NullSerialCable{},
	}
}

func (ui *UI) Framebuffer() chan<- image.Image {
	return ui.fbChan
}

func (ui *UI) Log(msg string, args ...any) {
	ui.logger.Printf(msg+"\n", args...)
}

func (ui *UI) LogErr(msg string, args ...any) {
	ui.Log("ERROR: "+msg, args...)
}

func (ui *UI) LogWarn(msg string, args ...any) {
	ui.Log("WARN: "+msg, args...)
}

func (ui *UI) Exited() <-chan bool {
	return ui.exitedChan
}

func (ui *UI) SetLogger(logger *log.Logger) {
	ui.logger = logger
}

func (ui *UI) SerialCable() devices.SerialCable {
	return ui.serialCable
}

func (ui *UI) AttachSerialCable(serialCable devices.SerialCable) {
	ui.serialCable = serialCable
}

func (ui *UI) Update() error {
	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) {
	select {
	case fbImage := <-ui.fbChan:
		ui.framebufferImage = ebiten.NewImageFromImage(fbImage)
	default:
		// do nothing
	}

	if ui.framebufferImage != nil {
		screen.DrawImage(ui.framebufferImage, &ebiten.DrawImageOptions{})
	}
}

func (ui *UI) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (ui *UI) Run(console hardware.Console) error {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("gogo-gb, the go-getting GB emulator")

	if console == nil {
		return errors.New("console cannot be nil")
	}

	go func() {
		ui.Log("Starting console main loop")
		if err := console.Run(ui); err != nil {
			ui.LogErr("unexpected error occurred during runtime: %w", err)
			return
		}
	}()

	defer close(ui.exitedChan)

	ui.Log("Handing over to ebiten")
	return ebiten.RunGame(ui)
}
