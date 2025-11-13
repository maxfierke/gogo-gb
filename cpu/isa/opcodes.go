package isa

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
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
	Increment bool   `json:"increment,omitempty"`
	Decrement bool   `json:"decrement,omitempty"`
	Bytes     int    `json:"bytes,omitempty"`
}

func (operand *Operand) String() string {
	name := operand.Name

	if operand.Increment {
		name = name + "+"
	} else if operand.Decrement {
		name = name + "-"
	}

	if !operand.Immediate {
		return fmt.Sprintf("(%s)", name)
	}

	return name
}

type Opcode struct {
	Addr       uint8
	CbPrefixed bool
	Mnemonic   string       `json:"mnemonic"`
	Bytes      int          `json:"bytes"`
	Cycles     []int        `json:"cycles"`
	Operands   []Operand    `json:"operands"`
	Immediate  bool         `json:"immediate"`
	Flags      OperandFlags `json:"flags"`
}

func (opcode *Opcode) String() string {
	operands := make([]string, 0, len(opcode.Operands))

	for i := range opcode.Operands {
		operandText := opcode.Operands[i].String()
		operands = append(operands, operandText)
	}

	return fmt.Sprintf(
		"0x%02X %s %s",
		opcode.Addr,
		opcode.Mnemonic,
		strings.Join(operands, ", "),
	)
}

type OpcodesJSON struct {
	Unprefixed map[string]*Opcode `json:"unprefixed"`
	CbPrefixed map[string]*Opcode `json:"cbprefixed"`
}

type Opcodes struct {
	Unprefixed map[uint8]*Opcode
	CbPrefixed map[uint8]*Opcode
}

func (opcodes *Opcodes) InstructionFromByte(addr uint16, value byte, prefixed bool) (*Instruction, bool) {
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
		Addr:   addr,
		Opcode: opcode,
	}, true
}

func (opcodes *Opcodes) DebugPrint(w io.Writer) {
	fmt.Fprint(w, "== Opcodes ==\n")

	fmt.Fprintf(w, "=== 8-bit opcodes: \n\n")

	unprefixedOpcodes := slices.Collect(maps.Keys(opcodes.Unprefixed))
	slices.Sort(unprefixedOpcodes)

	for _, k := range unprefixedOpcodes {
		fmt.Fprintf(w, "%s\n", opcodes.Unprefixed[k].String())
	}

	cbPrefixedOpcodes := slices.Collect(maps.Keys(opcodes.CbPrefixed))
	slices.Sort(cbPrefixedOpcodes)

	fmt.Fprintf(w, "\n=== 16-bit opcodes: \n\n")
	for _, k := range cbPrefixedOpcodes {
		fmt.Fprintf(w, "0xCB %s\n", opcodes.CbPrefixed[k].String())
	}
}

func LoadOpcodes() (*Opcodes, error) {
	return parseOpcodeJson(opcodeJsonBytes)
}

//go:embed opcodes.json
var opcodeJsonBytes []byte

func parseOpcodeJson(jsonBytes []byte) (*Opcodes, error) {
	var opcodesJSON OpcodesJSON
	err := json.Unmarshal(jsonBytes, &opcodesJSON)
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
		opcode.CbPrefixed = true
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
