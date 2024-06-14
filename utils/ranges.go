package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
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
