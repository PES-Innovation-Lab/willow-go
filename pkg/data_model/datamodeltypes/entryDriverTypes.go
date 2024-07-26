package datamodeltypes

import (
	"errors"
	"log"
	"sort"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	kdtree "github.com/rishitc/go-kd-tree"
	"golang.org/x/exp/constraints"
)

type KDTreeStorage[PreFingerPrint, FingerPrint string, K constraints.Unsigned] struct {
	KDTree *kdtree.KDTree[kdnode.Key]

	Opts struct {
		Namespace         types.NamespaceId
		SubspaceScheme    SubspaceScheme
		PayloadScheme     PayloadScheme
		PathParams        types.PathParams[K]
		FingerprintScheme FingerprintScheme[PreFingerPrint, FingerPrint]
	}
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Get(Subspace types.SubspaceId, Path types.Path) (types.Position3d, error) {
	subspaceRange := types.Range[types.SubspaceId]{
		Start:   Subspace,
		End:     utils.SuccessorSubspaceId(Subspace),
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

	res := kdnode.Query(k.KDTree, range3d)
	if len(res) > 1 {

		log.Fatalln("get returned multiple nodes")
	}
	switch len(res) {
	case 0:
		return types.Position3d{}, nil
	case 1:
		return types.Position3d{
			Subspace: res[0].Subspace,
			Time:     res[0].Timestamp,
			Path:     res[0].Path}, nil
	default:
		log.Fatalln("get returned multiple nodes")
		return types.Position3d{}, errors.New("get returned multiple nodes")
	}
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Insert(Subspace types.SubspaceId, Path types.Path, Timestamp uint64) error {
	newVal := kdnode.Key{
		Subspace:  Subspace,
		Path:      Path,
		Timestamp: Timestamp,
	}
	if !k.KDTree.Add(newVal) {
		return errors.New("error inserting the node into the KD tree")
	}
	return nil
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Query(QueryRange types.Range3d) []kdnode.Key {
	return kdnode.Query(k.KDTree, QueryRange)
}

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Remove(entry types.Position3d) bool {

	NodeToDelete := kdnode.Key{
		Subspace:  entry.Subspace,
		Timestamp: entry.Time,
		Path:      entry.Path,
	}

	return k.KDTree.Delete(NodeToDelete)
}

// TODO :- Not Fullproof, check triplestorage.ts implementation for further additions
func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) GetInterestRange(areaOfInterest types.AreaOfInterest) types.Range3d {
	newRange := utils.AreaTo3dRange(
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

func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) Summarise(Range types.Range3d) struct {
	FingerPrint string
	Size        uint64
} {
	var size uint64
	var fingerPrint string
	valuesInRange := k.Query(Range)
	for _, val := range valuesInRange {
		size += 1
		fingerPrint = xorStrings(fingerPrint, val.Fingerprint)
	}
	return struct {
		FingerPrint string
		Size        uint64
	}{
		FingerPrint: fingerPrint,
		Size:        size,
	}
}

/* Used for splitting a 3dRange*/
func (k *KDTreeStorage[PreFingerPrint, FingerPrint, K]) SplitRange(Range types.Range3d, size int) (types.Range3d, types.Range3d) {
	entries := k.Query(Range)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})
	//find median
	mid := len(entries) / 2
	median := entries[mid].Timestamp

	//split the range
	leftRange := types.Range3d{
		SubspaceRange: Range.SubspaceRange,
		PathRange:     Range.PathRange,
		TimeRange: types.Range[uint64]{
			Start:   Range.TimeRange.Start,
			End:     median,
			OpenEnd: false,
		},
	}
	rightRange := types.Range3d{
		SubspaceRange: Range.SubspaceRange,
		PathRange:     Range.PathRange,
		TimeRange: types.Range[uint64]{
			Start:   median,
			End:     Range.TimeRange.End,
			OpenEnd: Range.TimeRange.OpenEnd,
		},
	}

	return leftRange, rightRange
}

func xorStrings(a, b string) string {
	// ensure both strings have the same length
	// they should always be the same length 🙄
	// if len(a) > len(b) {
	// 	b += string(make([]byte, len(a)-len(b)))
	// } else if len(b) > len(a) {
	// 	a += string(make([]byte, len(b)-len(a)))
	// }

	if len(a) != len(b) {
		log.Fatal("Hashes of payloads are of different length 😨, fingerprinting.go, line 63")
	}

	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return string(result)
}
