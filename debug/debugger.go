package debug

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/mem"
	"github.com/maxfierke/gogo-gb/ppu"
)

var ErrSoftBreakExit = errors.New("soft-break exit")

type SoftBreakpointAction string

const (
	SoftBreakpointActionBreak      SoftBreakpointAction = "break"
	SoftBreakpointActionScreenshot SoftBreakpointAction = "screenshot"
	SoftBreakpointActionExit       SoftBreakpointAction = "exit"
)

type DebuggerOptions struct {
	// AttachOnBoot determines whether the debugger should attach itself before code starts executing
	AttachOnBoot bool

	// EnableSoftBreakpoints enables breaking on LD B, B
	EnableSoftBreakpoints bool

	// SoftBreakpointAction defines the action to take for software breakpoints
	SoftBreakpointAction SoftBreakpointAction

	// ScreenshotPath defines the path for screenshots
	ScreenshotPath string
}

func (opts *DebuggerOptions) onSoftBreak(ppu *ppu.PPU) error {
	switch opts.SoftBreakpointAction {
	case SoftBreakpointActionScreenshot:
		err := takeScreenshot(ppu.Draw(), opts.ScreenshotPath)
		if err != nil {
			return fmt.Errorf("taking screenshot: %w", err)
		}

		return ErrSoftBreakExit
	case SoftBreakpointActionExit:
		return ErrSoftBreakExit
	default:
		return nil
	}
}

type newDebuggerFunc func(*DebuggerOptions) Debugger

var debuggers = map[string]newDebuggerFunc{
	"gameboy-doctor": NewGBDoctorDebugger,
	"none":           NewNullDebugger,
	"":               NewNullDebugger,
}

type Debugger interface {
	mem.MemHandler

	Attach(cpu *cpu.CPU, mmu *mem.MMU)
	OnDecode(cpu *cpu.CPU, mmu *mem.MMU) error
	OnExecute(cpu *cpu.CPU, mmu *mem.MMU)
	OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU)
	Setup(cpu *cpu.CPU, mmu *mem.MMU, cart *cart.Cartridge, ppu *ppu.PPU)
}

func NewDebugger(name string, opts *DebuggerOptions) (Debugger, error) {
	if debugger, ok := debuggers[name]; ok {
		return debugger(opts), nil
	}

	return nil, fmt.Errorf("unrecognized debugger: %v", name)
}

type NullDebugger struct{}

func NewNullDebugger(*DebuggerOptions) Debugger {
	return &NullDebugger{}
}

func (nd *NullDebugger) Attach(cpu *cpu.CPU, mmu *mem.MMU) {}
func (nd *NullDebugger) OnDecode(cpu *cpu.CPU, mmu *mem.MMU) error {
	return nil
}
func (nd *NullDebugger) OnExecute(cpu *cpu.CPU, mmu *mem.MMU)   {}
func (nd *NullDebugger) OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU) {}

func (nd *NullDebugger) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return mem.ReadPassthrough()
}

func (nd *NullDebugger) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WritePassthrough()
}

func (nd *NullDebugger) Setup(cpu *cpu.CPU, mmu *mem.MMU, cart *cart.Cartridge, ppu *ppu.PPU) {}

func takeScreenshot(img image.Image, path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating screenshot file: %w", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("encoding png: %w", err)
	}

	return nil
}
