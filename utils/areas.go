package utils

import (
	"fmt"
	"math"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// Define the options struct
type Options[SubspaceType constraints.Ordered] struct {
	SuccessorSubspace      types.SuccessorFn[SubspaceType]
	MaxPathLength          int
	MaxComponentCount      int
	MaxPathComponentLength int
	MinimalSubspace        SubspaceType
}

type EncodeAreaOpts[SubspaceId constraints.Ordered] struct {
	EncodeSubspace func(subspace SubspaceId) []byte
	OrderSubspace  types.TotalOrder[SubspaceId]
	PathScheme     types.PathParams[SubspaceId]
}

/** The full area is the Area including all Entries. */
func FullArea[SubspaceId constraints.Ordered]() types.Area[SubspaceId] {
	return types.Area[SubspaceId]{Subspace_id: nil, Path: nil, Times: types.Range[uint64]{Start: 0, End: nil}}
}

/** The subspace area is the Area include all entries with a given subspace ID. */
func SubspaceArea[SubspaceId constraints.Ordered](subspaceId SubspaceId) types.Area[SubspaceId] {
	return types.Area[SubspaceId]{Subspace_id: nil, Path: nil, Times: types.Range[uint64]{Start: 0, End: nil}}
}

/** Return whether a subspace ID is included by an `Area`. */
func IsSubspaceIncludedInArea[SubspaceType constraints.Ordered](orderSubspace types.TotalOrder[SubspaceType], area types.Area[SubspaceType], subspace SubspaceType) bool {
	if area.Subspace_id == nil {
		return true
	}

	return orderSubspace(*area.Subspace_id, subspace) == 0 //===used here in ts, neeed to see if the functionality remains the same
}

/** Return whether a 3d position is included by an `Area`. */
func IsIncludedArea[SubspaceType constraints.Ordered](orderSubspace types.TotalOrder[SubspaceType], area types.Area[SubspaceType], position types.Position3d[SubspaceType]) bool {
	if !IsSubspaceIncludedInArea(orderSubspace, area, position.Subspace) {
		return false
	}
	if !IsIncludedRange(OrderTimestamp, area.Times, position.Time) {
		return false
	}
	if !IsPathPrefixed(area.Path, position.Path) {
		return false
	}
	return true
}

/** Return whether an area is fully included by another area. */
func AreaIsIncluded[SubspaceType constraints.Ordered](orderSubspace types.TotalOrder[SubspaceType], inner, outer types.Area[SubspaceType]) bool {
	if outer.Subspace_id != nil && inner.Subspace_id == nil {
		return false
	}
	if outer.Subspace_id != nil && inner.Subspace_id != nil && orderSubspace(*outer.Subspace_id, *inner.Subspace_id) != 0 {
		return false
	}
	if !IsPathPrefixed(outer.Path, inner.Path) {
		return false
	}
	if !RangeisIncluded(OrderTimestamp, outer.Times, inner.Times) {
		return false
	}
	return true
}

/** Return the intersection of two areas, for which there may be none. */
func IntersectArea[SubspaceType constraints.Ordered](orderSubspace types.TotalOrder[SubspaceType], a, b types.Area[SubspaceType]) *types.Area[SubspaceType] {
	if a.Subspace_id != nil && b.Subspace_id != nil && orderSubspace(*a.Subspace_id, *b.Subspace_id) != 0 {
		return nil
	}

	isPrefixA := IsPathPrefixed(a.Path, b.Path)
	isPrefixB := IsPathPrefixed(b.Path, a.Path)

	if !isPrefixA && !isPrefixB {
		return nil
	}

	timeIntersection := IntersectRange(OrderTimestamp, a.Times, b.Times)

	if timeIntersection == nil {
		return nil
	}

	if isPrefixA {
		return &types.Area[SubspaceType]{Subspace_id: a.Subspace_id, Path: b.Path, Times: *timeIntersection}
	}

	return &types.Area[SubspaceType]{Subspace_id: a.Subspace_id, Path: a.Path, Times: *timeIntersection}
}

/** Convert an `Area` to a `Range3d`. */
func AreaTo3dRange[SubspaceType constraints.Ordered](opts Options[SubspaceType], area types.Area[SubspaceType]) types.Range3D[SubspaceType] {
	var subspace_range types.Range[SubspaceType]
	if area.Subspace_id == nil {
		subspace_range = types.Range[SubspaceType]{Start: opts.MinimalSubspace, End: nil}
	} else {
		subspace_range = types.Range[SubspaceType]{
			Start: *area.Subspace_id,
			End:   opts.SuccessorSubspace(*area.Subspace_id)}
	}
	var path_range types.Range[SubspaceType]
	path_range = types.Range[SubspaceType]{
		Start: SubspaceType(area.Path),
		End:   SuccessorPrefix(area.Path) || nil,
	}
	//FIX PATH_RANGE
	return types.Range3D[SubspaceType]{Subspaces: subspace_range, Paths: path_range, Times: area.Times}
}

// Define a constant for a really big integer (2^64 in this case)
const REALLY_BIG_INT uint64 = 18446744073709551601

/** `Math.min`, but for `BigInt`. */
// bigIntMin returns the minimum of two big.Int values
func bigIntMin(a, b uint64) uint64 {
	if a > b {
		return b
	}
	// Check for overflow (a - b might overflow uint64)
	if a-b > math.MaxUint64 {
		return 0 // Or handle overflow appropriately
	}
	return a
}

/** Encode an `Area` relative to known outer `Area`.
 *
 * https://willowprotocol.org/specs/encodings/index.html#enc_area_in_area
 */
func EncodeAreaInArea[SubspaceId constraints.Ordered](opts EncodeAreaOpts[SubspaceId], inner, outer types.Area[SubspaceId]) []byte {
	if !AreaIsIncluded[SubspaceId](opts.OrderSubspace, inner, outer) {
		fmt.Errorf("Inner is not included by outer")
	}

	var innerEnd uint64

	if inner.Times.End == nil {
		innerEnd = REALLY_BIG_INT
	} else {
		innerEnd = *inner.Times.End
	}

	var outerEnd uint64

	if outer.Times.End == nil {
		outerEnd = REALLY_BIG_INT
	} else {
		outerEnd = *outer.Times.End
	}

	startDiff := bigIntMin(
		inner.Times.Start-outer.Times.Start, outerEnd-inner.Times.Start,
	)

	endDiff := bigIntMin(
		innerEnd-inner.Times.Start, outerEnd-innerEnd,
	)

	flags := 0x0

	isSubspaceSame := (inner.Subspace_id == nil && outer.Subspace_id == nil) || (inner.Subspace_id != nil && outer.Subspace_id != nil && (opts.OrderSubspace(*inner.Subspace_id, *outer.Subspace_id) == 0))

	if !isSubspaceSame {
		flags |= 0x80
	}

	if inner.Times.End == nil {
		flags |= 0x40
	}

	if startDiff == (inner.Times.Start - outer.Times.Start) {
		flags |= 0x20
	}

	if endDiff == (innerEnd - inner.Times.Start) {
		flags |= 0x10
	}

	startDiffCompactWidth := compactWidth(startDiff) //to be done in encoding

	if startDiffCompactWidth == 4 || startDiffCompactWidth == 8 {
		flags |= 0x8
	}

	if startDiffCompactWidth == 2 || startDiffCompactWidth == 8 {
		flags |= 0x4
	}

	endDiffCompactWidth := compactWidth(endDiff)
}
