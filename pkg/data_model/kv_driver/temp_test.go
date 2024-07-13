package kv_driver

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TeestEncodeKey(t *testing.T) {
	type args struct {
		timestamp  uint64
		subspaceId uint64
		pathParams types.PathParams[uint64]
		path       types.Path
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "simple case",
			args: args{
				timestamp:  123456789,
				subspaceId: 1,
				pathParams: types.PathParams[uint64]{MaxComponentCount: 10, MaxComponentLength: 10, MaxPathLength: 100},
				path:       types.Path{{0x01, 0x02}, {0x03, 0x04}},
			},
			want:    append(utils.BigIntToBytes(123456789), append(utils.EncodePath(types.PathParams[uint64]{MaxComponentCount: 10, MaxComponentLength: 10, MaxPathLength: 100}, types.Path{{0x01, 0x02}, {0x03, 0x04}}), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}...)...),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeKey(tt.args.timestamp, tt.args.subspaceId, tt.args.pathParams, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
