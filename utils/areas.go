package utils

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// Define the options struct
type Options[K constraints.Unsigned] struct {
	MinimalSubspace        types.SubspaceId
	SuccessorSubspace      types.SuccessorFn[types.SubspaceId]
	MaxPathLength          K
	MaxComponentCount      K
	MaxPathComponentLength K
}

type EntryOpts[K constraints.Unsigned] struct {
	DecodeStreamSubspace      func(bytes *GrowingBytes) chan types.SubspaceId
	DecodeStreamPayloadDigest func(bytes *GrowingBytes) chan types.PayloadDigest
	PathScheme                types.PathParams[K]
}

type EncodeAreaOpts[Params constraints.Unsigned] struct {
	EncodeSubspace func(subspace types.SubspaceId) []byte
	OrderSubspace  types.TotalOrder[types.SubspaceId]
	PathScheme     types.PathParams[Params]
}

type EncodeAreaInAreaLengthOptions[Params constraints.Unsigned] struct {
	EncodeSubspaceIdLength func(subspace types.SubspaceId) int
	OrderSubspace          types.TotalOrder[types.SubspaceId]
	PathScheme             types.PathParams[Params]
}

type DecodeAreaInAreaOptions[Params constraints.Unsigned] struct {
	DecodeSubspaceId func(encoded []byte) (types.SubspaceId, error)
	PathScheme       types.PathParams[Params]
}

type Result struct {
	Err  error
	Area types.Area
}

type DecodeStreamAreaInAreaOptions[Params constraints.Unsigned] struct {
	PathScheme           types.PathParams[Params]
	DecodeStreamSubspace EncodingScheme[types.SubspaceId]
}

type EncodeEntryInNamespaceAreaOptions[Params constraints.Unsigned] struct {
	EncodeSubspaceId    func(subspace types.SubspaceId) []byte
	EncodePayloadDigest func(digest types.PayloadDigest) []byte
	PathScheme          types.PathParams[Params]
}

func concat(byteSlices ...[]byte) []byte {
	var result []byte
	for _, b := range byteSlices {
		result = append(result, b...)
	}
	return result
}

func isEmpty(path types.Path) bool {
	return len(path) == 0
}

/** The full area is the Area including all Entries. */
func FullArea() types.Area {
	return types.Area{Any_subspace: true, Path: nil, Times: types.Range[uint64]{Start: 0, End: 0, OpenEnd: true}}
}

/** The subspace area is the Area include all entries with a given subspace ID. */
func SubspaceArea(subspaceId types.SubspaceId) types.Area {
	return types.Area{Subspace_id: subspaceId, Any_subspace: false, Path: nil, Times: types.Range[uint64]{Start: 0, End: 0, OpenEnd: true}}
}

/** Return whether a subspace ID is included by an `Area`. */
func IsSubspaceIncludedInArea(orderSubspace types.TotalOrder[types.SubspaceId], area types.Area, subspace types.SubspaceId) bool {
	if area.Any_subspace {
		return true
	}

	return orderSubspace(area.Subspace_id, subspace) == 0 //===used here in ts, need to see if the functionality remains the same
}

/** Return whether a 3d position is included by an `Area`. */
func IsIncludedArea(orderSubspace types.TotalOrder[types.SubspaceId], area types.Area, position types.Position3d) bool {
	if !IsSubspaceIncludedInArea(orderSubspace, area, position.Subspace) {
		return false
	}
	if !IsIncludedRange(OrderTimestamp, area.Times, position.Time) {
		return false
	}
	res, _ := IsPathPrefixed(area.Path, position.Path)
	return res
}

/** Return whether an area is fully included by another area. */
/** Inner is the area being tested for inclusion. */
/** Outer is the area which we are testing for inclusion within. */
func AreaIsIncluded(orderSubspace types.TotalOrder[types.SubspaceId], inner, outer types.Area) bool {
	if !outer.Any_subspace && inner.Any_subspace {
		return false
	}
	if !outer.Any_subspace && !inner.Any_subspace && orderSubspace(outer.Subspace_id, inner.Subspace_id) != 0 {
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
func IntersectArea(orderSubspace types.TotalOrder[types.SubspaceId], a, b types.Area) *types.Area {
	if !a.Any_subspace && !b.Any_subspace && orderSubspace(a.Subspace_id, b.Subspace_id) != 0 {
		return nil
	}

	isPrefixA, _ := IsPathPrefixed(a.Path, b.Path) // a.pathPrefix is being checked if it's a prefix of b.pathPrefix
	isPrefixB, _ := IsPathPrefixed(b.Path, a.Path) // b.pathPrefix is being checked if it's a prefix of a.pathPrefix

	if !isPrefixA && !isPrefixB {
		return nil
	}

	choice, timeIntersection := IntersectRange(OrderTimestamp, a.Times, b.Times)

	if !choice {
		return nil
	}

	if isPrefixA {
		return &types.Area{Subspace_id: a.Subspace_id, Path: b.Path, Times: timeIntersection} // we put b.Path here, as a.Path is it's prefix, which means that there's no use of putting a.Path
	}

	return &types.Area{Subspace_id: a.Subspace_id, Path: a.Path, Times: timeIntersection}
}

/** Convert an `Area` to a `Range3d`. */
//THIS FUNCTION NEEDS TO BE FIXED
func AreaTo3dRange[Params constraints.Unsigned](opts Options[Params], area types.Area) types.Range3d {
	var subspace_range types.Range[types.SubspaceId]
	if !area.Any_subspace {
		sucSubspace := opts.SuccessorSubspace(area.Subspace_id)
		if !reflect.DeepEqual(sucSubspace, []byte{}) {
			subspace_range = types.Range[types.SubspaceId]{
				Start:   area.Subspace_id,
				End:     sucSubspace, // NEED TO CHANGE THE SUCCESSOR DEFINITION IN ORDER
				OpenEnd: false,
			}
		} else {
			subspace_range = types.Range[types.SubspaceId]{
				Start:   area.Subspace_id,
				OpenEnd: true,
			}
		}

	} else {
		subspace_range = types.Range[types.SubspaceId]{Start: opts.MinimalSubspace, OpenEnd: true}
	}
	var path_range types.Range[types.Path]

	// Create a copy of area.Path to preserve its original value
	startPath := make(types.Path, len(area.Path))
	copy(startPath, area.Path)

	end := SuccessorPrefix(area.Path, types.PathParams[Params]{
		MaxComponentCount:  opts.MaxComponentCount,
		MaxComponentLength: opts.MaxPathComponentLength,
		MaxPathLength:      opts.MaxPathLength,
	}) // Use the copied startPath
	var choice bool
	if isEmpty(end) {
		end = types.Path{}
		choice = true
	} else {
		choice = false
	}

	path_range = types.Range[types.Path]{
		Start:   startPath,
		End:     end,
		OpenEnd: choice,
	}

	return types.Range3d{SubspaceRange: subspace_range, PathRange: path_range, TimeRange: area.Times}
}

// Define a constant for a really big integer (2^64 in this case)
const REALLY_BIG_INT uint64 = 18446744073709551601

/** `Math.min`, but for `BigInt`. */
// bigIntMin returns the minimum of two big.Int values
func bigIntMin(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

/** Encode an `Area` relative to known outer `Area`.
 *
 * https://willowprotocol.org/specs/encodings/index.html#enc_area_in_area
 */
func EncodeAreaInArea[Params constraints.Unsigned](opts EncodeAreaOpts[Params], inner, outer types.Area) []byte {
	if !AreaIsIncluded(opts.OrderSubspace, inner, outer) {
		fmt.Println("Inner is not included by outer")
	}

	var innerEnd uint64

	if inner.Times.OpenEnd {
		innerEnd = REALLY_BIG_INT
	} else {
		innerEnd = inner.Times.End
	}

	var outerEnd uint64

	if outer.Times.OpenEnd {
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

	flags := 0x0

	isSubspaceSame := (inner.Any_subspace && outer.Any_subspace) || (!inner.Any_subspace && !outer.Any_subspace && (opts.OrderSubspace(inner.Subspace_id, outer.Subspace_id) == 0))

	if !isSubspaceSame {
		flags |= 0x80
	}

	if inner.Times.OpenEnd {
		flags |= 0x40
	}

	if startDiff == (inner.Times.Start - outer.Times.Start) {
		flags |= 0x20
	}

	if endDiff == (innerEnd - inner.Times.Start) {
		flags |= 0x10
	}

	startDiffCompactWidth := GetWidthMax64Int(startDiff)

	if startDiffCompactWidth == 4 || startDiffCompactWidth == 8 {
		flags |= 0x8
	}

	if startDiffCompactWidth == 2 || startDiffCompactWidth == 8 {
		flags |= 0x4
	}

	endDiffCompactWidth := GetWidthMax64Int(endDiff)

	if endDiffCompactWidth == 4 || endDiffCompactWidth == 8 {
		flags |= 0x2
	}

	if endDiffCompactWidth == 2 || endDiffCompactWidth == 8 {
		flags |= 0x1
	}

	flagByte := []byte{byte(flags)}

	startDiffBytes := EncodeIntMax64(startDiff)
	var endDiffBytes []byte
	if inner.Times.OpenEnd {
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
func EncodeAreaInAreaLength[Params constraints.Unsigned](opts EncodeAreaInAreaLengthOptions[Params], inner, outer types.Area) int {
	isSubspaceSame := (inner.Any_subspace && outer.Any_subspace) || (!inner.Any_subspace && !outer.Any_subspace && (opts.OrderSubspace(inner.Subspace_id, outer.Subspace_id) == 0))

	var subspaceLen int
	if isSubspaceSame {
		subspaceLen = 0
	} else {
		subspaceLen = len(inner.Subspace_id)
	}

	pathLen := EncodePathRelativeLength(opts.PathScheme, inner.Path, outer.Path) // ask where this is written

	var innerEnd uint64

	if !inner.Times.OpenEnd {
		innerEnd = REALLY_BIG_INT
	} else {
		innerEnd = inner.Times.End
	}

	var outerEnd uint64

	if !outer.Times.OpenEnd {
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

	if inner.Times.OpenEnd {
		endDiffLen = 0
	} else {
		endDiffLen = GetWidthMax64Int(endDiff)
	}

	return 1 + subspaceLen + pathLen + startDiffLen + endDiffLen
}

func DecodeAreaInArea[Params constraints.Unsigned](opts DecodeAreaInAreaOptions[Params], encodedInner []byte, outer types.Area) (types.Area, error) {
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

		startDiff, err := DecodeIntMax64(subarray)
		if err != nil {
			return types.Area{}, fmt.Errorf("error decoding start diff: %w", err)
		}

		path := DecodeRelativePath[Params](opts.PathScheme, encodedInner[pathPos:], outer.Path)
		subspacePos := pathPos + EncodePathRelativeLength(opts.PathScheme, path, outer.Path)
		var subspaceId types.SubspaceId
		if includeInnerSubspaceId {
			subspaceId, _ = opts.DecodeSubspaceId(encodedInner[subspacePos:])
		} else {
			subspaceId = outer.Subspace_id
		}
		var innerStart uint64
		if addStartDiff {
			innerStart = outer.Times.Start + startDiff
		} else {
			innerStart = outer.Times.Start - startDiff
		}
		return types.Area{Path: path, Subspace_id: subspaceId, Times: types.Range[uint64]{Start: innerStart, End: REALLY_BIG_INT, OpenEnd: true}}, nil // just recheck the return of Subspace_id
	}
	endDiffPos := 1 + startDiffWidth
	pathPos := endDiffPos + endDiffWidth

	startDiff, err := DecodeIntMax64(encodedInner[1:endDiffPos])
	if err != nil {
		return types.Area{}, fmt.Errorf("error decoding start diff: %w", err)
	}
	endDiff, err := DecodeIntMax64(encodedInner[endDiffPos:pathPos])
	if err != nil {
		return types.Area{}, fmt.Errorf("error decoding end diff: %w", err)
	}
	path := DecodeRelativePath[Params](opts.PathScheme, encodedInner[pathPos:], outer.Path)
	subspacePos := pathPos + EncodePathRelativeLength(opts.PathScheme, path, outer.Path)
	var subspaceId types.SubspaceId
	if includeInnerSubspaceId {
		subspaceId, _ = opts.DecodeSubspaceId(encodedInner[subspacePos:])
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

	return types.Area{Path: path, Subspace_id: subspaceId, Times: types.Range[uint64]{Start: innerStart, End: innerEnd, OpenEnd: false}}, nil
}

var compactWidthEndMasks = map[int]int{
	1: 0x0,
	2: 0x1,
	4: 0x2,
	8: 0x3,
}

func DecodeStreamAreaInArea[Params constraints.Unsigned](
	opts DecodeStreamAreaInAreaOptions[Params],
	bytes *GrowingBytes,
	outer types.Area,
) (types.Area, error) {
	accumulatedBytes := bytes.NextAbsolute(1)
	flags := accumulatedBytes[0]

	includeInnerSybspaceId := (flags & 0x80) == 0x80
	hasOpenEnd := (flags & 0x40) == 0x40
	addStartDiff := (flags & 0x20) == 0x20
	addEndDiff := (flags & 0x10) == 0x10
	startDiffWidth := math.Pow(2, float64((0x3 & flags >> 2)))
	endDiffWidth := math.Pow(2, float64((0x3 & flags)))
	var subSpaceId types.SubspaceId
	var timeReturnStart uint64

	bytes.Prune(1)

	if hasOpenEnd {
		accumulatedBytes = bytes.NextAbsolute(int(startDiffWidth))
		startDiff, err := DecodeIntMax64(accumulatedBytes[0:int(startDiffWidth)])
		if err != nil {
			return types.Area{}, fmt.Errorf("error decoding startdiff: %v", err)
		}
		bytes.Prune(int(startDiffWidth))
		path := DecodeRelPathStream(opts.PathScheme, bytes, outer.Path)
		if includeInnerSybspaceId {
			var err error
			subSpaceId = <-opts.DecodeStreamSubspace.DecodeStream(bytes)
			if subSpaceId == nil {
				return types.Area{}, fmt.Errorf("error decoding subspace: %v", err)
			}
		} else {
			subSpaceId = outer.Subspace_id
		}
		if addStartDiff {
			timeReturnStart = outer.Times.Start + startDiff
		} else {
			timeReturnStart = outer.Times.Start - startDiff
		}
		return types.Area{
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

	startDiff, err := DecodeIntMax64(accumulatedBytes[0:int(startDiffWidth)])
	if err != nil {
		return types.Area{}, fmt.Errorf("error decoding startdiff: %v", err)
	}
	bytes.Prune(int(startDiffWidth))

	accumulatedBytes = bytes.NextAbsolute(int(endDiffWidth))
	endDif, err := DecodeIntMax64(accumulatedBytes[0:int(endDiffWidth)])
	if err != nil {
		return types.Area{}, fmt.Errorf("error decoding enddiff: %v", err)
	}
	bytes.Prune(int(endDiffWidth))

	path := DecodeRelPathStream(opts.PathScheme, bytes, outer.Path)
	if includeInnerSybspaceId {
		subSpaceId = <-opts.DecodeStreamSubspace.DecodeStream(bytes)
		if subSpaceId == nil {
			return types.Area{}, fmt.Errorf("error decoding subspace: %v", err)
		}
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

	return types.Area{
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

func EncodeEntryInNamespaceArea[Params constraints.Unsigned](
	opts EncodeEntryInNamespaceAreaOptions[Params],
	entry types.Entry,
	outer types.Area,
) []byte {
	var timeDiff uint64
	if outer.Times.OpenEnd {
		timeDiff = entry.Timestamp - outer.Times.Start
	} else {
		timeDiff = bigIntMin(entry.Timestamp-outer.Times.Start, outer.Times.End-entry.Timestamp)
	}

	var isSubspaceAnyFlag int
	if outer.Any_subspace {
		isSubspaceAnyFlag = 0x80
	} else {
		isSubspaceAnyFlag = 0x0
	}

	var addTimeToStartOrSubtractFromEndFlag int
	if outer.Times.OpenEnd {
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
	if outer.Any_subspace {
		encodedSubspace = []byte{}
	} else {
		encodedSubspace = opts.EncodeSubspaceId(entry.Subspace_id)
	}

	encodedPath := EncodeRelativePath[Params](opts.PathScheme, entry.Path, outer.Path)

	encodedTimeDiff := GetWidthMax64Int(timeDiff)

	encodedTimeDiffBytes := []byte(strconv.Itoa(encodedTimeDiff))

	encodedPayloadLength := GetWidthMax64Int(entry.Payload_length)

	encodedPayloadLengthBytes := []byte(strconv.Itoa(encodedPayloadLength))

	encodedPayloadDigest := opts.EncodePayloadDigest(entry.Payload_digest)

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
func DecodeStreamEntryInNamespaceArea[T constraints.Unsigned](opts EntryOpts[T], bytes *GrowingBytes, outer types.Area, nameSpaceId types.NamespaceId) (types.Entry, error) {
	accumulatedBytes := bytes.NextAbsolute(1)
	header := accumulatedBytes[0]

	isSubspaceEncoded := (header & 0x80) == 0x80
	addToStartOrSubtractFromEnd := (header & 0x40) == 0x40
	compactWidthTimeDiff := math.Pow(2, float64(header&0x30>>4))
	compactWidthPayloadLength := math.Pow(2, float64(header&0xc>>2))

	bytes.Prune(1)
	var subspaceId types.SubspaceId

	if isSubspaceEncoded {
		subspaceId = <-opts.DecodeStreamSubspace(bytes)
	} else if !outer.Any_subspace {
		subspaceId = outer.Subspace_id
	} else {
		return types.Entry{}, fmt.Errorf("entry was encoded relative to area")
	}

	path := DecodeRelPathStream(opts.PathScheme, bytes, outer.Path)
	accumulatedBytes = bytes.NextAbsolute(int(compactWidthTimeDiff))

	timeDiff, err := DecodeIntMax64(accumulatedBytes[0:int(compactWidthTimeDiff)])
	if err != nil {
		return types.Entry{}, err
	}
	accumulatedBytes = bytes.NextAbsolute(int(compactWidthTimeDiff))
	payloadLength, err := DecodeIntMax64(accumulatedBytes[int(compactWidthTimeDiff) : int(compactWidthTimeDiff)+int(compactWidthPayloadLength)])
	if err != nil {
		return types.Entry{}, err
	}
	bytes.Prune(int(compactWidthTimeDiff) + int(compactWidthPayloadLength))
	payloadDigest := opts.DecodeStreamPayloadDigest(bytes)

	var timeStamp uint64

	if addToStartOrSubtractFromEnd {
		timeStamp = outer.Times.Start + timeDiff
	} else if !outer.Times.OpenEnd {
		timeStamp = outer.Times.End - timeDiff
	} else {
		return types.Entry{}, fmt.Errorf("entry was encoded relative to area with concrete time end")
	}
	return types.Entry{
		Namespace_id:   nameSpaceId,
		Subspace_id:    subspaceId,
		Path:           path,
		Payload_digest: <-payloadDigest,
		Payload_length: payloadLength,
		Timestamp:      timeStamp,
	}, nil
}
