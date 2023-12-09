package isa

import (
	"encoding/hex"
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
	if !operand.Immediate {
		return fmt.Sprintf("(%s)", operand.Name)
	}
	return operand.Name
}

type Opcode struct {
	Addr      uint8
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

	return fmt.Sprintf("%s %s %s", opcode.Mnemonic, strings.Join(operands, ", "), comment)
}

type OpcodesJSON struct {
	Unprefixed map[string]*Opcode `json:"unprefixed"`
	CbPrefixed map[string]*Opcode `json:"cbprefixed"`
}

type Opcodes struct {
	Unprefixed map[uint8]*Opcode
	CbPrefixed map[uint8]*Opcode
}

func (opcodes *Opcodes) InstructionFromByte(value byte, prefixed bool) (*Instruction, bool) {
	var opcode *Opcode
	var present bool

	if prefixed {
		opcode, present = opcodes.CbPrefixed[value]
	} else {
		opcode, present = opcodes.Unprefixed[value]
	}

	if !present {
		return nil, present
	}

	return &Instruction{
		Addr:   uint16(value),
		Opcode: opcode,
	}, true
}

func LoadOpcodes(path string) (*Opcodes, error) {
	jsonBytes, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var opcodesJSON OpcodesJSON
	err = json.Unmarshal(jsonBytes, &opcodesJSON)
	if err != nil {
		return nil, err
	}

	opcodes := Opcodes{
		Unprefixed: map[uint8]*Opcode{},
		CbPrefixed: map[uint8]*Opcode{},
	}

	for k := range opcodesJSON.Unprefixed {
		addr, err := parseOpcodeAddr(k)
		if err != nil {
			return nil, err
		}

		opcode := opcodesJSON.Unprefixed[k]
		opcode.Addr = addr
		opcodes.Unprefixed[addr] = opcode
	}

	for k := range opcodesJSON.CbPrefixed {
		addr, err := parseOpcodeAddr(k)
		if err != nil {
			return nil, err
		}

		opcode := opcodesJSON.CbPrefixed[k]
		opcode.Addr = addr
		opcodes.CbPrefixed[addr] = opcode
	}

	return &opcodes, nil
}

func parseOpcodeAddr(key string) (uint8, error) {
	hexStr, _ := strings.CutPrefix(key, "0x")

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0x00, err
	}

	return decoded[0], nil
}
