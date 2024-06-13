package types

import (
	"golang.org/x/exp/constraints"
)

type Area[SubspaceId constraints.Ordered] struct {
	subspace_id SubspaceId
	path        Path
	times       Range[uint64]
}

type AreaOfInterest[SubspaceId constraints.Ordered] struct {
	area      Area[SubspaceId]
	max_count uint64
	max_size  uint64
}
