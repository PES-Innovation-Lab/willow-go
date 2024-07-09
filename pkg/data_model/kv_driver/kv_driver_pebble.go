package data_model

import (
	"github.com/cockroachdb/pebble"
	"golang.org/x/exp/constraints"
)

type KeyPart KvPart

type KvPart interface {
	[]byte | constraints.Ordered
}

type KvKey[T KvPart] struct {
	key []T
}

type KvBatch[T KvPart] struct {
	Set    func(key KvKey[T], value any) error
	Get    func(key KvKey[T]) (any, error)
	Commit func() error
}

type KvDriver[T KvPart] struct {
	Db     *pebble.DB
	Close  func() error
	Get    func(key KvKey[T]) (any, error)
	Set    func(key KvKey[T], value any) error
	Delete func(key KvKey[T]) error
	List   func(selector struct {
		Start  KvKey[T]
		End    KvKey[T]
		Prefix KvKey[T]
	}, opts struct {
		Reverse   bool
		Limit     uint
		BatchSize uint
	}) ([]struct {
		Value any
		Key   KvKey[T]
	}, error)
	Clear func(opts struct {
		Start  KvKey[T]
		End    KvKey[T]
		Prefix KvKey[T]
	}) error
	Batch func() (KvBatch[T], error)
}
