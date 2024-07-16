package kv_driver

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func DriverPrefixesOf[K constraints.Unsigned](Path types.Path, pathParams types.PathParams[K], kdt *Kdtree.KDTree[Kdtree.KDNodeKey]) []Kdtree.KDNodeKey {
	prefixes := utils.PrefixesOf(Path)
	prefixes = prefixes[1:(len(prefixes) - 1)]

	var results []Kdtree.KDNodeKey
	var nothing types.SubspaceId

	for _, prefix := range prefixes {
		subspaceRange := types.Range[types.SubspaceId]{
			Start:   nothing,
			End:     nothing,
			OpenEnd: true,
		}

		pathRange := types.Range[types.Path]{
			Start:   prefix,
			End:     utils.SuccessorPath(prefix, pathParams),
			OpenEnd: false,
		}

		timeRange := types.Range[uint64]{
			Start:   0,
			End:     2,
			OpenEnd: true,
		}

		range3d := types.Range3d{
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

func PrefixedBy[K constraints.Unsigned](Subspace types.SubspaceId, Path types.Path, PathParams types.PathParams[K], kdt *(Kdtree.KDTree[Kdtree.KDNodeKey])) []Kdtree.KDNodeKey {
	// var nothing T

	subspaceRange := types.Range[types.SubspaceId]{
		Start:   Subspace,
		End:     Subspace,
		OpenEnd: false,
	}

	pathRange := types.Range[types.Path]{
		Start:   Path,
		End:     utils.SuccessorPrefix(Path, PathParams),
		OpenEnd: false,
	}

	timeRange := types.Range[uint64]{
		Start:   0,
		End:     2,
		OpenEnd: true,
	}

	range3d := types.Range3d{
		SubspaceRange: subspaceRange,
		PathRange:     pathRange,
		TimeRange:     timeRange,
	}
	fmt.Println(range3d, kdt)
	return Kdtree.Query(kdt, range3d)
}
