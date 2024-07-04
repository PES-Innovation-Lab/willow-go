package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
	"testing"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

func EncodeSubspaceIdLength[SubspaceId, T constraints.Ordered](subspace SubspaceId) uint64 {
	// Replace this with your actual implementation logic
	// This function should calculate the encoded length of the given subspace identifier.

	// Example (if SubspaceId is a string):
	// For simplicity, assuming the encoded length is the length of the string.
	return len(bits.Len(subspace))

	// Adjust the logic based on your encoding scheme for SubspaceId.
}

func TestIntersectRange(t *testing.T) {
	// Define a TotalOrder function for int
	order := func(a, b int) types.Rel {
		if a < b {
			return types.Less
		} else if a > b {
			return types.Greater
		}
		return types.Equal
	}

	t.Run("BothRangesOpenEnded", func(t *testing.T) {
		a := types.Range[int]{
			Start:   10,
			OpenEnd: true,
		}
		b := types.Range[int]{
			Start:   5,
			OpenEnd: true,
		}

		intersected, result := IntersectRange(order, a, b)
		assert.True(t, intersected)
		assert.Equal(t, types.Range[int]{Start: 10, OpenEnd: true}, result)
	})

	t.Run("AOpenEndedBNotOpenEnded", func(t *testing.T) {
		a := types.Range[int]{
			Start:   10,
			OpenEnd: true,
		}
		b := types.Range[int]{
			Start:   5,
			End:     15,
			OpenEnd: false,
		}

		intersected, result := IntersectRange(order, a, b)
		assert.True(t, intersected)
		assert.Equal(t, types.Range[int]{Start: 10, End: 15, OpenEnd: false}, result)
	})

	t.Run("BothRangesClosedEnded", func(t *testing.T) {
		a := types.Range[int]{
			Start:   10,
			End:     20,
			OpenEnd: false,
		}
		b := types.Range[int]{
			Start:   15,
			End:     25,
			OpenEnd: false,
		}

		intersected, result := IntersectRange(order, a, b)
		assert.True(t, intersected)
		assert.Equal(t, types.Range[int]{Start: 15, End: 20, OpenEnd: false}, result)
	})

	t.Run("NoIntersection", func(t *testing.T) {
		a := types.Range[int]{
			Start:   10,
			End:     15,
			OpenEnd: false,
		}
		b := types.Range[int]{
			Start:   20,
			End:     25,
			OpenEnd: false,
		}

		intersected, result := IntersectRange(order, a, b)
		assert.False(t, intersected)
		assert.Equal(t, types.Range[int]{}, result)
	})
}

func EncodeSubspaceId(subspaceId uint64) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int64(subspaceId)); err != nil {
		return nil
	}
	return buf.Bytes()
}

// DecodeSubspaceId takes a []byte and returns an int.
func DecodeSubspaceId(encoded []byte) int {
	buf := bytes.NewReader(encoded)
	var subspaceId int64
	if err := binary.Read(buf, binary.BigEndian, &subspaceId); err != nil {
		return 0
	}
	return int(subspaceId)
}

func createPath(paths ...string) types.Path {
	var p types.Path
	for _, path := range paths {
		p = append(p, []byte(path))
	}
	return p
}

func TestEncodeDecodeRange3dRelative(t *testing.T) {
	order := func(a, b uint64) types.Rel {
		if a < b {
			return types.Less
		} else if a > b {
			return types.Greater
		}
		return types.Equal
	}

	pathScheme := types.PathParams[uint64]{
		MaxComponentCount:  4,
		MaxComponentLength: 8,
		MaxPathLength:      32,
	}

	now := uint64(time.Now().Unix())
	tests := []struct {
		name string
		r    types.Range3d[uint64]
		r1   types.Range3d[uint64]
		want types.Range3d[uint64]
	}{
		{
			name: "Test1",
			r: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 1, End: 3, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: createPath("path1"), End: createPath("path1"), OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: now, End: now + 1000, OpenEnd: false},
			},
			r1: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 1, End: 3, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: createPath("path1"), End: createPath("path20"), OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: now - 1000, End: now - 500, OpenEnd: false},
			},
			want: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 1, End: 2, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: createPath("path1"), End: createPath("path1"), OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: now, End: now + 1000, OpenEnd: false},
			},
		},
		{
			name: "Test2",
			r: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 13, End: 14, OpenEnd: true},
				PathRange:     types.Range[types.Path]{Start: createPath("path5"), End: createPath("path5"), OpenEnd: true},
				TimeRange:     types.Range[uint64]{Start: now + 2000, End: now + 3000, OpenEnd: true},
			},
			r1: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 13, End: 14, OpenEnd: true},
				PathRange:     types.Range[types.Path]{Start: createPath("path5"), End: createPath("path20"), OpenEnd: true},
				TimeRange:     types.Range[uint64]{Start: now + 4000, End: now + 6000, OpenEnd: true},
			},
			want: types.Range3d[uint64]{
				SubspaceRange: types.Range[uint64]{Start: 13, End: 14, OpenEnd: true},
				PathRange:     types.Range[types.Path]{Start: createPath("path5"), End: createPath("path5"), OpenEnd: true},
				TimeRange:     types.Range[uint64]{Start: now + 2000, End: now + 3000, OpenEnd: true},
			},
		},
	}

	for _, cases := range tests {
		encoded := EncodeRange3dRelative(order, EncodeSubspaceId, pathScheme, cases.r, cases.r1)
		fmt.Printf("%v+", encoded)
		decode := DecodeRange3dRelative(DecodeSubspaceId, EncodeSubspaceIdLength, pathScheme, encoded, cases.r1)

		fmt.Printf("+v \n +v", decode, cases.want)
	}
}
