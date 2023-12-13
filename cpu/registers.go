package cpu

const (
	zeroFlagBit      = 7
	subtractFlagBit  = 6
	halfCarryFlagBit = 5
	carryFlagBit     = 4
)

type Flags struct {
	Zero      bool
	Subtract  bool
	HalfCarry bool
	Carry     bool
}

func (flags *Flags) Read() uint8 {
	var (
		zero      uint8
		subtract  uint8
		halfCarry uint8
		carry     uint8
	)

	if flags.Zero {
		zero = 1 << zeroFlagBit
	} else {
		zero = 0
	}

	if flags.Subtract {
		subtract = 1 << subtractFlagBit
	} else {
		subtract = 0
	}

	if flags.HalfCarry {
		halfCarry = 1 << halfCarryFlagBit
	} else {
		halfCarry = 0
	}

	if flags.Carry {
		carry = 1 << carryFlagBit
	} else {
		carry = 0
	}

	return (zero | subtract | halfCarry | carry)
}

func (flags *Flags) Write(value uint8) {
	flags.Zero = ((value >> zeroFlagBit) & 0b1) != 0
	flags.Subtract = ((value >> subtractFlagBit) & 0b1) != 0
	flags.HalfCarry = ((value >> halfCarryFlagBit) & 0b1) != 0
	flags.Carry = ((value >> carryFlagBit) & 0b1) != 0
}

type Registerable interface {
	uint8 | uint16
}

type Register[T Registerable] struct {
	name  string
	value T
}

func (reg *Register[T]) Read() T {
	return reg.value
}

func (reg *Register[T]) Write(value T) {
	reg.value = value
}

func (reg *Register[T]) Inc(value T) {
	reg.value += value
}

func (reg *Register[T]) Dec(value T) {
	reg.value -= value
}

type RWByte interface {
	Read() byte
	Write(value byte)
}

type RWTwoByte interface {
	Read() uint16
	Write(value uint16)
}

type CompoundRegister struct {
	name string
	high RWByte
	low  RWByte
}

func (reg *CompoundRegister) Read() uint16 {
	return (uint16(reg.high.Read()) << 8) | uint16(reg.low.Read())
}

func (reg *CompoundRegister) Write(value uint16) {
	reg.high.Write(uint8((value & 0xFF00) >> 8))
	reg.low.Write(uint8(value & 0xFF))
}

func (reg *CompoundRegister) Inc(value uint16) {
	reg.Write(reg.Read() + value)
}

func (reg *CompoundRegister) Dec(value uint16) {
	reg.Write(reg.Read() - value)
}

type Registers struct {
	A *Register[uint8]
	B *Register[uint8]
	C *Register[uint8]
	D *Register[uint8]
	E *Register[uint8]
	F *Flags
	H *Register[uint8]
	L *Register[uint8]

	AF *CompoundRegister
	BC *CompoundRegister
	DE *CompoundRegister
	HL *CompoundRegister
}

func NewRegisters() Registers {
	a := &Register[uint8]{name: "A", value: 0x00}
	b := &Register[uint8]{name: "B", value: 0x00}
	c := &Register[uint8]{name: "C", value: 0x00}
	d := &Register[uint8]{name: "D", value: 0x00}
	e := &Register[uint8]{name: "E", value: 0x00}
	f := &Flags{
		Zero:      false,
		Subtract:  false,
		HalfCarry: false,
		Carry:     false,
	}
	h := &Register[uint8]{name: "H", value: 0x00}
	l := &Register[uint8]{name: "L", value: 0x00}

	return Registers{
		A:  a,
		B:  b,
		C:  c,
		D:  d,
		E:  e,
		F:  f,
		H:  h,
		L:  l,
		AF: &CompoundRegister{name: "AF", high: a, low: f},
		BC: &CompoundRegister{name: "BC", high: b, low: c},
		DE: &CompoundRegister{name: "DE", high: d, low: e},
		HL: &CompoundRegister{name: "HL", high: h, low: l},
	}
}

func (regs *Registers) Reset() {
	regs.AF.Write(0x0000)
	regs.BC.Write(0x0000)
	regs.DE.Write(0x0000)
	regs.HL.Write(0x0000)
}
