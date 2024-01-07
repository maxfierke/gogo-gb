package debug

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/devices"
	"github.com/maxfierke/gogo-gb/mem"
)

type GBDoctorDebugger struct{}

func NewGBDoctorDebugger() *GBDoctorDebugger {
	return &GBDoctorDebugger{}
}

func (gbd *GBDoctorDebugger) Setup(cpu *cpu.CPU, mmu *mem.MMU) {
	cpu.ResetToBootROM()
	gbd.printState(cpu, mmu)
}

func (gbd *GBDoctorDebugger) OnDecode(cpu *cpu.CPU, mmu *mem.MMU) {}

func (gbd *GBDoctorDebugger) OnExecute(cpu *cpu.CPU, mmu *mem.MMU) {
	gbd.printState(cpu, mmu)
}

func (gbd *GBDoctorDebugger) OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU) {}

func (gbd *GBDoctorDebugger) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == devices.REG_LCD_LY {
		return mem.ReadReplace(0x90) // gameboy-doctor needs a stubbed out LCD
	}

	return mem.ReadPassthrough()
}

func (gbd *GBDoctorDebugger) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WritePassthrough()
}

func (gbd *GBDoctorDebugger) printState(cpu *cpu.CPU, mmu *mem.MMU) {
	fmt.Printf(
		"A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X\n",
		cpu.Reg.A.Read(),
		cpu.Reg.F.Read(),
		cpu.Reg.B.Read(),
		cpu.Reg.C.Read(),
		cpu.Reg.D.Read(),
		cpu.Reg.E.Read(),
		cpu.Reg.H.Read(),
		cpu.Reg.L.Read(),
		cpu.SP.Read(),
		cpu.PC.Read(),
		mmu.Read8(cpu.PC.Read()),
		mmu.Read8(cpu.PC.Read()+1),
		mmu.Read8(cpu.PC.Read()+2),
		mmu.Read8(cpu.PC.Read()+3),
	)
}
