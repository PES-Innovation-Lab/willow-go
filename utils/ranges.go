package utils

import (
	"fmt"
	"math"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// orderRangePair orders two Range structs based on their end values.
func OrderRangePair[T types.OrderableGeneric](a, b types.Range[T]) (types.Range[T], types.Range[T]) {
	if (!a.OpenEnd && !b.OpenEnd) ||
		(a.OpenEnd && b.OpenEnd) ||
		(a.OpenEnd && !b.OpenEnd) {
		return a, b
	}
	return b, a
}

/** Returns whether the range's end is greater than its start. */
func IsValidRange[T types.OrderableGeneric](order types.TotalOrder[T], r types.Range[T]) bool {
	if r.OpenEnd {
		return true
	}
	startEndOrder := order(r.Start, r.End)

	return startEndOrder < 0
}

func IsIncludedRange[T types.OrderableGeneric](order types.TotalOrder[T], r types.Range[T], value T) bool {
	var gteStart bool = order(value, r.Start) >= 0
	if r.OpenEnd || !gteStart {
		return gteStart
	}
	var ltEnd bool = order(value, r.End) == -1
	return ltEnd
}

func IntersectRange[T types.OrderableGeneric](order types.TotalOrder[T], a, b types.Range[T]) (bool, types.Range[T]) {
	if !IsValidRange(order, a) || !IsValidRange(order, b) {
		fmt.Println("Paths are not valid paths... BOZO CAN'T EVEN PASS PATHS PROPERLY AHH")
	}

	a, b = OrderRangePair(a, b)

	// case when both ranges are open-ended
	if a.OpenEnd && b.OpenEnd {
		start := a.Start
		if order(a.Start, b.Start) <= 0 {
			start = b.Start
		}
		return true, types.Range[T]{
			Start:   start,
			OpenEnd: true,
		}
	}

	// case when only 'a' is open-ended
	if a.OpenEnd && !b.OpenEnd {
		aStartBStartOrder := order(a.Start, b.Start)
		aStartBEndOrder := order(a.Start, b.End)

		if aStartBStartOrder <= 0 {
			return true, b
		} else if aStartBStartOrder > 0 && aStartBEndOrder < 0 {
			return true, types.Range[T]{
				Start: a.Start,
				End:   b.End,
			}
		}

		return false, types.Range[T]{}
	}

	// case when both ranges are closed-ended
	if !a.OpenEnd && !b.OpenEnd {
		min := a
		max := b
		if order(a.Start, b.Start) >= 0 {
			min, max = b, a
		}

		// reject if min's end is less than or equal to max's start
		if order(min.End, max.Start) <= 0 {
			return false, types.Range[T]{}
		}

		// reject if max's start is greater than or equal to min's end
		// if order(max.Start, min.End) >= 0 {
		// 	return false, types.Range[T]{}
		// }

		return true, types.Range[T]{
			Start: max.Start,
			End: func() T {
				if order(min.End, max.End) < 0 {
					return min.End
				}
				return max.End
			}(),
			OpenEnd: false,
		}
	}

	return false, types.Range[T]{}
}

func RangeIsIncluded[T types.OrderableGeneric](order types.TotalOrder[T], parentRange types.Range[T], childRange types.Range[T]) bool {
	if childRange.OpenEnd && !parentRange.OpenEnd {
		return false
	} else {
		gteStart := order(childRange.Start, parentRange.Start) >= 0
		if parentRange.OpenEnd {
			return gteStart
		} else if !gteStart {
			return false
		}

		return order(childRange.End, parentRange.End) <= 0

	}
}

func IsValidRange3d[SubspaceId types.OrderableGeneric](OrderSubspace types.TotalOrder[SubspaceId], r types.Range3d[SubspaceId]) bool {
	if !IsValidRange(OrderTimestamp, r.TimeRange) {
		return false
	}
	if !IsValidRange(OrderPath, r.PathRange) {
		return false
	}
	if !IsValidRange(OrderSubspace, r.SubspaceRange) {
		return false
	}
	return true
}

func IsIncluded3d[SubspaceId types.OrderableGeneric](orderSubspace types.TotalOrder[SubspaceId], r types.Range3d[SubspaceId], position types.Position3d[SubspaceId]) bool {
	if !IsIncludedRange(OrderTimestamp, r.TimeRange, position.Time) {
		return false
	}
	if !IsIncludedRange(OrderPath, r.PathRange, position.Path) {
		return false
	}
	if !IsIncludedRange(orderSubspace, r.SubspaceRange, position.Subspace) {
		return false
	}
	return true
}

func IntersectRange3d[SubspaceId types.OrderableGeneric](OrderSubspace types.TotalOrder[SubspaceId], a types.Range3d[SubspaceId], b types.Range3d[SubspaceId]) (bool, types.Range3d[SubspaceId]) {
	ok, intersectionTimestamp := IntersectRange(OrderTimestamp, a.TimeRange, b.TimeRange)
	if !ok {
		return false, types.Range3d[SubspaceId]{}
	}
	ok, intersectionSubspace := IntersectRange(OrderSubspace, a.SubspaceRange, b.SubspaceRange)
	if !ok {
		return false, types.Range3d[SubspaceId]{}
	}
	ok, intersectionPath := IntersectRange(OrderPath, a.PathRange, b.PathRange)
	if !ok {
		return false, types.Range3d[SubspaceId]{}
	}

	return true, types.Range3d[SubspaceId]{
		TimeRange:     intersectionTimestamp,
		PathRange:     intersectionPath,
		SubspaceRange: intersectionSubspace,
	}
}

func IsEqualRangeValue[T types.OrderableGeneric](order types.TotalOrder[T], a types.Range[T], isStartA bool, b types.Range[T], isStartB bool) bool {
	if a.OpenEnd && b.OpenEnd {
		return true
	}

	var x, y T

	switch isStartA {
	case true:
		x = a.Start
	case false:
		x = a.End
	}

	switch isStartB {
	case true:
		y = b.Start
	case false:
		y = b.End
	}

	if !a.OpenEnd && !b.OpenEnd && order(x, y) == 0 {
		return true
	}
	return false
}

func AbsDiffuint64(a uint64, b uint64) uint64 {
	/* return absolute value of a - b uint64 values*/
	if a < b {
		return (b - a)
	} else {
		return (a - b)
	}
}

func EncodeRange3dRelative[SubspaceId types.OrderableGeneric, T constraints.Unsigned](
	orderSubspace types.TotalOrder[SubspaceId],
	encodeSubspaceId func(subspace SubspaceId) []byte,
	pathScheme types.PathParams[T],
	r types.Range3d[SubspaceId],
	ref types.Range3d[SubspaceId],
) []byte {
	start_to_start := AbsDiffuint64(r.TimeRange.Start, ref.TimeRange.Start)

	if ref.TimeRange.OpenEnd {
		ref.TimeRange.End = math.MaxUint64
		if r.TimeRange.OpenEnd {
			r.TimeRange.End = math.MaxUint64
		}
	}

	start_to_end := AbsDiffuint64(r.TimeRange.Start, ref.TimeRange.End)
	end_to_start := AbsDiffuint64(r.TimeRange.End, ref.TimeRange.Start)
	end_to_end := AbsDiffuint64(r.TimeRange.End, ref.TimeRange.End)
	start_time_diff := min(start_to_start, start_to_end)
	end_time_diff := min(end_to_start, end_to_end)

	var encoding1 byte = 0x00
	var encoding2 byte = 0x00

	var subspaceIdEncodingStart []byte
	var subspaceIdEncodingEnd []byte

	var pathEncodingStart []byte
	var pathEncodingEnd []byte

	// Encode byte 1

	// encoding bits 0, 1
	if IsEqualRangeValue(orderSubspace, r.SubspaceRange, true, ref.SubspaceRange, true) {
		encoding1 = encoding1 | 0x40
	} else if IsEqualRangeValue(orderSubspace, r.SubspaceRange, true, ref.SubspaceRange, false) {
		encoding1 = encoding1 | 0x80
	} else {
		encoding1 = encoding1 | 0xC0
		subspaceIdEncodingStart = encodeSubspaceId(r.SubspaceRange.Start)
	}

	// eoncoding bits at 2, 3
	if r.SubspaceRange.OpenEnd {
		// do nothing
	} else if IsEqualRangeValue(orderSubspace, r.SubspaceRange, false, ref.SubspaceRange, true) {
		encoding1 = encoding2 | 0x10
	} else if IsEqualRangeValue(orderSubspace, r.SubspaceRange, false, ref.SubspaceRange, false) {
		encoding1 = encoding1 | 0x20
	} else {
		encoding1 = encoding1 | 0x30
		subspaceIdEncodingEnd = encodeSubspaceId(r.SubspaceRange.End)
	}

	// encoding bit 4
	prefixStartStart, _ := CommonPrefix(r.PathRange.Start, ref.PathRange.Start)
	prefixStartEnd, _ := CommonPrefix(r.PathRange.Start, ref.PathRange.End)

	if len(prefixStartStart) >= len(prefixStartEnd) {
		encoding1 = encoding1 | 0x08
		pathEncodingStart = EncodeRelativePath(pathScheme, r.PathRange.Start, ref.PathRange.Start)
	} else {
		pathEncodingStart = EncodeRelativePath(pathScheme, r.PathRange.Start, ref.PathRange.End)
	}

	// encoding bit 5
	if r.PathRange.OpenEnd {
		encoding1 = encoding1 | 0x04
	}

	// encoding bit 6
	prefixEndStart, _ := CommonPrefix(r.PathRange.End, ref.PathRange.Start)
	prefixEndEnd, _ := CommonPrefix(r.PathRange.End, ref.PathRange.End)
	if len(prefixEndStart) >= len(prefixEndEnd) {
		encoding1 = encoding1 | 0x02
		pathEncodingEnd = EncodeRelativePath(pathScheme, r.PathRange.End, ref.PathRange.Start)
	} else {
		pathEncodingEnd = EncodeRelativePath(pathScheme, r.PathRange.End, ref.PathRange.End)
	}
	if r.PathRange.OpenEnd {
		encoding1 = encoding1 & 0xFD
	}

	// encoding big 7
	if r.TimeRange.OpenEnd {
		encoding1 = encoding1 | 0x01
	}

	// encoding byte 2

	// encoding bit 8 (0)
	if start_to_start <= start_to_end {
		encoding2 = encoding2 | 0x80
	}

	// encoding bit 9 (1)
	if (encoding2 & 0x80) == 0x80 {
		if r.TimeRange.Start >= ref.TimeRange.Start {
			encoding2 = encoding2 | 0x40
		}
	} else {
		if r.TimeRange.Start >= ref.TimeRange.End {
			encoding2 = encoding2 | 0x40
		}
	}

	// encoding bit 10, 11 (2,3)
	_compactWidth := GetWidthMax64Int(start_time_diff)

	switch _compactWidth {
	case 2:
		encoding2 = encoding2 | 0x10
	case 4:
		encoding2 = encoding2 | 0x20
	case 8:
		encoding2 = encoding2 | 0x30
	}

	// encoding bit 12 (4)
	if end_to_start <= end_to_end {
		encoding2 = encoding2 | 0x08
	}
	// encoding bit 13 (5)
	if (encoding2 & 0x08) == 0x08 {
		if r.TimeRange.End >= ref.TimeRange.Start {
			encoding2 = encoding2 | 0x04
		}
	} else {
		if r.TimeRange.End >= ref.TimeRange.End {
			encoding2 = encoding2 | 0x04
		}
	}

	// encoding bit 14, 15 (6,7)
	switch GetWidthMax64Int(end_time_diff) {
	case 2:
		encoding2 = encoding2 | 0x01
	case 4:
		encoding2 = encoding2 | 0x02
	case 8:
		encoding2 = encoding2 | 0x03
	}

	var end_time_diff_encoding []byte
	if !r.TimeRange.OpenEnd && ref.TimeRange.OpenEnd {
		end_time_diff_encoding = EncodeIntMax64(end_time_diff)
	} else {
		end_time_diff_encoding = EncodeIntMax64(end_to_start)
	}
	// remaining encoding information ->
	start_time_diff_encoding := EncodeIntMax64(start_time_diff)

	EncodedArray := concat(
		[]byte{encoding1, encoding2},
		subspaceIdEncodingStart,
		subspaceIdEncodingEnd,
		pathEncodingStart,
		pathEncodingEnd,
		start_time_diff_encoding,
		end_time_diff_encoding,
	)
	return EncodedArray
}

func DecodeStreamRange3dRelative[SubspaceId constraints.Ordered, K constraints.Unsigned](
	DecodeStreamSubspaceId func(bytes *GrowingBytes) SubspaceId,
	pathScheme types.PathParams[K],
	bytes *GrowingBytes,
	ref types.Range3d[SubspaceId],
) (types.Range3d[SubspaceId], error) {
	accumulatedBytes := bytes.NextAbsolute(2)
	firstByte, secondByte := accumulatedBytes[0], accumulatedBytes[1]

	var isSubspaceStartEncoded string

	switch true {
	case (firstByte & 0xc0) == 0xc0:
		isSubspaceStartEncoded = "yes"
	case (firstByte & 0x80) == 0x80:
		isSubspaceStartEncoded = "ref_end"
	case (firstByte & 0x40) == 0x40:
		isSubspaceStartEncoded = "ref_start"
	default:
		isSubspaceStartEncoded = "invalid"
	}
	if isSubspaceStartEncoded == "invalid" {
		return types.Range3d[SubspaceId]{}, fmt.Errorf("invalid subspace")
	}

	var isSubspaceEndEncoded string

	switch true {
	case (firstByte & 0x30) == 0x30:
		isSubspaceStartEncoded = "yes"
	case (firstByte & 0x20) == 0x20:
		isSubspaceStartEncoded = "ref_end"
	case (firstByte & 0x10) == 0x10:
		isSubspaceStartEncoded = "ref_start"
	default:
		isSubspaceStartEncoded = "open"
	}

	isPathStartRelativeToRefStart := (firstByte & 0x8) == 0x8

	isRangePathEndOpen := (firstByte & 0x4) == 0x40

	isPathEndEncodedRelToRefStart := (firstByte & 0x2) == 0x2

	isTimeEndOpen := (firstByte & 0x1) == 0x1

	encodeTimeStartRelToRefTimeStart := (secondByte & 0x80) == 0x80

	addStartTimeDiff := (secondByte & 0x40) == 0x40

	compactWidthStartTimeDiff := math.Pow(2, float64((secondByte&0x30)>>4))

	encodeTimeEndRelToRefStart := (secondByte & 0x8) == 0x8

	addEndTimeDiff := (secondByte & 0x4) == 0x4

	compactWidthEndTimeDiff := math.Pow(2, float64(secondByte&0x3))

	bytes.Prune(2)

	var subspaceStart SubspaceId

	switch isSubspaceStartEncoded {
	case "ref_start":
		subspaceStart = ref.SubspaceRange.Start
	case "ref_end":
		if !ref.SubspaceRange.OpenEnd {
			subspaceStart = ref.SubspaceRange.End
		} else {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("start value cannot be open ended")
		}
	case "yes":
		subspaceStart = DecodeStreamSubspaceId(bytes)
	}

	var subspaceEnd SubspaceId
	var subspaceOpenEnd bool

	switch isSubspaceEndEncoded {
	case "open":
		subspaceEnd = subspaceStart
		subspaceOpenEnd = true
	case "ref_start":
		subspaceEnd = ref.SubspaceRange.Start
		subspaceOpenEnd = false
	case "ref_end":
		subspaceEnd = ref.SubspaceRange.End
		subspaceOpenEnd = false
	case "yes":
		subspaceEnd = DecodeStreamSubspaceId(bytes)
		subspaceOpenEnd = false
	}

	var pathStart types.Path

	if isPathStartRelativeToRefStart {
		pathStart = DecodeRelPathStream(pathScheme, bytes, ref.PathRange.Start)
	} else {
		if ref.PathRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the start of a path range cannot be encoded relative to an open end")
		}
		pathStart = DecodeRelPathStream(pathScheme, bytes, ref.PathRange.End)
	}

	var pathEnd types.Path
	var pathOpenEnd bool

	if isRangePathEndOpen {
		pathEnd = types.Path{}
		pathOpenEnd = true
	} else if isPathEndEncodedRelToRefStart {
		pathEnd = DecodeRelPathStream(pathScheme, bytes, ref.PathRange.Start)
	} else {
		if ref.PathRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the end of a path range cannot be encoded relative to an open end")
		}
		pathEnd = DecodeRelPathStream(pathScheme, bytes, ref.PathRange.End)
	}
	accumulatedBytes = bytes.NextAbsolute(int(compactWidthStartTimeDiff))

	startTimeDiff, err := DecodeIntMax64(accumulatedBytes[0:int(compactWidthStartTimeDiff)])
	if err != nil {
		return types.Range3d[SubspaceId]{}, fmt.Errorf("could not decode startTime")
	}

	bytes.Prune(int(compactWidthStartTimeDiff))

	var timeStart uint64

	if encodeTimeStartRelToRefTimeStart {
		if addStartTimeDiff {
			timeStart = ref.TimeRange.Start + uint64(startTimeDiff)
		} else {
			timeStart = ref.TimeRange.Start - uint64(startTimeDiff)
		}
	} else {
		if ref.TimeRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the start of a time cannot be open ended")
		}
		if addStartTimeDiff {
			timeStart = ref.TimeRange.End + uint64(startTimeDiff)
		} else {
			timeStart = ref.TimeRange.End - uint64(startTimeDiff)
		}
	}
	var timeEnd uint64
	var timeOpenEnd bool

	if isTimeEndOpen {
		timeOpenEnd = true
		timeEnd = timeStart
	} else {
		accumulatedBytes = bytes.NextAbsolute(int(compactWidthEndTimeDiff))

		endTimeDiff, err := DecodeIntMax64(accumulatedBytes[0:int(compactWidthEndTimeDiff)])
		if err != nil {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("could not decode end time difference")
		}

		if encodeTimeEndRelToRefStart {
			if addEndTimeDiff {
				timeEnd = ref.TimeRange.End + uint64(endTimeDiff)
			} else {
				timeEnd = ref.TimeRange.End - uint64(endTimeDiff)
			}
		} else {
			if ref.TimeRange.OpenEnd {
				return types.Range3d[SubspaceId]{}, fmt.Errorf("end of timerange cannot be encoded relative to open end")
			}
			if addEndTimeDiff {
				timeEnd = ref.TimeRange.End + uint64(endTimeDiff)
			} else {
				timeEnd = ref.TimeRange.End - uint64(endTimeDiff)
			}
		}
	}
	return types.Range3d[SubspaceId]{
		SubspaceRange: types.Range[SubspaceId]{
			Start:   subspaceStart,
			End:     subspaceEnd,
			OpenEnd: subspaceOpenEnd,
		},
		PathRange: types.Range[types.Path]{
			Start:   pathStart,
			End:     pathEnd,
			OpenEnd: pathOpenEnd,
		},
		TimeRange: types.Range[uint64]{
			Start:   timeStart,
			End:     timeEnd,
			OpenEnd: timeOpenEnd,
		},
	}, nil

}

/** Decode an {@linkcode Range3d} relative to another `Range3d` from {@linkcode GrowingBytes}. */
func DecodeRange3dRelative[SubspaceId types.OrderableGeneric, T constraints.Unsigned](
	decodeSubspaceId func(encoded []byte) SubspaceId,
	encodedSubspacIdLength func(subspace SubspaceId) T,
	pathScheme types.PathParams[T],
	encoded []byte,
	ref types.Range3d[SubspaceId],
) (types.Range3d[SubspaceId], error) {
	firstByte, secondByte := encoded[0], encoded[1]

	//Decoding the first byte
	//Bit 0,1
	var isSubspaceStartEncoded string
	switch true {
	case (firstByte & 0xc0) == 0xc0:
		isSubspaceStartEncoded = "yes"

	case (firstByte & 0x80) == 0x80:
		isSubspaceStartEncoded = "ref_end"

	case (firstByte & 0x40) == 0x40:
		isSubspaceStartEncoded = "ref_start"
	default:
		isSubspaceStartEncoded = "invalid"
	}

	if isSubspaceStartEncoded == "invalid" {
		return types.Range3d[SubspaceId]{}, fmt.Errorf("invalid 3d range relative to relative 3d range encoding")
	}

	//Bit 2,3
	var isSubspaceEndEncoded string
	switch true {
	case (firstByte & 0x30) == 0x30:
		isSubspaceEndEncoded = "yes"

	case (firstByte & 0x20) == 0x20:
		isSubspaceEndEncoded = "ref_end"

	case (firstByte & 0x10) == 0x10:
		isSubspaceEndEncoded = "ref_start"
	default:
		isSubspaceEndEncoded = "open"
	}

	//Bit 4
	isPathStartRelativeToRefStart := (firstByte & 0x8) == 0x8

	//Bit 5
	isRangePathEndOpen := (firstByte & 0x4) == 0x4

	//Bit 6
	isPathEndEncodedRelToRefStart := (firstByte & 0x2) == 0x2

	//Bit 7
	isTimeEndOpen := (firstByte & 0x1) == 0x1

	//Decoding the second byte

	//Bit 8
	encodeTimeStartRelToRefTimeStart := (secondByte & 0x80) == 0x80
	//Bit 9
	addStartTimeDiff := (secondByte & 0x40) == 0x40
	//Bit 10-11
	compactWidthStartTimeDiff := T(math.Pow(2, float64((secondByte&0x30)>>4)))
	//Bit 12
	encodeTimeEndRelToRefStart := (secondByte & 0x8) == 0x8
	//Bit 13
	addEndTimeDiff := (secondByte & 0x4) == 0x4
	//Bit 14-15
	compactWidthEndTimeDiff := T(math.Pow(2, float64(secondByte&0x3)))

	var position T = 2

	//Subspace Start
	var SubspaceStart SubspaceId

	switch isSubspaceStartEncoded {
	case "ref_start":
		SubspaceStart = ref.SubspaceRange.Start

	case "ref_end":
		if !ref.SubspaceRange.OpenEnd {
			SubspaceStart = ref.SubspaceRange.End
		} else {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the start value of an encoded range cannot be that of the reference end (open)")
		}

	case "yes":
		SubspaceStart = decodeSubspaceId(encoded[position:])
		position += encodedSubspacIdLength(SubspaceStart)
	}

	//Subspace End
	var SubspaceEnd SubspaceId
	var isSubspaceOpenEnd bool = false
	switch isSubspaceEndEncoded {
	case "open":
		isSubspaceOpenEnd = true

	case "ref_start":
		SubspaceEnd = ref.SubspaceRange.Start

	case "ref_end":
		SubspaceEnd = ref.SubspaceRange.End

	case "yes":
		SubspaceEnd = decodeSubspaceId(encoded[position:])
		position += encodedSubspacIdLength(SubspaceEnd)

	}

	//Path Start
	var PathStart types.Path

	if isPathStartRelativeToRefStart {
		PathStart = DecodeRelativePath(
			pathScheme,
			encoded[position:],
			ref.PathRange.Start,
		)
		position += T(EncodePathRelativeLength(
			pathScheme,
			PathStart,
			ref.PathRange.Start,
		))
	} else {
		if ref.PathRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the start of a path range cannot be encoded relative to an open end")
		}

		PathStart = DecodeRelativePath(
			pathScheme,
			encoded[position:],
			ref.PathRange.End,
		)
		position += T(EncodePathRelativeLength(
			pathScheme,
			PathStart,
			ref.PathRange.End,
		))
	}

	//Path End

	var PathEnd types.Path
	var isPathOpenEnd bool = false
	if isRangePathEndOpen {
		isPathOpenEnd = true
	} else if isPathEndEncodedRelToRefStart {
		PathEnd = DecodeRelativePath(
			pathScheme,
			encoded[position:],
			ref.PathRange.Start,
		)
		position += T(EncodePathRelativeLength(
			pathScheme,
			PathEnd,
			ref.PathRange.Start,
		))
	} else {
		if ref.PathRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the end of a path range cannot be encoded relative to an open end")
		}

		PathEnd = DecodeRelativePath(
			pathScheme,
			encoded[position:],
			ref.PathRange.End,
		)
		position += T(EncodePathRelativeLength(
			pathScheme,
			PathEnd,
			ref.PathRange.End,
		))

	}

	//Time Start

	StartTimeDiff, err := DecodeIntMax64(encoded[position : position+compactWidthStartTimeDiff])
	if err != nil {
		return types.Range3d[SubspaceId]{}, fmt.Errorf("could not decode starttimediff")
	}
	position += compactWidthStartTimeDiff

	var timeStart uint64

	if encodeTimeStartRelToRefTimeStart {
		if addStartTimeDiff {
			timeStart = ref.TimeRange.Start + StartTimeDiff
		} else {
			timeStart = ref.TimeRange.Start - StartTimeDiff
		}
	} else {
		if ref.TimeRange.OpenEnd {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("the start of a time range cannot be encoded relative to an open end")
		}
		if addStartTimeDiff {
			timeStart = ref.TimeRange.End + StartTimeDiff
		} else {
			timeStart = ref.TimeRange.End - StartTimeDiff
		}
	}

	//Time End
	var timeEnd uint64
	var isTimeOpenEnd bool = false

	if isTimeEndOpen {
		isTimeOpenEnd = true
	} else {
		EndTimeDiff, err := DecodeIntMax64(encoded[position : position+compactWidthEndTimeDiff])
		if err != nil {
			return types.Range3d[SubspaceId]{}, fmt.Errorf("could not decode endtimediff")
		}

		if encodeTimeEndRelToRefStart {
			if addEndTimeDiff {
				timeStart = ref.TimeRange.Start + EndTimeDiff
			} else {
				timeStart = ref.TimeRange.Start - EndTimeDiff
			}
		} else {
			if ref.TimeRange.OpenEnd {
				return types.Range3d[SubspaceId]{}, fmt.Errorf("the start of a time range cannot be encoded relative to an open end")
			}
			if addEndTimeDiff {
				timeStart = ref.TimeRange.End + EndTimeDiff
			} else {
				timeStart = ref.TimeRange.End - EndTimeDiff
			}
		}
	}

	return types.Range3d[SubspaceId]{
		SubspaceRange: types.Range[SubspaceId]{
			Start:   SubspaceStart,
			End:     SubspaceEnd,
			OpenEnd: isSubspaceOpenEnd,
		},
		PathRange: types.Range[types.Path]{
			Start:   PathStart,
			End:     PathEnd,
			OpenEnd: isPathOpenEnd,
		},
		TimeRange: types.Range[uint64]{
			Start:   timeStart,
			End:     timeEnd,
			OpenEnd: isTimeOpenEnd,
		},
	}, nil
}

func DefaultRange3d[SubspaceId constraints.Ordered](defaultSubspaceId SubspaceId) types.Range3d[SubspaceId] {
	return types.Range3d[SubspaceId]{
		SubspaceRange: types.Range[SubspaceId]{
			Start:   defaultSubspaceId,
			OpenEnd: true,
		},
		PathRange: types.Range[types.Path]{
			Start:   types.Path{},
			OpenEnd: true,
		},
		TimeRange: types.Range[uint64]{
			Start:   0,
			OpenEnd: true,
		},
	}
}
