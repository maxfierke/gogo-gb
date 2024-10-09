package hardware

import (
	"fmt"
	"io"
	"log"
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
	lcd       *devices.LCD
	serial    *devices.SerialPort
	timer     *devices.Timer

	// Non-components
	bootROM         *devices.BootROM
	debugger        debug.Debugger
	debuggerHandler mem.MemHandlerHandle
}

func NewDMG(opts ...DMGOption) (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
	}

	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)
	echo := mem.NewEchoRegion()
	unmapped := mem.NewUnmappedRegion()

	dmg := &DMG{
		cpu:       cpu,
		mmu:       mmu,
		cartridge: cart.NewCartridge(),
		debugger:  debug.NewNullDebugger(),
		ic:        devices.NewInterruptController(),
		lcd:       devices.NewLCD(),
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

	mmu.AddHandler(mem.MemRegion{Start: 0xFF01, End: 0xFF02}, dmg.serial) // Serial Port (Control & Data)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF04, End: 0xFF07}, dmg.timer)  // Timer (not RTC)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF40, End: 0xFF4B}, dmg.lcd)    // LCD control registers

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

func (dmg *DMG) LoadCartridge(r *cart.Reader) error {
	return dmg.cartridge.LoadCartridge(r)
}

func (dmg *DMG) DebugPrint(logger *log.Logger) {
	dmg.cartridge.DebugPrint(logger)
}

func (dmg *DMG) Step() error {
	dmg.debugger.OnDecode(dmg.cpu, dmg.mmu)

	cycles, err := dmg.cpu.Step(dmg.mmu)
	if err != nil {
		return fmt.Errorf("Unexpected error while executing instruction: %w", err)
	}

	cycles += dmg.cpu.PollInterrupts(dmg.mmu, dmg.ic)

	dmg.timer.Step(cycles, dmg.ic)
	dmg.serial.Step(cycles, dmg.ic)

	dmg.debugger.OnExecute(dmg.cpu, dmg.mmu)

	return nil
}

func (dmg *DMG) Run(host devices.HostInterface) error {
	framebuffer := host.Framebuffer()
	defer close(framebuffer)

	dmg.serial.AttachCable(host.SerialCable())
	dmg.debugger.Setup(dmg.cpu, dmg.mmu)

	hostExit := host.Exited()

	clockRate := time.NewTicker(time.Second / DMG_CPU_HZ)
	defer clockRate.Stop()

	fakeVBlank := time.NewTicker(time.Second / 60)
	defer fakeVBlank.Stop()

	for {
		select {
		case <-hostExit:
			return nil
		case <-fakeVBlank.C:
			framebuffer <- dmg.lcd.Draw()
		default:
			// Do nothing
		}

		if err := dmg.Step(); err != nil {
			return err
		}
		<-clockRate.C
	}
}
