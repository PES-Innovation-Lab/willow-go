package kv_driver

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type PrefixDriver[PathParamValue constraints.Unsigned] struct{}

func (PD *PrefixDriver[PathParamValue]) DriverPrefixesOf(Subspace types.SubspaceId, Path types.Path, pathParams types.PathParams[PathParamValue], kdt *Kdtree.KDTree[Kdtree.KDNodeKey]) []Kdtree.KDNodeKey {
	prefixes := utils.PrefixesOf(Path)
	prefixes = prefixes[1:(len(prefixes) - 1)]

	var results []Kdtree.KDNodeKey
	// var nothing types.SubspaceId

	for _, prefix := range prefixes {
		subspaceRange := types.Range[types.SubspaceId]{
			Start:   Subspace,
			End:     utils.SuccessorSubspaceId(Subspace),
			OpenEnd: false,
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
	return results
}

func (PD *PrefixDriver[PathParamValue]) PrefixedBy(Subspace types.SubspaceId, Path types.Path, PathParams types.PathParams[PathParamValue], kdt *(Kdtree.KDTree[Kdtree.KDNodeKey])) []Kdtree.KDNodeKey {
	// var nothing T
	subspaceRange := types.Range[types.SubspaceId]{
		Start:   Subspace,
		End:     utils.SuccessorSubspaceId(Subspace),
		OpenEnd: false,
	}

	pathRange := types.Range[types.Path]{
		Start:   utils.SuccessorPath(Path, PathParams),
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
	res := Kdtree.Query(kdt, range3d)

	return res
}
