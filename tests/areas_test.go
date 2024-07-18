package tests

// import (
// 	"errors"
// 	"fmt"
// 	"reflect"
// 	"strings"
// 	"testing"

// 	"github.com/PES-Innovation-Lab/willow-go/types"
// 	"github.com/PES-Innovation-Lab/willow-go/utils"
// )

// func SuccessorSubspace(subspace int64) (int64, bool) {
// 	return subspace + 1, true
// }

// // func TestAreaTo3dRange(t *testing.T) {
// // 	_ := []struct {
// // 		name string
// // 		opts utils.Options[int64]
// // 		area types.Area[int64]
// // 		want types.Range3d[int64]
// // 	}{
// // 		{
// // 			name: "Test Case 1: Closed Time Range",
// // 			opts: utils.Options[int64]{
// // 				MinimalSubspace:   300,
// // 				SuccessorSubspace: SuccessorSubspace,
// // 			},
// // 			area: types.Area[int64]{
// // 				Path:        types.Path{[]byte{0, 0, 0, 0}},
// // 				Subspace_id: 1,
// // 				Times:       types.Range[uint64]{Start: 500, End: 1000, OpenEnd: false},
// // 			},
// // 			want: types.Range3d[int64]{
// // 				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
// // 				PathRange:     types.Range[types.Path]{Start: types.Path{[]byte{0, 0, 0, 0}}, End: types.Path{[]byte{0, 0, 0, 1}}, OpenEnd: false},
// // 				TimeRange:     types.Range[uint64]{Start: 500, End: 1000, OpenEnd: false},
// // 			},
// // 		},
// // 		{
// // 			name: "Test Case 2: Open End Time Range",
// // 			opts: utils.Options[int64]{
// // 				MinimalSubspace:   400,
// // 				SuccessorSubspace: SuccessorSubspace,
// // 			},
// // 			area: types.Area[int64]{
// // 				Path:        types.Path{},
// // 				Subspace_id: 1,
// // 				Times:       types.Range[uint64]{Start: 500, End: 0, OpenEnd: true},
// // 			},
// // 			want: types.Range3d[int64]{
// // 				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
// // 				PathRange:     types.Range[types.Path]{Start: types.Path{}, End: types.Path{}, OpenEnd: true},
// // 				TimeRange:     types.Range[uint64]{Start: 500, End: 0, OpenEnd: true},
// // 			},
// // 		},
// // 		{
// // 			name: "Test Case 3: Closed End Time Range",
// // 			opts: utils.Options[int64]{
// // 				MinimalSubspace:   100,
// // 				SuccessorSubspace: SuccessorSubspace,
// // 			},
// // 			area: types.Area[int64]{
// // 				Path:        types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
// // 				Subspace_id: 1,
// // 				Times:       types.Range[uint64]{Start: 7, End: 13, OpenEnd: false},
// // 			},
// // 			want: types.Range3d[int64]{
// // 				SubspaceRange: types.Range[int64]{Start: 1, End: 2, OpenEnd: false},
// // 				PathRange:     types.Range[types.Path]{Start: types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}, End: types.Path{[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}, OpenEnd: false},
// // 				TimeRange:     types.Range[uint64]{Start: 7, End: 13, OpenEnd: false},
// // 			},
// // 		},
// // 	}
// // 	// for _, tt := range tests {
// // 	// 	t.Run(tt.name, func(t *testing.T) {
// // 	// 		if got := utils.AreaTo3dRange(tt.opts, tt.area, tt.PathParams); !reflect.DeepEqual(got, tt.want) {
// // 	// 			t.Errorf("AreaTo3dRange() = %v, want %v", got, tt.want)
// // 	// 		}
// // 	// 	})
// // 	// }
// // }

// // Helper function to convert a string path to types.Path
// func NewPath(path string) types.Path {
// 	components := strings.Split(path, "/")
// 	pathBytes := make(types.Path, len(components))
// 	for i, component := range components {
// 		pathBytes[i] = []byte(component)
// 	}
// 	return pathBytes
// }

// func TestEncodeAreaInArea(t *testing.T) {
// 	type args struct {
// 		opts  utils.EncodeAreaOpts[uint64]
// 		inner types.Area[uint64]
// 		outer types.Area[uint64]
// 	}

// 	areaInAreaVectors := []struct {
// 		inner types.Area[uint64]
// 		outer types.Area[uint64]
// 	}{
// 		{
// 			inner: types.Area[uint64]{
// 				Subspace_id: 1,
// 				Times: types.Range[uint64]{
// 					Start:   10,
// 					End:     20,
// 					OpenEnd: false,
// 				},
// 				Path: NewPath("inner/path"),
// 			},
// 			outer: types.Area[uint64]{
// 				Subspace_id: 1,
// 				Times: types.Range[uint64]{
// 					Start:   0,
// 					End:     30,
// 					OpenEnd: false,
// 				},
// 				Path: NewPath("inner/path2"),
// 			},
// 		},
// 		{
// 			inner: types.Area[uint64]{
// 				Subspace_id: 1,
// 				Times: types.Range[uint64]{
// 					Start:   15,
// 					End:     utils.REALLY_BIG_INT,
// 					OpenEnd: true,
// 				},
// 				Path: NewPath("inner/long/path"),
// 			},
// 			outer: types.Area[uint64]{
// 				Subspace_id: 1,
// 				Times: types.Range[uint64]{
// 					Start:   0,
// 					End:     50,
// 					OpenEnd: false,
// 				},
// 				Path: NewPath("inner/long/path2"),
// 			},
// 		},
// 	}

// 	for _, vector := range areaInAreaVectors {
// 		fmt.Println("Nai")
// 		t.Run("Encode and decode areas", func(t *testing.T) {
// 			opts := utils.EncodeAreaOpts[uint64]{
// 				PathScheme: types.PathParams[uint64]{
// 					MaxComponentCount:  255,
// 					MaxComponentLength: 255,
// 					MaxPathLength:      255,
// 				},
// 				EncodeSubspace: func(v uint64) []byte {
// 					return []byte{byte(v)}
// 				},
// 				OrderSubspace: func(a, b uint64) types.Rel {
// 					if a < b {
// 						return types.Less
// 					} else if a > b {
// 						return types.Greater
// 					}
// 					return types.Equal
// 				},
// 			}

// 			encoded := utils.EncodeAreaInArea(opts, vector.inner, vector.outer)
// 			t.Logf("Encoded bytes: %v", encoded)

// 			decoded, _ := utils.DecodeAreaInArea(utils.DecodeAreaInAreaOptions[uint64]{
// 				PathScheme: types.PathParams[uint64]{
// 					MaxComponentCount:  255,
// 					MaxComponentLength: 255,
// 					MaxPathLength:      255,
// 				},
// 				DecodeSubspaceId: func(encoded []byte) (uint64, error) {
// 					if len(encoded) == 0 {
// 						return 0, errors.New("encoded data is empty")
// 					}
// 					return uint64(encoded[0]), nil
// 				},
// 			}, encoded, vector.outer)
// 			t.Logf("Decoded area: %+v", decoded)

// 			if !reflect.DeepEqual(vector.inner, decoded) {
// 				t.Errorf("Test failed. Expected: %+v, got: %+v", vector.inner, decoded)
// 			}
// 		})
// 	}
// }
