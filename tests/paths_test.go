package tests

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TestPrefixesOf(t *testing.T) {
	type PrefixesOfVector struct {
		Path     types.Path
		Prefixes []types.Path
	}

	PrefixesOfVectors := []PrefixesOfVector{
		{
			Path: types.Path{},
			Prefixes: []types.Path{
				{},
			},
		},
		{
			Path: types.Path{make([]byte, 2)},
			Prefixes: []types.Path{
				{},
				{make([]byte, 2)},
			},
		},
		{
			Path: types.Path{
				make([]byte, 2),
				make([]byte, 3),
				make([]byte, 4),
			},
			Prefixes: []types.Path{
				{},
				{make([]byte, 2)},
				{make([]byte, 2), make([]byte, 3)},
				{make([]byte, 2), make([]byte, 3), make([]byte, 4)},
			},
		},
	}

	for _, vector := range PrefixesOfVectors {
		actual := utils.PrefixesOf(vector.Path)
		expected := vector.Prefixes
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("For path %v, expected prefixes %v, but got %v", vector.Path, expected, actual)
		}
	}
}

func TestIsValidPath(t *testing.T) {
	type ValidPathVector struct {
		Path               types.Path
		MaxComponentCount  uint
		MaxComponentLength uint
		MaxPathLength      uint
		ExpectedResult     bool
	}

	// Define your validPathVectors as a slice of ValidPathVector
	validPathVectors := []ValidPathVector{
		{
			Path:               types.Path{{0}},
			MaxComponentCount:  1,
			MaxComponentLength: 1,
			MaxPathLength:      1,
			ExpectedResult:     true,
		},
		{
			Path:               types.Path{{0}},
			MaxComponentCount:  0,
			MaxComponentLength: 0,
			MaxPathLength:      0,
			ExpectedResult:     false,
		},
		{
			Path:               types.Path{{0}, {0}},
			MaxComponentCount:  1,
			MaxComponentLength: 1,
			MaxPathLength:      2,
			ExpectedResult:     false,
		},
		{
			Path:               types.Path{{0}, {0, 255}},
			MaxComponentCount:  2,
			MaxComponentLength: 1,
			MaxPathLength:      3,
			ExpectedResult:     false,
		},
		{
			Path:               types.Path{{0}, {0, 255}},
			MaxComponentCount:  2,
			MaxComponentLength: 2,
			MaxPathLength:      1,
			ExpectedResult:     false,
		},
	}

	for _, val := range validPathVectors {
		valid, _ := utils.IsValidPath[uint](val.Path, types.PathParams[uint]{
			MaxComponentcount:  val.MaxComponentCount,
			MaxComponentLength: val.MaxComponentLength,
			MaxPathLength:      val.MaxPathLength,
		})
		if valid != val.ExpectedResult {
			t.Errorf("Test IsValid failed!!")
		}
	}
}

func TestIsPathPrefixed(t *testing.T) {
	type PrefixPathVector struct {
		Path           types.Path
		Prefix         types.Path
		ExpectedResult bool
	}

	prefixVectors := []PrefixPathVector{
		{
			Path:           types.Path{make([]byte, 1)},
			Prefix:         types.Path{make([]byte, 1)},
			ExpectedResult: true,
		},
		{
			Path:           types.Path{{0}, {2}},
			Prefix:         types.Path{make([]byte, 1)},
			ExpectedResult: true,
		},
		{
			Path:           types.Path{{0}, {2}},
			Prefix:         types.Path{make([]byte, 1)},
			ExpectedResult: true,
		},
		{
			Path:           types.Path{{1}, {2}, {3}},
			Prefix:         types.Path{{1}, {2}, {3}, {4}},
			ExpectedResult: false,
		},
	}

	for _, vector := range prefixVectors {
		result, _ := utils.IsPathPrefixed(vector.Prefix, vector.Path)
		if result != vector.ExpectedResult {
			t.Error("Test isPrefixed failed!!")
		}
	}
}

func TestCommonPrefix(t *testing.T) {
	type PrefixPathVector struct {
		Path1    types.Path
		Path2    types.Path
		Expected types.Path
	}
	prefixVectors := []PrefixPathVector{
		{
			Path1:    types.Path{{0}, {1}, {2}},
			Path2:    types.Path{{0}, {1}, {2}, {3}},
			Expected: types.Path{{0}, {1}, {2}},
		},
		{
			Path1:    types.Path{{0}},
			Path2:    types.Path{{0}},
			Expected: types.Path{{0}},
		},
		{
			Path1:    types.Path{{0}, {1}, {2}},
			Path2:    types.Path{{1}, {2}, {3}},
			Expected: types.Path{},
		},
		{
			Path1:    types.Path{{0}, {1}, {2}, {4}},
			Path2:    types.Path{{0}, {1}, {3}, {2}},
			Expected: types.Path{{0}, {1}},
		},
	}
	for _, vector := range prefixVectors {
		result, _ := utils.CommonPrefix(vector.Path1, vector.Path2)
		for index, component := range result {
			if utils.OrderBytes(result[index], component) != 0 {
				t.Errorf("Test failed! Expected")
			}
		}
	}
}

func TestEncodeDecodePrefix(t *testing.T) {
	pathParams := types.PathParams[uint8]{
		MaxComponentcount:  8,
		MaxPathLength:      8,
		MaxComponentLength: 8,
	}

}
