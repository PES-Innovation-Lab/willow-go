package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type KDTreeStorage[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	KDTree *Kdtree.KDTree[Kdtree.KDNodeKey]

	Opts struct {
		Namespace         types.NamespaceId
		SubspaceScheme    SubspaceScheme
		PayloadScheme     PayloadScheme
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[PreFingerPrint, FingerPrint]
		GetPayloadLength  func(digest types.PayloadDigest) uint64
	}

	/** Retrieve an entry at a subspace and path. */
	Get func(subspace types.SubspaceId, path types.Path) (struct {
		Entry         types.Entry
		AuthTokenHash types.PayloadDigest
	}, error)
	/** Insert a new entry. */
	Insert func(opts struct {
		Subspace      types.SubspaceId
		Path          types.Path
		PayloadDigest types.PayloadDigest
		Timestamp     uint64
		PayloadLength uint64
		AuthTokenHash types.PayloadDigest
	}) error

	/** Update the available payload bytes for a given entry. */

	UpdateAvailablePayload func(subspace types.SubspaceId, path types.Path) bool
	/** Remove an entry. */
	Remove func(entry types.Position3d) error
	// Used during sync.

	/** Summarise a given `Range3d` by mapping the included set of `Entry` to ` PreFingerprint`.  */
	Summarise func(range3d types.Range3d) struct {
		Fingerprint PreFingerPrint
		Size        uint64
	}
	/** Split a range into two smaller ranges. */
	SplitRange func(range3d types.Range3d, knownSize uint) []types.Range3d
	/** 3D Range Query **/
	Query func(range3d types.Range3d, reverse bool) []struct {
		Entry         types.Entry
		AuthTokenHash types.PayloadDigest
	}
}
