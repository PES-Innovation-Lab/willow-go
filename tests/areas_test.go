package tests

import (
	"cmp"
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

type Options[SubspaceType cmp.Ordered] struct {
	SuccessorSubspace      types.SuccessorFn[SubspaceType]
	MaxPathLength          int
	MaxComponentCount      int
	MaxPathComponentLength int
	MinimalSubspace        SubspaceType
}

type testCase struct {
	name     string
	opts     Options[int]
	area     types.Area[int]
	expected types.Range3d[int]
}

func TestAreaTo3dRange(t *testing.T) {
	testCases := []testCase{
		{
			name: "Any subspace, empty path",
			opts: Options[int]{
				MinimalSubspace: 1,
				SuccessorSubspace: func(x int) *int {
					y := x + 1
					return &y
				},
			},
			area: types.Area[int]{
				Any_subspace: true,
				Subspace_id:  0,
				Path:         types.Path{},
				Times:        types.Range[uint64]{Start: 10, End: 20, OpenEnd: false},
			},
			expected: types.Range3d[int]{
				SubspaceRange: types.Range[int]{Start: 1, End: 0, OpenEnd: true},
				PathRange:     types.Range[types.Path]{Start: types.Path{}, End: types.Path{}, OpenEnd: true},
				TimeRange:     types.Range[uint64]{Start: 10, End: 20, OpenEnd: false},
			},
		},
		{
			name: "Specific subspace, non-empty path",
			opts: Options[int]{
				MinimalSubspace: 1,
				SuccessorSubspace: func(x int) *int {
					y := x + 1
					return &y
				},
			},
			area: types.Area[int]{
				Any_subspace: false,
				Subspace_id:  2,
				Path:         types.Path{[]byte("1"), []byte("2"), []byte("3")},
				Times:        types.Range[uint64]{Start: 5, End: 15, OpenEnd: true},
			},
			expected: types.Range3d[int]{
				SubspaceRange: types.Range[int]{Start: 2, End: 3, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: types.Path{[]byte("1"), []byte("2"), []byte("3")}, End: types.Path{[]byte("1"), []byte("2"), []byte("4")}, OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: 5, End: 15, OpenEnd: true},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.AreaTo3dRange(tc.opts, tc.area)
			if !equalRange3d(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func equalRange3d[T cmp.Ordered](r1, r2 types.Range3d[T]) bool {
	return equalRange(r1.SubspaceRange, r2.SubspaceRange) &&
		equalRangePath(r1.PathRange, r2.PathRange) &&
		equalRange(r1.TimeRange, r2.TimeRange)
}

func equalRange[T cmp.Ordered](r1, r2 types.Range[T]) bool {
	return r1.Start == r2.Start && r1.End == r2.End && r1.OpenEnd == r2.OpenEnd
}

func equalRangePath(r1, r2 types.Range[types.Path]) bool {
	return reflect.DeepEqual(r1.Start, r2.Start) &&
		reflect.DeepEqual(r1.End, r2.End) &&
		r1.OpenEnd == r2.OpenEnd
}
