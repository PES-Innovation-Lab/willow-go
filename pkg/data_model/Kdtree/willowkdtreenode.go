package Kdtree

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type KDNodeKey[SubspaceId constraints.Ordered] struct {
	Timestamp uint64
	Subspace  SubspaceId
	Path      types.Path
}

func (lhs KDNodeKey[SubspaceId]) Order(rhs KDNodeKey[SubspaceId], dim int) Relation {

	switch dim {
	case 0:
		// Compare timestamps
		switch utils.OrderTimestamp(lhs.Timestamp, rhs.Timestamp) {
		case -1:
			return Lesser
		case 1:
			return Greater
		}
	case 1:
		// Compare subspace IDs
		if lhs.Subspace < rhs.Subspace {
			return Lesser
		} else if lhs.Subspace > rhs.Subspace {
			return Greater
		}
	case 2:
		switch utils.OrderPath(lhs.Path, rhs.Path) {
		case -1:
			return Lesser
		case 1:
			return Greater
		}

	}
	return Equal
}

func (lhs KDNodeKey[SubspaceId]) DistDim(rhs KDNodeKey[SubspaceId], dim int) int {
	switch dim {
	case 0:
		return int((lhs.Timestamp - rhs.Timestamp) * (lhs.Timestamp - rhs.Timestamp))
	case 1:
		// TODO for subspaces. Return all possible subsapceids in between
		return 1

	case 2:
		// TODO for paths. Return all possible paths in between

		return 1

	}

	return 1

}

func (lhs KDNodeKey[SubspaceId]) Dist(rhs KDNodeKey[SubspaceId]) int {
	// TODO for distances between keys (Size of Range Query)
	return 1
}

func (lhs KDNodeKey[SubspaceId]) Encode() []byte {
	// TODO for encoding keys to bytes. Need to encode path with pat param opts and new subspace encoding
	return []byte{}
}

func (lhs KDNodeKey[SubspaceId]) String() string {
	return fmt.Sprintf("[%v,%v,%v]", lhs.Timestamp, lhs.Subspace, lhs.Path)
}
