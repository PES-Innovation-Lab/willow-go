package kdnode

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	kdtree "github.com/rishitc/go-kd-tree"
)

type Key struct {
	Timestamp   uint64
	Subspace    types.SubspaceId
	Path        types.Path
	Fingerprint string
}

func (lhs Key) Order(rhs Key, dim int) kdtree.Relation {
	dimensions := 3
	for i := 0; i < dimensions; i++ {
		switch dim {
		case 0:
			// Compare timestamps
			switch utils.OrderTimestamp(lhs.Timestamp, rhs.Timestamp) {
			case -1:
				return kdtree.Lesser
			case 1:
				return kdtree.Greater
			}

		case 1:
			// Compare subspace IDs
			switch utils.OrderBytes(lhs.Subspace, rhs.Subspace) {
			case -1:
				return kdtree.Lesser
			case 1:
				return kdtree.Greater
			}

		case 2:
			switch utils.OrderPath(lhs.Path, rhs.Path) {
			case -1:
				return kdtree.Lesser
			case 1:
				return kdtree.Greater
			}

		}
		dim = (dim + 1) % dimensions
	}
	return kdtree.Equal
}

func (lhs Key) DistDim(rhs Key, dim int) int {
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

func (lhs Key) Dist(rhs Key) int {
	// Calculate the squared difference between timestamps
	timeDist := (lhs.Timestamp - rhs.Timestamp) * (lhs.Timestamp - rhs.Timestamp)

	// Use a simple constant distance for differing subspace IDs and paths
	subspaceDist := 0
	if utils.OrderBytes(lhs.Subspace, rhs.Subspace) != 0 {
		subspaceDist = 1
	}

	pathDist := 0
	if utils.OrderPath(lhs.Path, rhs.Path) != 0 {
		pathDist = 1
	}

	// Sum these values to get the overall distance
	return int(timeDist + uint64(subspaceDist+pathDist))
}

func (lhs Key) Encode() []byte {
	// TODO for encoding keys to bytes. Need to encode path with pat param opts and new subspace encoding
	return []byte{}
}

func (lhs Key) String() string {
	return fmt.Sprintf("[%v,%v,%v]", lhs.Timestamp, lhs.Subspace, lhs.Path)
}

// func ListNodes(r *kdtree.KdNode[key]) []Key {
// 	var res []Key
// 	ListHelper(r, &res)
// 	return res
// }

// func ListHelper(r *kdtree.KdNode[Key], res *[]Key) {
// 	if r == nil {
// 		return
// 	}

// 	*res = append(*res, r.Value)
// 	ListHelper(r.Left, res)
// 	ListHelper(r.Right, res)
// }
