package types

import (
	"golang.org/x/exp/constraints"
)

// All paths are of [][]byte type and do not satisfy constrainst.Ordered (direct < , > etc. comparisions)
// We do have other helper functions and methods to compare paths and prefixes though!
type Range[T constraints.Ordered | Path] struct {
	start T
	end   *T // use pointer to indicate open range if nil
}

type Range3D[SubspaceId constraints.Ordered] struct {
	subspaces Range[SubspaceId]
	paths     Range[Path]
	times     Range[uint64]
}
