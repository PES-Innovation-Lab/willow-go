package tests

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TestEncodeRange3dRelative(t *testing.T) {
	type args struct {
		orderSubspace    types.TotalOrder[types.SubspaceId]
		encodeSubspaceId func(subspace types.SubspaceId) []byte
		pathScheme       types.PathParams[uint8]
		r                types.Range3d
		ref              types.Range3d
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
