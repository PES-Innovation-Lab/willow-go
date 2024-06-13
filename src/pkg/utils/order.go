package utils

import (
	"math/big"
	"src/pkg/types"
)

// OrderBytes compares two byte slices.
func OrderBytes(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}

	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}

	return 0
}

// OrderTimestamp compares two big.Int values.
func OrderTimestamp(a, b *big.Int) int {
	return a.Cmp(b)
}

// OrderPath compares two types.Path values.
func OrderPath(a, b types.Path) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		order := OrderBytes(a[i], b[i])
		if order != 0 {
			return order
		}
	}

	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}

	return 0
}
