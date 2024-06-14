package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
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

// 	x, y := OrderRangePair(a, b)

// 	if x.End == nil && y.End == nil {
// 		start := x.Start
// 		if order(x.Start, y.Start) <= 0 {
// 			start = y.Start
// 		}

// 		// Return the new Range with calculated start and nil
// 		return types.Range[T]{Start: start, End: nil}
// 	} else if x.End == nil && y.End != nil {
// 		aStartBStartOrder := order(x.Start, y.Start)
// 		aStartBEndOrder := order(x.Start, y.End)

// 		if aStartBStartOrder <= 0 {
// 			return y
// 		} else if aStartBStartOrder > 0 && aStartBEndOrder < 0 {
// 			return types.Range[T]{Start: x.Start, End: y.End}
// 		}
// 		return //something here instead of null, to figure
// 	} else if x.End != nil && y.End != nil {
// 		var min, max types.Range[T]
// 		if order(x.Start, y.Start) < 0 {
// 			min = x
// 		} else {
// 			z=max.End
// 		}
// 		return types.Range[T]{Start: max.Start, End: z}
// 	}
// 	return nil//something like null
// }
//
// // func RangeisIncluded[T constraints.Ordered | types.Path](order types.TotalOrder, p, r types.Range[T]) bool {
// // 	if r.End == nil && p.End != nil {
// // 		return false
// // 	} else if p.End == nil {
// // 		return order(r.Start, p.Start) >= 0
// // 	}
//
// // 	gteStart := order(r.Start, p.Start) >= 0
//
// // 	if !gteStart {
// // 		return false
// // 	}
//
// // 	lteEnd := order(r.End, p.End) <= 0 //as ValueType in ts, check it out
//
// // 	return lteEnd
// // }
//
// // /*func IsValidRange3d [T constraints.Ordered | types.Path](orderSubspace types.TotalOrder, r types.Range3D[T]) bool {
// // 	if !IsValidRange(order)
// // } */
