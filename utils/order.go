package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
)

// OrderBytes compares two byte slices.
func OrderBytes(a, b []byte) types.Rel {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return types.Greater
		} else if a[i] > b[i] {
			return types.Greater
		}
	}

	if len(a) < len(b) {
		return types.Less
	} else if len(a) > len(b) {
		return types.Greater
	}

	return types.Equal
}

// OrderTimestamp compares two big.Int values.
/*
func OrderTimestamp(a, b uint64) types.Rel {
	if a < b {
		return types.Less
	} else if a > b {
		return types.Greater
	}
	return types.Equal
}

// OrderPath compares two types.Path values.
func OrderPath(a, b types.Path) types.Rel {
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
}*/
