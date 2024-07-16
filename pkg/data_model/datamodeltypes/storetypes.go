package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"

	"golang.org/x/exp/constraints"
)

type NamespaceScheme[K constraints.Unsigned] struct {
	utils.EncodingScheme[K]
	IsEqual            types.EqualityFn[types.NamespaceId]
	DefaultNamespaceId types.NamespaceId
}

type SubspaceScheme[K constraints.Unsigned] struct {
	utils.EncodingScheme[K]
	SuccessorSubspaceFn types.SuccessorFn[types.SubspaceId]
	Order               types.TotalOrder[types.SubspaceId]
	MinimalSubspaceId   types.SubspaceId
}

type PayloadScheme[PayloadDigest constraints.Ordered, K constraints.Unsigned] struct {
	utils.EncodingScheme[PayloadDigest]
	FromBytes            func(bytes []byte) chan PayloadDigest
	Order                types.TotalOrder[PayloadDigest]
	DefaultPayloadDigest PayloadDigest
}

type AuthorisationScheme[PayloadDigest constraints.Ordered, AuthorisationOpts any, AuthorisationToken string, K constraints.Unsigned] struct {
	Authorise        func(entry types.Entry[PayloadDigest], opts AuthorisationOpts) AuthorisationToken
	IsAuthoriseWrite func(entry types.Entry[PayloadDigest], token AuthorisationToken) bool
	TokenEncoding    utils.EncodingScheme[AuthorisationToken]
}

type FingerprintScheme[PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	FingerPrintSingleton func(entry LengthyEntry[PayloadDigest]) chan PreFingerPrint
	FingerPrintCombine   func(a, b PreFingerPrint) PreFingerPrint
	FingerPrintFinalise  func(fp PreFingerPrint) FingerPrint
	Neutral              PreFingerPrint
	NeutralFinalised     FingerPrint
	IsEqual              func(a, b FingerPrint) bool
	Encoding             utils.EncodingScheme[FingerPrint]
}

type StoreSchemes[PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string] struct {
	PathParams          types.PathParams[K]
	NamespaceScheme     NamespaceScheme[K]
	SubspaceScheme      SubspaceScheme[K]
	PayloadScheme       PayloadScheme[PayloadDigest, K]
	AuthorisationScheme AuthorisationScheme[PayloadDigest, AuthorisationOpts, AuthorisationToken, K]
	FingerprintScheme   FingerprintScheme[PayloadDigest, PreFingerPrint, FingerPrint, K]
}

// type StoreOpts[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts interface{}, AuthorisationToken string, T KvPart] struct {
// 	Namespace     NamespaceId
// 	Schemes       StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
// 	EntryDriver   EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
// 	PayloadDriver PayloadDriver[PayloadDigest, K]
// }

type Payload struct {
	Bytes           func() []byte
	BytesWithOffset func(offset int) ([]byte, error)
	Length          func() (uint64, error)
}

type EntryInput struct {
	Path      types.Path
	Subspace  types.SubspaceId
	Payload   []byte
	Timestamp uint64
}
type LengthyEntry[PayloadDigest constraints.Ordered] struct {
	Entry     types.Entry[PayloadDigest]
	Available uint64
}
