package cpu

import (
	"testing"

	"github.com/maxfierke/gogo-gb/mem"
)

func assertFlags(t *testing.T, cpu *CPU, zero bool, subtract bool, halfCarry bool, carry bool) {
	if cpu.Reg.F.Zero != zero {
		t.Errorf("Expected Z flag to be %t, but it was %t", zero, cpu.Reg.F.Zero)
	}

	if cpu.Reg.F.Subtract != subtract {
		t.Errorf("Expected N flag to be %t, but it was %t", subtract, cpu.Reg.F.Subtract)
	}

	if cpu.Reg.F.HalfCarry != halfCarry {
		t.Errorf("Expected HC flag to be %t, but it was %t", halfCarry, cpu.Reg.F.HalfCarry)
	}

	if cpu.Reg.F.Carry != carry {
		t.Errorf("Expected C flag to be %t, but it was %t", carry, cpu.Reg.F.Carry)
	}
}

func assertNextPC(t *testing.T, nextPC uint16, expectedNextPC uint16) {
	if nextPC != expectedNextPC {
		t.Errorf("Expected next PC value to be 0x%X, but got 0x%X", expectedNextPC, nextPC)
	}
}

var NULL_MMU = mem.NewMMU([]byte{})

func TestExecuteAdd8NonOverflowingTargetA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x87, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xE)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8NonOverflowingTargetC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x81, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xA)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8NonOverflowingTargetCWithCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x81, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xA)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8TargetBCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x80, false)

	cpu.Reg.A.Write(0xFC)
	cpu.Reg.B.Write(0x9)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x5)
	assertFlags(t, cpu, false, false, true, true)
}

func TestExecuteAdd16TargetHL(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x29, false)

	cpu.Reg.HL.Write(0x2331)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x4662)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd16TargetBCHalfCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x09, false)

	cpu.Reg.HL.Write(0x0300)
	cpu.Reg.BC.Write(0x0700)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0A00)
	assertFlags(t, cpu, false, false, true, false)
}

func TestExecuteAdd16TargetDECarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x19, false)

	cpu.Reg.HL.Write(0xF110)
	cpu.Reg.DE.Write(0x0FF0)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0100)
	assertFlags(t, cpu, false, false, true, true)
}

func TestExecuteInc8(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3C, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x8)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteInc8HalfCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3C, false)

	cpu.Reg.A.Write(0xF)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x10)
	assertFlags(t, cpu, false, false, true, false)
}

func TestExecuteInc8Overflow(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3C, false)

	cpu.Reg.A.Write(0xFF)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x00)
	assertFlags(t, cpu, true, false, true, false)
}

func TestExecuteInc16ByteOverflow(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x03, false)

	cpu.Reg.BC.Write(0xFF)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.BC.Read(), 0x100)
	assertRegEquals(t, cpu.Reg.B.Read(), 0x1)
	assertRegEquals(t, cpu.Reg.C.Read(), 0x0)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteInc16Overflow(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x03, false)

	cpu.Reg.BC.Write(0xFFFF)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.BC.Read(), 0x0)
	assertRegEquals(t, cpu.Reg.B.Read(), 0x0)
	assertRegEquals(t, cpu.Reg.C.Read(), 0x0)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteIncHLIndirect(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.Reg.HL.Write(0xFFF8)
	mmu.Write8(0xFFF8, 0x03)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x34, false)

	cpu.Execute(mmu, inst)

	assertRegEquals(t, ram[0xFFF8], 0x04)
}

func TestExecuteJump(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC3, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Zero = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xCA, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xDA, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD2, false)

	expectedNextPC := uint16(0xFB)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xDA, false)

	expectedNextPC := uint16(0xFB)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpNoZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC2, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD2, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpHL(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.HL.Write(0x02FC)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xE9, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecutePushPop(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.Reg.B.Write(0x4)
	cpu.Reg.C.Write(0x89)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC5, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x4)
	assertRegEquals(t, mmu.Read8(0xE), 0x89)
	assertRegEquals(t, cpu.SP.Read(), 0xE)

	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD1, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.D.Read(), 0x04)
	assertRegEquals(t, cpu.Reg.E.Read(), 0x89)
}

func TestExecuteRet(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.Reg.BC.Write(0x0489)
	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC5, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x4)
	assertRegEquals(t, mmu.Read8(0xE), 0x89)
	assertRegEquals(t, cpu.SP.Read(), 0xE)

	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC9, false)
	nextPc, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteRetFail(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.Reg.BC.Write(0x0489)
	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC5, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x4)
	assertRegEquals(t, mmu.Read8(0xE), 0x89)
	assertRegEquals(t, cpu.SP.Read(), 0xE)

	cpu.Reg.F.Carry = true

	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD0, false)
	nextPc, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x101)
}

func TestExecuteRst(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, 0xFFFF)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xFF, false)

	expectedNextPC := uint16(0x38)
	nextPC, _ := cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read16(0x0E), 0x101)
	assertRegEquals(t, cpu.SP.Read(), 0x0E)
	assertNextPC(t, nextPC, expectedNextPC)
}

func TestCPUReset(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.AF.Write(0x1234)
	cpu.Reg.BC.Write(0x5678)
	cpu.Reg.DE.Write(0x9ABC)
	cpu.Reg.HL.Write(0xDEF0)
	cpu.PC.Write(0x1010)
	cpu.SP.Write(0x0202)

	cpu.Reset()

	assertRegEquals(t, cpu.Reg.A.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.B.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.C.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.D.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.E.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.F.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.H.Read(), 0x00)
	assertRegEquals(t, cpu.Reg.L.Read(), 0x00)
	assertRegEquals(t, cpu.PC.Read(), 0x0000)
	assertRegEquals(t, cpu.SP.Read(), 0x0000)
}
