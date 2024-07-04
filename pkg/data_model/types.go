package datamodel

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

type AuthorisationScheme[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered, AuthorisationOpts interface{}, AuthorisationToken string, K constraints.Unsigned] struct {
	Authorise        func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], opts AuthorisationOpts) chan AuthorisationToken
	IsAuthoriseWrite func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], token AuthorisationToken) chan bool
	TokenEncoding    utils.EncodingScheme[AuthorisationToken, K]
}

type FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PrefingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	FinerprintSingleton func(entry Lengthy)
}
