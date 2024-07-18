package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"

	"golang.org/x/exp/constraints"
)

type NamespaceScheme struct {
	EncodingScheme 	   utils.EncodingScheme[types.NamespaceId]
	IsEqual            types.EqualityFn[types.NamespaceId]
	DefaultNamespaceId types.NamespaceId
}

type SubspaceScheme struct {
	EncodingScheme  	utils.EncodingScheme[types.SubspaceId]
	SuccessorSubspaceFn types.SuccessorFn[types.SubspaceId]
	Order               types.TotalOrder[types.SubspaceId]
	MinimalSubspaceId   types.SubspaceId
}

type PayloadScheme struct {
	EncodingScheme		 utils.EncodingScheme[types.PayloadDigest]
	FromBytes            func(bytes []byte) chan types.PayloadDigest
	Order                types.TotalOrder[types.PayloadDigest]
	DefaultPayloadDigest types.PayloadDigest
}

type AuthorisationScheme[AuthorisationOpts []byte, AuthorisationToken string] struct {
	Authorise        func(entry types.Entry, opts AuthorisationOpts) (AuthorisationToken, error)
	IsAuthoriseWrite func(entry types.Entry, token AuthorisationToken) bool
	TokenEncoding    utils.EncodingScheme[AuthorisationToken]
}

type FingerprintScheme[PreFingerPrint, FingerPrint constraints.Ordered] struct {
	FingerPrintSingleton func(entry LengthyEntry) chan PreFingerPrint
	FingerPrintCombine   func(a, b PreFingerPrint) PreFingerPrint
	FingerPrintFinalise  func(fp PreFingerPrint) FingerPrint
	Neutral              PreFingerPrint
	NeutralFinalised     FingerPrint
	IsEqual              func(a, b FingerPrint) bool
	Encoding             utils.EncodingScheme[FingerPrint]
}

type StoreSchemes[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts []byte, AuthorisationToken string] struct {
	PathParams          types.PathParams[K]
	NamespaceScheme     NamespaceScheme
	SubspaceScheme      SubspaceScheme
	PayloadScheme       PayloadScheme
	AuthorisationScheme AuthorisationScheme[AuthorisationOpts, AuthorisationToken]
	FingerprintScheme   FingerprintScheme[PreFingerPrint, FingerPrint]
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
type LengthyEntry struct {
	Entry     types.Entry
	Available uint64
}
