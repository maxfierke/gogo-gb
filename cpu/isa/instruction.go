package isa

import "fmt"

type Instruction struct {
	Addr   uint16
	Opcode *Opcode
}

func (opcodes *Opcodes) FromByte(value byte, prefixed bool) Instruction {
	key := fmt.Sprintf("0x%x", value)

	var opcode *Opcode

	if prefixed {
		opcode = opcodes.CbPrefixed[key]
	} else {
		opcode = opcodes.Unprefixed[key]
	}

	return Instruction{
		Addr:   uint16(value),
		Opcode: opcode,
	}
}
