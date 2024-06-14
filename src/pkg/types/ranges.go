package types

import (
	"golang.org/x/exp/constraints"
)

// All paths are of [][]byte type and do not satisfy constrainst.Ordered (direct < , > etc. comparisions)
// We do have other helper functions and methods to compare paths and prefixes though!
type Range[T constraints.Ordered | Path] struct {
	Start T
	End   *T // use pointer to indicate open range if nil
}

type Range3D[SubspaceId constraints.Ordered] struct {
	Subspaces Range[SubspaceId]
	Paths     Range[Path]
	Times     Range[uint64]
}

type Position3d[SubspaceId constraints.Ordered] struct {
	Subspace SubspaceId
	Path     Path
	Time     uint64
}
