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

type KvBatch[T KvPart] struct {
	Set    func(key KvKey[T], value KvValue) error
	Get    func(key KvKey[T]) (KvValue, error)
	Commit func() error
}

type ListOpts struct {
	Reverse   bool
	Limit     uint
	BatchSize uint
}

type ListSelector[T KvPart] struct {
	Start  KvKey[T]
	End    KvKey[T]
	Prefix KvKey[T]
}

type EntryIterator[T KvPart] struct {
	Value KvValue
	Key   KvKey[T]
}

type KvDriver[T KvPart] struct {
	Db     *pebble.DB
	Close  func(Db *pebble.DB) error
	Get    func(Db *pebble.DB, key KvKey[T]) (KvValue, error)
	Set    func(Db *pebble.DB, key KvKey[T], value KvValue) error
	Delete func(Db *pebble.DB, key KvKey[T]) error
	List   func(selector ListSelector[T], opts ListOpts) ([]EntryIterator[T], error)
	Clear  func(opts ListSelector[T]) error
	Batch  func() (KvBatch[T], error)
}
