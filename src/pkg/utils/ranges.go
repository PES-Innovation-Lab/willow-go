package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"golang.org/x/exp/constraints"
)

// Constants for open end and closed end representations
var OPEN_END = new(interface{})

// orderRangePair orders two Range structs based on their end values.
func OrderRangePair[T types.OrderableGeneric](a, b types.Range[T]) (types.Range[T], types.Range[T]) {
	if (a.End == nil && b.End == nil) ||
		(a.End != nil && b.End != nil) ||
		(a.End == nil && b.End != nil) {
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

func IntersectRange[T types.OrderableGeneric](order types.TotalOrder[T], a, b types.Range[T]) types.Range[T] {
	
	if !IsValidRange(order, a) || !IsValidRange(order, b) {
		fmt.Println("Paths are not valid paths... BOZO CAN'T EVEN PASS PATHS PROPERLY AHH")
	}

	a, b :=  OrderRangePair(a, b)
	if a.OpenEnd && b.OpenEnd {
		return types.Range[T] {
			Start: func() T { if order(a.Start, b.Start) <= 0 { return b.Start } else { return a.Start } }(),
			End: 0,
			OpenEnd: true,
		}
	} else if a.OpenEnd && !b.OpenEnd {
		aStartbStartOrder := order(a.Start, b.Start)
		aStartbEndOrder := order(a.Start, b.Start)

		if aStartbStartOrder <= 0 {
			return b
		} else if aStartbStartOrder > 0 && aStartbEndOrder < 0 {
			return types.Range[T] {
				Start: a.Start,
				End: b.End,
			}
		}
	}

func RangeIsIncluded[T types.OrderableGeneric](order types.TotalOrde, parentRange types.Range, childRange types.Range) bool{
	if childRange.OpenEnd && !parentRange.OpenEnd{
		return false
	}
	else{
		gteStart =order(childRange.start, parentRange.start)>=0
		if parentRange.OpenEnd{
			return gteStart
		}
		else if !gteStart{
			return false
		}

	return order(childRange.End, parentRange.End)<=0	
}


