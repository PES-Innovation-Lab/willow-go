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

	var PrefixesOfVectors = []PrefixesOfVector{
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
