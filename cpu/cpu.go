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

	opcodes *isa.Opcodes
}

func (cpu *CPU) Step(mmu *mem.MMU) uint8 {
	inst := cpu.fetchAndDecode(mmu)
	nextPc, cycles := cpu.Execute(mmu, inst)

	cpu.PC.Write(nextPc)

	return cycles
}

func (cpu *CPU) fetchAndDecode(mmu *mem.MMU) *isa.Instruction {
	// Fetch :)
	addr := cpu.PC.Read()
	opcodeByte := mmu.Read8(addr)
	prefixed := opcodeByte == 0xCB

	if prefixed {
		addr += 1
		opcodeByte = mmu.Read8(addr)
	}

	// Decode :D
	inst, exist := cpu.opcodes.InstructionFromByte(addr, opcodeByte, prefixed)

	if !exist {
		if prefixed {
			log.Fatalf("Unimplemented instruction found @ 0x%x: 0xCB%X", addr, opcodeByte)
		} else {
			log.Fatalf("Unimplemented instruction found @ 0x%x: 0x%X", addr, opcodeByte)
		}
	}

	return inst
}

func (cpu *CPU) Execute(mmu *mem.MMU, inst *isa.Instruction) (nextPC uint16, cycles uint8) {
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
		cpu.add8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x87:
		// ADD A, A
		cpu.add8(cpu.Reg.A, cpu.Reg.A.Read())
	case 0xC2:
		// JP NZ, a16
		return cpu.jump(mmu, opcode, !cpu.Reg.F.Zero)
	case 0xC3:
		// JP a16
		return cpu.jump(mmu, opcode, true)
	case 0xCA:
		// JP Z, a16
		return cpu.jump(mmu, opcode, cpu.Reg.F.Zero)
	case 0xD2:
		// JP NC, a16
		return cpu.jump(mmu, opcode, !cpu.Reg.F.Carry)
	case 0xDA:
		// JP C, a16
		return cpu.jump(mmu, opcode, cpu.Reg.F.Carry)
	case 0xE9:
		// JP HL
		return cpu.Reg.HL.Read(), uint8(opcode.Cycles[0])
	default:
		log.Fatalf("Unimplemented instruction @ 0x%X: %s", inst.Addr, opcode)
	}

	return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[0])
}

func NewCPU() (*CPU, error) {
	cpu := new(CPU)
	cpu.Reg = NewRegisters()
	cpu.PC = &Register[uint16]{name: "PC", value: 0x0000}
	cpu.SP = &Register[uint16]{name: "SP", value: 0x0000}

	opcodes, err := isa.LoadOpcodes()
	if err != nil {
		return nil, err
	}

	cpu.opcodes = opcodes

	return cpu, nil
}

func (cpu *CPU) Reset() {
	cpu.Reg.Reset()
	cpu.PC.Write(0x0000)
	cpu.SP.Write(0x0000)

	// TODO: Reset interrupts, etc.
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
	oldValue := reg.Read()
	newValue := oldValue + value
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = newValue < oldValue
	cpu.Reg.F.HalfCarry = isHalfCarry8(oldValue, value)

	return newValue
}

func (cpu *CPU) add16(reg RWTwoByte, value uint16) uint16 {
	newValue := reg.Read() + value
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = newValue < value
	cpu.Reg.F.HalfCarry = isHalfCarry16(newValue, value)

	return newValue
}

func (cpu *CPU) jump(mmu *mem.MMU, opcode *isa.Opcode, should_jump bool) (nextPC uint16, cycles uint8) {
	if should_jump {
		return mmu.Read16(cpu.PC.Read() + 1), uint8(opcode.Cycles[0])
	} else {
		return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[1])
	}
}

// Did the aVal carry over from the lower half of the byte to the upper half?
func isHalfCarry8(aVal uint8, bVal uint8) bool {
	fourBitMask := uint8(0xF)
	bitFourMask := uint8(0x10)
	return (((aVal & fourBitMask) + (bVal & fourBitMask)) & bitFourMask) == bitFourMask
}

// Did the aVal carry over from the lower half of the word to the upper half?
func isHalfCarry16(aVal uint16, bVal uint16) bool {
	twelveBitMask := uint16(0xFFFF)
	bitTwelveMask := uint16(0x1000)
	return (((aVal & twelveBitMask) + (bVal & twelveBitMask)) & bitTwelveMask) == bitTwelveMask
}
