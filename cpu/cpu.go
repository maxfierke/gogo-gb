package cpu

import (
	"log"

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

func (cpu *CPU) fetchAndDecode() *isa.Instruction {
	// Fetch :)
	opcodeByte := cpu.mmu.Read8(cpu.PC.Read())
	prefixed := opcodeByte == 0xCB

	if prefixed {
		opcodeByte = cpu.mmu.Read8(cpu.PC.Read() + 1)
	}

	// Decode :D
	inst, exist := cpu.opcodes.InstructionFromByte(opcodeByte, prefixed)

	if !exist {
		if prefixed {
			log.Fatalf("Unimplemented instruction found: 0xCB%X", opcodeByte)
		} else {
			log.Fatalf("Unimplemented instruction found: 0x%X", opcodeByte)
		}
	}

	return inst
}

func (cpu *CPU) Execute(inst *isa.Instruction) uint16 {
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
	case 0xC2:
		// JP NZ, a16
		return cpu.jump(!cpu.Reg.F.Zero)
	case 0xC3:
		// JP a16
		return cpu.jump(true)
	case 0xCA:
		// JP Z, a16
		return cpu.jump(cpu.Reg.F.Zero)
	case 0xD2:
		// JP NC, a16
		return cpu.jump(!cpu.Reg.F.Carry)
	case 0xDA:
		// JP C, a16
		return cpu.jump(cpu.Reg.F.Carry)
	case 0xE9:
		// JP HL
		return cpu.PC.Read() + cpu.Reg.HL.Read()
	default:
		log.Fatalf("Unimplemented instruction 0x%X %s", inst.Addr, opcode)
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

func (cpu *CPU) Reset() {
	cpu.Reg.AF.Write(0x0000)
	cpu.Reg.BC.Write(0x0000)
	cpu.Reg.DE.Write(0x0000)
	cpu.Reg.HL.Write(0x0000)
	cpu.PC.Write(0x0000)
	cpu.SP.Write(0x0000)

	// TODO: Reset memory, interrupts, etc.
}

// Reset CPU and registers to post-boot ROM state
// Mostly for gameboy-doctor usage
func (cpu *CPU) ResetToBootROM() {
	cpu.Reg.A.Write(0x01)
	cpu.Reg.F.Write(0xB0)
	cpu.Reg.B.Write(0x00)
	cpu.Reg.C.Write(0x13)
	cpu.Reg.D.Write(0x00)
	cpu.Reg.E.Write(0xD8)
	cpu.Reg.H.Write(0x01)
	cpu.Reg.L.Write(0x4D)
	cpu.SP.Write(0xFFFE)
	cpu.PC.Write(0x100)
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

// Are these necessary?
func overflowingAdd8(x, y uint8) (uint8, bool) {
	sum16 := uint16(x) + uint16(y)
	return x + y, uint8(sum16>>7) == 1
}

func overflowingAdd16(x, y uint16) (uint16, bool) {
	sum32 := uint32(x) + uint32(y)
	return x + y, uint16(sum32>>15) == 1
}

func (cpu *CPU) jump(should_jump bool) uint16 {
	if should_jump {
		// A little endianness conversion...
		lsb := uint16(cpu.mmu.Read8(cpu.PC.Read() + 1))
		msb := uint16(cpu.mmu.Read8(cpu.PC.Read() + 2))
		return msb<<8 | lsb
	} else {
		// All JP instructions are 3 bytes except 0xE9 (JP HL), which is handled separately
		return cpu.PC.Read() + 3
	}
}
