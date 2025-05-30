package host

import (
	"errors"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

const (
	FB_HEIGHT = 144
	FB_WIDTH  = 160
)

type UI struct {
	fbChan      chan image.Image
	frameChan   chan struct{}
	inputChan   chan devices.JoypadInputs
	logger      *log.Logger
	serialCable devices.SerialCable

	framebufferImage *ebiten.Image
}

var (
	_ Host        = (*UI)(nil)
	_ ebiten.Game = (*UI)(nil)
)

func NewUIHost() *UI {
	return &UI{
		fbChan:      make(chan image.Image),
		frameChan:   make(chan struct{}),
		inputChan:   make(chan devices.JoypadInputs),
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

func (ui *UI) RequestFrame() <-chan struct{} {
	return ui.frameChan
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

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		inputs.Up = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		inputs.Down = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		inputs.Left = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		inputs.Right = true
	}

	ui.inputChan <- inputs

	requestFrame := struct{}{}
	select {
	case ui.frameChan <- requestFrame:
		// Requested frame
	default:
		// Frame dropped?
	}

	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) {
	select {
	case fbImage := <-ui.fbChan:
		if ui.framebufferImage == nil {
			ui.framebufferImage = ebiten.NewImageFromImage(fbImage)
		} else {
			for x := range FB_WIDTH {
				for y := range FB_HEIGHT {
					ui.framebufferImage.Set(x, y, fbImage.At(x, y))
				}
			}
		}
		screen.DrawImage(ui.framebufferImage, &ebiten.DrawImageOptions{})
	default:
		// do nothing
	}
}

func (ui *UI) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return FB_WIDTH, FB_HEIGHT
}

func (ui *UI) Run(console hardware.Console) error {
	ebiten.SetWindowSize(480, 432)
	ebiten.SetWindowTitle("gogo-gb, the go-getting GB emulator")
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)

	// We only render full frames, so no need to waste resources clearing
	ebiten.SetScreenClearedEveryFrame(false)

	if console == nil {
		return errors.New("console cannot be nil")
	}

	go func() {
		ui.Log("starting console main loop")
		if err := console.Run(ui); err != nil {
			ui.LogErr("unexpected error occurred during runtime: %w", err)
			return
		}
	}()

	defer close(ui.inputChan)
	defer close(ui.frameChan)

	ui.Log("handing over main thread to ebiten for rendering")
	return ebiten.RunGame(ui)
}
