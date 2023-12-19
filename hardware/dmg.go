package hardware

import (
	"errors"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/mem"
)

var (
	ErrCartridgeAlreadyLoaded = errors.New("cartridge already loaded")
)

type DMG struct {
	cpu       *cpu.CPU
	mmu       *mem.MMU
	cartridge *cart.Cartridge
}

func NewDMG() (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
	}

	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.ResetToBootROM() // TODO: Load an actual boot ROOM

	return &DMG{
		cpu: cpu,
		mmu: mmu,
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
