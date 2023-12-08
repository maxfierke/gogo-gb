package cpu

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/maxfierke/gogo-gb/mem"
)

type CPU struct {
	Reg Registers
	PC  *Register[uint16]
	SP  *Register[uint16]
	mmu *mem.MMU

	opcodes *isa.Opcodes
}

func (cpu *CPU) Step() {
	inst := cpu.fetchAndDecode()
	nextPc := cpu.Execute(inst)

	cpu.PC.Write(nextPc)
}

func (cpu *CPU) fetchAndDecode() isa.Instruction {
	opcodeByte := cpu.mmu.Read8(cpu.PC.Read())
	prefixed := opcodeByte == 0xCB

	if prefixed {
		opcodeByte = cpu.mmu.Read8(cpu.PC.Read() + 1)
	}

	inst := cpu.opcodes.FromByte(opcodeByte, prefixed)
	return inst
}

func (cpu *CPU) Execute(inst isa.Instruction) uint16 {
	opcode := inst.Opcode

	switch opcode.Addr {
	case 0x00:
		// NOP
	case 0x09:
		// ADD HL, BC
		cpu.add16(cpu.Reg.HL, cpu.Reg.BC.Read())
	case 0x19:
		// ADD HL, DE
		cpu.add16(cpu.Reg.HL, cpu.Reg.DE.Read())
	case 0x29:
		// ADD HL, HL
		cpu.add16(cpu.Reg.HL, cpu.Reg.HL.Read())
	case 0x39:
		// ADD HL, SP
		cpu.add16(cpu.Reg.HL, cpu.SP.Read())
	case 0x80:
		// ADD A, B
		cpu.add8(cpu.Reg.A, cpu.Reg.B.Read())
	case 0x81:
		// ADD A, C
		cpu.add8(cpu.Reg.A, cpu.Reg.C.Read())
	case 0x82:
		// ADD A, D
		cpu.add8(cpu.Reg.A, cpu.Reg.D.Read())
	case 0x83:
		// ADD A, E
		cpu.add8(cpu.Reg.A, cpu.Reg.E.Read())
	case 0x84:
		// ADD A, H
		cpu.add8(cpu.Reg.A, cpu.Reg.H.Read())
	case 0x85:
		// ADD A, L
		cpu.add8(cpu.Reg.A, cpu.Reg.L.Read())
	case 0x86:
		// ADD A, (HL)
		cpu.add8(cpu.Reg.A, cpu.mmu.Read8(cpu.Reg.HL.Read()))
	case 0x87:
		// ADD A, A
		cpu.add8(cpu.Reg.A, cpu.Reg.A.Read())
	default:
		panic(fmt.Sprintf("Unimplemented instruction @ 0x%x: %s", inst.Addr, opcode))
	}

	return cpu.PC.Read() + uint16(inst.Opcode.Bytes)
}

func NewCpu() *CPU {
	cpu := new(CPU)
	a := &Register[uint8]{name: "A", value: 0x00}
	b := &Register[uint8]{name: "B", value: 0x00}
	c := &Register[uint8]{name: "C", value: 0x00}
	d := &Register[uint8]{name: "D", value: 0x00}
	e := &Register[uint8]{name: "E", value: 0x00}
	f := &Flags{
		Zero:      false,
		Subtract:  false,
		HalfCarry: false,
		Carry:     false,
	}
	h := &Register[uint8]{name: "H", value: 0x00}
	l := &Register[uint8]{name: "L", value: 0x00}

	cpu.Reg = Registers{
		A:  a,
		B:  b,
		C:  c,
		D:  d,
		E:  e,
		F:  f,
		H:  h,
		L:  l,
		AF: &CompoundRegister{name: "AF", high: a, low: f},
		BC: &CompoundRegister{name: "BC", high: b, low: c},
		DE: &CompoundRegister{name: "DE", high: d, low: e},
		HL: &CompoundRegister{name: "HL", high: h, low: l},
	}

	cpu.PC = &Register[uint16]{name: "PC", value: 0x0000}
	cpu.SP = &Register[uint16]{name: "SP", value: 0x0000}

	cpu.mmu = mem.NewMMU()

	return cpu
}

func (cpu *CPU) add8(reg RWByte, value uint8) uint8 {
	newValue, didCarry := overflowingAdd8(reg.Read(), value)
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = didCarry

	// Did the newValue carry over from the lower half of the byte to the upper half?
	cpu.Reg.F.HalfCarry = (newValue&0xF)+(value&0xF) > 0xF

	return newValue
}

func (cpu *CPU) add16(reg RWTwoByte, value uint16) uint16 {
	newValue, didCarry := overflowingAdd16(reg.Read(), value)
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = didCarry

	// Did the newValue carry over from the lower byte to the upper byte?
	cpu.Reg.F.HalfCarry = (newValue&0xFFF)+(value&0xFFF) > 0xFFF

	return newValue
}

func overflowingAdd8(x, y uint8) (uint8, bool) {
	sum16 := uint16(x) + uint16(y)
	return x + y, uint8(sum16>>7) == 1
}

func overflowingAdd16(x, y uint16) (uint16, bool) {
	sum32 := uint32(x) + uint32(y)
	return x + y, uint16(sum32>>15) == 1
}
