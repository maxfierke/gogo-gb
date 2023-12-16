package dmg

import (
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/mem"
)

type DMG struct {
	cpu *cpu.CPU
	mmu *mem.MMU
}

func NewDMG() (*DMG, error) {
	cpu, err := cpu.NewCPU()
	if err != nil {
		return nil, err
	}

	mmu := mem.NewMMU()

	return &DMG{
		cpu: cpu,
		mmu: mmu,
	}, nil
}
