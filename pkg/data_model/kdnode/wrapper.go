package kdnode

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	kdtree "github.com/rishitc/go-kd-tree"
)

// Queries the given tree for the given 3-d Range
func Query(kdt *(kdtree.KDTree[Key]), QueryRange types.Range3d) []Key {
	isInRange := func(k Key, d int) kdtree.RelativePosition {
		Timestamp, Subspace, Path := k.Timestamp, k.Subspace, k.Path
		switch d {
		case -1:
			Position := types.Position3d{
				Subspace: Subspace,
				Path:     Path,
				Time:     Timestamp,
			}
			if utils.IsIncluded3d(utils.OrderSubspace, QueryRange, Position) {
				return kdtree.InRange
			} else {
				return kdtree.BeforeRange // Returning anything other than `InRange` works here.
			}
		case 0:
			if utils.OrderTimestamp(Timestamp, QueryRange.TimeRange.Start) >= 0 {
				if QueryRange.TimeRange.OpenEnd || utils.OrderTimestamp(Timestamp, QueryRange.TimeRange.End) < 0 {
					return kdtree.InRange
				} else {
					return kdtree.AfterRange
				}
			} else {
				return kdtree.BeforeRange
			}
		case 1:
			if utils.OrderBytes(Subspace, QueryRange.SubspaceRange.Start) >= 0 {
				if QueryRange.SubspaceRange.OpenEnd || utils.OrderBytes(Subspace, QueryRange.SubspaceRange.End) < 0 {
					return kdtree.InRange
				} else {
					return kdtree.AfterRange
				}
			} else {
				return kdtree.BeforeRange
			}
		case 2:
			if utils.OrderPath(Path, QueryRange.PathRange.Start) >= 0 {
				if QueryRange.PathRange.OpenEnd || utils.OrderPath(Path, QueryRange.PathRange.End) < 0 {
					return kdtree.InRange
				} else {
					return kdtree.AfterRange
				}
			} else {
				return kdtree.BeforeRange
			}
		}
		panic(fmt.Sprintf("Invalid dimension provided: %v", d))
	}
	return kdt.Query(isInRange)
}
