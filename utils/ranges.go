package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/types"
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

	switch isStartB {
	case true:
		y = b.Start
	case false:
		y = b.End
		
	}


	if !a.OpenEnd && !b.OpenEnd && order(x,y) == 0{
		return true
	}
	return false
}
