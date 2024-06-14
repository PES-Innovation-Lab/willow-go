package types

import (
	"cmp"

	"golang.org/x/exp/constraints"
)

type Rel int

type OrderableGeneric interface {
	cmp.Ordered | Path
}

const (
	Less    Rel = -1
	Equal   Rel = 0
	Greater Rel = 1
)

// TotalOrder defines a total order over a given set.
type TotalOrder[T OrderableGeneric] func(a, b T) Rel

// SuccessorFn returns the succeeding value for a given value of a set.
type SuccessorFn[T constraints.Ordered] func(val T) *T
