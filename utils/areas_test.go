package utils

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeStreamAreaInArea(t *testing.T) {
	// Test case 1: Test with an open-ended range
	opts := DecodeStreamAreaInAreaOptions[uint64]{
		PathScheme:           nil,
		DecodeStreamSubspace: nil,
	}
	bytes := NewGrowingBytes(bytesChan([]byte{0x80, 0x02, 0x05, 0x01, 0x02, 0x03}))
	outer := types.Area[uint64]{
		Path:        types.Path{[]byte{0x01, 0x02}},
		Subspace_id: 1,
		Times: types.Range[uint64]{
			Start:   100,
			End:     200,
			OpenEnd: false,
		},
		Any_subspace: false,
	}

	area, err := DecodeStreamAreaInArea(opts, bytes, outer)
	assert.NoError(t, err)
	assert.Equal(t, types.Path{[]byte{0x01, 0x02}, []byte{0x03}}, area.Path)
	assert.Equal(t, uint64(1), area.Subspace_id)
	assert.Equal(t, uint64(105), area.Times.Start)
	assert.Equal(t, uint64(0), area.Times.End)
	assert.True(t, area.Times.OpenEnd)

	// Test case 2: Test with a closed range
	bytes = NewGrowingBytes(bytesChan([]byte{0x0C, 0x02, 0x05, 0x03, 0x01, 0x02, 0x03, 0x04}))
	area, err = DecodeStreamAreaInArea(opts, bytes, outer)
	assert.NoError(t, err)
	assert.Equal(t, types.Path{[]byte{0x01, 0x02}, []byte{0x03, 0x04}}, area.Path)
	assert.Equal(t, uint64(1), area.Subspace_id)
	assert.Equal(t, uint64(102), area.Times.Start)
	assert.Equal(t, uint64(107), area.Times.End)
	assert.False(t, area.Times.OpenEnd)

	// Add more test cases as needed
}

func bytesChan(bytes []byte) chan []byte {
	c := make(chan []byte)
	go func() {
		for i := 0; i < len(bytes); i++ {
			c <- []byte{bytes[i]}
		}
		// Don't close the channel here
	}()
	return c
}
