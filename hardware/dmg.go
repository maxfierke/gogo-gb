package hardware

import (
	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/debug"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

const DMG_RAM_SIZE = 0xFFFF + 1

type DMG struct {
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
	debugger  debug.Debugger
	ic        *devices.InterruptController
	lcd       *devices.LCD
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

	ram := make([]byte, DMG_RAM_SIZE)
	mmu := mem.NewMMU(ram)

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0xFFFF}, debugger)

	mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x7FFF}, cartridge) // MBCs ROM Banks
	mmu.AddHandler(mem.MemRegion{Start: 0xA000, End: 0xBFFF}, cartridge) // MBCs RAM Banks

	mmu.AddHandler(mem.MemRegion{Start: 0xFF40, End: 0xFF4B}, lcd) // LCD control registers

	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, ic)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, ic)

	return &DMG{
		cpu:       cpu,
		mmu:       mmu,
		cartridge: cartridge,
		debugger:  debugger,
		ic:        ic,
		lcd:       lcd,
	}, nil
}

func (dmg *DMG) LoadCartridge(r *cart.Reader) error {
	return dmg.cartridge.LoadCartridge(r)
}

func (dmg *DMG) DebugPrint() {
	dmg.cartridge.DebugPrint()
}

func (dmg *DMG) Step() bool {
	dmg.debugger.OnDecode(dmg.cpu, dmg.mmu)

	dmg.cpu.Step(dmg.mmu)

	return true
}

func (dmg *DMG) Run() {
	dmg.debugger.Setup(dmg.cpu, dmg.mmu)

	for dmg.Step() {
		dmg.debugger.OnExecute(dmg.cpu, dmg.mmu)
	}
}
