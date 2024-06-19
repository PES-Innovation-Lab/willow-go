package tests

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

func TestIsValidRange[T constraints.Ordered](t *testing.T) {
	assert := assert.New(t)

	// Test cases
	tests := []struct {
		name   string
		order  types.TotalOrder[T]
		r      types.Range[int]
		expect bool
	}{

		{"Open end range", OrderTimestamp, Range[int]{Start: 0, End: 10, OpenEnd: true}, true},
		{"Valid range", OrderTimestamp, Range[int]{Start: 0, End: 10, OpenEnd: false}, true},
		{"Invalid range", OrderTimestamp, Range[int]{Start: 10, End: 0, OpenEnd: false}, false},
		{"Equal start and end", OrderTimestamp, Range[int]{Start: 10, End: 10, OpenEnd: false}, false},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRange(tt.order, tt.r)
			assert.Equal(tt.expect, result)
		})
	}
}
