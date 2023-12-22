package hardware

import (
	"errors"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

var (
	ErrCartridgeAlreadyLoaded = errors.New("cartridge already loaded")
)

const DMGRamSize = 0xFFFF + 1

type DMG struct {
	cpu       *cpu.CPU
	mmu       *mem.MMU
	ic        *devices.InterruptController
	lcd       *devices.LCD
	cartridge *cart.Cartridge
}

func NewDMG() (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
	}

	ic := devices.NewInterruptController()
	lcd := devices.NewLCD()

	ram := make([]byte, DMGRamSize)
	mmu := mem.NewMMU(ram)

	mmu.AddHandler(mem.MemRegion{Start: 0xFFFF, End: 0xFFFF}, ic)
	mmu.AddHandler(mem.MemRegion{Start: 0xFF0F, End: 0xFF0F}, ic)

	cpu.ResetToBootROM() // TODO: Load an actual boot ROOM

	return &DMG{
		cpu: cpu,
		mmu: mmu,
		ic:  ic,
		lcd: lcd,
	}, nil
}

func (dmg *DMG) LoadCartridge(r *cart.Reader) error {
	if dmg.cartridge != nil {
		return ErrCartridgeAlreadyLoaded
	}

	cartridge, err := cart.NewCartridge(r)
	if err != nil {
		return err
	}

	dmg.cartridge = cartridge
	dmg.mmu.AddHandler(mem.MemRegion{Start: 0x0000, End: 0x7FFF}, dmg.cartridge) // MBCs ROM Banks
	dmg.mmu.AddHandler(mem.MemRegion{Start: 0x0A00, End: 0xBFFF}, dmg.cartridge) // MBCs RAM Banks

	return nil
}

func (dmg *DMG) DebugPrint() {
	if dmg.cartridge != nil {
		dmg.cartridge.DebugPrint()
	}
}

func (dmg *DMG) Step() bool {
	dmg.cpu.Step(dmg.mmu)

	return true
}

func (dmg *DMG) Run() {
	for dmg.Step() {

	}
}
