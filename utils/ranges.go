package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/types"
)

// Constants for open end and closed end representations
var OPEN_END = new(interface{})

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

	// Case when both ranges are closed-ended
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
		if order(max.Start, min.End) >= 0 {
			return false, types.Range[T]{}
		}

		return true, types.Range[T]{
			Start: max.Start,
			End: func() T {
				if order(min.End, max.End) < 0 {
					return min.End
				}
				return max.End
			}(),
		}
	}

	return false, types.Range[T]{}
}

/*
func RangeIsIncluded[T types.OrderableGeneric](order types.TotalOrder[T], parentRange types.Range[T], childRange types.Range[T]) bool{
	if childRange.OpenEnd && !parentRange.OpenEnd{
		return false
	} else {
		gteStart =order(childRange.start, parentRange.start)>=0
		if parentRange.OpenEnd{
			return gteStart
		} else if !gteStart{
			return false
		}

	return order(childRange.End, parentRange.End)<=0
}
} */
//func IsValidRange3d[SubspaceType types.OrderableGeneric](order types.types.TotalOrder)
