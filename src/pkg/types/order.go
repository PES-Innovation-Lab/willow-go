package types

// TotalOrder defines a total order over a given set.
type TotalOrder func(a, b int) int

// SuccessorFn returns the succeeding value for a given value of a set.
type SuccessorFn func(val int) *int
