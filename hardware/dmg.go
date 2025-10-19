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
	"github.com/maxfierke/gogo-gb/ppu"
	"github.com/maxfierke/gogo-gb/ppu/rendering"
)

const (
	DMG_RAM_SIZE = 0x10000

	dmgCyclesPerFrame uint = 70224
)

type DMG struct {
	// Components
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
	dma       *ppu.DMA
	ic        *devices.InterruptController
	joypad    *devices.Joypad
	ppu       *ppu.PPU
	serial    *devices.SerialPort
	timer     *devices.Timer

	// Non-components
	debugger        debug.Debugger
	debuggerHandler mem.MemHandlerHandle
}

var _ Console = (*DMG)(nil)

func NewDMG(opts ...ConsoleOption) (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, fmt.Errorf("constructing CPU: %w", err)
	}

	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	echo := mem.NewEchoRegion()
	unmapped := mem.NewUnmappedRegion()

	ic := devices.NewInterruptController()

	dmg := &DMG{
		cpu:       cpu,
		mmu:       mmu,
		cartridge: cart.NewCartridge(),
		debugger:  debug.NewNullDebugger(),
		dma:       ppu.NewDMA(),
		ic:        ic,
		joypad:    devices.NewJoypad(ic),
		ppu:       ppu.NewPPU(ic, rendering.Scanline),
		serial:    devices.NewSerialPort(),
		timer:     devices.NewTimer(),
	}

	for _, opt := range opts {
		err = opt(dmg, mmu)
		if err != nil {
			return nil, err
		}
	}

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x7FFF}, dmg.cartridge) // MBCs ROM Banks
	mmu.AddHandler(mem.MemRegion{Start: 0xA000, End: 0xBFFF}, dmg.cartridge) // MBCs RAM Banks

	mmu.AddHandler(mem.MemRegion{Start: 0xE000, End: 0xFDFF}, echo)     // Echo RAM (mirrors WRAM)
	mmu.AddHandler(mem.MemRegion{Start: 0xFEA0, End: 0xFEFF}, unmapped) // Nop writes, zero reads

	mmu.AddHandler(mem.MemRegion{Start: 0xFF00, End: 0xFF00}, dmg.joypad) // Joypad
	mmu.AddHandler(mem.MemRegion{Start: 0xFF01, End: 0xFF02}, dmg.serial) // Serial Port (Control & Data)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF04, End: 0xFF07}, dmg.timer)  // Timer (not RTC)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF40, End: 0xFF41}, dmg.ppu)    // LCD status, control registers
	mmu.AddHandler(mem.MemRegion{Start: 0xFF42, End: 0xFF45}, dmg.ppu)    // PPU registers
	mmu.AddHandler(mem.MemRegion{Start: 0xFF46, End: 0xFF46}, dmg.dma)    // DMA
	mmu.AddHandler(mem.MemRegion{Start: 0xFF47, End: 0xFF4B}, dmg.ppu)    // PPU registers

	mmu.AddHandler(mem.MemRegion{Start: 0x8000, End: 0x9FFF}, dmg.ppu) // VRAM tiles
	mmu.AddHandler(mem.MemRegion{Start: 0xFE00, End: 0xFE9F}, dmg.ppu) // OAM

	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, dmg.ic)   // Interrupts Requested
	mmu.AddHandler(mem.MemRegion{Start: 0xFF4D, End: 0xFF77}, unmapped) // CGB regs
	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, dmg.ic)   // Interrupts Enabled

	return dmg, nil
}

func (dmg *DMG) AttachCable(cable devices.SerialCable) {
	dmg.serial.AttachCable(cable)
}

func (dmg *DMG) AttachDebugger(debugger debug.Debugger) {
	dmg.detachDebugger()

	dmg.debuggerHandler = dmg.mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0xFFFF}, debugger)
	dmg.debugger = debugger
}

func (dmg *DMG) detachDebugger() {
	// Remove any existing handlers
	dmg.mmu.RemoveHandler(dmg.debuggerHandler)
	dmg.debugger = debug.NewNullDebugger()
}

func (dmg *DMG) SetupDebugger() {
	dmg.debugger.Setup(dmg.cpu, dmg.mmu, dmg.cartridge)
}

func (dmg *DMG) Debugger() debug.Debugger {
	return dmg.debugger
}

func (dmg *DMG) CartridgeHeader() cart.Header {
	if dmg.cartridge == nil {
		return cart.Header{}
	}

	return dmg.cartridge.Header
}

func (dmg *DMG) CyclesPerFrame() uint {
	return dmgCyclesPerFrame
}

func (dmg *DMG) LoadCartridge(r io.Reader) error {
	cartReader, err := cart.NewReader(r)
	if err != nil && !errors.Is(err, cart.ErrChecksum) {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	err = dmg.cartridge.LoadCartridge(cartReader)
	if err != nil {
		return fmt.Errorf("loading cartridge: %w", err)
	}

	return nil
}

func (dmg *DMG) Draw() image.Image {
	return dmg.ppu.Draw()
}

func (dmg *DMG) LoadSave(r io.Reader) error {
	err := dmg.cartridge.LoadSave(r)
	if err != nil {
		return fmt.Errorf("loading save: %w", err)
	}

	return nil
}

func (dmg *DMG) Save(w io.Writer) error {
	err := dmg.cartridge.Save(w)
	if err != nil {
		return fmt.Errorf("writing save: %w", err)
	}

	return nil
}

func (dmg *DMG) ReceiveInputs(inputs devices.JoypadInputs) {
	dmg.joypad.ReceiveInputs(inputs)
}

func (dmg *DMG) Step() (uint8, error) {
	dmg.debugger.OnDecode(dmg.cpu, dmg.mmu)

	var cycles uint8

	haltedPriorToExecute := dmg.cpu.IsHalted()

	cycles, err := dmg.cpu.Step(dmg.mmu)
	if err != nil {
		return 0, fmt.Errorf("unexpected error while executing instruction: %w", err)
	}

	// We're checking the halted state from _before_ the current instruction was
	// executed, because we want to trigger the OnExecute for the HALT instruction
	// itself, but not while halted, since the CPU isn't really executing during
	// this time.
	if !haltedPriorToExecute {
		dmg.debugger.OnExecute(dmg.cpu, dmg.mmu)
	}

	hasInterrupt, intCycles := dmg.cpu.PollInterrupts(dmg.mmu, dmg.ic)
	if hasInterrupt {
		cycles += intCycles
		dmg.debugger.OnInterrupt(dmg.cpu, dmg.mmu)
	}

	dmg.cartridge.Step(cycles)
	dmg.dma.Step(dmg.mmu, cycles)
	dmg.ppu.Step(dmg.mmu, cycles)
	dmg.timer.Step(cycles, dmg.ic)
	dmg.serial.Step(cycles, dmg.ic)

	return cycles, nil
}
