package hardware

import (
	"fmt"
	"image"
	"io"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

type Console interface {
	AttachCable(cable devices.SerialCable)
	AttachDebugger(debugger debug.Debugger)
	SetupDebugger()
	Debugger() debug.Debugger
	Draw() image.Image
	CartridgeHeader() cart.Header
	LoadCartridge(r io.Reader) error
	Save(w io.Writer) error
	LoadSave(r io.Reader) error
	Step() (uint8, error)
	ReceiveInputs(inputs devices.JoypadInputs)
}

type ConsoleOption func(console Console, mmu *mem.MMU) error

const BOOTROM_SIZE = 0x100

func WithBootROM(r io.Reader) ConsoleOption {
	return func(console Console, mmu *mem.MMU) error {
		rom := make([]byte, BOOTROM_SIZE)
		if _, err := r.Read(rom); err != nil {
			return fmt.Errorf("unable to load boot ROM: %w", err)
		}

		bootROM := devices.NewBootROM()
		bootROM.LoadROM(rom)

		mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x00FF}, bootROM)
		mmu.AddHandler(mem.MemRegion{Start: 0xFF50, End: 0xFF50}, bootROM)

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
		dmg := console.(*DMG) // TODO: Avoid the cast
		dmg.cpu.ResetToBootROM()
		return nil
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
		for frameCycles < cyclesPerFrame {
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
