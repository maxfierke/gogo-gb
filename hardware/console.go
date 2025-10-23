package hardware

import (
	"errors"
	"fmt"
	"image"
	"io"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
	"github.com/maxfierke/gogo-gb/ppu"
)

type Console interface {
	AttachCable(cable devices.SerialCable)
	AttachDebugger(debugger debug.Debugger)
	SetupDebugger()
	Debugger() debug.Debugger
	Draw() image.Image
	CartridgeHeader() cart.Header
	CyclesPerFrame() uint
	LoadCartridge(r io.Reader) error
	Save(w io.Writer) error
	LoadSave(r io.Reader) error
	Step() (uint8, error)
	SetRenderer(renderer ppu.RendererConstructor)
	ReceiveInputs(inputs devices.JoypadInputs)
}

type ConsoleOption func(console Console, mmu *mem.MMU) error

const (
	ConsoleModelDMG ConsoleModel = "dmg"
	ConsoleModelMGB ConsoleModel = "mgb"
	ConsoleModelCGB ConsoleModel = "cgb"
)

type ConsoleModel string

func WithBootROM(r io.Reader) ConsoleOption {
	return func(console Console, mmu *mem.MMU) error {
		var bootROM devices.BootROM

		switch console.(type) {
		case *DMG:
			bootROM = devices.NewDMGBootROM()
		case *CGB:
			bootROM = devices.NewCGBBootROM()
		default:
			return fmt.Errorf("unrecognized console")
		}

		err := bootROM.LoadROM(r)
		if err != nil {
			return fmt.Errorf("loading boot ROM: %w", err)
		}
		bootROM.AttachMemHandlers(mmu)

		return nil
	}
}

func WithDebugger(debugger debug.Debugger) ConsoleOption {
	return func(console Console, mmu *mem.MMU) error {
		console.AttachDebugger(debugger)
		return nil
	}
}

func WithFakeBootROM() ConsoleOption {
	return func(console Console, mmu *mem.MMU) error {
		if dmg, isDMG := console.(*DMG); isDMG {
			dmg.cpu.ResetToBootROM()
			return nil
		}

		return errors.New("WithFakeBootROM is only supported for DMG")
	}
}

func WithRenderer(renderer ppu.RendererConstructor) ConsoleOption {
	return func(console Console, mmu *mem.MMU) error {
		console.SetRenderer(renderer)
		return nil
	}
}

func NewConsole(model ConsoleModel, opts ...ConsoleOption) (Console, error) {
	switch model {
	case ConsoleModelDMG:
		dmg, err := NewDMG(opts...)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize DMG: %w", err)
		}

		return dmg, nil
	case ConsoleModelCGB:
		cgb, err := NewCGB(opts...)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize CGB: %w", err)
		}

		return cgb, nil
	default:
		return nil, fmt.Errorf("unrecognized model: %s", model)
	}
}

func Run(console Console, host devices.HostInterface) error {
	framebuffer := host.Framebuffer()
	defer close(framebuffer)

	console.AttachCable(host.SerialCable())
	console.SetupDebugger()

	go func() {
		for inputs := range host.JoypadInput() {
			console.ReceiveInputs(inputs)
		}
	}()

	for range host.RequestFrame() {
		var frameCycles uint
		for frameCycles < console.CyclesPerFrame() {
			cycles, err := console.Step()
			if err != nil {
				return err
			}
			frameCycles += uint(cycles)
		}

		framebuffer <- console.Draw()
	}

	return nil
}
