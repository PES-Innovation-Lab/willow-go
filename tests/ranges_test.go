package tests

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TestEncodeRange3dRelative(t *testing.T) {
	type args struct {
		orderSubspace    types.TotalOrder[uint64]
		encodeSubspaceId func(subspace uint64) []byte
		pathScheme       types.PathParams[uint8]
		r                types.Range3d[uint64]
		ref              types.Range3d[uint64]
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.EncodeRange3dRelative(tt.args.orderSubspace, tt.args.encodeSubspaceId, tt.args.pathScheme, tt.args.r, tt.args.ref); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeRange3dRelative() = %v, want %v", got, tt.want)
			}
		})
	}
}
