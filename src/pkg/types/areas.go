package types

import (
	"golang.org/x/exp/constraints"
)

type Area[SubspaceId constraints.Ordered] struct {
	Subspace_id *SubspaceId
	Path        Path
	Times       Range[uint64]
}

type AreaOfInterest[SubspaceId constraints.Ordered] struct {
	Area      Area[SubspaceId]
	Max_count uint64
	Max_size  uint64
}
