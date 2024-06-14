package types

import "golang.org/x/exp/constraints"

// TotalOrder defines a total order over a given set.
type TotalOrder func(a, b interface{}) int

// SuccessorFn returns the succeeding value for a given value of a set.
type SuccessorFn[T constraints.Ordered] func(val T) *T
