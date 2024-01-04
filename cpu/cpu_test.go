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

const testRamSize = 0xFFFF + 1

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

func TestExecuteAdd8SignedTargetSP(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xE8, false)

	cpu.SP.Write(0xFFFE)
	cpu.PC.Write(0x100)

	mmu.Write8(0x101, 0xFC) // -4

	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.SP.Read(), 0xFFFA)
	assertFlags(t, cpu, false, false, true, true)

	cpu.SP.Write(0xFFFA)
	mmu.Write8(0x101, 0x4)

	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.SP.Read(), 0xFFFE)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8NonOverflowingTargetANoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x8F, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xE)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdc8NonOverflowingTargetACarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x8F, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xF)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdc8NonOverflowingTargetC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x89, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xA)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdc8NonOverflowingTargetCWithCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x89, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xB)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdc8TargetBCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x88, false)

	cpu.Reg.A.Write(0xFC)
	cpu.Reg.B.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, true, false, true, true)
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

	cpu.Reg.HL.Write(0x4C00)
	cpu.Reg.BC.Write(0x4C00)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x9800)
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

func TestExecuteSub8NonUnderflowingTargetA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x97, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, true, true, false, false)
}

func TestExecuteSub8NonUnderflowingTargetC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x91, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x4)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteSub8NonUnderflowingTargetCWithCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x91, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x4)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteSub8TargetBCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x90, false)

	cpu.Reg.A.Write(0x4)
	cpu.Reg.B.Write(0x9)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xFB)
	assertFlags(t, cpu, false, true, true, true)
}

func TestExecuteSbc8NonUnderflowingTargetA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x9F, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, true, true, false, false)
}

func TestExecuteSbc8NonUnderflowingTargetAWithCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x9F, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xFF)
	assertFlags(t, cpu, false, true, true, true)
}

func TestExecuteSbc8NonUnderflowingTargetC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x99, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x4)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteSbc8NonUnderflowingTargetCWithCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x99, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x3)
	assertFlags(t, cpu, false, true, false, false)
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
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.HL.Write(0xFFF8)
	mmu.Write8(0xFFF8, 0x03)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x34, false)

	cpu.Execute(mmu, inst)

	assertRegEquals(t, ram[0xFFF8], 0x04)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteDec8(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3D, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x6)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteDec8HalfCarry(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3D, false)

	cpu.Reg.A.Write(0x80)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7F)
	assertFlags(t, cpu, false, true, true, false)
}

func TestExecuteDec8Underflow(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3D, false)

	cpu.Reg.A.Write(0x0)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xFF)
	assertFlags(t, cpu, false, true, true, false)
}

func TestExecuteDec16ByteUnderflow(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x0B, false)

	cpu.Reg.BC.Write(0x0000)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.BC.Read(), 0xFFFF)
	assertRegEquals(t, cpu.Reg.B.Read(), 0xFF)
	assertRegEquals(t, cpu.Reg.C.Read(), 0xFF)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteDecHLIndirect(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.HL.Write(0xFFF8)
	mmu.Write8(0xFFF8, 0x04)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x35, false)

	cpu.Execute(mmu, inst)

	assertRegEquals(t, ram[0xFFF8], 0x03)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteAnd(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)
	cpu.Reg.B.Write(0x3)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xA0, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x03)
	assertFlags(t, cpu, false, false, true, false)
}

func TestExecuteAndZeros(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x20)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xA0, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x00)
	assertFlags(t, cpu, true, false, true, false)
}

func TestExecuteAndA(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xA7, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7)
	assertFlags(t, cpu, false, false, true, false)
}

func TestExecuteBit(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x40, true)

	cpu.Reg.B.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)
	assertFlags(t, cpu, false, false, true, false)

	cpu.Reg.B.Write(0b1011_0100)

	cpu.Execute(NULL_MMU, inst)
	assertFlags(t, cpu, true, false, true, false)
}

func TestExecuteCcf(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.F.Zero = true
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.HalfCarry = true
	cpu.Reg.F.Carry = true

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3F, false)

	cpu.Execute(NULL_MMU, inst)

	assertFlags(t, cpu, true, false, false, false)

	cpu.Execute(NULL_MMU, inst)

	assertFlags(t, cpu, true, false, false, true)
}

func TestExecuteCompareNonUnderflowTarget(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)
	cpu.Reg.B.Write(0x3)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB8, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteCompareNonOverflowTargetWithCarry(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)
	cpu.Reg.B.Write(0x3)
	cpu.Reg.F.Carry = true

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB8, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7)
	assertFlags(t, cpu, false, true, false, false)
}

func TestExecuteCompareCarry(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x4)
	cpu.Reg.B.Write(0x9)
	cpu.Reg.F.Carry = true

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB8, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x4)
	assertFlags(t, cpu, false, true, true, true)
}

func TestExecuteCompareNonUnderflowTargetA(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xBF, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7)
	assertFlags(t, cpu, true, true, false, false)
}

func TestExecuteCpl(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0b1011_0100)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x2F, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0100_1011)
	assertFlags(t, cpu, false, true, true, false)
}

func TestExecuteOr(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x20)
	cpu.Reg.B.Write(0x3)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB0, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x23)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteOrWithZero(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x20)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB0, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x20)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteOrWithZeros(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x0)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB0, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, true, false, false, false)
}

func TestExecuteOrA(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xB7, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x7)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteXor(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x21)
	cpu.Reg.B.Write(0x3)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xA8, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x22)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteXorZeros(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x08)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xA8, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x08)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteXorA(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.A.Write(0x7)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xAF, false)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, true, false, false, false)
}

func TestExecuteCall(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	mmu.Write8(0x102, 0x04)
	mmu.Write8(0x101, 0x89)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xCD, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x01)
	assertRegEquals(t, mmu.Read8(0xE), 0x03)
	assertRegEquals(t, cpu.SP.Read(), 0xE)
	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteCallNotZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)
	cpu.Reg.F.Zero = true

	mmu.Write8(0x102, 0x04)
	mmu.Write8(0x101, 0x89)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC4, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x103)

	cpu.Reg.F.Zero = false
	nextPc, _, _ = cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x01)
	assertRegEquals(t, mmu.Read8(0xE), 0x03)
	assertRegEquals(t, cpu.SP.Read(), 0xE)
	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteCallZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)
	cpu.Reg.F.Zero = false

	mmu.Write8(0x102, 0x04)
	mmu.Write8(0x101, 0x89)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xCC, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x103)

	cpu.Reg.F.Zero = true
	nextPc, _, _ = cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x01)
	assertRegEquals(t, mmu.Read8(0xE), 0x03)
	assertRegEquals(t, cpu.SP.Read(), 0xE)
	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteCallNotCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)
	cpu.Reg.F.Carry = true

	mmu.Write8(0x102, 0x04)
	mmu.Write8(0x101, 0x89)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD4, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x103)

	cpu.Reg.F.Carry = false
	nextPc, _, _ = cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x01)
	assertRegEquals(t, mmu.Read8(0xE), 0x03)
	assertRegEquals(t, cpu.SP.Read(), 0xE)
	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteCallCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)
	cpu.Reg.F.Carry = false

	mmu.Write8(0x102, 0x04)
	mmu.Write8(0x101, 0x89)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xDC, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x103)

	cpu.Reg.F.Carry = true
	nextPc, _, _ = cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x01)
	assertRegEquals(t, mmu.Read8(0xE), 0x03)
	assertRegEquals(t, cpu.SP.Read(), 0xE)
	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteJump(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC3, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Zero = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xCA, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xDA, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD2, false)

	expectedNextPC := uint16(0xFB)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xDA, false)

	expectedNextPC := uint16(0xFB)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpNoZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC2, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0xFC)
	mmu.Write8(0xFA, 0x02)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD2, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpHL(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.HL.Write(0x02FC)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xE9, false)

	expectedNextPC := uint16(0x02FC)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpRel(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x18, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)

	mmu.Write8(0xF9, 0xFC)
	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x18, false)

	expectedNextPC = uint16(0xF6)
	nextPC, _, _ = cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpRelZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Zero = true

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x28, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpRelCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x38, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpRelCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = true

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x38, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteNoJumpRelNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = false

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x30, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpRelNoZero(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Zero = false

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x20, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteJumpRelNoCarry(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)
	cpu.Reg.F.Carry = false

	mmu.Write8(0xF9, 0x04)
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x30, false)

	expectedNextPC := uint16(0xFE)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteLD8RegToReg(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.B.Write(0x01)
	cpu.Reg.C.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x41, false)
	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.B.Read(), 0x33)
}

func TestExecuteLD8ImmToReg(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	cpu.Reg.A.Write(0x01)
	mmu.Write8(0xF9, 0x04)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3E, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x04)
}

func TestExecuteLD8IndirectToReg(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	cpu.Reg.A.Write(0x01)
	mmu.Write16(0xF9, 0x0620)
	mmu.Write16(0x0620, 0x04)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xFA, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x04)
}

func TestExecuteLD8HLIndirectToReg(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.B.Write(0x01)
	cpu.Reg.HL.Write(0x0620)
	mmu.Write16(0x0620, 0x04)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x46, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.B.Read(), 0x04)
}

func TestExecuteLD8RegToIndirect(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.BC.Write(0x0620)
	mmu.Write16(0x0620, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x02, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0x0620), 0x33)
}

func TestExecuteLD8RegToHLIndirectInc(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.HL.Write(0x0620)
	mmu.Write16(0x0620, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x22, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0x0620), 0x33)
	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0621)
}

func TestExecuteLD8RegToHLIndirectDec(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.HL.Write(0x0620)
	mmu.Write16(0x0620, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x32, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0x0620), 0x33)
	assertRegEquals(t, cpu.Reg.HL.Read(), 0x061F)
}

func TestExecuteLD8RegToIndirectImm(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write16(0xF9, 0x0620)
	mmu.Write16(0x0620, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xEA, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0x0620), 0x33)
}

func TestExecuteLD8RegToIndirectImmH(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0x20)
	mmu.Write8(0xFF20, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xE0, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xFF20), 0x33)
}

func TestExecuteLD8RegToIndirectC(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	cpu.Reg.C.Write(0x20)
	mmu.Write8(0xFF20, 0x04)
	cpu.Reg.A.Write(0x33)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xE2, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xFF20), 0x33)
}

func TestExecuteLD8ImmToRegH(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write8(0xF9, 0x04)
	mmu.Write8(0xFF04, 0x33)
	cpu.Reg.A.Write(0x01)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF0, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x33)
}

func TestExecuteLD8IndirectToRegH(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	cpu.Reg.A.Write(0x01)
	cpu.Reg.C.Write(0x20)
	mmu.Write16(0xFF20, 0x04)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF2, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x04)
}

func TestExecuteLD16RegToReg(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.SP.Write(0x0102)
	cpu.Reg.HL.Write(0x3322)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF9, false)
	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.SP.Read(), 0x3322)
}

func TestExecuteLD16ImmToReg(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	cpu.Reg.BC.Write(0x0201)
	mmu.Write16(0xF9, 0x0413)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x01, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.BC.Read(), 0x0413)
}

func TestExecuteLD16RegToIndirectImm(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0xF8)

	mmu.Write16(0xF9, 0x0620)
	mmu.Write16(0x0620, 0x0402)
	cpu.SP.Write(0x3322)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x08, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read16(0x0620), 0x3322)
}

func TestExecuteLD16SPPlusSignedImmToHL(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)
	cpu.PC.Write(0x100)

	cpu.SP.Write(0x0201)
	mmu.Write8(0x101, 0x2)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF8, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0203)

	mmu.Write8(0x101, 0xFD) // -2

	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF8, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x01FE)
}

func TestExecutePushPop(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
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
	ram := make([]byte, testRamSize)
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
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x0489)
}

func TestExecuteRetFail(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
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
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x101)
}

func TestExecuteRst(t *testing.T) {
	cpu, _ := NewCPU()
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xFF, false)

	expectedNextPC := uint16(0x38)
	nextPC, _, _ := cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read16(0x0E), 0x101)
	assertRegEquals(t, cpu.SP.Read(), 0x0E)
	assertNextPC(t, nextPC, expectedNextPC)
}

func TestExecuteDi(t *testing.T) {
	cpu, _ := NewCPU()
	cpu.ime = true

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xF3, false)

	cpu.Execute(NULL_MMU, inst)

	if cpu.ime {
		t.Errorf("Expected IME flag to be disabled, but it was enabled")
	}
}

func TestExecuteEi(t *testing.T) {
	cpu, _ := NewCPU()
	cpu.ime = false

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xFB, false)

	cpu.Execute(NULL_MMU, inst)

	if !cpu.ime {
		t.Errorf("Expected IME flag to be enabled, but it was disabled")
	}
}

func TestExecuteHalt(t *testing.T) {
	cpu, _ := NewCPU()
	cpu.halted = false

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x76, false)

	cpu.Execute(NULL_MMU, inst)

	if !cpu.halted {
		t.Errorf("Expected CPU to be halted")
	}
}

func TestExecuteReset(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x87, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b1011_0100)
}

func TestExecuteRetI(t *testing.T) {
	cpu, _ := NewCPU()
	cpu.ime = false
	ram := make([]byte, testRamSize)
	mmu := mem.NewMMU(ram)

	cpu.Reg.BC.Write(0x0489)
	cpu.PC.Write(0x100)
	cpu.SP.Write(0x10)

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC5, false)
	cpu.Execute(mmu, inst)

	assertRegEquals(t, mmu.Read8(0xF), 0x4)
	assertRegEquals(t, mmu.Read8(0xE), 0x89)
	assertRegEquals(t, cpu.SP.Read(), 0xE)

	inst, _ = cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xD9, false)
	nextPc, _, _ := cpu.Execute(mmu, inst)

	assertNextPC(t, nextPc, 0x0489)

	if !cpu.ime {
		t.Errorf("Expected IME flag to be enabled, but it was disabled")
	}
}

func TestExecuteRL(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x17, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0110_1010)
	assertFlags(t, cpu, false, false, false, true)

	cpu.Reg.A.Write(0b1011_0101)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0110_1011)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRLA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x17, false)

	cpu.Reg.A.Write(0x80)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRLC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x07, true)

	cpu.Reg.A.Write(0x80)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x1)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRLCA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x07, false)

	cpu.Reg.A.Write(0x80)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x1)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRR(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x1F, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0101_1010)
	assertFlags(t, cpu, false, false, false, true)

	cpu.Reg.A.Write(0b1011_0101)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b1101_1010)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRRA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x1F, false)

	cpu.Reg.A.Write(0x1)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x0)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRRC(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x0F, true)

	cpu.Reg.A.Write(0x1)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x80)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteRRCA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x0F, false)

	cpu.Reg.A.Write(0x1)
	cpu.Reg.F.Carry = true

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x80)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteScf(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.F.Zero = true
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.HalfCarry = true
	cpu.Reg.F.Carry = false

	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x37, false)

	cpu.Execute(NULL_MMU, inst)

	assertFlags(t, cpu, true, false, false, true)
}

func TestExecuteSet(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0xC7, true)

	cpu.Reg.A.Write(0b1011_0100)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0110_1011)
}

func TestExecuteSLA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x27, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0110_1010)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteSRA(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x2F, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b1101_1010)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteSRL(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x3F, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0101_1010)
	assertFlags(t, cpu, false, false, false, true)
}

func TestExecuteSwap(t *testing.T) {
	cpu, _ := NewCPU()
	inst, _ := cpu.opcodes.InstructionFromByte(cpu.PC.Read(), 0x37, true)

	cpu.Reg.A.Write(0b1011_0101)

	cpu.Execute(NULL_MMU, inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0b0101_1011)
	assertFlags(t, cpu, false, false, false, false)
}

func TestCPUReset(t *testing.T) {
	cpu, _ := NewCPU()

	cpu.Reg.AF.Write(0x1234)
	cpu.Reg.BC.Write(0x5678)
	cpu.Reg.DE.Write(0x9ABC)
	cpu.Reg.HL.Write(0xDEF0)
	cpu.PC.Write(0x1010)
	cpu.SP.Write(0x0202)
	cpu.ime = false
	cpu.halted = true

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

	if !cpu.ime {
		t.Errorf("Expected IME flag to be enabled, but it was disabled")
	}

	if cpu.halted {
		t.Errorf("Expected not to be halted")
	}
}
