package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type OperandFlags struct {
	Z string `json:"Z"`
	N string `json:"N"`
	H string `json:"H"`
	C string `json:"C"`
}

type Operand struct {
	Name      string `json:"name"`
	Immediate bool   `json:"immediate"`
	Bytes     int    `json:"bytes,omitempty"`
}

func (operand *Operand) String() string {
	return operand.Name
}

type Opcode struct {
	Mnemonic  string `json:"mnemonic"`
	Comment   string
	Bytes     int          `json:"bytes"`
	Cycles    []int        `json:"cycles"`
	Operands  []Operand    `json:"operands"`
	Immediate bool         `json:"immediate"`
	Flags     OperandFlags `json:"flags"`
}

func (opcode *Opcode) String() string {
	operands := make([]string, 0, len(opcode.Operands))
	for i := range opcode.Operands {
		operandText := opcode.Operands[i].String()
		operands = append(operands, operandText)
	}

	var comment string

	if opcode.Comment != "" {
		comment = fmt.Sprintf("; %s", opcode.Comment)
	}

	return fmt.Sprintf("%s %s %s", opcode.Mnemonic, strings.Join(operands, ","), comment)
}

type Opcodes struct {
	Unprefixed map[string]*Opcode `json:"unprefixed"`
	CbPrefixed map[string]*Opcode `json:"cbprefixed"`
}

func LoadOpcodes(path string) (*Opcodes, error) {
	jsonBytes, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var opcodes Opcodes
	err = json.Unmarshal(jsonBytes, &opcodes)
	if err != nil {
		return nil, err
	}

	return &opcodes, nil
}
