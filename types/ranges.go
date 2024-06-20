package types

import "cmp"

// All paths are of [][]byte type and do not satisfy constrainst.Ordered (direct < , > etc. comparisions)
// We do have other helper functions and methods to compare paths and prefixes though!
type Range[T OrderableGeneric] struct {
	Start, End T
	OpenEnd    bool
}

type SubspaceRange[T cmp.Ordered] struct {
	Value   T
	OpenEnd bool
}

type Range3d[SubspaceId OrderableGeneric] struct {
	SubspaceRange Range[SubspaceId]
	PathRange     Range[Path]
	TimeRange     Range[uint64]
}

type Position3d[SubspaceId OrderableGeneric] struct {
	Subspace SubspaceId
	Path     Path
	Time     uint64
}
