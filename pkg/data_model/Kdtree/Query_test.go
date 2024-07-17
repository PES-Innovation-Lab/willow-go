package Kdtree

import (
	"fmt"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
)

// Custom comparison function for KDNodeKey
// func compareKDNodeKey(a, b KDNodeKey) bool {
// 	return a.Timestamp == b.Timestamp &&
// 		utils.OrderSubspace(a.Subspace, b.Subspace) == 0 &&
// 		reflect.DeepEqual(a.Path, b.Path)
// }

func TestQuery(t *testing.T) {
	// Set up the KDTree with sample values
	kdtree := NewKDTreeWithValues[KDNodeKey](3, []KDNodeKey{
		{
			Timestamp: 500,
			Subspace:  []byte{0},
			Path:      types.Path{{0}},
		},
		{
			Timestamp: 600,
			Subspace:  []byte{0},
			Path:      types.Path{{2}, {10}, {99}},
		},
		{
			Timestamp: 700,
			Subspace:  []byte{0},
			Path:      types.Path{{1}},
		},
	})

	// Define the query range
	subspaceRange := types.Range[types.SubspaceId]{
		Start:   []byte{0},
		End:     []byte{10},
		OpenEnd: false,
	}

	pathRange := types.Range[types.Path]{
		Start:   types.Path{{0}},
		End:     types.Path{{1}},
		OpenEnd: false,
	}

	timeRange := types.Range[uint64]{
		Start:   0,
		End:     2000,
		OpenEnd: true,
	}

	range3d := types.Range3d{
		SubspaceRange: subspaceRange,
		PathRange:     pathRange,
		TimeRange:     timeRange,
	}

	// Execute the query
	res := Query(kdtree, range3d)

	fmt.Println(res)
	// Verify the results
	// expected := []Kdtree.KDNodeKey[uint64]{
	// 	{
	// 		Timestamp: 1000,
	// 		Subspace:  3,
	// 		Path:      types.Path{{5}, {6}, {7}},
	// 	},
	// 	{
	// 		Timestamp: 1100,
	// 		Subspace:  1,
	// 		Path:      types.Path{{6}},
	// 	},
	// }

	// if len(res) != len(expected) {
	// 	t.Fatalf("expected %d results, got %d", len(expected), len(res))
	// }

	// for i, exp := range expected {
	// 	if !compareKDNodeKey(res[i], exp) {
	// 		t.Errorf("expected result %d to be %v, got %v", i, exp, res[i])
	// 	}
	// }
}
