package kv_driver

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

// Custom comparison function for KDNodeKey
func compareKDNodeKey(a, b Kdtree.KDNodeKey) bool {
	return a.Timestamp == b.Timestamp &&
		utils.OrderSubspace(a.Subspace, b.Subspace) == 0 &&
		reflect.DeepEqual(a.Path, b.Path)
}

func TestPrefixesOf(t *testing.T) {
	// Set up the KDTree with sample values
	kdtree := Kdtree.NewKDTreeWithValues[Kdtree.KDNodeKey](3, []Kdtree.KDNodeKey{
		{Timestamp: 500, Subspace: []byte{0}, Path: types.Path{{0}}},
		{Timestamp: 600, Subspace: []byte{1}, Path: types.Path{{0}, {1}}},
		{Timestamp: 700, Subspace: []byte{0}, Path: types.Path{{1}}},
	})

	pathParams := types.PathParams[uint64]{
		MaxComponentCount:  50,
		MaxComponentLength: 200,
		MaxPathLength:      50,
	}

	// Define the path for the test
	path := types.Path{{0}, {1}, {2}, {50}}

	// Execute the PrefixesOf function
	res := DriverPrefixesOf(path, pathParams, kdtree)
	// Verify the results
	expected := []Kdtree.KDNodeKey{
		{Timestamp: 500, Subspace: []byte{0}, Path: types.Path{{0}}},
		{Timestamp: 600, Subspace: []byte{1}, Path: types.Path{{0}, {1}}},
	}

	if len(res) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(res))
	}

	for i, exp := range expected {
		if !compareKDNodeKey(res[i], exp) {
			t.Errorf("expected result %d to be %v, got %v", i, exp, res[i])
		}
	}
}

// func TestPrefixedBy(t *testing.T) {
// 	// Set up the KDTree with sample values
// 	kdtree := Kdtree.NewKDTreeWithValues[Kdtree.KDNodeKey[uint64]](3, []Kdtree.KDNodeKey[uint64]{
// 		{Timestamp: 500, Subspace: 0, Path: types.Path{{0}}},
// 		{Timestamp: 600, Subspace: 0, Path: types.Path{{0}, {10}, {99}}},
// 		{Timestamp: 700, Subspace: 0, Path: types.Path{{0}, {2}}},
// 	})

// 	pathParams := types.PathParams[uint64]{
// 		MaxComponentCount:  50,
// 		MaxComponentLength: 200,
// 		MaxPathLength:      50,
// 	}

// 	// Define the path for the test
// 	path := types.Path{{0}}
// 	subspace := 0

// 	// Execute the PrefixedBy function
// 	res := PrefixedBy(path, subspace, pathParams, kdtree)
// 	fmt.Println()
// 	// Verify the results
// 	expected := []Kdtree.KDNodeKey[uint64]{
// 		{Timestamp: 500, Subspace: 0, Path: types.Path{{0}}},
// 		{Timestamp: 600, Subspace: 0, Path: types.Path{{0}, {10}, {99}}},
// 		{Timestamp: 700, Subspace: 0, Path: types.Path{{0}, {2}}},
// 	}

// 	if len(res) != len(expected) {
// 		t.Fatalf("expected %d results, got %d", len(expected), len(res))
// 	}

// 	for i, exp := range expected {
// 		if !compareKDNodeKey(res[i], exp) {
// 			t.Errorf("expected result %d to be %v, got %v", i, exp, res[i])
// 		}
// 	}
// }
