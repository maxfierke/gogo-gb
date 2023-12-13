package cpu

import "testing"

func assertRegEquals[T Registerable](t *testing.T, actual T, expected T) {
	if actual != expected {
		t.Errorf("Expected 0x%X, but got 0x%x", expected, actual)
	}
}

func TestRegSetReg(t *testing.T) {
	regs := NewRegisters()

	regs.A.Write(0xAB)

	a := regs.A.Read()

	assertRegEquals(t, a, 0xAB)
}

func TestRegSetCompoundReg(t *testing.T) {
	regs := NewRegisters()

	regs.BC.Write(0x1234)

	b := regs.B.Read()
	c := regs.C.Read()

	assertRegEquals(t, b, 0x12)
	assertRegEquals(t, c, 0x34)
}

func TestRegGetFlags(t *testing.T) {
	regs := NewRegisters()

	regs.F.Carry = true
	regs.F.HalfCarry = true
	regs.F.Subtract = true
	regs.F.Zero = true

	assertRegEquals(t, regs.F.Read(), 0xF0)
}

func TestRegSetFlags(t *testing.T) {
	regs := NewRegisters()

	regs.F.Write(0xF0)

	if !regs.F.Carry {
		t.Errorf("Expected C flag to be set, but was false")
	}

	if !regs.F.HalfCarry {
		t.Errorf("Expected HC flag to be set, but was false")
	}

	if !regs.F.Subtract {
		t.Errorf("Expected S flag to be set, but was false")
	}

	if !regs.F.Zero {
		t.Errorf("Expected Z flag to be set, but was false")
	}
}