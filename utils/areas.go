package utils

import (
	"cmp"
	"fmt"
	"math"
	"strconv"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// Define the options struct
type Options[SubspaceType cmp.Ordered] struct {
	SuccessorSubspace      types.SuccessorFn[SubspaceType]
	MaxPathLength          int
	MaxComponentCount      int
	MaxPathComponentLength int
	MinimalSubspace        SubspaceType
}

type EncodeAreaOpts[SubspaceId constraints.Unsigned] struct {
	EncodeSubspace func(subspace SubspaceId) []byte
	OrderSubspace  types.TotalOrder[SubspaceId]
	PathScheme     types.PathParams[SubspaceId]
}

type EncodeAreaInAreaLengthOptions[SubspaceId constraints.Unsigned] struct {
	EncodeSubspaceIdLength func(subspace SubspaceId) int
	OrderSubspace          types.TotalOrder[SubspaceId]
	PathScheme             types.PathParams[SubspaceId]
}

type DecodeAreaInAreaOptions[SubspaceId constraints.Unsigned] struct {
	decodeSubspaceId EncodingScheme[SubspaceId, uint]
	PathScheme       types.PathParams[SubspaceId]
}

type Result[SubspaceId constraints.Unsigned] struct {
	Area types.Area[SubspaceId]
	Err  error
}

type DecodeStreamAreaInAreaOptions[SubspaceId constraints.Unsigned] struct {
	PathScheme           types.PathParams[SubspaceId]
	DecodeStreamSubspace EncodingScheme[SubspaceId, uint]
}

type EncodeEntryInNamespaceAreaOptions[SubspaceId constraints.Unsigned, PayloadDigest any] struct {
	encodeSubspaceId    func(subspace SubspaceId) []byte
	encodePayloadDigest func(digest PayloadDigest) []byte
	pathScheme          types.PathParams[SubspaceId]
}

func concat(byteSlices ...[]byte) []byte {
	var result []byte
	for _, b := range byteSlices {
		result = append(result, b...)
	}
	return result
}

/** The full area is the Area including all Entries. */
func FullArea[SubspaceId cmp.Ordered]() types.Area[SubspaceId] {
	return types.Area[SubspaceId]{Subspace_id: SubspaceId(0), Any_subspace: true, Path: nil, Times: types.Range[uint64]{Start: 0, End: 0, OpenEnd: true}}
}

/** The subspace area is the Area include all entries with a given subspace ID. */
func SubspaceArea[SubspaceId cmp.Ordered](subspaceId SubspaceId) types.Area[SubspaceId] {
	return types.Area[SubspaceId]{Subspace_id: subspaceId, Any_subspace: false, Path: nil, Times: types.Range[uint64]{Start: 0, End: 0, OpenEnd: true}}
}

/** Return whether a subspace ID is included by an `Area`. */
func IsSubspaceIncludedInArea[SubspaceType cmp.Ordered](orderSubspace types.TotalOrder[SubspaceType], area types.Area[SubspaceType], subspace SubspaceType) bool {
	if area.Any_subspace == true {
		return true
	}

	return orderSubspace(area.Subspace_id, subspace) == 0 //===used here in ts, need to see if the functionality remains the same
}

/** Return whether a 3d position is included by an `Area`. */
func IsIncludedArea[SubspaceType cmp.Ordered](orderSubspace types.TotalOrder[SubspaceType], area types.Area[SubspaceType], position types.Position3d[SubspaceType]) bool {
	if !IsSubspaceIncludedInArea(orderSubspace, area, position.Subspace) {
		return false
	}
	if !IsIncludedRange(OrderTimestamp, area.Times, position.Time) {
		return false
	}
	res, _ := IsPathPrefixed(area.Path, position.Path)
	if !res {
		return false
	}
	return true
}

/** Return whether an area is fully included by another area. */
/** Inner is the area being tested for inclusion. */
/** Outer is the area which we are testing for inclusion within. */
func AreaIsIncluded[SubspaceType cmp.Ordered](orderSubspace types.TotalOrder[SubspaceType], inner, outer types.Area[SubspaceType]) bool {
	if outer.Any_subspace != true && inner.Any_subspace == true {
		return false
	}
	if outer.Any_subspace != true && inner.Any_subspace != true && orderSubspace(outer.Subspace_id, inner.Subspace_id) != 0 {
		return false
	}
	res, _ := IsPathPrefixed(outer.Path, inner.Path)
	if !res {
		return false
	}
	if !RangeIsIncluded(OrderTimestamp, outer.Times, inner.Times) {
		return false
	}
	return true
}

/** Return the intersection of two areas, for which there may be none. */
func IntersectArea[SubspaceType cmp.Ordered](orderSubspace types.TotalOrder[SubspaceType], a, b types.Area[SubspaceType]) *types.Area[SubspaceType] {
	if a.Any_subspace != true && b.Any_subspace != true && orderSubspace(a.Subspace_id, b.Subspace_id) != 0 {
		return nil
	}

	isPrefixA, _ := IsPathPrefixed(a.Path, b.Path) // a.pathPrefix is being checked if it's a prefix of b.pathPrefix
	isPrefixB, _ := IsPathPrefixed(b.Path, a.Path) // b.pathPrefix is being checked if it's a prefix of a.pathPrefix

	if !isPrefixA && !isPrefixB {
		return nil
	}

	choice, timeIntersection := IntersectRange(OrderTimestamp, a.Times, b.Times)

	if choice == false {
		return nil
	}

	if isPrefixA {
		return &types.Area[SubspaceType]{Subspace_id: a.Subspace_id, Path: b.Path, Times: timeIntersection} //we put b.Path here, as a.Path is it's prefix, which means that there's no use of putting a.Path
	}

	return &types.Area[SubspaceType]{Subspace_id: a.Subspace_id, Path: a.Path, Times: timeIntersection}
}

/** Convert an `Area` to a `Range3d`. */
//THIS FUNCTION NEEDS TO BE FIXED
func AreaTo3dRange[T cmp.Ordered](opts Options[T], area types.Area[T]) types.Range3d[T] {
	var subspace_range types.Range[T]
	if area.Any_subspace == true {
		subspace_range = types.Range[T]{Start: opts.MinimalSubspace, End: T(0), OpenEnd: true}
	} else {
		subspace_range = types.Range[T]{
			Start:   area.Subspace_id,
			End:     *opts.SuccessorSubspace(area.Subspace_id), // NEED TO CHANGE THE SUCCESSOR DEFINITION IN ORDER
			OpenEnd: false,
		}
	}
	path_range := types.Range[types.Path]{
		Start:   area.Path,
		End:     SuccessorPrefix(area.Path),
		OpenEnd: false,
	}
	// FIX PATH_RANGE
	return types.Range3d[T]{SubspaceRange: subspace_range, PathRange: path_range, TimeRange: area.Times}
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
func EncodeAreaInArea[SubspaceId constraints.Unsigned](opts EncodeAreaOpts[SubspaceId], inner, outer types.Area[SubspaceId]) []byte {
	if !AreaIsIncluded[SubspaceId](opts.OrderSubspace, inner, outer) {
		fmt.Errorf("Inner is not included by outer")
	}

	var innerEnd uint64

	if inner.Times.OpenEnd == false {
		innerEnd = REALLY_BIG_INT
	} else {
		innerEnd = inner.Times.End
	}

	var outerEnd uint64

	if outer.Times.OpenEnd == false {
		outerEnd = REALLY_BIG_INT
	} else {
		outerEnd = outer.Times.End
	}

	startDiff := bigIntMin(
		inner.Times.Start-outer.Times.Start, outerEnd-inner.Times.Start,
	)

	endDiff := bigIntMin(
		innerEnd-inner.Times.Start, outerEnd-innerEnd,
	)

	flags := byte(0x0)

	isSubspaceSame := (inner.Any_subspace == true && outer.Any_subspace == true) || (inner.Any_subspace != true && outer.Any_subspace != true && (opts.OrderSubspace(inner.Subspace_id, outer.Subspace_id) == 0))

	if !isSubspaceSame {
		flags |= byte(0x80)
	}

	if inner.Times.OpenEnd == false {
		flags |= byte(0x40)
	}

	if startDiff == (inner.Times.Start - outer.Times.Start) {
		flags |= byte(0x20)
	}

	if endDiff == (innerEnd - inner.Times.Start) {
		flags |= byte(0x10)
	}

	startDiffCompactWidth := GetWidthMax64Int(startDiff)

	if startDiffCompactWidth == 4 || startDiffCompactWidth == 8 {
		flags |= byte(0x8)
	}

	if startDiffCompactWidth == 2 || startDiffCompactWidth == 8 {
		flags |= byte(0x4)
	}

	endDiffCompactWidth := GetWidthMax64Int(endDiff)

	if endDiffCompactWidth == 4 || endDiffCompactWidth == 8 {
		flags |= byte(0x2)
	}

	if endDiffCompactWidth == 2 || endDiffCompactWidth == 8 {
		flags |= byte(0x1)
	}

	flagByte := []byte{flags}

	startDiffBytes := EncodeIntMax64(startDiff)
	var endDiffBytes []byte
	if inner.Times.OpenEnd == false {
		endDiffBytes = []byte{}
	} else {
		endDiffBytes = EncodeIntMax64(endDiff)
	}

	relativePathBytes := EncodeRelativePath(opts.PathScheme, inner.Path, outer.Path) // the function to be implemented in path

	var subspaceIdBytes []byte
	if isSubspaceSame {
		subspaceIdBytes = []byte{}
	} else {
		subspaceIdBytes = opts.EncodeSubspace(inner.Subspace_id)
	}

	result := concat(flagByte, startDiffBytes, endDiffBytes, relativePathBytes, subspaceIdBytes)

	return result
}

/** The length of an encoded area in area. */
func EncodeAreaInAreaLength[SubspaceId constraints.Unsigned](opts EncodeAreaInAreaLengthOptions[SubspaceId], inner, outer types.Area[SubspaceId]) int {
	isSubspaceSame := (inner.Any_subspace == true && outer.Any_subspace == true) || (inner.Any_subspace != true && outer.Any_subspace != true && (opts.OrderSubspace(inner.Subspace_id, outer.Subspace_id) == 0))

	var subspaceLen int
	if isSubspaceSame {
		subspaceLen = 0
	} else {
		subspaceLen = opts.EncodeSubspaceIdLength(inner.Subspace_id)
	}

	pathLen := EncodePathRelativeLength(opts.PathScheme, inner.Path, outer.Path) // ask where this is written

	var innerEnd uint64

	if inner.Times.OpenEnd == false {
		innerEnd = REALLY_BIG_INT
	} else {
		innerEnd = inner.Times.End
	}

	var outerEnd uint64

	if outer.Times.OpenEnd == false {
		outerEnd = REALLY_BIG_INT
	} else {
		outerEnd = outer.Times.End
	}

	startDiff := bigIntMin(
		inner.Times.Start-outer.Times.Start, outerEnd-inner.Times.Start,
	)

	endDiff := bigIntMin(
		innerEnd-inner.Times.Start, outerEnd-innerEnd,
	)

	startDiffLen := GetWidthMax64Int(startDiff)

	var endDiffLen int

	if inner.Times.OpenEnd == true {
		endDiffLen = 0
	} else {
		endDiffLen = GetWidthMax64Int(endDiff)
	}

	return 1 + subspaceLen + pathLen + startDiffLen + endDiffLen
}

func DecodeAreaInArea[SubspaceId constraints.Unsigned](opts DecodeAreaInAreaOptions[SubspaceId], encodedInner []byte, outer types.Area[SubspaceId]) types.Area[SubspaceId] {
	flags := encodedInner[0]
	includeInnerSubspaceId := (flags & 0x80) == 0x80
	hasOpenEnd := (flags & 0x40) == 0x40
	addStartDiff := (flags & 0x20) == 0x20
	addEndDiff := (flags & 0x10) == 0x10
	startDiffWidth := int(math.Pow(2, float64(0x3&(flags>>2))))
	endDiffWidth := int(math.Pow(2, float64(0x3&(flags))))

	if hasOpenEnd {
		pathPos := 1 + startDiffWidth
		subarray := encodedInner[1:pathPos]
		startDiff, _ := DecodeIntMax64(subarray)
		path := DecodeRelativePath[SubspaceId](opts.PathScheme, encodedInner[pathPos:], outer.Path)
		subspacePos := pathPos + EncodePathRelativeLength(opts.PathScheme, path, outer.Path)
		var subspaceId SubspaceId
		if includeInnerSubspaceId {
			subspaceId, _ = opts.decodeSubspaceId.Decode(encodedInner[subspacePos:])
		} else {
			subspaceId = outer.Subspace_id
		}
		var innerStart uint64
		if addStartDiff {
			innerStart = outer.Times.Start + startDiff
		} else {
			innerStart = outer.Times.Start - startDiff
		}
		return types.Area[SubspaceId]{Path: path, Subspace_id: subspaceId, Times: types.Range[uint64]{Start: innerStart, End: 0, OpenEnd: true}} // just recheck the return of Subspace_id
	}
	endDiffPos := 1 + startDiffWidth
	pathPos := endDiffPos + endDiffWidth

	startDiff, _ := DecodeIntMax64(encodedInner[1:endDiffPos])
	endDiff, _ := DecodeIntMax64(encodedInner[endDiffPos:pathPos])
	path := DecodeRelativePath[SubspaceId](opts.PathScheme, encodedInner[pathPos:], outer.Path)
	subspacePos := pathPos + EncodePathRelativeLength(opts.PathScheme, path, outer.Path)
	var subspaceId SubspaceId
	if includeInnerSubspaceId {
		subspaceId, _ = opts.decodeSubspaceId.Decode(encodedInner[subspacePos:])
	} else {
		subspaceId = outer.Subspace_id
	}
	var innerStart uint64
	if addStartDiff {
		innerStart = outer.Times.Start + startDiff
	} else {
		innerStart = outer.Times.Start - startDiff
	}
	var innerEnd uint64
	if addEndDiff {
		innerEnd = innerStart + endDiff
	} else {
		innerEnd = outer.Times.End - endDiff
	}

	return types.Area[SubspaceId]{Path: path, Subspace_id: subspaceId, Times: types.Range[uint64]{Start: innerStart, End: innerEnd, OpenEnd: false}}
}

var compactWidthEndMasks = map[int]int{
	1: 0x0,
	2: 0x1,
	4: 0x2,
	8: 0x3,
}

func DecodeStreamAreaInArea[SubspaceId constraints.Unsigned](opts DecodeStreamAreaInAreaOptions[SubspaceId], bytes *GrowingBytes, outer types.Area[SubspaceId]) (types.Area[SubspaceId], error) {
	// TO-DO finish
	accumulatedBytes := bytes.NextAbsolute(1)
	flags := accumulatedBytes[0]

	includeInnerSybspaceId := (flags & 0x80) == 0x80
	hasOpenEnd := (flags & 0x40) == 0x40
	addStartDiff := (flags & 0x20) == 0x20
	addEndDiff := (flags & 0x10) == 0x10
	startDiffWidth := math.Pow(2, float64((0x3 & flags >> 2)))
	endDiffWidth := math.Pow(2, float64((0x3 & flags)))
	var subSpaceId SubspaceId
	var timeReturnStart uint64

	bytes.Prune(1)

	if hasOpenEnd {
		accumulatedBytes = bytes.NextAbsolute(int(startDiffWidth))
		startDiff, _ := DecodeIntMax64(accumulatedBytes[0:int(startDiffWidth)])
		bytes.Prune(int(startDiffWidth))
		path := DecodeRelPathStream(opts.PathScheme, bytes, outer.Path)
		if includeInnerSybspaceId {
			subSpaceId, _ = opts.DecodeStreamSubspace.DrcodeStream(bytes)
		} else {
			subSpaceId = outer.Subspace_id
		}
		if addStartDiff {
			timeReturnStart = outer.Times.Start + startDiff
		} else {
			timeReturnStart = outer.Times.Start - startDiff
		}
		return types.Area[SubspaceId]{
			Path:        path,
			Subspace_id: subSpaceId,
			Times: types.Range[uint64]{
				Start:   timeReturnStart,
				End:     0,
				OpenEnd: true,
			},
			Any_subspace: false,
		}, nil
	}
	accumulatedBytes = bytes.NextAbsolute(int(startDiffWidth))

	startDiff, _ := DecodeIntMax64(accumulatedBytes[0:int(startDiffWidth)])
	bytes.Prune(int(startDiffWidth))

	accumulatedBytes = bytes.NextAbsolute(int(endDiffWidth))
	endDif, _ := DecodeIntMax64(accumulatedBytes[0:int(endDiffWidth)])
	bytes.Prune(int(endDiffWidth))

	path := DecodeRelPathStream(opts.PathScheme, bytes, outer.Path)
	if includeInnerSybspaceId {
		subSpaceId, _ = opts.DecodeStreamSubspace.DrcodeStream(bytes)
	} else {
		subSpaceId = outer.Subspace_id
	}
	if addStartDiff {
		timeReturnStart = outer.Times.Start + startDiff
	} else {
		timeReturnStart = outer.Times.Start - startDiff
	}
	var timeReturnEnd uint64
	if addEndDiff {
		timeReturnEnd = timeReturnStart + endDif
	} else {
		timeReturnEnd = outer.Times.End - endDif
	}

	return types.Area[SubspaceId]{
		Path:        path,
		Subspace_id: subSpaceId,
		Times: types.Range[uint64]{
			Start:   timeReturnStart,
			End:     timeReturnEnd,
			OpenEnd: false,
		},
		Any_subspace: false,
	}, nil
}

/** Encode an {@linkcode Entry} relative to an {@linkcode Area}. */

func EncodeEntryInNamespaceArea[NamespaceId, SubspaceId, PayloadDigest constraints.Unsigned](opts EncodeEntryInNamespaceAreaOptions[SubspaceId, PayloadDigest], entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], outer types.Area[SubspaceId]) []byte {
	var timeDiff uint64
	if outer.Times.OpenEnd == true {
		timeDiff = entry.Timestamp - outer.Times.Start
	} else {
		timeDiff = bigIntMin(entry.Timestamp-outer.Times.Start, outer.Times.End-entry.Timestamp)
	}

	var isSubspaceAnyFlag int
	if outer.Any_subspace == true {
		isSubspaceAnyFlag = 0x80
	} else {
		isSubspaceAnyFlag = 0x0
	}

	var addTimeToStartOrSubtractFromEndFlag int
	if outer.Times.OpenEnd == true {
		addTimeToStartOrSubtractFromEndFlag = 0x40
	} else {
		if entry.Timestamp-outer.Times.Start <= outer.Times.End {
			addTimeToStartOrSubtractFromEndFlag = 0x40
		} else {
			addTimeToStartOrSubtractFromEndFlag = 0x0
		}
	}

	compactWidthFlagsTimeDiff := compactWidthEndMasks[GetWidthMax64Int(timeDiff)] << 4

	compactWidthFlagsPayloadLength := compactWidthEndMasks[GetWidthMax64Int(entry.Payload_length)] << 2

	header := isSubspaceAnyFlag | addTimeToStartOrSubtractFromEndFlag | compactWidthFlagsTimeDiff | compactWidthFlagsPayloadLength

	headerBytes := []byte(strconv.Itoa(header))

	var encodedSubspace []byte
	if outer.Any_subspace == true {
		encodedSubspace = []byte{}
	} else {
		encodedSubspace = opts.encodeSubspaceId(entry.Subspace_id)
	}

	encodedPath := EncodeRelativePath[SubspaceId](opts.pathScheme, entry.Path, outer.Path)

	encodedTimeDiff := GetWidthMax64Int(timeDiff)

	encodedTimeDiffBytes := []byte(strconv.Itoa(encodedTimeDiff))

	encodedPayloadLength := GetWidthMax64Int(entry.Payload_length)

	encodedPayloadLengthBytes := []byte(strconv.Itoa(encodedPayloadLength))

	encodedPayloadDigest := opts.encodePayloadDigest(entry.Payload_digest)

	result := concat(
		headerBytes,
		encodedSubspace,
		encodedPath,
		encodedTimeDiffBytes,
		encodedPayloadLengthBytes,
		encodedPayloadDigest,
	)

	return result
}

/** Decode an Entry relative to a namespace area from {@linkcode GrowingBytes}. */
