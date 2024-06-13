package types

import (
	"golang.org/x/exp/constraints"
)

type Range[T constraints.Ordered | Path] struct {
	Start T
	End   *T // Use pointer to indicate open range if nil
}

type Range3D[SubspaceId constraints.Ordered] struct {
	Subspaces Range[SubspaceId]
	Paths     Range[Path]
	Times     Range[uint64]
}
