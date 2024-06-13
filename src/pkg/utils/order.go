package utils

func OrderBytes(first []byte, second []byte) int {
	if len(first) < len(second) {
		return -1
	} else if len(second) < len(first) {
		return 1
	}

	for index, firstBytes := range first {
		if firstBytes < second[index] {
			return -1
		} else if firstBytes > second[index] {
			return 1
		}
	}
	return 0
}
