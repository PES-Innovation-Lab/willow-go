package types

import "golang.org/x/exp/constraints"

// TODO: Depends on encoding tyopes

type PathParams[T constraints.Signed] struct {
	MaxComponentcount  T
	MaxComponentLength T
	MaxPathLength      T
}
