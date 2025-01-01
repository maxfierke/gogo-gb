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
	inputChan   chan devices.JoypadInputs
	logger      *log.Logger
	exitedChan  chan bool
	serialCable devices.SerialCable

	framebufferImage *ebiten.Image
}

var _ Host = (*UI)(nil)

func NewUIHost() *UI {
	return &UI{
		fbChan:      make(chan image.Image, 3),
		inputChan:   make(chan devices.JoypadInputs),
		exitedChan:  make(chan bool),
		logger:      log.Default(),
		serialCable: &devices.NullSerialCable{},
	}
}

func (ui *UI) Framebuffer() chan<- image.Image {
	return ui.fbChan
}

func (ui *UI) JoypadInput() <-chan devices.JoypadInputs {
	return ui.inputChan
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
	var inputs devices.JoypadInputs

	if ebiten.IsKeyPressed(ebiten.KeyX) {
		inputs.A = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		inputs.B = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		inputs.Start = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyShiftRight) {
		inputs.Select = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		inputs.Up = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		inputs.Down = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		inputs.Left = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		inputs.Right = true
	}

	ui.inputChan <- inputs

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

	defer close(ui.inputChan)
	defer close(ui.exitedChan)

	ui.Log("Handing over to ebiten")
	return ebiten.RunGame(ui)
}
