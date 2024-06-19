package utils

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func SuccessorSubspace(subspace int64) (int64, bool) {
	return subspace + 1, true
}

func TestAreaTo3dRange(t *testing.T) {
	tests := []struct {
		name string
		opts utils.Options[int64]
		area types.Area[int64]
		want types.Range3d[int64]
	}{
		{
			name: "Test Case 1: Closed Time Range",
			opts: utils.Options[int64]{
				MinimalSubspace:   300,
				SuccessorSubspace: SuccessorSubspace,
			},
			area: types.Area[int64]{
				Path:        types.Path{[]byte{0, 0, 0, 0}},
				Subspace_id: 1,
				Times:       types.Range[uint64]{Start: 500, End: 1000, OpenEnd: false},
			},
			want: types.Range3d[int64]{
				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: types.Path{[]byte{0, 0, 0, 0}}, End: types.Path{[]byte{0, 0, 0, 1}}, OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: 500, End: 1000, OpenEnd: false},
			},
		},
		{
			name: "Test Case 2: Open End Time Range",
			opts: utils.Options[int64]{
				MinimalSubspace:   400,
				SuccessorSubspace: SuccessorSubspace,
			},
			area: types.Area[int64]{
				Path:        types.Path{},
				Subspace_id: 1,
				Times:       types.Range[uint64]{Start: 500, End: 0, OpenEnd: true},
			},
			want: types.Range3d[int64]{
				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: types.Path{}, End: types.Path{}, OpenEnd: true},
				TimeRange:     types.Range[uint64]{Start: 500, End: 0, OpenEnd: true},
			},
		},
		{
			name: "Test Case 3: Closed End Time Range",
			opts: utils.Options[int64]{
				MinimalSubspace:   100,
				SuccessorSubspace: SuccessorSubspace,
			},
			area: types.Area[int64]{
				Path:        types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
				Subspace_id: 1,
				Times:       types.Range[uint64]{Start: 7, End: 13, OpenEnd: false},
			},
			want: types.Range3d[int64]{
				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
				PathRange:     types.Range[types.Path]{Start: types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}, End: types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}, OpenEnd: false},
				TimeRange:     types.Range[uint64]{Start: 7, End: 13, OpenEnd: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.AreaTo3dRange(tt.opts, tt.area); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AreaTo3dRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
