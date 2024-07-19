package Kdtree

import (
	"fmt"
	"sync"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

type KDNodeKey struct {
	Timestamp   uint64
	Subspace    types.SubspaceId
	Path        types.Path
	Fingerprint string
}

func (lhs KDNodeKey) Order(rhs KDNodeKey, dim int) Relation {
	dimensions := 3
	for i := 0; i < dimensions; i++ {
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
			switch utils.OrderBytes(lhs.Subspace, rhs.Subspace) {
			case -1:
				return Lesser
			case 1:
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
		dim = (dim + 1) % dimensions
	}
	return Equal
}

func (lhs KDNodeKey) DistDim(rhs KDNodeKey, dim int) int {
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

func (lhs KDNodeKey) Dist(rhs KDNodeKey) int {
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

func (lhs KDNodeKey) Encode() []byte {
	// TODO for encoding keys to bytes. Need to encode path with pat param opts and new subspace encoding
	return []byte{}
}

func (lhs KDNodeKey) String() string {
	return fmt.Sprintf("[%v,%v,%v]", lhs.Timestamp, lhs.Subspace, lhs.Path)
}

// Queries the given tree for the given 3-d Range
func Query(kdt *(KDTree[KDNodeKey]), QueryRange types.Range3d) []KDNodeKey {

	dim := 0
	var res []KDNodeKey
	if kdt == nil {
		return res
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	QueryHelper(kdt.Root, QueryRange, dim, &res, &mu, &wg)
	wg.Wait() // Wait for all goroutines to finish
	return res
}

// A helper function to recursively query for range
func QueryHelper(Node *KdNode[KDNodeKey], QueryRange types.Range3d, dim int, res *[]KDNodeKey, mu *sync.Mutex, wg *sync.WaitGroup) {
	if Node == nil {
		return
	}
	Timestamp, Subspace, Path := Node.Value.Timestamp, Node.Value.Subspace, Node.Value.Path
	Position := types.Position3d{
		Subspace: Subspace,
		Path:     Path,
		Time:     Timestamp,
	}

	inRange := utils.IsIncluded3d(utils.OrderSubspace, QueryRange, Position) //PLEASE CHANGE THE ordersubspace Call i have jugaad for now ~Samarth

	switch dim % 3 {
	case 0:
		if utils.OrderTimestamp(Timestamp, QueryRange.TimeRange.Start) >= 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Left, QueryRange, dim+1, res, mu, wg)
			}()
		}
		if QueryRange.TimeRange.OpenEnd || utils.OrderTimestamp(Timestamp, QueryRange.TimeRange.End) < 0 {
			if inRange {
				fmt.Println("case 1 Adding")
				mu.Lock()
				*res = append(*res, Node.Value)
				mu.Unlock()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Right, QueryRange, dim+1, res, mu, wg)
			}()
		}
	case 1:
		if utils.OrderBytes(Subspace, QueryRange.SubspaceRange.Start) >= 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Left, QueryRange, dim+1, res, mu, wg)
			}()
		}
		if QueryRange.SubspaceRange.OpenEnd || utils.OrderBytes(Subspace, QueryRange.SubspaceRange.End) < 0 || utils.OrderBytes(QueryRange.SubspaceRange.Start, QueryRange.SubspaceRange.End) == 0 {
			if inRange {
				fmt.Println("case 2 Adding")
				mu.Lock()
				*res = append(*res, Node.Value)
				mu.Unlock()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Right, QueryRange, dim+1, res, mu, wg)
			}()
		}
	case 2:
		if utils.OrderPath(Path, QueryRange.PathRange.Start) >= 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Left, QueryRange, dim+1, res, mu, wg)
			}()
		}
		if QueryRange.PathRange.OpenEnd || utils.OrderPath(Path, QueryRange.PathRange.End) < 0 {
			if inRange {
				fmt.Println("case 3 Adding")
				mu.Lock()
				*res = append(*res, Node.Value)
				mu.Unlock()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				QueryHelper(Node.Right, QueryRange, dim+1, res, mu, wg)
			}()
		}
	}
}

func ListNodes(r *KdNode[KDNodeKey]) []KDNodeKey {
	var res []KDNodeKey
	ListHelper(r, &res)
	return res
}

func ListHelper(r *KdNode[KDNodeKey], res *[]KDNodeKey) {
	if r == nil {
		return
	}

	*res = append(*res, r.Value)
	ListHelper(r.Left, res)
	ListHelper(r.Right, res)
}
