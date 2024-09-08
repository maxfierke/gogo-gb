package debug

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/mem"
)

type Debugger interface {
	Setup(cpu *cpu.CPU, mmu *mem.MMU)
	OnDecode(cpu *cpu.CPU, mmu *mem.MMU)
	OnExecute(cpu *cpu.CPU, mmu *mem.MMU)
	OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU)
	OnRead(mmu *mem.MMU, addr uint16) mem.MemRead
	OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite
}

func NewDebugger(name string) (Debugger, error) {
	switch name {
	case "gameboy-doctor":
		return NewGBDoctorDebugger(), nil
	case "interactive":
		return NewInteractiveDebugger()
	case "none":
		return NewNullDebugger(), nil
	default:
		return nil, fmt.Errorf("unrecognized debugger: %v", name)
	}
}

type NullDebugger struct{}

func NewNullDebugger() *NullDebugger {
	return &NullDebugger{}
}

func (nd *NullDebugger) Setup(cpu *cpu.CPU, mmu *mem.MMU)       {}
func (nd *NullDebugger) OnDecode(cpu *cpu.CPU, mmu *mem.MMU)    {}
func (nd *NullDebugger) OnExecute(cpu *cpu.CPU, mmu *mem.MMU)   {}
func (nd *NullDebugger) OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU) {}

func (nd *NullDebugger) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return mem.ReadPassthrough()
}

func (nd *NullDebugger) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WritePassthrough()
}
