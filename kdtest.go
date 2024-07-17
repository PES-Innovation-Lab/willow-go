package main

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
)

func Kdtest() {
	var testTree *Kdtree.KDTree[Kdtree.KDNodeKey] = Kdtree.NewKDTreeWithValues[Kdtree.KDNodeKey](3, []Kdtree.KDNodeKey{
		{Timestamp: 500,
			Subspace: []byte{0},
			Path:     types.Path{{0}, {2}},
		},
		// }, {
		// 	Timestamp: 600,
		// 	Subspace:  []byte{1},
		// 	Path:      types.Path{{1}, {3}},
		// },
		// {
		// 	Timestamp: 700,
		// 	Subspace:  []byte{2},
		// 	Path:      types.Path{{2}, {4}},
		// }, {Timestamp: 800, Subspace: []byte{0}, Path: types.Path{{3}}},
		// {Timestamp: 900, Subspace: []byte{2}, Path: types.Path{{4}, {5}}},
		// {Timestamp: 1000, Subspace: []byte{3}, Path: types.Path{{5}, {6}, {7}}},
		// {Timestamp: 1100, Subspace: []byte{1}, Path: types.Path{{6}}},
		// {Timestamp: 1200, Subspace: []byte{2}, Path: types.Path{{7}, {8}}},
		/*{Timestamp: 1300, Subspace: 3, Path: types.Path{{8}, {9}, {10}}},
		{Timestamp: 1400, Subspace: 0, Path: types.Path{{9}}},
		{Timestamp: 1500, Subspace: 3, Path: types.Path{{10}, {11}}},
		{Timestamp: 1600, Subspace: 1, Path: types.Path{{11}, {12}, {13}}},
		{Timestamp: 1700, Subspace: 4, Path: types.Path{{12}}},
		{Timestamp: 1800, Subspace: 2, Path: types.Path{{13}, {14}}},
		{Timestamp: 1900, Subspace: 1, Path: types.Path{{14}, {15}, {16}}},
		{Timestamp: 2000, Subspace: 3, Path: types.Path{{15}}},
		{Timestamp: 2100, Subspace: 4, Path: types.Path{{16}, {17}}},
		{Timestamp: 2200, Subspace: 2, Path: types.Path{{17}, {18}, {19}}},
		{Timestamp: 2300, Subspace: 1, Path: types.Path{{18}}},
		{Timestamp: 2400, Subspace: 0, Path: types.Path{{19}, {20}}}, */
	})
	fmt.Println(testTree)
	ok := testTree.Delete(Kdtree.KDNodeKey{Timestamp: 500, Subspace: []byte{0}, Path: types.Path{{0}, {2}}})
	fmt.Println(testTree, ok)
}

func main() {
	Kdtest()
}
