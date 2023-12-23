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

	ime     bool
	halted  bool
	opcodes *isa.Opcodes
}

func NewCPU() (*CPU, error) {
	cpu := new(CPU)
	cpu.Reg = NewRegisters()
	cpu.PC = &Register[uint16]{name: "PC", value: 0x0000}
	cpu.SP = &Register[uint16]{name: "SP", value: 0x0000}
	cpu.halted = false
	cpu.ime = true

	opcodes, err := isa.LoadOpcodes()
	if err != nil {
		return nil, err
	}

	cpu.opcodes = opcodes

	return cpu, nil
}

func (cpu *CPU) Step(mmu *mem.MMU) uint8 {
	if cpu.halted {
		// HALT is 4 cycles
		return 4
	}

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
			log.Fatalf("Unimplemented instruction found @ 0x%04X: 0xCB%02X", addr, opcodeByte)
		} else {
			log.Fatalf("Unimplemented instruction found @ 0x%04X: 0x%02X", addr, opcodeByte)
		}
	}

	return inst
}

func (cpu *CPU) Execute(mmu *mem.MMU, inst *isa.Instruction) (nextPC uint16, cycles uint8) {
	opcode := inst.Opcode

	switch opcode.Addr {
	case 0x00:
		// NOP
	case 0x01:
		// LD BC, n16
		cpu.load16(cpu.Reg.BC, cpu.readNext16(mmu))
	case 0x02:
		// LD (BC), A
		cpu.load8Indirect(mmu, cpu.Reg.BC.Read(), cpu.Reg.A)
	case 0x03:
		// INC BC
		cpu.Reg.BC.Inc(1)
	case 0x04:
		// INC B
		cpu.inc8(cpu.Reg.B)
	case 0x05:
		// DEC B
		cpu.dec8(cpu.Reg.B)
	case 0x06:
		// LD B, n8
		cpu.load8(cpu.Reg.B, cpu.readNext8(mmu))
	case 0x08:
		// LD (a16), SP
		cpu.load16Indirect(mmu, cpu.readNext16(mmu), cpu.SP)
	case 0x09:
		// ADD HL, BC
		cpu.add16(cpu.Reg.HL, cpu.Reg.BC.Read())
	case 0x0A:
		// LD A, (BC)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.Reg.BC.Read()))
	case 0x0B:
		// DEC BC
		cpu.Reg.BC.Dec(1)
	case 0x0C:
		// INC C
		cpu.inc8(cpu.Reg.C)
	case 0x0D:
		// DEC C
		cpu.dec8(cpu.Reg.C)
	case 0x0E:
		// LD C, n8
		cpu.load8(cpu.Reg.C, cpu.readNext8(mmu))
	case 0x11:
		// LD DE, n16
		cpu.load16(cpu.Reg.DE, cpu.readNext16(mmu))
	case 0x12:
		// LD (DE), A
		cpu.load8Indirect(mmu, cpu.Reg.DE.Read(), cpu.Reg.A)
	case 0x13:
		// INC DE
		cpu.Reg.DE.Inc(1)
	case 0x14:
		// INC D
		cpu.inc8(cpu.Reg.D)
	case 0x15:
		// DEC D
		cpu.dec8(cpu.Reg.D)
	case 0x18:
		// JR e8
		return cpu.jump_rel(mmu, opcode, true)
	case 0x19:
		// ADD HL, DE
		cpu.add16(cpu.Reg.HL, cpu.Reg.DE.Read())
	case 0x1A:
		// LD A, (DE)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.Reg.DE.Read()))
	case 0x1B:
		// DEC DE
		cpu.Reg.DE.Dec(1)
	case 0x1C:
		// INC E
		cpu.inc8(cpu.Reg.E)
	case 0x1D:
		// DEC E
		cpu.dec8(cpu.Reg.E)
	case 0x20:
		// JR NZ, e8
		return cpu.jump_rel(mmu, opcode, !cpu.Reg.F.Zero)
	case 0x21:
		// LD HL, n16
		cpu.load16(cpu.Reg.HL, cpu.readNext16(mmu))
	case 0x22:
		// LD (HL+), A
		cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.A)
		cpu.Reg.HL.Inc(1)
	case 0x23:
		// INC HL
		cpu.Reg.HL.Inc(1)
	case 0x24:
		// INC H
		cpu.inc8(cpu.Reg.H)
	case 0x25:
		// DEC H
		cpu.dec8(cpu.Reg.H)
	case 0x26:
		// LD H, n8
		cpu.load8(cpu.Reg.H, cpu.readNext8(mmu))
	case 0x28:
		// JR Z, e8
		return cpu.jump_rel(mmu, opcode, cpu.Reg.F.Zero)
	case 0x29:
		// ADD HL, HL
		cpu.add16(cpu.Reg.HL, cpu.Reg.HL.Read())
	case 0x2A:
		// LDI A, (HL+)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()))
		cpu.Reg.HL.Inc(1)
	case 0x2B:
		// DEC HL
		cpu.Reg.HL.Dec(1)
	case 0x2C:
		// INC L
		cpu.inc8(cpu.Reg.L)
	case 0x2D:
		// DEC L
		cpu.dec8(cpu.Reg.L)
	case 0x2E:
		// LD L, n8
		cpu.load8(cpu.Reg.L, cpu.readNext8(mmu))
	case 0x30:
		// JR NC, e8
		return cpu.jump_rel(mmu, opcode, !cpu.Reg.F.Carry)
	case 0x31:
		// LD SP, n16
		cpu.load16(cpu.SP, cpu.readNext16(mmu))
	case 0x32:
		// LD (HL-), A
		cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.A)
		cpu.Reg.HL.Dec(1)
	case 0x33:
		// INC SP
		cpu.SP.Inc(1)
	case 0x34:
		// INC (HL)
		value := mmu.Read8(cpu.Reg.HL.Read())
		fauxReg := &Register[uint8]{name: "(HL)", value: value}
		cpu.add8(fauxReg, 1)
		mmu.Write8(cpu.Reg.HL.Read(), fauxReg.Read())
	case 0x35:
		// DEC (HL)
		value := mmu.Read8(cpu.Reg.HL.Read())
		fauxReg := &Register[uint8]{name: "(HL)", value: value}
		cpu.sub8(fauxReg, 1)
		mmu.Write8(cpu.Reg.HL.Read(), fauxReg.Read())
	case 0x38:
		// JR C, e8
		return cpu.jump_rel(mmu, opcode, cpu.Reg.F.Carry)
	case 0x39:
		// ADD HL, SP
		cpu.add16(cpu.Reg.HL, cpu.SP.Read())
	case 0x3A:
		// LD A, (HL-)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()))
		cpu.Reg.HL.Dec(1)
	case 0x3B:
		// DEC SP
		cpu.SP.Dec(1)
	case 0x3C:
		// INC A
		cpu.inc8(cpu.Reg.A)
	case 0x3D:
		// DEC A
		cpu.dec8(cpu.Reg.A)
	case 0x3E:
		// LD A, n8
		cpu.load8(cpu.Reg.A, cpu.readNext8(mmu))
	case 0x40:
		// LD B, B
		cpu.load8(cpu.Reg.B, cpu.Reg.B.Read())
	case 0x41:
		// LD B, C
		cpu.load8(cpu.Reg.B, cpu.Reg.C.Read())
	case 0x42:
		// LD B, D
		cpu.load8(cpu.Reg.B, cpu.Reg.D.Read())
	case 0x43:
		// LD B, E
		cpu.load8(cpu.Reg.B, cpu.Reg.E.Read())
	case 0x44:
		// LD B, H
		cpu.load8(cpu.Reg.B, cpu.Reg.H.Read())
	case 0x45:
		// LD B, L
		cpu.load8(cpu.Reg.B, cpu.Reg.L.Read())
	case 0x46:
		// LD B, (HL)
		cpu.load8(cpu.Reg.B, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x47:
		// LD B, A
		cpu.load8(cpu.Reg.B, cpu.Reg.A.Read())
	case 0x48:
		// LD C, B
		cpu.load8(cpu.Reg.C, cpu.Reg.B.Read())
	case 0x49:
		// LD C, C
		cpu.load8(cpu.Reg.C, cpu.Reg.C.Read())
	case 0x4A:
		// LD C, D
		cpu.load8(cpu.Reg.C, cpu.Reg.D.Read())
	case 0x4B:
		// LD C, E
		cpu.load8(cpu.Reg.C, cpu.Reg.E.Read())
	case 0x4C:
		// LD C, H
		cpu.load8(cpu.Reg.C, cpu.Reg.H.Read())
	case 0x4D:
		// LD C, L
		cpu.load8(cpu.Reg.C, cpu.Reg.L.Read())
	case 0x4E:
		// LD C, (HL)
		cpu.load8(cpu.Reg.C, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x4F:
		// LD C, A
		cpu.load8(cpu.Reg.C, cpu.Reg.A.Read())
	case 0x50:
		// LD D, B
		cpu.load8(cpu.Reg.D, cpu.Reg.B.Read())
	case 0x51:
		// LD D, C
		cpu.load8(cpu.Reg.D, cpu.Reg.C.Read())
	case 0x52:
		// LD D, D
		cpu.load8(cpu.Reg.D, cpu.Reg.D.Read())
	case 0x53:
		// LD D, E
		cpu.load8(cpu.Reg.D, cpu.Reg.E.Read())
	case 0x54:
		// LD D, H
		cpu.load8(cpu.Reg.D, cpu.Reg.H.Read())
	case 0x55:
		// LD D, L
		cpu.load8(cpu.Reg.D, cpu.Reg.L.Read())
	case 0x56:
		// LD D, (HL)
		cpu.load8(cpu.Reg.D, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x57:
		// LD D, A
		cpu.load8(cpu.Reg.D, cpu.Reg.A.Read())
	case 0x58:
		// LD E, B
		cpu.load8(cpu.Reg.E, cpu.Reg.B.Read())
	case 0x59:
		// LD E, C
		cpu.load8(cpu.Reg.E, cpu.Reg.C.Read())
	case 0x5A:
		// LD E, D
		cpu.load8(cpu.Reg.E, cpu.Reg.D.Read())
	case 0x5B:
		// LD E, E
		cpu.load8(cpu.Reg.E, cpu.Reg.E.Read())
	case 0x5C:
		// LD E, H
		cpu.load8(cpu.Reg.E, cpu.Reg.H.Read())
	case 0x5D:
		// LD E, L
		cpu.load8(cpu.Reg.E, cpu.Reg.L.Read())
	case 0x5E:
		// LD E, (HL)
		cpu.load8(cpu.Reg.E, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x5F:
		// LD E, A
		cpu.load8(cpu.Reg.E, cpu.Reg.A.Read())
	case 0x60:
		// LD H, B
		cpu.load8(cpu.Reg.H, cpu.Reg.B.Read())
	case 0x61:
		// LD H, C
		cpu.load8(cpu.Reg.H, cpu.Reg.C.Read())
	case 0x62:
		// LD H, D
		cpu.load8(cpu.Reg.H, cpu.Reg.D.Read())
	case 0x63:
		// LD H, E
		cpu.load8(cpu.Reg.H, cpu.Reg.E.Read())
	case 0x64:
		// LD H, H
		cpu.load8(cpu.Reg.H, cpu.Reg.H.Read())
	case 0x65:
		// LD H, L
		cpu.load8(cpu.Reg.H, cpu.Reg.L.Read())
	case 0x66:
		// LD H, (HL)
		cpu.load8(cpu.Reg.H, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x67:
		// LD H, A
		cpu.load8(cpu.Reg.H, cpu.Reg.A.Read())
	case 0x68:
		// LD L, B
		cpu.load8(cpu.Reg.L, cpu.Reg.B.Read())
	case 0x69:
		// LD L, C
		cpu.load8(cpu.Reg.L, cpu.Reg.C.Read())
	case 0x6A:
		// LD L, D
		cpu.load8(cpu.Reg.L, cpu.Reg.D.Read())
	case 0x6B:
		// LD L, E
		cpu.load8(cpu.Reg.L, cpu.Reg.E.Read())
	case 0x6C:
		// LD L, H
		cpu.load8(cpu.Reg.L, cpu.Reg.H.Read())
	case 0x6D:
		// LD L, L
		cpu.load8(cpu.Reg.L, cpu.Reg.L.Read())
	case 0x6E:
		// LD L, (HL)
		cpu.load8(cpu.Reg.L, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x6F:
		// LD L, A
		cpu.load8(cpu.Reg.L, cpu.Reg.A.Read())
	case 0x76:
		// HALT
		cpu.halted = true
	case 0x78:
		// LD A, B
		cpu.load8(cpu.Reg.A, cpu.Reg.B.Read())
	case 0x79:
		// LD A, C
		cpu.load8(cpu.Reg.A, cpu.Reg.C.Read())
	case 0x7A:
		// LD A, D
		cpu.load8(cpu.Reg.A, cpu.Reg.D.Read())
	case 0x7B:
		// LD A, E
		cpu.load8(cpu.Reg.A, cpu.Reg.E.Read())
	case 0x7C:
		// LD A, H
		cpu.load8(cpu.Reg.A, cpu.Reg.H.Read())
	case 0x7D:
		// LD A, L
		cpu.load8(cpu.Reg.A, cpu.Reg.L.Read())
	case 0x7E:
		// LD A, (HL)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x7F:
		// LD A, A
		cpu.load8(cpu.Reg.A, cpu.Reg.A.Read())
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
	case 0x90:
		// SUB A, B
		cpu.sub8(cpu.Reg.A, cpu.Reg.B.Read())
	case 0x91:
		// SUB A, C
		cpu.sub8(cpu.Reg.A, cpu.Reg.C.Read())
	case 0x92:
		// SUB A, D
		cpu.sub8(cpu.Reg.A, cpu.Reg.D.Read())
	case 0x93:
		// SUB A, E
		cpu.sub8(cpu.Reg.A, cpu.Reg.E.Read())
	case 0x94:
		// SUB A, H
		cpu.sub8(cpu.Reg.A, cpu.Reg.H.Read())
	case 0x95:
		// SUB A, L
		cpu.sub8(cpu.Reg.A, cpu.Reg.L.Read())
	case 0x96:
		// SUB A, (HL)
		cpu.sub8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()))
	case 0x97:
		// SUB A, A
		cpu.sub8(cpu.Reg.A, cpu.Reg.A.Read())
	case 0xC0:
		// RET NZ
		return cpu.ret(mmu, opcode, !cpu.Reg.F.Zero)
	case 0xC1:
		// POP BC
		cpu.Reg.BC.Write(cpu.pop(mmu))
	case 0xC2:
		// JP NZ, a16
		return cpu.jump(mmu, opcode, !cpu.Reg.F.Zero)
	case 0xC3:
		// JP a16
		return cpu.jump(mmu, opcode, true)
	case 0xC4:
		// CALL NZ, a16
		return cpu.call(mmu, opcode, !cpu.Reg.F.Zero)
	case 0xC5:
		// PUSH BC
		cpu.push(mmu, cpu.Reg.BC.Read())
	case 0xC7:
		// RST 00H
		return cpu.rst(mmu, opcode, 0x00)
	case 0xC8:
		// RET Z
		return cpu.ret(mmu, opcode, cpu.Reg.F.Zero)
	case 0xC9:
		// RET
		return cpu.ret(mmu, opcode, true)
	case 0xCA:
		// JP Z, a16
		return cpu.jump(mmu, opcode, cpu.Reg.F.Zero)
	case 0xCC:
		// CALL Z, a16
		return cpu.call(mmu, opcode, cpu.Reg.F.Zero)
	case 0xCD:
		// CALL a16
		return cpu.call(mmu, opcode, true)
	case 0xCF:
		// RST 08H
		return cpu.rst(mmu, opcode, 0x08)
	case 0xD0:
		// RET NC
		return cpu.ret(mmu, opcode, !cpu.Reg.F.Carry)
	case 0xD1:
		// POP DE
		cpu.Reg.DE.Write(cpu.pop(mmu))
	case 0xD2:
		// JP NC, a16
		return cpu.jump(mmu, opcode, !cpu.Reg.F.Carry)
	case 0xD4:
		// CALL NC, a16
		return cpu.call(mmu, opcode, !cpu.Reg.F.Carry)
	case 0xD5:
		// PUSH DE
		cpu.push(mmu, cpu.Reg.DE.Read())
	case 0xD6:
		// SUB A, n8
		cpu.sub8(cpu.Reg.A, cpu.readNext8(mmu))
	case 0xD7:
		// RST 10H
		return cpu.rst(mmu, opcode, 0x10)
	case 0xD8:
		// RET C
		return cpu.ret(mmu, opcode, cpu.Reg.F.Carry)
	case 0xD9:
		// RETI
		cpu.ime = true
		return cpu.ret(mmu, opcode, true)
	case 0xDA:
		// JP C, a16
		return cpu.jump(mmu, opcode, cpu.Reg.F.Carry)
	case 0xDC:
		// CALL C, a16
		return cpu.call(mmu, opcode, cpu.Reg.F.Carry)
	case 0xDF:
		// RST 18H
		return cpu.rst(mmu, opcode, 0x18)
	case 0xE0:
		// LDH (a8), A
		addr := 0xFF00 + uint16(cpu.readNext8(mmu))
		cpu.load8Indirect(mmu, addr, cpu.Reg.A)
	case 0xE1:
		// POP HL
		cpu.Reg.HL.Write(cpu.pop(mmu))
	case 0xE2:
		// LD (0xFF00 + C), A
		addr := 0xFF00 + uint16(cpu.Reg.C.Read())
		cpu.load8Indirect(mmu, addr, cpu.Reg.A)
	case 0xE5:
		// PUSH HL
		cpu.push(mmu, cpu.Reg.HL.Read())
	case 0xE7:
		// RST 20H
		return cpu.rst(mmu, opcode, 0x20)
	case 0xE9:
		// JP HL
		return cpu.Reg.HL.Read(), uint8(opcode.Cycles[0])
	case 0xEA:
		// LD (a16), A
		cpu.load8Indirect(mmu, cpu.readNext16(mmu), cpu.Reg.A)
	case 0xEF:
		// RST 28H
		return cpu.rst(mmu, opcode, 0x28)
	case 0xF0:
		// LDH A, (a8)
		addr := 0xFF00 + uint16(cpu.readNext8(mmu))
		value := mmu.Read8(addr)
		cpu.load8(cpu.Reg.A, value)
	case 0xF1:
		// POP AF
		cpu.Reg.AF.Write(cpu.pop(mmu))
	case 0xF2:
		// LD A, (0xFF00 + C)
		addr := 0xFF00 + uint16(cpu.Reg.C.Read())
		value := mmu.Read8(addr)
		cpu.load8(cpu.Reg.A, value)
	case 0xF3:
		// DI
		cpu.ime = false
	case 0xF5:
		// PUSH AF
		cpu.push(mmu, cpu.Reg.AF.Read())
	case 0xF7:
		// RST 30H
		return cpu.rst(mmu, opcode, 0x30)
	case 0xF9:
		// LD SP, HL
		cpu.load16(cpu.SP, cpu.Reg.HL.Read())
	case 0xFB:
		// EI
		cpu.ime = true
	case 0xFA:
		// LD A, (a16)
		cpu.load8(cpu.Reg.A, mmu.Read8(cpu.readNext16(mmu)))
	case 0xFF:
		// RST 38H
		return cpu.rst(mmu, opcode, 0x38)
	default:
		log.Fatalf("Unimplemented instruction @ %s", inst)
	}

	return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[0])
}

func (cpu *CPU) Reset() {
	cpu.Reg.Reset()
	cpu.PC.Write(0x0000)
	cpu.SP.Write(0x0000)
	cpu.ime = true
	cpu.halted = false
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
	cpu.ime = true
	cpu.halted = false
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

func (cpu *CPU) inc8(reg RWByte) uint8 {
	oldValue := reg.Read()
	newValue := oldValue + 1
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = isHalfCarry8(oldValue, 1)

	return newValue
}

func (cpu *CPU) sub8(reg RWByte, value uint8) uint8 {
	oldValue := reg.Read()
	newValue := oldValue - value
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.Carry = newValue > oldValue
	cpu.Reg.F.HalfCarry = isHalfCarry8(newValue, value)

	return newValue
}

func (cpu *CPU) dec8(reg RWByte) uint8 {
	oldValue := reg.Read()
	newValue := oldValue - 1
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.HalfCarry = isHalfCarry8(newValue, 1)

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

func (cpu *CPU) call(mmu *mem.MMU, opcode *isa.Opcode, should_jump bool) (nextPC uint16, cycles uint8) {
	nextPC = cpu.PC.Read() + uint16(opcode.Bytes)

	if should_jump {
		cpu.push(mmu, nextPC)
		return mmu.Read16(cpu.PC.Read() + 1), uint8(opcode.Cycles[0])
	}

	return nextPC, uint8(opcode.Cycles[0])
}

func (cpu *CPU) load8(reg RWByte, value byte) {
	reg.Write(value)
}

func (cpu *CPU) load8Indirect(mmu *mem.MMU, addr uint16, reg RWByte) {
	mmu.Write8(addr, reg.Read())
}

func (cpu *CPU) load16(reg RWTwoByte, value uint16) {
	reg.Write(value)
}

func (cpu *CPU) load16Indirect(mmu *mem.MMU, addr uint16, reg RWTwoByte) {
	mmu.Write16(addr, reg.Read())
}

func (cpu *CPU) jump(mmu *mem.MMU, opcode *isa.Opcode, should_jump bool) (nextPC uint16, cycles uint8) {
	if should_jump {
		return cpu.readNext16(mmu), uint8(opcode.Cycles[0])
	} else {
		return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[1])
	}
}

func (cpu *CPU) jump_rel(mmu *mem.MMU, opcode *isa.Opcode, should_jump bool) (nextPC uint16, cycles uint8) {
	nextPC = cpu.PC.Read() + uint16(opcode.Bytes)

	if should_jump {
		nextPcDiff := int8(cpu.readNext8(mmu))
		return nextPC + uint16(nextPcDiff), uint8(opcode.Cycles[0])
	} else {
		return nextPC, uint8(opcode.Cycles[1])
	}
}

func (cpu *CPU) pop(mmu *mem.MMU) uint16 {
	value := mmu.Read16(cpu.SP.Read())
	cpu.SP.Inc(2)
	return value
}

func (cpu *CPU) push(mmu *mem.MMU, value uint16) {
	cpu.SP.Dec(2)
	mmu.Write16(cpu.SP.Read(), value)
}

func (cpu *CPU) readNext8(mmu *mem.MMU) byte {
	return mmu.Read8(cpu.PC.Read() + 1)
}

func (cpu *CPU) readNext16(mmu *mem.MMU) uint16 {
	return mmu.Read16(cpu.PC.Read() + 1)
}

func (cpu *CPU) ret(mmu *mem.MMU, opcode *isa.Opcode, should_jump bool) (nextPC uint16, cycles uint8) {
	if should_jump {
		return cpu.pop(mmu), uint8(opcode.Cycles[0])
	} else {
		return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[1])
	}
}

func (cpu *CPU) rst(mmu *mem.MMU, opcode *isa.Opcode, value byte) (uint16, uint8) {
	cpu.push(mmu, cpu.PC.Read()+1)
	return uint16(value), uint8(opcode.Cycles[0])
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
