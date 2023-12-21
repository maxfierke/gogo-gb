package isa

import "fmt"

type Instruction struct {
	Addr   uint16
	Opcode *Opcode
}

func (ins *Instruction) String() string {
	return fmt.Sprintf("0x%04X    %s", ins.Addr, ins.Opcode)
}
