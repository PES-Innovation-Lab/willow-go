package encoding

var CompactWidthEndMasks = map[int]int{
	1: 0x0,
	2: 0x1,
	4: 0x2,
	8: 0x3,
}

func CompactWidthOr(byte int, compactWidth int) int {
	return byte | CompactWidthEndMasks[compactWidth]
}
