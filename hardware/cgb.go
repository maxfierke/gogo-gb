package hardware

import (
	"errors"
	"fmt"
	"image"
	"io"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	cgbCyclesPerFrame uint = dmgCyclesPerFrame * 2
)

type CGB struct {
	// Components
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
	dma       *devices.DMA
	ic        *devices.InterruptController
	joypad    *devices.Joypad
	ppu       *devices.PPU
	serial    *devices.SerialPort
	timer     *devices.Timer

	// Non-components
	debugger        debug.Debugger
	debuggerHandler mem.MemHandlerHandle
}

var _ Console = (*CGB)(nil)

func NewCGB(opts ...ConsoleOption) (*CGB, error) {
	cgbCpu, err := cpu.NewCPU()
	if err != nil {
		return nil, fmt.Errorf("constructing CPU: %w", err)
	}

	// Enable CGB CPU Features
	err = cgbCpu.EnableFeature(cpu.FeatureDoubleSpeed)
	if err != nil {
		return nil, fmt.Errorf("enabling double-speed CPU feature: %w", err)
	}

	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	echo := mem.NewEchoRegion()
	unmapped := mem.NewUnmappedRegion()

	ic := devices.NewInterruptController()

	cgb := &CGB{
		cpu:       cgbCpu,
		mmu:       mmu,
		cartridge: cart.NewCartridge(),
		debugger:  debug.NewNullDebugger(),
		dma:       devices.NewDMA(),
		ic:        ic,
		joypad:    devices.NewJoypad(ic),
		ppu:       devices.NewPPU(ic),
		serial:    devices.NewSerialPort(),
		timer:     devices.NewTimer(),
	}

	for _, opt := range opts {
		err = opt(cgb, mmu)
		if err != nil {
			return nil, err
		}
	}

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x7FFF}, cgb.cartridge) // MBCs ROM Banks
	mmu.AddHandler(mem.MemRegion{Start: 0xA000, End: 0xBFFF}, cgb.cartridge) // MBCs RAM Banks

	mmu.AddHandler(mem.MemRegion{Start: 0xE000, End: 0xFDFF}, echo)     // Echo RAM (mirrors WRAM)
	mmu.AddHandler(mem.MemRegion{Start: 0xFEA0, End: 0xFEFF}, unmapped) // Nop writes, zero reads

	mmu.AddHandler(mem.MemRegion{Start: 0xFF00, End: 0xFF00}, cgb.joypad) // Joypad
	mmu.AddHandler(mem.MemRegion{Start: 0xFF01, End: 0xFF02}, cgb.serial) // Serial Port (Control & Data)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF04, End: 0xFF07}, cgb.timer)  // Timer (not RTC)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF40, End: 0xFF41}, cgb.ppu)    // LCD status, control registers
	mmu.AddHandler(mem.MemRegion{Start: 0xFF42, End: 0xFF45}, cgb.ppu)    // PPU registers
	mmu.AddHandler(mem.MemRegion{Start: 0xFF46, End: 0xFF46}, cgb.dma)    // DMA
	mmu.AddHandler(mem.MemRegion{Start: 0xFF47, End: 0xFF4B}, cgb.ppu)    // PPU registers

	mmu.AddHandler(mem.MemRegion{Start: 0xFF4F, End: 0xFF4F}, unmapped) // VRAM Bank Select
	mmu.AddHandler(mem.MemRegion{Start: 0xFF51, End: 0xFF55}, unmapped) // VRAM DMA

	mmu.AddHandler(mem.MemRegion{Start: 0xFF56, End: 0xFF56}, unmapped) // IR Port

	mmu.AddHandler(mem.MemRegion{Start: 0xFF68, End: 0xFF6B}, unmapped) // BG/OBJ Palettes
	mmu.AddHandler(mem.MemRegion{Start: 0xFF6C, End: 0xFF6C}, unmapped) // OBJ Priority Mode
	mmu.AddHandler(mem.MemRegion{Start: 0xFF70, End: 0xFF70}, unmapped) // WRAM Bank Select

	mmu.AddHandler(mem.MemRegion{Start: 0xFF72, End: 0xFF73}, unmapped) // Unknown, should be R/W on CGB
	mmu.AddHandler(mem.MemRegion{Start: 0xFF74, End: 0xFF74}, unmapped) // Unknown, should be R/W on CGB
	mmu.AddHandler(mem.MemRegion{Start: 0xFF75, End: 0xFF75}, unmapped) // Unknown, bits 4-6 should be R/W on CGB

	mmu.AddHandler(mem.MemRegion{Start: 0xFF4D, End: 0xFF4D}, cgb.cpu) // CPU Speed Switch

	mmu.AddHandler(mem.MemRegion{Start: 0x8000, End: 0x9FFF}, cgb.ppu) // VRAM tiles
	mmu.AddHandler(mem.MemRegion{Start: 0xFE00, End: 0xFE9F}, cgb.ppu) // OAM

	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, cgb.ic) // Interrupts Requested
	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, cgb.ic) // Interrupts Enabled

	return cgb, nil
}

func (cgb *CGB) AttachCable(cable devices.SerialCable) {
	cgb.serial.AttachCable(cable)
}

func (cgb *CGB) AttachDebugger(debugger debug.Debugger) {
	cgb.detachDebugger()

	cgb.debuggerHandler = cgb.mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0xFFFF}, debugger)
	cgb.debugger = debugger
}

func (cgb *CGB) detachDebugger() {
	// Remove any existing handlers
	cgb.mmu.RemoveHandler(cgb.debuggerHandler)
	cgb.debugger = debug.NewNullDebugger()
}

func (cgb *CGB) SetupDebugger() {
	cgb.debugger.Setup(cgb.cpu, cgb.mmu, cgb.cartridge)
}

func (cgb *CGB) Debugger() debug.Debugger {
	return cgb.debugger
}

func (cgb *CGB) CartridgeHeader() cart.Header {
	if cgb.cartridge == nil {
		return cart.Header{}
	}

	return cgb.cartridge.Header
}

func (cgb *CGB) CyclesPerFrame() uint {
	if cgb.cpu.IsDoubleSpeed() {
		return cgbCyclesPerFrame
	}

	return dmgCyclesPerFrame
}

func (cgb *CGB) LoadCartridge(r io.Reader) error {
	cartReader, err := cart.NewReader(r)
	if err != nil && !errors.Is(err, cart.ErrChecksum) {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	err = cgb.cartridge.LoadCartridge(cartReader)
	if err != nil {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	return nil
}

func (cgb *CGB) Draw() image.Image {
	return cgb.ppu.Draw()
}

func (cgb *CGB) LoadSave(r io.Reader) error {
	err := cgb.cartridge.LoadSave(r)
	if err != nil {
		return fmt.Errorf("loading save: %w", err)
	}

	return nil
}

func (cgb *CGB) Save(w io.Writer) error {
	err := cgb.cartridge.Save(w)
	if err != nil {
		return fmt.Errorf("writing save: %w", err)
	}

	return nil
}

func (cgb *CGB) ReceiveInputs(inputs devices.JoypadInputs) {
	cgb.joypad.ReceiveInputs(inputs)
}

func (cgb *CGB) Step() (uint8, error) {
	cgb.debugger.OnDecode(cgb.cpu, cgb.mmu)

	var cycles uint8

	haltedPriorToExecute := cgb.cpu.IsHalted()

	cycles, err := cgb.cpu.Step(cgb.mmu)
	if err != nil {
		return 0, fmt.Errorf("unexpected error while executing instruction: %w", err)
	}

	// We're checking the halted state from _before_ the current instruction was
	// executed, because we want to trigger the OnExecute for the HALT instruction
	// itself, but not while halted, since the CPU isn't really executing during
	// this time.
	if !haltedPriorToExecute {
		cgb.debugger.OnExecute(cgb.cpu, cgb.mmu)
	}

	hasInterrupt, intCycles := cgb.cpu.PollInterrupts(cgb.mmu, cgb.ic)
	if hasInterrupt {
		cycles += intCycles
		cgb.debugger.OnInterrupt(cgb.cpu, cgb.mmu)
	}

	cgb.cartridge.Step(cycles)
	cgb.dma.Step(cgb.mmu, cycles)
	cgb.ppu.Step(cycles)
	cgb.timer.Step(cycles, cgb.ic)
	cgb.serial.Step(cycles, cgb.ic)

	return cycles, nil
}
