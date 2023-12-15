package cpu

import (
	"testing"
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

func TestExecuteAdd8NonOverflowingTargetA(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x87, false)

	cpu.Reg.A.Write(0x7)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xE)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8NonOverflowingTargetC(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x81, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xA)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8NonOverflowingTargetCWithCarry(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x81, false)

	cpu.Reg.A.Write(0x7)
	cpu.Reg.C.Write(0x3)
	cpu.Reg.F.Carry = true

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0xA)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd8TargetBCarry(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x80, false)

	cpu.Reg.A.Write(0xFC)
	cpu.Reg.B.Write(0x9)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.A.Read(), 0x5)
	assertFlags(t, cpu, false, false, true, true)
}

func TestExecuteAdd16TargetHL(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x29, false)

	cpu.Reg.HL.Write(0x2331)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x4662)
	assertFlags(t, cpu, false, false, false, false)
}

func TestExecuteAdd16TargetBCHalfCarry(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x09, false)

	cpu.Reg.HL.Write(0x0300)
	cpu.Reg.BC.Write(0x0700)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0A00)
	assertFlags(t, cpu, false, false, true, false)
}

func TestExecuteAdd16TargetDECarry(t *testing.T) {
	cpu, _ := NewCpu()
	inst, _ := cpu.opcodes.InstructionFromByte(0x19, false)

	cpu.Reg.HL.Write(0xF110)
	cpu.Reg.DE.Write(0x0FF0)

	cpu.Execute(inst)

	assertRegEquals(t, cpu.Reg.HL.Read(), 0x0100)
	assertFlags(t, cpu, false, false, true, true)
}

func TestCPUReset(t *testing.T) {
	cpu, _ := NewCpu()

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
