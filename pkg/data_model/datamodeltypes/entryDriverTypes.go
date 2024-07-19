package datamodeltypes

import (
	"errors"
	"fmt"
	"log"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
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
	}

	/** Retrieve an entry at a subspace and path. */
	// Get func(subspace types.SubspaceId, path types.Path) (struct {
	// 	Entry         types.Entry
	// 	AuthTokenHash types.PayloadDigest
	// }, error)
	// /** Insert a new entry. */
	// Insert func(opts struct {
	// 	Subspace      types.SubspaceId
	// 	Path          types.Path
	// 	PayloadDigest types.PayloadDigest
	// 	Timestamp     uint64
	// 	PayloadLength uint64
	// 	AuthTokenHash types.PayloadDigest
	// }) error

	// /** Update the available payload bytes for a given entry. */

	// UpdateAvailablePayload func(subspace types.SubspaceId, path types.Path) bool
	// /** Remove an entry. */
	// Remove func(entry types.Position3d) error
	// // Used during sync.

	// /** Summarise a given `Range3d` by mapping the included set of `Entry` to ` PreFingerprint`.  */
	// Summarise func(range3d types.Range3d) struct {
	// 	Fingerprint PreFingerPrint
	// 	Size        uint64
	// }
	// /** Split a range into two smaller ranges. */
	// SplitRange func(range3d types.Range3d, knownSize uint) []types.Range3d
	// /** 3D Range Query **/
	// Query func(range3d types.Range3d, reverse bool) []struct {
	// 	Entry         types.Entry
	// 	AuthTokenHash types.PayloadDigest
	// }
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Get(Subspace types.SubspaceId, Path types.Path) types.Position3d {
	subspaceRange := types.Range[types.SubspaceId]{
		Start:   Subspace,
		End:     Subspace,
		OpenEnd: false,
	}

	pathRange := types.Range[types.Path]{
		Start:   Path,
		End:     utils.SuccessorPath(Path, k.Opts.PathParams),
		OpenEnd: false,
	}

	timeRange := types.Range[uint64]{
		Start:   0,
		End:     2,
		OpenEnd: true,
	}

	range3d := types.Range3d{
		SubspaceRange: subspaceRange,
		PathRange:     pathRange,
		TimeRange:     timeRange,
	}

	res := Kdtree.Query(k.KDTree, range3d)
	fmt.Println(res)
	if len(res) > 1 {
		log.Fatalln("get returned multiple nodes")
	}
	switch len(res) {
	case 0:
		return types.Position3d{}
	case 1:
		return types.Position3d{
			Subspace: res[0].Subspace,
			Time:     res[0].Timestamp,
			Path:     res[0].Path}
	default:
		log.Fatalln("get returned multiple nodes")
		return types.Position3d{}
	}
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Insert(Subspace types.SubspaceId, Path types.Path, Timestamp uint64) error {
	newVal := Kdtree.KDNodeKey{
		Subspace:  Subspace,
		Path:      Path,
		Timestamp: Timestamp,
	}
	if !k.KDTree.Add(newVal) {
		return errors.New("error inserting the node into the KD tree")
	}
	return nil
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Query(QueryRange types.Range3d) []Kdtree.KDNodeKey {
	return Kdtree.Query(k.KDTree, QueryRange)
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Remove(entry types.Position3d) bool {

	NodeToDelete := Kdtree.KDNodeKey{
		Subspace:  entry.Subspace,
		Timestamp: entry.Time,
		Path:      entry.Path,
	}

	return k.KDTree.Delete(NodeToDelete)
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) GetInterestRange(areaOfInterest types.AreaOfInterest) types.Range3d {
	newRange := utils.AreaTo3dRange[K](
		utils.Options[K]{
			MinimalSubspace:        k.Opts.SubspaceScheme.MinimalSubspaceId,
			SuccessorSubspace:      k.Opts.SubspaceScheme.SuccessorSubspaceFn,
			MaxPathLength:          k.Opts.PathParams.MaxPathLength,
			MaxComponentCount:      k.Opts.PathParams.MaxComponentCount,
			MaxPathComponentLength: k.Opts.PathParams.MaxComponentLength,
		}, areaOfInterest.Area,
	)
	return newRange
}
