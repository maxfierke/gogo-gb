package hardware

import (
	"fmt"
	"io"
	"time"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	DMG_BOOTROM_SIZE = 0x100
	DMG_CPU_HZ       = 4194304
	DMG_RAM_SIZE     = 0x10000
)

type DMGOption func(dmg *DMG) error

func WithBootROM(r io.Reader) DMGOption {
	return func(dmg *DMG) error {
		rom := make([]byte, DMG_BOOTROM_SIZE)
		if _, err := r.Read(rom); err != nil {
			return fmt.Errorf("unable to load boot ROM: %w", err)
		}

		dmg.bootROM = devices.NewBootROM()
		dmg.bootROM.LoadROM(rom)

		dmg.mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x00FF}, dmg.bootROM)
		dmg.mmu.AddHandler(mem.MemRegion{Start: 0xFF50, End: 0xFF50}, dmg.bootROM)

		return nil
	}
}

func WithDebugger(debugger debug.Debugger) DMGOption {
	return func(dmg *DMG) error {
		dmg.AttachDebugger(debugger)
		return nil
	}
}

func WithFakeBootROM() DMGOption {
	return func(dmg *DMG) error {
		dmg.cpu.ResetToBootROM()
		return nil
	}
}

type DMG struct {
	// Components
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
	ic        *devices.InterruptController
	joypad    *devices.Joypad
	ppu       *devices.PPU
	serial    *devices.SerialPort
	timer     *devices.Timer

	// Non-components
	bootROM         *devices.BootROM
	debugger        debug.Debugger
	debuggerHandler mem.MemHandlerHandle
}

var _ Console = (*DMG)(nil)

func NewDMG(opts ...DMGOption) (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
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
		ic:        ic,
		joypad:    devices.NewJoypad(ic),
		ppu:       devices.NewPPU(ic),
		serial:    devices.NewSerialPort(),
		timer:     devices.NewTimer(),
	}

	for _, opt := range opts {
		err = opt(dmg)
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
	mmu.AddHandler(mem.MemRegion{Start: 0xFF42, End: 0xFF4B}, dmg.ppu)    // PPU registers

	mmu.AddHandler(mem.MemRegion{Start: 0x8000, End: 0x9FFF}, dmg.ppu) // VRAM tiles
	mmu.AddHandler(mem.MemRegion{Start: 0xFE00, End: 0xFE9F}, dmg.ppu) // OAM

	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, dmg.ic)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF4D, End: 0xFF77}, unmapped) // CGB regs
	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, dmg.ic)

	return dmg, nil
}

func (dmg *DMG) AttachDebugger(debugger debug.Debugger) {
	dmg.DetachDebugger()

	dmg.debuggerHandler = dmg.mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0xFFFF}, debugger)
	dmg.debugger = debugger
}

func (dmg *DMG) DetachDebugger() {
	// Remove any existing handlers
	dmg.mmu.RemoveHandler(dmg.debuggerHandler)
	dmg.debugger = debug.NewNullDebugger()
}

func (dmg *DMG) CartridgeHeader() cart.Header {
	if dmg.cartridge == nil {
		return cart.Header{}
	}

	return dmg.cartridge.Header
}

func (dmg *DMG) LoadCartridge(r *cart.Reader) error {
	err := dmg.cartridge.LoadCartridge(r)
	if err != nil {
		return fmt.Errorf("dmg: loading cartridge: %w", err)
	}

	return nil
}

func (dmg *DMG) LoadSave(r io.Reader) error {
	err := dmg.cartridge.LoadSave(r)
	if err != nil {
		return fmt.Errorf("dmg: loading save: %w", err)
	}

	return nil
}

func (dmg *DMG) Save(w io.Writer) error {
	err := dmg.cartridge.Save(w)
	if err != nil {
		return fmt.Errorf("dmg: writing save: %w", err)
	}

	return nil
}

func (dmg *DMG) DebugPrint(w io.Writer) {
	dmg.cartridge.DebugPrint(w)
}

func (dmg *DMG) Step() error {
	dmg.debugger.OnDecode(dmg.cpu, dmg.mmu)

	var cycles uint8

	haltedPriorToExecute := dmg.cpu.IsHalted()

	cycles, err := dmg.cpu.Step(dmg.mmu)
	if err != nil {
		return fmt.Errorf("Unexpected error while executing instruction: %w", err)
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

	dmg.ppu.Step(cycles)
	dmg.timer.Step(cycles, dmg.ic)
	dmg.serial.Step(cycles, dmg.ic)

	return nil
}

func (dmg *DMG) Run(host devices.HostInterface) error {
	framebuffer := host.Framebuffer()
	defer close(framebuffer)

	dmg.serial.AttachCable(host.SerialCable())
	dmg.debugger.Setup(dmg.cpu, dmg.mmu)

	hostExit := host.Exited()

	go func() {
		for inputs := range host.JoypadInput() {
			dmg.joypad.ReceiveInputs(inputs)
		}
	}()

	cyclesPerFrame := DMG_CPU_HZ / 4 / 60
	ticker := time.NewTicker(time.Second / 60)

	for range ticker.C {
		for i := 0; i < cyclesPerFrame; i++ {
			if err := dmg.Step(); err != nil {
				return err
			}
		}

		framebuffer <- dmg.ppu.Draw()

		select {
		case <-hostExit:
			return nil
		default:
			// Do nothing
		}
	}

	return nil
}
