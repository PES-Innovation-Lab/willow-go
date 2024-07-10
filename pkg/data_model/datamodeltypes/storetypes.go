package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
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

type NamespaceScheme[NamespaceId constraints.Ordered, K constraints.Unsigned] struct {
	utils.EncodingScheme[NamespaceId, K]
	IsEqual            types.EqualityFn[NamespaceId]
	DefaultNamespaceId NamespaceId
}

type SubspaceScheme[SubspaceId constraints.Ordered, K constraints.Unsigned] struct {
	utils.EncodingScheme[SubspaceId, K]
	SuccessorSubspaceFn types.SuccessorFn[SubspaceId]
	Order               types.TotalOrder[SubspaceId]
	MinimalSubspaceId   SubspaceId
}

type PayloadScheme[PayloadDigest constraints.Ordered, K constraints.Unsigned] struct {
	utils.EncodingScheme[PayloadDigest, K]
	FromBytes            func(bytes []byte) chan PayloadDigest
	Order                types.TotalOrder[PayloadDigest]
	DefaultPayloadDigest PayloadDigest
}

type AuthorisationScheme[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered, AuthorisationOpts interface{}, AuthorisationToken string, K constraints.Unsigned] struct {
	Authorise        func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], opts AuthorisationOpts) chan AuthorisationToken
	IsAuthoriseWrite func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], token AuthorisationToken) chan bool
	TokenEncoding    utils.EncodingScheme[AuthorisationToken, K]
}

type FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	FingerPrintSingleton func(entry LengthyEntry[NamespaceId, SubspaceId, PayloadDigest]) chan PreFingerPrint
	FingerPrintCombine   func(a, b PreFingerPrint) PreFingerPrint
	FingerPrintFinalise  func(fp PreFingerPrint) FingerPrint
	neutral              PreFingerPrint
	neutralFinalised     FingerPrint
	isEqual              func(a, b FingerPrint) bool
	encoding             utils.EncodingScheme[FingerPrint, K]
}

type StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string] struct {
	PathParams          types.PathParams[K]
	NamespaceScheme     NamespaceScheme[NamespaceId, K]
	SubspaceScheme      SubspaceScheme[SubspaceId, K]
	PayloadScheme       PayloadScheme[PayloadDigest, K]
	AuthorisationScheme AuthorisationScheme[NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts, AuthorisationToken, K]
	FingerprintScheme   FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
}

type StoreOpts[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string] struct {
	Namespace     NamespaceId
	Schemes       StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
	EntryDriver   EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint]
	PayloadDriver PayloadDriver[PayloadDigest, K]
}

type Payload struct {
	Bytes           func() []byte
	BytesWithOffset func(offset int) ([]byte, error)
	Length          func() (uint64, error)
}

type EntryInput[SubspacePublicKey constraints.Ordered] struct {
	Path      types.Path
	Subspace  SubspacePublicKey
	Payload   []byte
	Timestamp uint64
}
type LengthyEntry[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	entry     types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	Available uint64
}
