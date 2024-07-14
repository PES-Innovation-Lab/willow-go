package kv_driver

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func DriverPrefixesOf[T constraints.Ordered](Path types.Path, pathParams types.PathParams[uint64], kdt *Kdtree.KDTree[Kdtree.KDNodeKey[T]]) []Kdtree.KDNodeKey[T] {
	prefixes := utils.PrefixesOf(Path)
	prefixes = prefixes[1:(len(prefixes) - 1)]

	var results []Kdtree.KDNodeKey[T]
	var nothing T

	for _, prefix := range prefixes {
		subspaceRange := types.Range[T]{
			Start:   nothing,
			End:     nothing,
			OpenEnd: true,
		}

		pathRange := types.Range[types.Path]{
			Start:   prefix,
			End:     utils.SuccessorPath[uint64](prefix, pathParams),
			OpenEnd: false,
		}

		timeRange := types.Range[uint64]{
			Start:   0,
			End:     2,
			OpenEnd: true,
		}

		range3d := types.Range3d[T]{
			SubspaceRange: subspaceRange,
			PathRange:     pathRange,
			TimeRange:     timeRange,
		}

		queryResults := Kdtree.Query(kdt, range3d)
		results = append(results, queryResults...)
	}

	fmt.Println(results)
	return results
}

func PrefixedBy[T constraints.Ordered](Path types.Path, PathParams types.PathParams[uint64], kdt *(Kdtree.KDTree[Kdtree.KDNodeKey[T]])) []Kdtree.KDNodeKey[T] {
	var nothing T

	subspaceRange := types.Range[T]{
		Start:   nothing,
		End:     nothing,
		OpenEnd: true,
	}

	pathRange := types.Range[types.Path]{
		Start:   Path,
		End:     utils.SuccessorPrefix[uint64](Path, PathParams),
		OpenEnd: false,
	}

	timeRange := types.Range[uint64]{
		Start:   0,
		End:     2,
		OpenEnd: true,
	}

	range3d := types.Range3d[T]{
		SubspaceRange: subspaceRange,
		PathRange:     pathRange,
		TimeRange:     timeRange,
	}
	fmt.Println(range3d, kdt)
	return Kdtree.Query(kdt, range3d)
}
