package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"golang.org/x/exp/constraints"
)

// Constants for open end and closed end representations
var OPEN_END = new(interface{})

// orderRangePair orders two Range structs based on their end values.
func OrderRangePair[T constraints.Ordered | types.Path](a, b types.Range[T]) (types.Range[T], types.Range[T]) {
	if (a.End == nil && b.End == nil) ||
		(a.End != nil && b.End != nil) ||
		(a.End == nil && b.End != nil) {
		return a, b
	}

	return b, a
}

/** Returns whether the range's end is greater than its start. */
func IsValidRange[T constraints.Ordered](order types.TotalOrder, r types.Range[T]) bool {
	if r.end == OPEN_END {
		return true
	}

	startEndOrder := order(r.Start, r.End)

	return startEndOrder < 0
}

/** Returns whether a `Value` is included by a given `Range`. */
func IsIncludedRange[T constraints.Ordered | types.Path](order types.TotalOrder, r types.Range[T], value T) bool {
	gteStart := order(value, r.Start) >= 0

	if r.End == nil || !gteStart {
		return gteStart
	}

	ltEnd := order(value, r.End) == -1

	return ltEnd
}

func IntersectRange[T constraints.Ordered | types.Path](order types.TotalOrder, a, b types.Range[T]) types.Range[T] {
	
	if err := IsValidRange(order, a); err != nil {
		fmt.Println("Error with range a:", err)
	}
	if err := IsValidRange(order, b); err != nil {
		fmt.Println("Error with range b:", err) // Expected
	}

	x, y := OrderRangePair(a, b)

	if x.End == nil && y.End == nil {
		start := x.Start
		if order(x.Start, y.Start) <= 0 {
			start = y.Start
		}

		// Return the new Range with calculated start and nil
		return types.Range[T]{Start: start, End: nil}
	} else if x.End == nil && y.End != nil {
		aStartBStartOrder := order(x.Start, y.Start)
		aStartBEndOrder := order(x.Start, y.End)

		if aStartBStartOrder <= 0 {
			return y
		} else if aStartBStartOrder > 0 && aStartBEndOrder < 0 {
			return types.Range[T]{Start: x.Start, End: y.End}
		}
		return //something here instead of null, to figure
	} else if x.End != nil && y.End != nil {
		var min, max types.Range[T]
		if order(x.Start, y.Start) < 0 {
			min = x
		} else {
			min = y
		}
		if (x, min){
			max = y
		} else {
			max = x
		}

		// reject if min's end is lte max's start
		if order(min.End, max.Start) <= 0 {
			return //something like null
		}

		// reject if max's start is gte min's end
		if order(max.Start, min.End) >= 0 {
			return //something like null
		}
		var z types.Range[T]
		if order(min.End, max.End) < 0 { //as ValueType in ts, find out why and if it has to implemented in go
			z=min.End
		} else {
			z=max.End
		}
		return types.Range[T]{Start: max.Start, End: z}
	}
	return //something like null
}

func RangeisIncluded[T constraints.Ordered | types.Path](order types.TotalOrder, p, r types.Range[T]) bool {
	if (r.End == nil && p.End != nil){
		return false
	} else if (p.End == nil){
		return order(r.Start, p.Start) >= 0
	}

	gteStart := order(r.Start, p.Start) >= 0

	if !gteStart {
		return false
	}

	lteEnd := order(r.End, p.End) <= 0 //as ValueType in ts, check it out

	return lteEnd
}

func IsValidRange3d [T constraints.Ordered | types.Path](orderSubspace types.TotalOrder, r types.Range3D[T]) bool {
	if !IsValidRange(order)
}