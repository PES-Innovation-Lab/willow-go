package decoding

func CompactWidthFromEndOfByte(bytes int) int {
	if (bytes & 0x3) == 0x3 {
		return 8
	} else if (bytes & 0x2) == 0x2 {
		return 4
	} else if (bytes & 0x1) == 0x1 {
		return 2
	}
	return 1
}
