package host

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/hardware"
)

type UI struct {
	console     hardware.Console
	logger      *log.Logger
	serialCable devices.SerialCable
}

var _ Host = (*UI)(nil)

func NewUIHost() *UI {
	return &UI{
		logger:      log.Default(),
		serialCable: &devices.NullSerialCable{},
	}
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

func (ui *UI) SetConsole(console hardware.Console) {
	ui.console = console
}

func (ui *UI) Update() error {
	if ui.console != nil {
		return ui.console.Step()
	}

	return ebiten.Termination
}

func (ui *UI) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, 160, 144, color.RGBA{R: 127, G: 127, B: 127}, false)
	ebitenutil.DebugPrint(screen, "gogo-gb!!!")
}

func (ui *UI) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (ui *UI) Run() error {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("gogo-gb, the go-getting GB emulator")
	return ebiten.RunGame(ui)
}
