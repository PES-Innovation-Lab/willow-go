package datamodeltypes

import (
	"github.com/cockroachdb/pebble"
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

type KvDriver struct {
	Db            *pebble.DB
	Close         func(Db *pebble.DB) error
	Get           func(Db *pebble.DB, key []byte) ([]byte, error)
	Set           func(Db *pebble.DB, key []byte, value []byte) error
	Delete        func(Db *pebble.DB, key []byte) error
	Clear         func(Db *pebble.DB) error
	ListAllValues func(Db *pebble.DB) ([]struct {
		Key   []byte
		Value []byte
	}, error)
	Batch func(Db *pebble.DB) (*pebble.Batch, error)
}
