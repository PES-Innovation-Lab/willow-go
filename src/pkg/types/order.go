package types

import (
	"golang.org/x/exp/constraints"
)

// TotalOrder defines a total order over a given set.
type TotalOrder[T constraints.Ordered] func(a, b T) int

// SuccessorFn returns the succeeding value for a given value of a set.
type SuccessorFn[T constraints.Ordered] func(val T) *T
