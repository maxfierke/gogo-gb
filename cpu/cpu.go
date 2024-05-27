package cpu

import (
	"fmt"
	"math/bits"

	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/maxfierke/gogo-gb/devices"
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

func (cpu *CPU) Step(mmu *mem.MMU) (uint8, error) {
	if cpu.halted {
		// HALT is 4 cycles
		return 4, nil
	}

	inst, err := cpu.fetchAndDecode(mmu)
	if err != nil {
		return 0, err
	}

	nextPc, cycles, err := cpu.Execute(mmu, inst)
	cpu.PC.Write(nextPc)

	return cycles, err
}

func (cpu *CPU) fetchAndDecode(mmu *mem.MMU) (*isa.Instruction, error) {
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
			return nil, fmt.Errorf("unimplemented instruction found @ 0x%04X: 0xCB%02X", addr, opcodeByte)
		} else {
			return nil, fmt.Errorf("unimplemented instruction found @ 0x%04X: 0x%02X", addr, opcodeByte)
		}
	}

	return inst, nil
}

func (cpu *CPU) Execute(mmu *mem.MMU, inst *isa.Instruction) (nextPC uint16, cycles uint8, err error) {
	opcode := inst.Opcode

	if opcode.CbPrefixed {
		switch opcode.Addr {
		case 0x00:
			// RLC B
			cpu.Reg.B.Write(cpu.rotl(cpu.Reg.B.Read(), true, false))
		case 0x01:
			// RLC C
			cpu.Reg.C.Write(cpu.rotl(cpu.Reg.C.Read(), true, false))
		case 0x02:
			// RLC D
			cpu.Reg.D.Write(cpu.rotl(cpu.Reg.D.Read(), true, false))
		case 0x03:
			// RLC E
			cpu.Reg.E.Write(cpu.rotl(cpu.Reg.E.Read(), true, false))
		case 0x04:
			// RLC H
			cpu.Reg.H.Write(cpu.rotl(cpu.Reg.H.Read(), true, false))
		case 0x05:
			// RLC L
			cpu.Reg.L.Write(cpu.rotl(cpu.Reg.L.Read(), true, false))
		case 0x06:
			// RLC (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.rotl(mmu.Read8(addr), true, false))
		case 0x07:
			// RLC A
			cpu.Reg.A.Write(cpu.rotl(cpu.Reg.A.Read(), true, false))
		case 0x08:
			// RRC B
			cpu.Reg.B.Write(cpu.rotr(cpu.Reg.B.Read(), true, false))
		case 0x09:
			// RRC C
			cpu.Reg.C.Write(cpu.rotr(cpu.Reg.C.Read(), true, false))
		case 0x0A:
			// RRC D
			cpu.Reg.D.Write(cpu.rotr(cpu.Reg.D.Read(), true, false))
		case 0x0B:
			// RRC E
			cpu.Reg.E.Write(cpu.rotr(cpu.Reg.E.Read(), true, false))
		case 0x0C:
			// RRC H
			cpu.Reg.H.Write(cpu.rotr(cpu.Reg.H.Read(), true, false))
		case 0x0D:
			// RRC L
			cpu.Reg.L.Write(cpu.rotr(cpu.Reg.L.Read(), true, false))
		case 0x0E:
			// RRC (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.rotr(mmu.Read8(addr), true, false))
		case 0x0F:
			// RRC A
			cpu.Reg.A.Write(cpu.rotr(cpu.Reg.A.Read(), true, false))
		case 0x10:
			// RL B
			cpu.Reg.B.Write(cpu.rotl(cpu.Reg.B.Read(), true, true))
		case 0x11:
			// RL C
			cpu.Reg.C.Write(cpu.rotl(cpu.Reg.C.Read(), true, true))
		case 0x12:
			// RL D
			cpu.Reg.D.Write(cpu.rotl(cpu.Reg.D.Read(), true, true))
		case 0x13:
			// RL E
			cpu.Reg.E.Write(cpu.rotl(cpu.Reg.E.Read(), true, true))
		case 0x14:
			// RL H
			cpu.Reg.H.Write(cpu.rotl(cpu.Reg.H.Read(), true, true))
		case 0x15:
			// RL L
			cpu.Reg.L.Write(cpu.rotl(cpu.Reg.L.Read(), true, true))
		case 0x16:
			// RL (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.rotl(mmu.Read8(addr), true, true))
		case 0x17:
			// RL A
			cpu.Reg.A.Write(cpu.rotl(cpu.Reg.A.Read(), true, true))
		case 0x18:
			// RR B
			cpu.Reg.B.Write(cpu.rotr(cpu.Reg.B.Read(), true, true))
		case 0x19:
			// RR C
			cpu.Reg.C.Write(cpu.rotr(cpu.Reg.C.Read(), true, true))
		case 0x1A:
			// RR D
			cpu.Reg.D.Write(cpu.rotr(cpu.Reg.D.Read(), true, true))
		case 0x1B:
			// RR E
			cpu.Reg.E.Write(cpu.rotr(cpu.Reg.E.Read(), true, true))
		case 0x1C:
			// RR H
			cpu.Reg.H.Write(cpu.rotr(cpu.Reg.H.Read(), true, true))
		case 0x1D:
			// RR L
			cpu.Reg.L.Write(cpu.rotr(cpu.Reg.L.Read(), true, true))
		case 0x1E:
			// RR (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.rotr(mmu.Read8(addr), true, true))
		case 0x1F:
			// RR A
			cpu.Reg.A.Write(cpu.rotr(cpu.Reg.A.Read(), true, true))
		case 0x20:
			// SLA B
			cpu.Reg.B.Write(cpu.sla(cpu.Reg.B.Read()))
		case 0x21:
			// SLA C
			cpu.Reg.C.Write(cpu.sla(cpu.Reg.C.Read()))
		case 0x22:
			// SLA D
			cpu.Reg.D.Write(cpu.sla(cpu.Reg.D.Read()))
		case 0x23:
			// SLA E
			cpu.Reg.E.Write(cpu.sla(cpu.Reg.E.Read()))
		case 0x24:
			// SLA H
			cpu.Reg.H.Write(cpu.sla(cpu.Reg.H.Read()))
		case 0x25:
			// SLA L
			cpu.Reg.L.Write(cpu.sla(cpu.Reg.L.Read()))
		case 0x26:
			// SLA (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.sla(mmu.Read8(addr)))
		case 0x27:
			// SLA A
			cpu.Reg.A.Write(cpu.sla(cpu.Reg.A.Read()))
		case 0x28:
			// SRA B
			cpu.Reg.B.Write(cpu.sra(cpu.Reg.B.Read()))
		case 0x29:
			// SRA C
			cpu.Reg.C.Write(cpu.sra(cpu.Reg.C.Read()))
		case 0x2A:
			// SRA D
			cpu.Reg.D.Write(cpu.sra(cpu.Reg.D.Read()))
		case 0x2B:
			// SRA E
			cpu.Reg.E.Write(cpu.sra(cpu.Reg.E.Read()))
		case 0x2C:
			// SRA H
			cpu.Reg.H.Write(cpu.sra(cpu.Reg.H.Read()))
		case 0x2D:
			// SRA L
			cpu.Reg.L.Write(cpu.sra(cpu.Reg.L.Read()))
		case 0x2E:
			// SRA (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.sra(mmu.Read8(addr)))
		case 0x2F:
			// SRA A
			cpu.Reg.A.Write(cpu.sra(cpu.Reg.A.Read()))
		case 0x30:
			// SWAP B
			cpu.swap(cpu.Reg.B)
		case 0x31:
			// SWAP C
			cpu.swap(cpu.Reg.C)
		case 0x32:
			// SWAP D
			cpu.swap(cpu.Reg.D)
		case 0x33:
			// SWAP E
			cpu.swap(cpu.Reg.E)
		case 0x34:
			// SWAP H
			cpu.swap(cpu.Reg.H)
		case 0x35:
			// SWAP L
			cpu.swap(cpu.Reg.L)
		case 0x36:
			// SWAP (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.swap(&cell)
			mmu.Write8(addr, cell.Read())
		case 0x37:
			// SWAP A
			cpu.swap(cpu.Reg.A)
		case 0x38:
			// SRL B
			cpu.Reg.B.Write(cpu.srl(cpu.Reg.B.Read()))
		case 0x39:
			// SRL C
			cpu.Reg.C.Write(cpu.srl(cpu.Reg.C.Read()))
		case 0x3A:
			// SRL D
			cpu.Reg.D.Write(cpu.srl(cpu.Reg.D.Read()))
		case 0x3B:
			// SRL E
			cpu.Reg.E.Write(cpu.srl(cpu.Reg.E.Read()))
		case 0x3C:
			// SRL H
			cpu.Reg.H.Write(cpu.srl(cpu.Reg.H.Read()))
		case 0x3D:
			// SRL L
			cpu.Reg.L.Write(cpu.srl(cpu.Reg.L.Read()))
		case 0x3E:
			// SRL (HL)
			addr := cpu.Reg.HL.Read()
			mmu.Write8(addr, cpu.srl(mmu.Read8(addr)))
		case 0x3F:
			// SRL A
			cpu.Reg.A.Write(cpu.srl(cpu.Reg.A.Read()))
		case 0x40:
			// BIT 0, B
			cpu.testBit(0, cpu.Reg.B)
		case 0x41:
			// BIT 0, C
			cpu.testBit(0, cpu.Reg.C)
		case 0x42:
			// BIT 0, D
			cpu.testBit(0, cpu.Reg.D)
		case 0x43:
			// BIT 0, E
			cpu.testBit(0, cpu.Reg.E)
		case 0x44:
			// BIT 0, H
			cpu.testBit(0, cpu.Reg.H)
		case 0x45:
			// BIT 0, L
			cpu.testBit(0, cpu.Reg.L)
		case 0x46:
			// BIT 0, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(0, &cell)
		case 0x47:
			// BIT 0, A
			cpu.testBit(0, cpu.Reg.A)
		case 0x48:
			// BIT 1, B
			cpu.testBit(1, cpu.Reg.B)
		case 0x49:
			// BIT 1, C
			cpu.testBit(1, cpu.Reg.C)
		case 0x4A:
			// BIT 1, D
			cpu.testBit(1, cpu.Reg.D)
		case 0x4B:
			// BIT 1, E
			cpu.testBit(1, cpu.Reg.E)
		case 0x4C:
			// BIT 1, H
			cpu.testBit(1, cpu.Reg.H)
		case 0x4D:
			// BIT 1, L
			cpu.testBit(1, cpu.Reg.L)
		case 0x4E:
			// BIT 1, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(1, &cell)
		case 0x4F:
			// BIT 1, A
			cpu.testBit(1, cpu.Reg.A)
		case 0x50:
			// BIT 2, B
			cpu.testBit(2, cpu.Reg.B)
		case 0x51:
			// BIT 2, C
			cpu.testBit(2, cpu.Reg.C)
		case 0x52:
			// BIT 2, D
			cpu.testBit(2, cpu.Reg.D)
		case 0x53:
			// BIT 2, E
			cpu.testBit(2, cpu.Reg.E)
		case 0x54:
			// BIT 2, H
			cpu.testBit(2, cpu.Reg.H)
		case 0x55:
			// BIT 2, L
			cpu.testBit(2, cpu.Reg.L)
		case 0x56:
			// BIT 2, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(2, &cell)
		case 0x57:
			// BIT 2, A
			cpu.testBit(2, cpu.Reg.A)
		case 0x58:
			// BIT 3, B
			cpu.testBit(3, cpu.Reg.B)
		case 0x59:
			// BIT 3, C
			cpu.testBit(3, cpu.Reg.C)
		case 0x5A:
			// BIT 3, D
			cpu.testBit(3, cpu.Reg.D)
		case 0x5B:
			// BIT 3, E
			cpu.testBit(3, cpu.Reg.E)
		case 0x5C:
			// BIT 3, H
			cpu.testBit(3, cpu.Reg.H)
		case 0x5D:
			// BIT 3, L
			cpu.testBit(3, cpu.Reg.L)
		case 0x5E:
			// BIT 3, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(3, &cell)
		case 0x5F:
			// BIT 3, A
			cpu.testBit(3, cpu.Reg.A)
		case 0x60:
			// BIT 4, B
			cpu.testBit(4, cpu.Reg.B)
		case 0x61:
			// BIT 4, C
			cpu.testBit(4, cpu.Reg.C)
		case 0x62:
			// BIT 4, D
			cpu.testBit(4, cpu.Reg.D)
		case 0x63:
			// BIT 4, E
			cpu.testBit(4, cpu.Reg.E)
		case 0x64:
			// BIT 4, H
			cpu.testBit(4, cpu.Reg.H)
		case 0x65:
			// BIT 4, L
			cpu.testBit(4, cpu.Reg.L)
		case 0x66:
			// BIT 4, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(4, &cell)
		case 0x67:
			// BIT 4, A
			cpu.testBit(4, cpu.Reg.A)
		case 0x68:
			// BIT 5, B
			cpu.testBit(5, cpu.Reg.B)
		case 0x69:
			// BIT 5, C
			cpu.testBit(5, cpu.Reg.C)
		case 0x6A:
			// BIT 5, D
			cpu.testBit(5, cpu.Reg.D)
		case 0x6B:
			// BIT 5, E
			cpu.testBit(5, cpu.Reg.E)
		case 0x6C:
			// BIT 5, H
			cpu.testBit(5, cpu.Reg.H)
		case 0x6D:
			// BIT 5, L
			cpu.testBit(5, cpu.Reg.L)
		case 0x6E:
			// BIT 5, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(5, &cell)
		case 0x6F:
			// BIT 5, A
			cpu.testBit(5, cpu.Reg.A)
		case 0x70:
			// BIT 6, B
			cpu.testBit(6, cpu.Reg.B)
		case 0x71:
			// BIT 6, C
			cpu.testBit(6, cpu.Reg.C)
		case 0x72:
			// BIT 6, D
			cpu.testBit(6, cpu.Reg.D)
		case 0x73:
			// BIT 6, E
			cpu.testBit(6, cpu.Reg.E)
		case 0x74:
			// BIT 6, H
			cpu.testBit(6, cpu.Reg.H)
		case 0x75:
			// BIT 6, L
			cpu.testBit(6, cpu.Reg.L)
		case 0x76:
			// BIT 6, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(6, &cell)
		case 0x77:
			// BIT 6, A
			cpu.testBit(6, cpu.Reg.A)
		case 0x78:
			// BIT 7, B
			cpu.testBit(7, cpu.Reg.B)
		case 0x79:
			// BIT 7, C
			cpu.testBit(7, cpu.Reg.C)
		case 0x7A:
			// BIT 7, D
			cpu.testBit(7, cpu.Reg.D)
		case 0x7B:
			// BIT 7, E
			cpu.testBit(7, cpu.Reg.E)
		case 0x7C:
			// BIT 7, H
			cpu.testBit(7, cpu.Reg.H)
		case 0x7D:
			// BIT 7, L
			cpu.testBit(7, cpu.Reg.L)
		case 0x7E:
			// BIT 7, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.testBit(7, &cell)
		case 0x7F:
			// BIT 7, A
			cpu.testBit(7, cpu.Reg.A)
		case 0x80:
			// RES 0, B
			cpu.resetBit(0, cpu.Reg.B)
		case 0x81:
			// RES 0, C
			cpu.resetBit(0, cpu.Reg.C)
		case 0x82:
			// RES 0, D
			cpu.resetBit(0, cpu.Reg.D)
		case 0x83:
			// RES 0, E
			cpu.resetBit(0, cpu.Reg.E)
		case 0x84:
			// RES 0, H
			cpu.resetBit(0, cpu.Reg.H)
		case 0x85:
			// RES 0, L
			cpu.resetBit(0, cpu.Reg.L)
		case 0x86:
			// RES 0, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(0, &cell)
			mmu.Write8(addr, cell.Read())
		case 0x87:
			// RES 0, A
			cpu.resetBit(0, cpu.Reg.A)
		case 0x88:
			// RES 1, B
			cpu.resetBit(1, cpu.Reg.B)
		case 0x89:
			// RES 1, C
			cpu.resetBit(1, cpu.Reg.C)
		case 0x8A:
			// RES 1, D
			cpu.resetBit(1, cpu.Reg.D)
		case 0x8B:
			// RES 1, E
			cpu.resetBit(1, cpu.Reg.E)
		case 0x8C:
			// RES 1, H
			cpu.resetBit(1, cpu.Reg.H)
		case 0x8D:
			// RES 1, L
			cpu.resetBit(1, cpu.Reg.L)
		case 0x8E:
			// RES 1, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(1, &cell)
			mmu.Write8(addr, cell.Read())
		case 0x8F:
			// RES 1, A
			cpu.resetBit(1, cpu.Reg.A)
		case 0x90:
			// RES 2, B
			cpu.resetBit(2, cpu.Reg.B)
		case 0x91:
			// RES 2, C
			cpu.resetBit(2, cpu.Reg.C)
		case 0x92:
			// RES 2, D
			cpu.resetBit(2, cpu.Reg.D)
		case 0x93:
			// RES 2, E
			cpu.resetBit(2, cpu.Reg.E)
		case 0x94:
			// RES 2, H
			cpu.resetBit(2, cpu.Reg.H)
		case 0x95:
			// RES 2, L
			cpu.resetBit(2, cpu.Reg.L)
		case 0x96:
			// RES 2, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(2, &cell)
			mmu.Write8(addr, cell.Read())
		case 0x97:
			// RES 2, A
			cpu.resetBit(2, cpu.Reg.A)
		case 0x98:
			// RES 3, B
			cpu.resetBit(3, cpu.Reg.B)
		case 0x99:
			// RES 3, C
			cpu.resetBit(3, cpu.Reg.C)
		case 0x9A:
			// RES 3, D
			cpu.resetBit(3, cpu.Reg.D)
		case 0x9B:
			// RES 3, E
			cpu.resetBit(3, cpu.Reg.E)
		case 0x9C:
			// RES 3, H
			cpu.resetBit(3, cpu.Reg.H)
		case 0x9D:
			// RES 3, L
			cpu.resetBit(3, cpu.Reg.L)
		case 0x9E:
			// RES 3, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(3, &cell)
			mmu.Write8(addr, cell.Read())
		case 0x9F:
			// RES 3, A
			cpu.resetBit(3, cpu.Reg.A)
		case 0xA0:
			// RES 4, B
			cpu.resetBit(4, cpu.Reg.B)
		case 0xA1:
			// RES 4, C
			cpu.resetBit(4, cpu.Reg.C)
		case 0xA2:
			// RES 4, D
			cpu.resetBit(4, cpu.Reg.D)
		case 0xA3:
			// RES 4, E
			cpu.resetBit(4, cpu.Reg.E)
		case 0xA4:
			// RES 4, H
			cpu.resetBit(4, cpu.Reg.H)
		case 0xA5:
			// RES 4, L
			cpu.resetBit(4, cpu.Reg.L)
		case 0xA6:
			// RES 4, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(4, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xA7:
			// RES 4, A
			cpu.resetBit(4, cpu.Reg.A)
		case 0xA8:
			// RES 5, B
			cpu.resetBit(5, cpu.Reg.B)
		case 0xA9:
			// RES 5, C
			cpu.resetBit(5, cpu.Reg.C)
		case 0xAA:
			// RES 5, D
			cpu.resetBit(5, cpu.Reg.D)
		case 0xAB:
			// RES 5, E
			cpu.resetBit(5, cpu.Reg.E)
		case 0xAC:
			// RES 5, H
			cpu.resetBit(5, cpu.Reg.H)
		case 0xAD:
			// RES 5, L
			cpu.resetBit(5, cpu.Reg.L)
		case 0xAE:
			// RES 5, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(5, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xAF:
			// RES 5, A
			cpu.resetBit(5, cpu.Reg.A)
		case 0xB0:
			// RES 6, B
			cpu.resetBit(6, cpu.Reg.B)
		case 0xB1:
			// RES 6, C
			cpu.resetBit(6, cpu.Reg.C)
		case 0xB2:
			// RES 6, D
			cpu.resetBit(6, cpu.Reg.D)
		case 0xB3:
			// RES 6, E
			cpu.resetBit(6, cpu.Reg.E)
		case 0xB4:
			// RES 6, H
			cpu.resetBit(6, cpu.Reg.H)
		case 0xB5:
			// RES 6, L
			cpu.resetBit(6, cpu.Reg.L)
		case 0xB6:
			// RES 6, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(6, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xB7:
			// RES 6, A
			cpu.resetBit(6, cpu.Reg.A)
		case 0xB8:
			// RES 7, B
			cpu.resetBit(7, cpu.Reg.B)
		case 0xB9:
			// RES 7, C
			cpu.resetBit(7, cpu.Reg.C)
		case 0xBA:
			// RES 7, D
			cpu.resetBit(7, cpu.Reg.D)
		case 0xBB:
			// RES 7, E
			cpu.resetBit(7, cpu.Reg.E)
		case 0xBC:
			// RES 7, H
			cpu.resetBit(7, cpu.Reg.H)
		case 0xBD:
			// RES 7, L
			cpu.resetBit(7, cpu.Reg.L)
		case 0xBE:
			// RES 7, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.resetBit(7, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xBF:
			// RES 7, A
			cpu.resetBit(7, cpu.Reg.A)
		case 0xC0:
			// SET 0, B
			cpu.setBit(0, cpu.Reg.B)
		case 0xC1:
			// SET 0, C
			cpu.setBit(0, cpu.Reg.C)
		case 0xC2:
			// SET 0, D
			cpu.setBit(0, cpu.Reg.D)
		case 0xC3:
			// SET 0, E
			cpu.setBit(0, cpu.Reg.E)
		case 0xC4:
			// SET 0, H
			cpu.setBit(0, cpu.Reg.H)
		case 0xC5:
			// SET 0, L
			cpu.setBit(0, cpu.Reg.L)
		case 0xC6:
			// SET 0, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(0, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xC7:
			// SET 0, A
			cpu.setBit(0, cpu.Reg.A)
		case 0xC8:
			// SET 1, B
			cpu.setBit(1, cpu.Reg.B)
		case 0xC9:
			// SET 1, C
			cpu.setBit(1, cpu.Reg.C)
		case 0xCA:
			// SET 1, D
			cpu.setBit(1, cpu.Reg.D)
		case 0xCB:
			// SET 1, E
			cpu.setBit(1, cpu.Reg.E)
		case 0xCC:
			// SET 1, H
			cpu.setBit(1, cpu.Reg.H)
		case 0xCD:
			// SET 1, L
			cpu.setBit(1, cpu.Reg.L)
		case 0xCE:
			// SET 1, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(1, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xCF:
			// SET 1, A
			cpu.setBit(1, cpu.Reg.A)
		case 0xD0:
			// SET 2, B
			cpu.setBit(2, cpu.Reg.B)
		case 0xD1:
			// SET 2, C
			cpu.setBit(2, cpu.Reg.C)
		case 0xD2:
			// SET 2, D
			cpu.setBit(2, cpu.Reg.D)
		case 0xD3:
			// SET 2, E
			cpu.setBit(2, cpu.Reg.E)
		case 0xD4:
			// SET 2, H
			cpu.setBit(2, cpu.Reg.H)
		case 0xD5:
			// SET 2, L
			cpu.setBit(2, cpu.Reg.L)
		case 0xD6:
			// SET 2, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(2, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xD7:
			// SET 2, A
			cpu.setBit(2, cpu.Reg.A)
		case 0xD8:
			// SET 3, B
			cpu.setBit(3, cpu.Reg.B)
		case 0xD9:
			// SET 3, C
			cpu.setBit(3, cpu.Reg.C)
		case 0xDA:
			// SET 3, D
			cpu.setBit(3, cpu.Reg.D)
		case 0xDB:
			// SET 3, E
			cpu.setBit(3, cpu.Reg.E)
		case 0xDC:
			// SET 3, H
			cpu.setBit(3, cpu.Reg.H)
		case 0xDD:
			// SET 3, L
			cpu.setBit(3, cpu.Reg.L)
		case 0xDE:
			// SET 3, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(3, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xDF:
			// SET 3, A
			cpu.setBit(3, cpu.Reg.A)
		case 0xE0:
			// SET 4, B
			cpu.setBit(4, cpu.Reg.B)
		case 0xE1:
			// SET 4, C
			cpu.setBit(4, cpu.Reg.C)
		case 0xE2:
			// SET 4, D
			cpu.setBit(4, cpu.Reg.D)
		case 0xE3:
			// SET 4, E
			cpu.setBit(4, cpu.Reg.E)
		case 0xE4:
			// SET 4, H
			cpu.setBit(4, cpu.Reg.H)
		case 0xE5:
			// SET 4, L
			cpu.setBit(4, cpu.Reg.L)
		case 0xE6:
			// SET 4, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(4, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xE7:
			// SET 4, A
			cpu.setBit(4, cpu.Reg.A)
		case 0xE8:
			// SET 5, B
			cpu.setBit(5, cpu.Reg.B)
		case 0xE9:
			// SET 5, C
			cpu.setBit(5, cpu.Reg.C)
		case 0xEA:
			// SET 5, D
			cpu.setBit(5, cpu.Reg.D)
		case 0xEB:
			// SET 5, E
			cpu.setBit(5, cpu.Reg.E)
		case 0xEC:
			// SET 5, H
			cpu.setBit(5, cpu.Reg.H)
		case 0xED:
			// SET 5, L
			cpu.setBit(5, cpu.Reg.L)
		case 0xEE:
			// SET 5, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(5, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xEF:
			// SET 5, A
			cpu.setBit(5, cpu.Reg.A)
		case 0xF0:
			// SET 6, B
			cpu.setBit(6, cpu.Reg.B)
		case 0xF1:
			// SET 6, C
			cpu.setBit(6, cpu.Reg.C)
		case 0xF2:
			// SET 6, D
			cpu.setBit(6, cpu.Reg.D)
		case 0xF3:
			// SET 6, E
			cpu.setBit(6, cpu.Reg.E)
		case 0xF4:
			// SET 6, H
			cpu.setBit(6, cpu.Reg.H)
		case 0xF5:
			// SET 6, L
			cpu.setBit(6, cpu.Reg.L)
		case 0xF6:
			// SET 6, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(6, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xF7:
			// SET 6, A
			cpu.setBit(6, cpu.Reg.A)
		case 0xF8:
			// SET 7, B
			cpu.setBit(7, cpu.Reg.B)
		case 0xF9:
			// SET 7, C
			cpu.setBit(7, cpu.Reg.C)
		case 0xFA:
			// SET 7, D
			cpu.setBit(7, cpu.Reg.D)
		case 0xFB:
			// SET 7, E
			cpu.setBit(7, cpu.Reg.E)
		case 0xFC:
			// SET 7, H
			cpu.setBit(7, cpu.Reg.H)
		case 0xFD:
			// SET 7, L
			cpu.setBit(7, cpu.Reg.L)
		case 0xFE:
			// SET 7, (HL)
			addr := cpu.Reg.HL.Read()
			cell := ByteCell{value: mmu.Read8(addr)}
			cpu.setBit(7, &cell)
			mmu.Write8(addr, cell.Read())
		case 0xFF:
			// SET 7, A
			cpu.setBit(7, cpu.Reg.A)
		default:
			return 0, 0, fmt.Errorf("unimplemented instruction @ %s", inst)
		}
	} else {
		switch opcode.Addr {
		case 0x00:
			// NOP
		case 0xD3, 0xDB, 0xDD, 0xE3, 0xE4, 0xEB, 0xEC, 0xED, 0xFC, 0xFD:
			// ILLEGAL instructions
			// These would hang on real hardware, so we'll error out here
			return 0, 0, fmt.Errorf("illegal opcode used @ %s", inst)
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
		case 0x07:
			// RLCA
			value := cpu.Reg.A.Read()
			newValue := bits.RotateLeft8(value, 1)
			cpu.Reg.F.Zero = false
			cpu.Reg.F.Subtract = false
			cpu.Reg.F.HalfCarry = false
			cpu.Reg.F.Carry = (value & 0x80) != 0x0
			cpu.Reg.A.Write(newValue)
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
		case 0x0F:
			// RRCA
			value := cpu.Reg.A.Read()
			newValue := bits.RotateLeft8(value, -1)
			cpu.Reg.F.Zero = false
			cpu.Reg.F.Subtract = false
			cpu.Reg.F.HalfCarry = false
			cpu.Reg.F.Carry = (value & 0x1) != 0x0
			cpu.Reg.A.Write(newValue)
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
		case 0x16:
			// LD D, n8
			cpu.load8(cpu.Reg.D, cpu.readNext8(mmu))
		case 0x17:
			// RLA
			cpu.Reg.A.Write(cpu.rotl(cpu.Reg.A.Read(), false, true))
		case 0x18:
			// JR e8
			return cpu.jumpRel(mmu, opcode, true)
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
		case 0x1E:
			// LD E, n8
			cpu.load8(cpu.Reg.E, cpu.readNext8(mmu))
		case 0x1F:
			// RRA
			cpu.Reg.A.Write(cpu.rotr(cpu.Reg.A.Read(), false, true))
		case 0x20:
			// JR NZ, e8
			return cpu.jumpRel(mmu, opcode, !cpu.Reg.F.Zero)
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
		case 0x27:
			// DAA
			cpu.daa()
		case 0x28:
			// JR Z, e8
			return cpu.jumpRel(mmu, opcode, cpu.Reg.F.Zero)
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
		case 0x2F:
			// CPL
			cpu.cpl()
		case 0x30:
			// JR NC, e8
			return cpu.jumpRel(mmu, opcode, !cpu.Reg.F.Carry)
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
			cell := ByteCell{value: value}
			cpu.inc8(&cell)
			mmu.Write8(cpu.Reg.HL.Read(), cell.Read())
		case 0x35:
			// DEC (HL)
			value := mmu.Read8(cpu.Reg.HL.Read())
			cell := ByteCell{value: value}
			cpu.dec8(&cell)
			mmu.Write8(cpu.Reg.HL.Read(), cell.Read())
		case 0x36:
			// LD (HL), n8
			value := cpu.readNext8(mmu)
			cell := ByteCell{value: value}
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), &cell)
		case 0x37:
			// SCF
			cpu.Reg.F.Subtract = false
			cpu.Reg.F.HalfCarry = false
			cpu.Reg.F.Carry = true
		case 0x38:
			// JR C, e8
			return cpu.jumpRel(mmu, opcode, cpu.Reg.F.Carry)
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
		case 0x3F:
			// CCF
			cpu.Reg.F.Subtract = false
			cpu.Reg.F.HalfCarry = false
			cpu.Reg.F.Carry = !cpu.Reg.F.Carry
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
		case 0x70:
			// LD (HL), B
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.B)
		case 0x71:
			// LD (HL), C
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.C)
		case 0x72:
			// LD (HL), D
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.D)
		case 0x73:
			// LD (HL), E
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.E)
		case 0x74:
			// LD (HL), H
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.H)
		case 0x75:
			// LD (HL), L
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.L)
		case 0x76:
			// HALT
			cpu.halted = true
		case 0x77:
			// LD (HL), A
			cpu.load8Indirect(mmu, cpu.Reg.HL.Read(), cpu.Reg.A)
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
			cpu.add8(cpu.Reg.A, cpu.Reg.B.Read(), false)
		case 0x81:
			// ADD A, C
			cpu.add8(cpu.Reg.A, cpu.Reg.C.Read(), false)
		case 0x82:
			// ADD A, D
			cpu.add8(cpu.Reg.A, cpu.Reg.D.Read(), false)
		case 0x83:
			// ADD A, E
			cpu.add8(cpu.Reg.A, cpu.Reg.E.Read(), false)
		case 0x84:
			// ADD A, H
			cpu.add8(cpu.Reg.A, cpu.Reg.H.Read(), false)
		case 0x85:
			// ADD A, L
			cpu.add8(cpu.Reg.A, cpu.Reg.L.Read(), false)
		case 0x86:
			// ADD A, (HL)
			cpu.add8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()), false)
		case 0x87:
			// ADD A, A
			cpu.add8(cpu.Reg.A, cpu.Reg.A.Read(), false)
		case 0x88:
			// ADC A, B
			cpu.add8(cpu.Reg.A, cpu.Reg.B.Read(), true)
		case 0x89:
			// ADC A, C
			cpu.add8(cpu.Reg.A, cpu.Reg.C.Read(), true)
		case 0x8A:
			// ADC A, D
			cpu.add8(cpu.Reg.A, cpu.Reg.D.Read(), true)
		case 0x8B:
			// ADC A, E
			cpu.add8(cpu.Reg.A, cpu.Reg.E.Read(), true)
		case 0x8C:
			// ADC A, H
			cpu.add8(cpu.Reg.A, cpu.Reg.H.Read(), true)
		case 0x8D:
			// ADC A, L
			cpu.add8(cpu.Reg.A, cpu.Reg.L.Read(), true)
		case 0x8E:
			// ADC A, (HL)
			cpu.add8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()), true)
		case 0x8F:
			// ADC A, A
			cpu.add8(cpu.Reg.A, cpu.Reg.A.Read(), true)
		case 0x90:
			// SUB A, B
			cpu.sub8(cpu.Reg.A, cpu.Reg.B.Read(), false)
		case 0x91:
			// SUB A, C
			cpu.sub8(cpu.Reg.A, cpu.Reg.C.Read(), false)
		case 0x92:
			// SUB A, D
			cpu.sub8(cpu.Reg.A, cpu.Reg.D.Read(), false)
		case 0x93:
			// SUB A, E
			cpu.sub8(cpu.Reg.A, cpu.Reg.E.Read(), false)
		case 0x94:
			// SUB A, H
			cpu.sub8(cpu.Reg.A, cpu.Reg.H.Read(), false)
		case 0x95:
			// SUB A, L
			cpu.sub8(cpu.Reg.A, cpu.Reg.L.Read(), false)
		case 0x96:
			// SUB A, (HL)
			cpu.sub8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()), false)
		case 0x97:
			// SUB A, A
			cpu.sub8(cpu.Reg.A, cpu.Reg.A.Read(), false)
		case 0x98:
			// SBC A, B
			cpu.sub8(cpu.Reg.A, cpu.Reg.B.Read(), true)
		case 0x99:
			// SBC A, C
			cpu.sub8(cpu.Reg.A, cpu.Reg.C.Read(), true)
		case 0x9A:
			// SBC A, D
			cpu.sub8(cpu.Reg.A, cpu.Reg.D.Read(), true)
		case 0x9B:
			// SBC A, E
			cpu.sub8(cpu.Reg.A, cpu.Reg.E.Read(), true)
		case 0x9C:
			// SBC A, H
			cpu.sub8(cpu.Reg.A, cpu.Reg.H.Read(), true)
		case 0x9D:
			// SBC A, L
			cpu.sub8(cpu.Reg.A, cpu.Reg.L.Read(), true)
		case 0x9E:
			// SBC A, (HL)
			cpu.sub8(cpu.Reg.A, mmu.Read8(cpu.Reg.HL.Read()), true)
		case 0x9F:
			// SBC A, A
			cpu.sub8(cpu.Reg.A, cpu.Reg.A.Read(), true)
		case 0xA0:
			// AND A, B
			cpu.and(cpu.Reg.B.Read())
		case 0xA1:
			// AND A, C
			cpu.and(cpu.Reg.C.Read())
		case 0xA2:
			// AND A, D
			cpu.and(cpu.Reg.D.Read())
		case 0xA3:
			// AND A, E
			cpu.and(cpu.Reg.E.Read())
		case 0xA4:
			// AND A, H
			cpu.and(cpu.Reg.H.Read())
		case 0xA5:
			// AND A, L
			cpu.and(cpu.Reg.L.Read())
		case 0xA6:
			// AND A, (HL)
			cpu.and(mmu.Read8(cpu.Reg.HL.Read()))
		case 0xA7:
			// AND A, A
			cpu.and(cpu.Reg.A.Read())
		case 0xA8:
			// XOR A, B
			cpu.xor(cpu.Reg.B.Read())
		case 0xA9:
			// XOR A, C
			cpu.xor(cpu.Reg.C.Read())
		case 0xAA:
			// XOR A, D
			cpu.xor(cpu.Reg.D.Read())
		case 0xAB:
			// XOR A, E
			cpu.xor(cpu.Reg.E.Read())
		case 0xAC:
			// XOR A, H
			cpu.xor(cpu.Reg.H.Read())
		case 0xAD:
			// XOR A, L
			cpu.xor(cpu.Reg.L.Read())
		case 0xAE:
			// XOR A, (HL)
			cpu.xor(mmu.Read8(cpu.Reg.HL.Read()))
		case 0xAF:
			// XOR A, A
			cpu.xor(cpu.Reg.A.Read())
		case 0xB0:
			// OR A, B
			cpu.or(cpu.Reg.B.Read())
		case 0xB1:
			// OR A, C
			cpu.or(cpu.Reg.C.Read())
		case 0xB2:
			// OR A, D
			cpu.or(cpu.Reg.D.Read())
		case 0xB3:
			// OR A, E
			cpu.or(cpu.Reg.E.Read())
		case 0xB4:
			// OR A, H
			cpu.or(cpu.Reg.H.Read())
		case 0xB5:
			// OR A, L
			cpu.or(cpu.Reg.L.Read())
		case 0xB6:
			// OR A, (HL)
			cpu.or(mmu.Read8(cpu.Reg.HL.Read()))
		case 0xB7:
			// OR A, A
			cpu.or(cpu.Reg.A.Read())
		case 0xB8:
			// CP A, B
			cpu.compare(cpu.Reg.B.Read())
		case 0xB9:
			// CP A, C
			cpu.compare(cpu.Reg.C.Read())
		case 0xBA:
			// CP A, D
			cpu.compare(cpu.Reg.D.Read())
		case 0xBB:
			// CP A, E
			cpu.compare(cpu.Reg.E.Read())
		case 0xBC:
			// CP A, H
			cpu.compare(cpu.Reg.H.Read())
		case 0xBD:
			// CP A, L
			cpu.compare(cpu.Reg.L.Read())
		case 0xBE:
			// CP A, (HL)
			cpu.compare(mmu.Read8(cpu.Reg.HL.Read()))
		case 0xBF:
			// CP A, A
			cpu.compare(cpu.Reg.A.Read())
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
		case 0xC6:
			// ADD A, n8
			cpu.add8(cpu.Reg.A, cpu.readNext8(mmu), false)
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
		case 0xCE:
			// ADC A, n8
			cpu.add8(cpu.Reg.A, cpu.readNext8(mmu), true)
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
			cpu.sub8(cpu.Reg.A, cpu.readNext8(mmu), false)
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
		case 0xDE:
			// SBC A, n8
			cpu.sub8(cpu.Reg.A, cpu.readNext8(mmu), true)
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
		case 0xE6:
			// AND A, n8
			cpu.and(cpu.readNext8(mmu))
		case 0xE7:
			// RST 20H
			return cpu.rst(mmu, opcode, 0x20)
		case 0xE8:
			// ADD SP, e8
			cpu.add8Signed(cpu.SP, cpu.readNext8(mmu))
		case 0xE9:
			// JP HL
			return cpu.Reg.HL.Read(), uint8(opcode.Cycles[0]), nil
		case 0xEA:
			// LD (a16), A
			cpu.load8Indirect(mmu, cpu.readNext16(mmu), cpu.Reg.A)
		case 0xEE:
			// XOR A, n8
			cpu.xor(cpu.readNext8(mmu))
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
		case 0xF6:
			// OR A, n8
			cpu.or(cpu.readNext8(mmu))
		case 0xF7:
			// RST 30H
			return cpu.rst(mmu, opcode, 0x30)
		case 0xF8:
			// LD HL, SP+e8
			cell := WordCell{value: cpu.SP.Read()}
			cpu.add8Signed(&cell, cpu.readNext8(mmu))
			cpu.load16(cpu.Reg.HL, cell.Read())
		case 0xF9:
			// LD SP, HL
			cpu.load16(cpu.SP, cpu.Reg.HL.Read())
		case 0xFB:
			// EI
			cpu.ime = true
		case 0xFA:
			// LD A, (a16)
			cpu.load8(cpu.Reg.A, mmu.Read8(cpu.readNext16(mmu)))
		case 0xFE:
			// CP A, n8
			cpu.compare(cpu.readNext8(mmu))
		case 0xFF:
			// RST 38H
			return cpu.rst(mmu, opcode, 0x38)
		default:
			return 0, 0, fmt.Errorf("unimplemented instruction @ %s", inst)
		}
	}

	return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[0]), nil
}

func (cpu *CPU) PollInterrupts(mmu *mem.MMU, ic *devices.InterruptController) uint8 {
	if cpu.ime {
		interrupt := ic.ConsumeRequest()
		if interrupt == devices.INT_NONE {
			return 0
		}

		// Disable interrupts while we process this one
		cpu.ime = false

		// Jump to interrupt handler
		cpu.push(mmu, cpu.PC.Read())
		cpu.PC.Write(uint16(interrupt))

		// If we were halted, we're not now!
		cpu.halted = false

		// Consuming an IRQ is 20 cycles (Or 5 M-cycles)
		// ref: https://gbdev.io/pandocs/Interrupts.html#interrupt-handling
		return 20
	} else if cpu.halted {
		if interrupt := ic.NextRequest(); interrupt != 0 {
			// Wakey-wakey
			cpu.halted = false
		}

		return 0
	} else {
		// Ignore
		return 0
	}
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

func (cpu *CPU) add8(reg RWByte, value uint8, withCarry bool) uint8 {
	oldValue := reg.Read()
	newValue := oldValue + value

	carry := uint8(0)
	if withCarry && cpu.Reg.F.Carry {
		carry = 1
		newValue += carry
	}

	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = isCarry8(oldValue, value, carry)
	cpu.Reg.F.HalfCarry = isHalfCarry8(oldValue, value, carry)

	return newValue
}

func (cpu *CPU) add16(reg RWTwoByte, value uint16) uint16 {
	oldValue := reg.Read()
	newValue := oldValue + value
	reg.Write(newValue)

	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = isCarry16(oldValue, value)
	cpu.Reg.F.HalfCarry = isHalfCarry16(oldValue, value)

	return newValue
}

func (cpu *CPU) add8Signed(reg RWTwoByte, value uint8) uint16 {
	oldValue := reg.Read()

	var delta uint16
	if (value & 0x80) != 0 {
		delta = (0xFF00 | uint16(value))
	} else {
		delta = uint16(value)
	}

	newValue := oldValue + delta

	reg.Write(newValue)

	carryMask := uint16(1<<8) - 1

	cpu.Reg.F.Zero = false
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = (oldValue&carryMask)+(delta&carryMask) > carryMask
	cpu.Reg.F.HalfCarry = isHalfCarry8(uint8(oldValue), value, 0)

	return newValue
}

func (cpu *CPU) inc8(reg RWByte) uint8 {
	oldValue := reg.Read()
	newValue := oldValue + 1
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = isHalfCarry8(oldValue, 1, 0)

	return newValue
}

func (cpu *CPU) sub8(reg RWByte, value uint8, withCarry bool) uint8 {
	oldValue := reg.Read()
	newValue := oldValue - value

	carry := uint8(0)
	if withCarry && cpu.Reg.F.Carry {
		carry = 1
		newValue -= carry
	}

	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.Carry = isCarry8(newValue, value, carry)
	cpu.Reg.F.HalfCarry = isHalfCarry8(newValue, value, carry)

	return newValue
}

func (cpu *CPU) dec8(reg RWByte) uint8 {
	oldValue := reg.Read()
	newValue := oldValue - 1
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.HalfCarry = isHalfCarry8(newValue, 1, 0)

	return newValue
}

func (cpu *CPU) and(value byte) byte {
	andValue := cpu.Reg.A.Read() & value
	cpu.Reg.A.Write(andValue)

	cpu.Reg.F.Zero = andValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = false
	cpu.Reg.F.HalfCarry = true

	return andValue
}

func (cpu *CPU) compare(compareValue byte) {
	value := cpu.Reg.A.Read()

	cpu.Reg.F.Zero = value == compareValue
	cpu.Reg.F.Subtract = true
	cpu.Reg.F.Carry = value < compareValue
	cpu.Reg.F.HalfCarry = (value & 0xF) < (compareValue & 0xF)
}

func (cpu *CPU) cpl() {
	value := cpu.Reg.A.Read()
	newValue := value ^ 0xFF

	cpu.Reg.A.Write(newValue)

	cpu.Reg.F.Subtract = true
	cpu.Reg.F.HalfCarry = true
}

// Based on https://github.com/rylev/DMG-01/blob/70fcfac0cfaf8214d03c972ef4509d4c66a44089/lib-dmg-01/src/cpu/mod.rs#L1337
// because I frankly could not find info on how this instruction works
func (cpu *CPU) daa() {
	carry := false
	value := cpu.Reg.A.Read()
	newValue := value

	if !cpu.Reg.F.Subtract {
		if cpu.Reg.F.Carry || value > 0x99 {
			carry = true
			newValue += 0x60
		}

		if cpu.Reg.F.HalfCarry || ((value & 0xF) > 0x9) {
			newValue += 0x06
		}
	} else if cpu.Reg.F.Carry {
		carry = true

		if cpu.Reg.F.HalfCarry {
			newValue += 0x9A
		} else {
			newValue += 0xA0
		}
	} else if cpu.Reg.F.HalfCarry {
		newValue += 0xFA
	}

	cpu.Reg.A.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = carry
}

func (cpu *CPU) or(value byte) byte {
	orValue := cpu.Reg.A.Read() | value
	cpu.Reg.A.Write(orValue)

	cpu.Reg.F.Zero = orValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = false
	cpu.Reg.F.HalfCarry = false

	return orValue
}

func (cpu *CPU) xor(value byte) byte {
	xorValue := cpu.Reg.A.Read() ^ value
	cpu.Reg.A.Write(xorValue)

	cpu.Reg.F.Zero = xorValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.Carry = false
	cpu.Reg.F.HalfCarry = false

	return xorValue
}

func (cpu *CPU) call(mmu *mem.MMU, opcode *isa.Opcode, shouldJump bool) (nextPC uint16, cycles uint8, err error) {
	nextPC = cpu.PC.Read() + uint16(opcode.Bytes)

	if shouldJump {
		cpu.push(mmu, nextPC)
		return mmu.Read16(cpu.PC.Read() + 1), uint8(opcode.Cycles[0]), nil
	}

	return nextPC, uint8(opcode.Cycles[1]), nil
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

func (cpu *CPU) jump(mmu *mem.MMU, opcode *isa.Opcode, shouldJump bool) (nextPC uint16, cycles uint8, err error) {
	if shouldJump {
		return cpu.readNext16(mmu), uint8(opcode.Cycles[0]), nil
	} else {
		return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[1]), nil
	}
}

func (cpu *CPU) jumpRel(mmu *mem.MMU, opcode *isa.Opcode, shouldJump bool) (nextPC uint16, cycles uint8, err error) {
	nextPC = cpu.PC.Read() + uint16(opcode.Bytes)

	if shouldJump {
		nextPcDiff := int8(cpu.readNext8(mmu))
		return nextPC + uint16(nextPcDiff), uint8(opcode.Cycles[0]), nil
	} else {
		return nextPC, uint8(opcode.Cycles[1]), nil
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

func (cpu *CPU) ret(mmu *mem.MMU, opcode *isa.Opcode, shouldJump bool) (nextPC uint16, cycles uint8, err error) {
	if shouldJump {
		return cpu.pop(mmu), uint8(opcode.Cycles[0]), nil
	} else {
		return cpu.PC.Read() + uint16(opcode.Bytes), uint8(opcode.Cycles[1]), nil
	}
}

func (cpu *CPU) rotl(value byte, zero bool, throughCarry bool) byte {
	carryBit := byte(0x0)

	if throughCarry && cpu.Reg.F.Carry {
		carryBit = 1
	} else if !throughCarry {
		carryBit = (value >> 7)
	}

	newValue := (value << 1) | carryBit
	cpu.Reg.F.Zero = zero && newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = (value & 0x80) != 0x0
	return newValue
}

func (cpu *CPU) rotr(value byte, zero bool, throughCarry bool) byte {
	carryBit := byte(0x0)

	if throughCarry && cpu.Reg.F.Carry {
		carryBit = 1 << 7
	} else if !throughCarry {
		carryBit = (value << 7)
	}

	newValue := carryBit | (value >> 1)
	cpu.Reg.F.Zero = zero && newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = (value & 0x1) != 0x0
	return newValue
}

func (cpu *CPU) rst(mmu *mem.MMU, opcode *isa.Opcode, value byte) (nextPC uint16, cycles uint8, err error) {
	cpu.push(mmu, cpu.PC.Read()+1)
	return uint16(value), uint8(opcode.Cycles[0]), nil
}

func (cpu *CPU) sla(value byte) byte {
	newValue := value << 1
	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = (value & 0x80) == 0x80
	return newValue
}

func (cpu *CPU) sra(value byte) byte {
	msb := value & 0x80
	newValue := msb | (value >> 1)
	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = (value & 0x1) == 0x1
	return newValue
}

func (cpu *CPU) srl(value byte) byte {
	newValue := value >> 1
	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = (value & 0x1) == 0x1
	return newValue
}

func (cpu *CPU) swap(reg RWByte) {
	value := reg.Read()
	high := value & 0xF0
	low := value & 0xF
	newValue := (low << 4) | (high >> 4)
	reg.Write(newValue)

	cpu.Reg.F.Zero = newValue == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = false
	cpu.Reg.F.Carry = false
}

func (cpu *CPU) testBit(bit uint8, reg RWByte) {
	mask := uint8(1 << bit)
	value := reg.Read()

	cpu.Reg.F.Zero = (value & mask) == 0
	cpu.Reg.F.Subtract = false
	cpu.Reg.F.HalfCarry = true
}

func (cpu *CPU) setBit(bit uint8, reg RWByte) {
	mask := uint8(1 << bit)
	value := reg.Read()

	reg.Write(value | mask)
}

func (cpu *CPU) resetBit(bit uint8, reg RWByte) {
	mask := ^uint8(1 << bit)
	value := reg.Read()

	reg.Write(value & mask)
}

// Did the aVal carry over from the lower 4 bits to the upper 4 bits?
func isHalfCarry8(aVal uint8, bVal uint8, carry uint8) bool {
	fourBitMask := uint(0xF)
	return ((uint(aVal) & fourBitMask) + (uint(bVal) & fourBitMask) + uint(carry)) > fourBitMask
}

func isCarry8(aVal uint8, bVal uint8, carry uint8) bool {
	byteMask := uint(0xFF)
	return ((uint(aVal) & byteMask) + (uint(bVal) & byteMask) + uint(carry)) > byteMask
}

// Did the aVal carry over from the lower 4 bits of the top byte in the word to the upper 4 bits?
func isHalfCarry16(aVal uint16, bVal uint16) bool {
	twelveBitMask := uint(0xFFF)
	return ((uint(aVal) & twelveBitMask) + (uint(bVal) & twelveBitMask)) > twelveBitMask
}

func isCarry16(aVal uint16, bVal uint16) bool {
	wordMask := uint(0xFFFF)
	return ((uint(aVal) & wordMask) + (uint(bVal) & wordMask)) > wordMask
}
