package decoding

func CompactWidthFromEndOfByte(byte int) int {
	if (byte & 0x3) == 0x3 {
		return 8
	} else if (byte & 0x2) == 0x2 {
		return 4
	} else if (byte & 0x1) == 0x1 {
		return 2
	}

	return 1
}
