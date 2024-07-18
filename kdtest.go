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
	})
	fmt.Println(testTree)
	ok := testTree.Delete(Kdtree.KDNodeKey{Timestamp: 500, Subspace: []byte{0}, Path: types.Path{{0}, {2}}})
	fmt.Println(testTree, ok)
}

func main() {
	Kdtest()
}
