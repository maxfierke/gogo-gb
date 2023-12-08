package cpu

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

	for k := range opcodes.Unprefixed {
		addr, err := parseOpcodeAddr(k)
		if err != nil {
			return nil, err
		}

		opcodes.Unprefixed[k].Addr = addr
	}

	for k := range opcodes.CbPrefixed {
		addr, err := parseOpcodeAddr(k)
		if err != nil {
			return nil, err
		}

		opcodes.CbPrefixed[k].Addr = addr
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

type Instruction struct {
	Addr   uint16
	Opcode *Opcode
}
