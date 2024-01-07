package hardware

import (
	"io"
	"log"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

const DMG_RAM_SIZE = 0xFFFF + 1

type DMG struct {
	// Components
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
	ic        *devices.InterruptController
	lcd       *devices.LCD
	serial    *devices.SerialPort

	// Non-components
	debugger debug.Debugger
	logger   *log.Logger
}

func NewDMG() (*DMG, error) {
	debugger := debug.NewNullDebugger()
	return NewDMGDebug(debugger)
}

func NewDMGDebug(debugger debug.Debugger) (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
	}

	cartridge := cart.NewCartridge()
	ic := devices.NewInterruptController()
	lcd := devices.NewLCD()
	serial := devices.NewSerialPort()

	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0xFFFF}, debugger)

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x7FFF}, cartridge) // MBCs ROM Banks
	mmu.AddHandler(mem.MemRegion{Start: 0xA000, End: 0xBFFF}, cartridge) // MBCs RAM Banks

	mmu.AddHandler(mem.MemRegion{Start: 0xFF40, End: 0xFF4B}, lcd) // LCD control registers

	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, ic)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, ic)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF01, End: 0xFF02}, serial)

	return &DMG{
		cpu:       cpu,
		mmu:       mmu,
		cartridge: cartridge,
		debugger:  debugger,
		ic:        ic,
		lcd:       lcd,
		serial:    serial,
	}, nil
}

func (dmg *DMG) LoadCartridge(r *cart.Reader) error {
	return dmg.cartridge.LoadCartridge(r)
}

func (dmg *DMG) DebugPrint(logger *log.Logger) {
	dmg.cartridge.DebugPrint(logger)
}

func (dmg *DMG) Step() bool {
	dmg.debugger.OnDecode(dmg.cpu, dmg.mmu)

	cycles, err := dmg.cpu.Step(dmg.mmu)
	if err != nil {
		dmg.logger.Printf("Unexpected error while executing instruction: %v\n", err)
		return false
	}

	cycles += dmg.cpu.PollInterrupts(dmg.mmu, dmg.ic)

	dmg.serial.Step(uint(cycles), dmg.ic)

	return true
}

func (dmg *DMG) Run() {
	dmg.debugger.Setup(dmg.cpu, dmg.mmu)

	for dmg.Step() {
		dmg.debugger.OnExecute(dmg.cpu, dmg.mmu)
	}
}

func (dmg *DMG) SetLogger(logger *log.Logger) {
	dmg.logger = logger
}

func (dmg *DMG) SetSerialReader(serial io.Reader) {
	dmg.serial.SetReader(serial)
}

func (dmg *DMG) SetSerialWriter(serial io.Writer) {
	dmg.serial.SetWriter(serial)
}
