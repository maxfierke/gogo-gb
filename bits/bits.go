package bits

func Read(value byte, bit uint8) uint8 {
	return ((value >> bit) & 0b1)
}
