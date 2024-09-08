package debug

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/maxfierke/gogo-gb/cpu"
	"github.com/maxfierke/gogo-gb/mem"
	"github.com/samber/lo"
)

var ErrConsoleNotAttached = errors.New("debugger must be attached to a running console")

var registerNames = []string{
	"A", "F", "B", "C", "D", "E", "H", "L",
	"AF", "BC", "DE", "HL", "SP", "PC",
}

type InteractiveDebugger struct {
	breakpoints []uint16
	isStepping  bool
	shell       *ishell.Shell
}

var _ Debugger = (*InteractiveDebugger)(nil)

func NewInteractiveDebugger() (*InteractiveDebugger, error) {
	debugger := &InteractiveDebugger{
		breakpoints: []uint16{},
	}

	shell := ishell.New()
	shell.Println("gogo-gb interactive debugger")

	shell.AddCmd(&ishell.Cmd{
		Name:    "break",
		Aliases: []string{"br", "b"},
		Help:    "Set breakpoint at address",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("must provide an address"))
			}

			addr, err := parseAddr(c.Args[0])
			if err != nil {
				c.Err(fmt.Errorf("parsing addr %w:", err))
				return
			}

			if !slices.Contains(debugger.breakpoints, addr) {
				debugger.breakpoints = append(debugger.breakpoints, addr)
				c.Printf("added breakpoint @ 0x%02X\n", addr)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "continue",
		Aliases: []string{"c"},
		Help:    "Continue execution until next breakpoint",
		Func: func(c *ishell.Context) {
			debugger.isStepping = false
			c.Stop()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "step",
		Aliases: []string{"s"},
		Help:    "Execute the next instruction",
		Func: func(c *ishell.Context) {
			debugger.isStepping = true
			c.Stop()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "examine",
		Aliases: []string{"x"},
		Help:    "Examine value at address",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("must provide an address"))
				return
			}

			if slices.Contains(registerNames, c.Args[0]) {
				regName := c.Args[0]

				cpu, err := getCPU(c)
				if err != nil {
					c.Err(fmt.Errorf("accessing cpu: %w", err))
					return
				}

				switch regName {
				case "AF", "BC", "DE", "HL":
					c.Printf("%s: %04X\n", regName, getCompoundRegister(cpu, regName).Read())
				case "SP":
					c.Printf("%s: %04X\n", regName, cpu.SP.Read())
				case "PC":
					c.Printf("%s: %04X\n", regName, cpu.PC.Read())
				default:
					c.Printf("%s: %02X\n", regName, getRegister(cpu, regName).Read())
				}
			} else {
				addr, err := parseAddr(c.Args[0])
				if err != nil {
					c.Err(fmt.Errorf("parsing addr %w:", err))
					return
				}

				mmu, err := getMMU(c)
				if err != nil {
					c.Err(fmt.Errorf("accessing mmu: %w", err))
					return
				}

				if len(c.Args) > 1 && c.Args[1] == "16" {
					c.Printf("0x%04X: %04X\n", addr, mmu.Read16(addr))
				} else {
					c.Printf("0x%04X: %02X\n", addr, mmu.Read8(addr))
				}
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "disassemble",
		Aliases: []string{"d", "di", "dis"},
		Help:    "Disassemble at address",
		Func: func(c *ishell.Context) {
			var addr uint16

			cpu, err := getCPU(c)
			if err != nil {
				c.Err(fmt.Errorf("accessing cpu: %w", err))
				return
			}

			mmu, err := getMMU(c)
			if err != nil {
				c.Err(fmt.Errorf("accessing mmu: %w", err))
				return
			}

			if len(c.Args) == 0 {
				addr = cpu.PC.Read()
			} else {
				addr, err = parseAddr(c.Args[0])
				if err != nil {
					c.Err(fmt.Errorf("parsing addr %w:", err))
					return
				}
			}

			inst, err := cpu.FetchAndDecode(mmu, addr)
			if err != nil {
				c.Err(fmt.Errorf("disassembling: %w", err))
				return
			}

			operands := make([]string, 0, len(inst.Opcode.Operands))

			operandOffset := 1
			for _, operand := range inst.Opcode.Operands {
				switch {
				case operand.Bytes == 1 && operand.Immediate:
					operands = append(operands, fmt.Sprintf("$%02X", mmu.Read8(addr+1)))
				case operand.Bytes == 1 && !operand.Immediate:
					operands = append(operands, fmt.Sprintf("($%02X)", mmu.Read8(addr+1)))
				case operand.Bytes == 2 && operand.Immediate:
					operands = append(operands, fmt.Sprintf("$%04X", mmu.Read16(addr+1)))
				case operand.Bytes == 2 && !operand.Immediate:
					operands = append(operands, fmt.Sprintf("($%04X)", mmu.Read16(addr+1)))
				default:
					operands = append(operands, operand.String())
				}

				operandOffset += operand.Bytes
			}

			c.Printf(
				"0x%04X    0x%02X %s %s\n",
				inst.Addr,
				inst.Opcode.Addr,
				inst.Opcode.Mnemonic,
				strings.Join(operands, ", "),
			)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "registers",
		Aliases: []string{"r", "regs"},
		Help:    "Print registers and other CPU state",
		Func: func(c *ishell.Context) {
			cpu, err := getCPU(c)
			if err != nil {
				c.Err(fmt.Errorf("accessing cpu: %w", err))
				return
			}

			mmu, err := getMMU(c)
			if err != nil {
				c.Err(fmt.Errorf("accessing mmu: %w", err))
				return
			}

			debugger.printState(cpu, mmu)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "list",
		Aliases: []string{"ls"},
		Help:    "List set breakpoints",
		Func: func(c *ishell.Context) {
			c.Println("Active breakpoints:")
			for i, breakpoint := range debugger.breakpoints {
				c.Printf("* %d: 0x%02X\n", i, breakpoint)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "clear",
		Aliases: []string{"cl"},
		Help:    "Clear breakpoint at address",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("must provide an address"))
				return
			}

			addr, err := parseAddr(c.Args[0])
			if err != nil {
				c.Err(fmt.Errorf("parsing addr %w:", err))
				return
			}

			if !slices.Contains(debugger.breakpoints, addr) {
				debugger.breakpoints = lo.Filter(debugger.breakpoints, func(b uint16, _ int) bool {
					return b == addr
				})
				c.Printf("cleared breakpoint @ 0x%02X\n", addr)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:    "reset",
		Aliases: []string{"res"},
		Help:    "Reset the CPU",
		Func: func(c *ishell.Context) {
			cpu, err := getCPU(c)
			if err != nil {
				c.Err(fmt.Errorf("accessing cpu: %w", err))
				return
			}

			cpu.Reset()
		},
	})

	debugger.shell = shell

	return debugger, nil
}

func (i *InteractiveDebugger) OnDecode(cpu *cpu.CPU, mmu *mem.MMU) {
	if slices.Contains(i.breakpoints, cpu.PC.Read()) {
		i.shell.Printf("reached 0x%02X\n", cpu.PC.Read())
		i.attachShell(cpu, mmu)
	} else if i.isStepping {
		i.attachShell(cpu, mmu)
	}
}

func (i *InteractiveDebugger) OnExecute(cpu *cpu.CPU, mmu *mem.MMU) {
}

func (i *InteractiveDebugger) OnInterrupt(cpu *cpu.CPU, mmu *mem.MMU) {
}

func (i *InteractiveDebugger) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	return mem.ReadPassthrough()
}

func (i *InteractiveDebugger) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	return mem.WritePassthrough()
}

func (i *InteractiveDebugger) Setup(cpu *cpu.CPU, mmu *mem.MMU) {
	i.attachShell(cpu, mmu)
}

func (i *InteractiveDebugger) attachShell(cpu *cpu.CPU, mmu *mem.MMU) {
	i.shell.Set("cpu", cpu)
	i.shell.Set("mmu", mmu)

	// Remove references to CPU & MMU once we're done, since control will pass
	// back to the console and we shouldn't hold onto references while emulation
	// is running
	defer i.shell.Del("cpu")
	defer i.shell.Del("mmu")

	i.shell.Run()
}

func (i *InteractiveDebugger) printState(cpu *cpu.CPU, mmu *mem.MMU) {
	i.shell.Printf("Registers:\n")
	i.shell.Printf(
		" A: %02X    F: %02X    AF: %04X\n",
		cpu.Reg.A.Read(),
		cpu.Reg.F.Read(),
		cpu.Reg.AF.Read(),
	)
	i.shell.Printf(
		" B: %02X    C: %02X    BC: %04X\n",
		cpu.Reg.B.Read(),
		cpu.Reg.C.Read(),
		cpu.Reg.BC.Read(),
	)
	i.shell.Printf(
		" D: %02X    E: %02X    DE: %04X\n",
		cpu.Reg.D.Read(),
		cpu.Reg.E.Read(),
		cpu.Reg.DE.Read(),
	)
	i.shell.Printf(
		" H: %02X    L: %02X    HL: %04X\n",
		cpu.Reg.H.Read(),
		cpu.Reg.L.Read(),
		cpu.Reg.HL.Read(),
	)
	i.shell.Printf("Flags:\n")
	i.shell.Printf(
		" Z: %t N: %t H: %t C: %t\n",
		cpu.Reg.F.Zero,
		cpu.Reg.F.Subtract,
		cpu.Reg.F.HalfCarry,
		cpu.Reg.F.Carry,
	)
	i.shell.Printf("Program state:\n")
	i.shell.Printf(
		"SP: %04X PC: %04X PCMEM: %02X,%02X,%02X,%02X\n",
		cpu.SP.Read(),
		cpu.PC.Read(),
		mmu.Read8(cpu.PC.Read()),
		mmu.Read8(cpu.PC.Read()+1),
		mmu.Read8(cpu.PC.Read()+2),
		mmu.Read8(cpu.PC.Read()+3),
	)
}

func parseAddr(addrString string) (uint16, error) {
	addrString = strings.TrimPrefix(addrString, "$")
	addrString = strings.TrimPrefix(addrString, "0x")

	parsedAddr, err := strconv.ParseUint(addrString, 16, 16)
	if err != nil {
		return 0, err
	}
	return uint16(parsedAddr), nil
}

func getCPU(c *ishell.Context) (*cpu.CPU, error) {
	cpu, ok := c.Get("cpu").(*cpu.CPU)
	if !ok {
		return nil, ErrConsoleNotAttached
	}
	return cpu, nil
}

func getMMU(c *ishell.Context) (*mem.MMU, error) {
	mmu, ok := c.Get("mmu").(*mem.MMU)
	if !ok {
		return nil, ErrConsoleNotAttached
	}

	return mmu, nil
}

func getRegister(cpu *cpu.CPU, regName string) *cpu.Register[uint8] {
	switch regName {
	case "A":
		return cpu.Reg.A
	case "B":
		return cpu.Reg.B
	case "C":
		return cpu.Reg.C
	case "D":
		return cpu.Reg.D
	case "E":
		return cpu.Reg.E
	case "H":
		return cpu.Reg.H
	case "L":
		return cpu.Reg.L
	default:
		panic("tried to access non-existent register")
	}
}

func getCompoundRegister(cpu *cpu.CPU, regName string) *cpu.CompoundRegister {
	switch regName {
	case "AF":
		return cpu.Reg.AF
	case "BC":
		return cpu.Reg.BC
	case "DE":
		return cpu.Reg.DE
	case "HL":
		return cpu.Reg.HL
	}

	panic("tried to access non-existent register")
}
