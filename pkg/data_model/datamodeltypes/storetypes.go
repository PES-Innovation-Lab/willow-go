package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"

	"golang.org/x/exp/constraints"
)

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

type AuthorisationScheme[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered, AuthorisationOpts any, AuthorisationToken string, K constraints.Unsigned] struct {
	Authorise        func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], opts AuthorisationOpts) chan AuthorisationToken
	IsAuthoriseWrite func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], token AuthorisationToken) chan bool
	TokenEncoding    utils.EncodingScheme[AuthorisationToken, K]
}

type FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	FingerPrintSingleton func(entry LengthyEntry[NamespaceId, SubspaceId, PayloadDigest]) chan PreFingerPrint
	FingerPrintCombine   func(a, b PreFingerPrint) PreFingerPrint
	FingerPrintFinalise  func(fp PreFingerPrint) FingerPrint
	Neutral              PreFingerPrint
	NeutralFinalised     FingerPrint
	IsEqual              func(a, b FingerPrint) bool
	Encoding             utils.EncodingScheme[FingerPrint, K]
}

type StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string] struct {
	PathParams          types.PathParams[K]
	NamespaceScheme     NamespaceScheme[NamespaceId, K]
	SubspaceScheme      SubspaceScheme[SubspaceId, K]
	PayloadScheme       PayloadScheme[PayloadDigest, K]
	AuthorisationScheme AuthorisationScheme[NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts, AuthorisationToken, K]
	FingerprintScheme   FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
}

type StoreOpts[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string, T KvPart] struct {
	Namespace     NamespaceId
	Schemes       StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
	EntryDriver   EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
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
	Entry     types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	Available uint64
}
