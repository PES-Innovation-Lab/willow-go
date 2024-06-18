package types

import "cmp"

type Area[SubspaceId cmp.Ordered] struct {
	Subspace_id  SubspaceId
	Any_subspace bool
	Path         Path
	Times        Range[uint64]
}

type AreaOfInterest[SubspaceId cmp.Ordered] struct {
	Area      Area[SubspaceId]
	Max_count uint64
	Max_size  uint64
}
