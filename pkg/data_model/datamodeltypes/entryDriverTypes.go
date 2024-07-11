package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, T KvPart, K constraints.Unsigned] struct {
	KDTreeStorage           KDTreeStorage[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	MakeStorage             func(namespace NamespaceId)
	PayloadReferenceCounter PayloadReferenceCounter[PayloadDigest]
	GetPayloadLength        func(digest PayloadDigest) uint64
	Opts                    struct {
		KVDriver          KvDriver[T]
		NamespaceScheme   NamespaceScheme[NamespaceId, K]
		SubspaceScheme    SubspaceScheme[NamespaceId, K]
		PayloadScheme     PayloadScheme[PayloadDigest, K]
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
	}
}

type PayloadReferenceCounter[PayloadDigest constraints.Ordered] interface {
	Increment(payloadDigest PayloadDigest) chan uint
	Decrement(payloadDigest PayloadDigest) chan uint
	Count(payloadDigest PayloadDigest) chan uint
}

type KDTreeStorage[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, T KvPart, K constraints.Unsigned] struct {
	KVDriver KvDriver[T]

	Opts struct {
		Namespace         NamespaceId
		SubspaceScheme    SubspaceScheme[NamespaceId, K]
		PayloadScheme     PayloadScheme[PayloadDigest, K]
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
		GetPayloadLength  func(digest PayloadDigest) uint64
	}

	/** Retrieve an entry at a subspace and path. */
	Get func(subspace SubspaceId, path types.Path) chan struct {
		Entry         types.Entry[NamespaceId, SubspaceId, PayloadDigest]
		AuthTokenHash PayloadDigest
	}
	/** Insert a new entry. */
	Insert func(opts struct {
		Subspace      SubspaceId
		Path          types.Path
		PayloadDigest PayloadDigest
		Timestamp     uint64
		PayloadLength uint64
		AuthTokenHash PayloadDigest
	}) chan error

	/** Update the available payload bytes for a given entry. */

	UpdateAvailablePayload func(subspace SubspaceId, path types.Path) chan bool
	/** Remove an entry. */
	Remove func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest]) chan error
	// Used during sync.

	/** Summarise a given `Range3d` by mapping the included set of `Entry` to ` PreFingerprint`.  */
	Summarise func(range3d types.Range3d[SubspaceId]) chan struct {
		Fingerprint PreFingerPrint
		Size        uint64
	}
	/** Split a range into two smaller ranges. */
	SplitRange func(range3d types.Range3d[SubspaceId], knownSize uint) chan []types.Range3d[SubspaceId]
	/** 3D Range Query **/
	Query func(range3d types.Range3d[SubspaceId], reverse bool) []struct {
		Entry         types.Entry[NamespaceId, SubspaceId, PayloadDigest]
		AuthTokenHash PayloadDigest
	}
}
