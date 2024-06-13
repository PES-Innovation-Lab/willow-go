package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"github.com/PES-Innovation-Lab/willow-go/src/pkg/utils"
	"golang.org/x/exp/constraints"
)

// checks if range end is greater than range start
func IsValidRange[T constraints.Ordered | types.Path](r types.Range[T]) bool {
	if r.End == nil {
		// open range, always valid
		return true
	}
	// checks if T is Path, if yes, stores it in startPath and compares
	if startPath, ok := any(r.Start).(Path); ok {
		endPath := any(*r.End).(Path)
		return utils.OrderPath(endPath, startPath) < 1
	}

	// normal comparison for constraints.Ordered types
	return *r.End >= r.Start
}
