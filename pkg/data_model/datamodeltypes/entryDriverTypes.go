package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, T KvPart, K constraints.Unsigned] struct {
	MakeStorage             func(namespace NamespaceId) KDTreeStorage[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	PayloadReferenceCounter PayloadReferenceCounter[PayloadDigest]
	GetPayloadLength        func(digest PayloadDigest) uint64
	Opts                    struct {
		KVDriver          KvDriver
		NamespaceScheme   NamespaceScheme[NamespaceId, K]
		SubspaceScheme    SubspaceScheme[NamespaceId, K]
		PayloadScheme     PayloadScheme[PayloadDigest, K]
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
	}
}

type PayloadReferenceCounter[PayloadDigest constraints.Ordered] interface {
	Increment(payloadDigest PayloadDigest) uint
	Decrement(payloadDigest PayloadDigest) uint
	Count(payloadDigest PayloadDigest) uint
}

type KDTreeStorage[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, T KvPart, K constraints.Unsigned] struct {
	KVDriver KvDriver

	KDTree *Kdtree.KDTree[Kdtree.KDNodeKey[SubspaceId]]

	Opts struct {
		Namespace         NamespaceId
		SubspaceScheme    SubspaceScheme[NamespaceId, K]
		PayloadScheme     PayloadScheme[PayloadDigest, K]
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K]
		GetPayloadLength  func(digest PayloadDigest) uint64
	}

	/** Retrieve an entry at a subspace and path. */
	Get func(subspace SubspaceId, path types.Path) (struct {
		Entry         types.Entry[NamespaceId, SubspaceId, PayloadDigest]
		AuthTokenHash PayloadDigest
	}, error)
	/** Insert a new entry. */
	Insert func(opts struct {
		Subspace      SubspaceId
		Path          types.Path
		PayloadDigest PayloadDigest
		Timestamp     uint64
		PayloadLength uint64
		AuthTokenHash PayloadDigest
	}) error

	/** Update the available payload bytes for a given entry. */

	UpdateAvailablePayload func(subspace SubspaceId, path types.Path) bool
	/** Remove an entry. */
	Remove func(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest]) error
	// Used during sync.

	/** Summarise a given `Range3d` by mapping the included set of `Entry` to ` PreFingerprint`.  */
	Summarise func(range3d types.Range3d[SubspaceId]) chan struct {
		Fingerprint PreFingerPrint
		Size        uint64
	}
	/** Split a range into two smaller ranges. */
	SplitRange func(range3d types.Range3d[SubspaceId], knownSize uint) []types.Range3d[SubspaceId]
	/** 3D Range Query **/
	Query func(range3d types.Range3d[SubspaceId], reverse bool) []struct {
		Entry         types.Entry[NamespaceId, SubspaceId, PayloadDigest]
		AuthTokenHash PayloadDigest
	}
}
