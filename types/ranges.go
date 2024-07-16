package types

// All paths are of [][]byte type and do not satisfy constrainst.Ordered (direct < , > etc. comparisions)
// We do have other helper functions and methods to compare paths and prefixes though!
type Range[T OrderableGeneric] struct {
	Start, End T
	OpenEnd    bool
}

type Range3d struct {
	SubspaceRange Range[SubspaceId]
	PathRange     Range[Path]
	TimeRange     Range[uint64]
}

type Position3d struct {
	Subspace SubspaceId
	Path     Path
	Time     uint64
}
