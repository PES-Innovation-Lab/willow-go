package kdnode

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	kdtree "github.com/rishitc/go-kd-tree"
)

func compareKey(a, b Key) bool {
	return a.Timestamp == b.Timestamp &&
		utils.OrderSubspace(a.Subspace, b.Subspace) == 0 &&
		reflect.DeepEqual(a.Path, b.Path)
}

func sortKeys(keys []Key) {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Timestamp != keys[j].Timestamp {
			return keys[i].Timestamp < keys[j].Timestamp
		}
		if !reflect.DeepEqual(keys[i].Subspace, keys[j].Subspace) {
			return utils.OrderSubspace(keys[i].Subspace, keys[j].Subspace) < 0
		}
		return reflect.DeepEqual(keys[i].Path, keys[j].Path)
	})
}

var TestPathParams types.PathParams[uint8] = types.PathParams[uint8]{
	MaxComponentCount:  50,
	MaxComponentLength: 50,
	MaxPathLength:      50,
}

func setupKDTree() *kdtree.KDTree[Key] {
	return kdtree.NewKDTreeWithValues[Key](3, []Key{
		{
			Subspace:    []byte("david"),
			Timestamp:   140,
			Path:        types.Path{{0}},
			Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
		},
		{
			Subspace:    []byte("betty"),
			Timestamp:   100,
			Path:        types.Path{{6}},
			Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
		},
		{
			Subspace:    []byte("alpha"),
			Timestamp:   500,
			Path:        types.Path{{3}},
			Fingerprint: "XYZABC123456",
		},
		{
			Subspace:    []byte("beta"),
			Timestamp:   200,
			Path:        types.Path{{2}},
			Fingerprint: "XYZABC654321",
		},
		{
			Subspace:    []byte("gamma"),
			Timestamp:   300,
			Path:        types.Path{{1}},
			Fingerprint: "XYZABC987654",
		},
		{
			Subspace:    []byte("delta"),
			Timestamp:   600,
			Path:        types.Path{{4}},
			Fingerprint: "XYZABC321789",
		},
		{
			Subspace:    []byte("epsilon"),
			Timestamp:   700,
			Path:        types.Path{{5}},
			Fingerprint: "XYZABC654987",
		},
		{
			Subspace:    []byte("zeta"),
			Timestamp:   800,
			Path:        types.Path{{7}},
			Fingerprint: "XYZABC789123",
		},
		{
			Subspace:    []byte("eta"),
			Timestamp:   900,
			Path:        types.Path{{8}},
			Fingerprint: "XYZABC456123",
		},
		{
			Subspace:    []byte("theta"),
			Timestamp:   1000,
			Path:        types.Path{{9}},
			Fingerprint: "XYZABC789654",
		},
	})
}

func TestQuery(t *testing.T) {
	kdtree := setupKDTree()

	tests := []struct {
		name     string
		range3d  types.Range3d
		expected []Key
	}{
		{
			name: "Whole range",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("a"),
					End:     []byte("z"),
					OpenEnd: true,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     utils.SuccessorPath(types.Path{{9}}, TestPathParams),
					OpenEnd: true,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     2000,
					OpenEnd: true,
				},
			},
			expected: kdtree.Values(),
		},
		{
			name: "Single Subspace",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("david"),
					End:     utils.SuccessorSubspaceId([]byte("david")),
					OpenEnd: false,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     utils.SuccessorPath(types.Path{{0}}, TestPathParams),
					OpenEnd: false,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     2000,
					OpenEnd: true,
				},
			},
			expected: []Key{
				{
					Subspace:    []byte("david"),
					Timestamp:   140,
					Path:        types.Path{{0}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},
			},
		},
		{
			name: "No matching nodes",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("x"),
					End:     []byte("y"),
					OpenEnd: false,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{10}},
					End:     types.Path{{20}},
					OpenEnd: false,
				},
				TimeRange: types.Range[uint64]{
					Start:   2000,
					End:     3000,
					OpenEnd: false,
				},
			},
			expected: []Key{},
		},
		{
			name: "Open end range - open subspace",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     nil,
					OpenEnd: true,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     types.Path{{9}},
					OpenEnd: false,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     2000,
					OpenEnd: false,
				},
			},
			expected: []Key{
				{
					Subspace:    []byte("alpha"),
					Timestamp:   500,
					Path:        types.Path{{3}},
					Fingerprint: "XYZABC123456",
				},
				{
					Subspace:    []byte("beta"),
					Timestamp:   200,
					Path:        types.Path{{2}},
					Fingerprint: "XYZABC654321",
				},
				{
					Subspace:    []byte("delta"),
					Timestamp:   600,
					Path:        types.Path{{4}},
					Fingerprint: "XYZABC321789",
				},
				{
					Subspace:    []byte("epsilon"),
					Timestamp:   700,
					Path:        types.Path{{5}},
					Fingerprint: "XYZABC654987",
				},
				{
					Subspace:    []byte("gamma"),
					Timestamp:   300,
					Path:        types.Path{{1}},
					Fingerprint: "XYZABC987654",
				},
				{
					Subspace:    []byte("david"),
					Timestamp:   140,
					Path:        types.Path{{0}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},
				{
					Subspace:    []byte("betty"),
					Timestamp:   100,
					Path:        types.Path{{6}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},

				{
					Subspace:    []byte("zeta"),
					Timestamp:   800,
					Path:        types.Path{{7}},
					Fingerprint: "XYZABC789123",
				},
				{
					Subspace:    []byte("eta"),
					Timestamp:   900,
					Path:        types.Path{{8}},
					Fingerprint: "XYZABC456123",
				},
			},
		},
		{
			name: "Open end range - open path",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     []byte("eta"),
					OpenEnd: false,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     nil,
					OpenEnd: true,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     2000,
					OpenEnd: false,
				},
			},
			expected: []Key{
				{
					Subspace:    []byte("alpha"),
					Timestamp:   500,
					Path:        types.Path{{3}},
					Fingerprint: "XYZABC123456",
				},
				{
					Subspace:    []byte("beta"),
					Timestamp:   200,
					Path:        types.Path{{2}},
					Fingerprint: "XYZABC654321",
				},
				{
					Subspace:    []byte("delta"),
					Timestamp:   600,
					Path:        types.Path{{4}},
					Fingerprint: "XYZABC321789",
				},
				{
					Subspace:    []byte("epsilon"),
					Timestamp:   700,
					Path:        types.Path{{5}},
					Fingerprint: "XYZABC654987",
				},
				{
					Subspace:    []byte("david"),
					Timestamp:   140,
					Path:        types.Path{{0}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},
				{
					Subspace:    []byte("betty"),
					Timestamp:   100,
					Path:        types.Path{{6}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},
			},
		},
		{
			name: "Open end range - open time",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     []byte("theta"),
					OpenEnd: false,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     types.Path{{5}},
					OpenEnd: false,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     1,
					OpenEnd: true,
				},
			},
			expected: []Key{
				{
					Subspace:    []byte("alpha"),
					Timestamp:   500,
					Path:        types.Path{{3}},
					Fingerprint: "XYZABC123456",
				},
				{
					Subspace:    []byte("beta"),
					Timestamp:   200,
					Path:        types.Path{{2}},
					Fingerprint: "XYZABC654321",
				},
				{
					Subspace:    []byte("delta"),
					Timestamp:   600,
					Path:        types.Path{{4}},
					Fingerprint: "XYZABC321789",
				},
				{
					Subspace:    []byte("gamma"),
					Timestamp:   300,
					Path:        types.Path{{1}},
					Fingerprint: "XYZABC987654",
				},
				{
					Subspace:    []byte("david"),
					Timestamp:   140,
					Path:        types.Path{{0}},
					Fingerprint: "AHSIDAIWDAWDNAINSDJNAWD",
				},
			},
		},
		{
			name: "Same subspace",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     []byte("alpha"),
					OpenEnd: false,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     types.Path{{5}},
					OpenEnd: true,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     1,
					OpenEnd: true,
				},
			},
			expected: []Key{},
		},
		{
			name: "Same Path",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     []byte("delta"),
					OpenEnd: true,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     types.Path{{0}},
					OpenEnd: false,
				},
				TimeRange: types.Range[uint64]{
					Start:   0,
					End:     1,
					OpenEnd: true,
				},
			},
			expected: []Key{},
		},
		{
			name: "Same Time",
			range3d: types.Range3d{
				SubspaceRange: types.Range[types.SubspaceId]{
					Start:   []byte("alpha"),
					End:     []byte("delta"),
					OpenEnd: true,
				},
				PathRange: types.Range[types.Path]{
					Start:   types.Path{{0}},
					End:     types.Path{{0}},
					OpenEnd: true,
				},
				TimeRange: types.Range[uint64]{
					Start:   300,
					End:     300,
					OpenEnd: false,
				},
			},
			expected: []Key{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Query(kdtree, tt.range3d)
			fmt.Println(res, tt.expected)
			if len(res) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(res))
			}

			sortKeys(res)
			sortKeys(tt.expected)

			for i, exp := range tt.expected {
				if !compareKey(res[i], exp) {
					t.Errorf("expected result %d to be %v, got %v", i, exp, res[i])
				}
			}
		})
	}
}
