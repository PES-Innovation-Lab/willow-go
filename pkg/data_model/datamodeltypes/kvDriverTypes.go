package datamodeltypes

import (
	"golang.org/x/exp/constraints"
)

type KvPart interface {
	[]byte | constraints.Ordered
}

type KvValue any

type KvKey[T KvPart] struct {
	Key []T
}

type ListOpts struct {
	Reverse   bool
	Limit     uint
	BatchSize uint
}