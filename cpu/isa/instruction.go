package isa

type Instruction struct {
	Addr   uint16
	Opcode *Opcode
}
