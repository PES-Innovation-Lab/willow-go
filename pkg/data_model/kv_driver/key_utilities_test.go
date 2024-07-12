package kv_driver

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TestEncodeKey(t *testing.T) {
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

func Test_encodeSubspaceId(t *testing.T) {
	type args struct {
		subspace uint64
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "was a uint8 case",
			args:    args{subspace: uint64(1)},
			want:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			wantErr: false,
		},
		{
			name:    "big.Int case",
			args:    args{subspace: 123456789},
			want:    utils.BigIntToBytes(123456789),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeSubspaceId(tt.args.subspace)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeSubspaceId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeSubspaceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_DecodeKey(t *testing.T) {
	type args struct {
		encodedKey []byte
		pathParams types.PathParams[uint64]
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		want1   uint64
		want2   types.Path
		wantErr bool
	}{
		{
			name: "simple case",
			args: args{
				encodedKey: append(utils.BigIntToBytes(123456789), append(utils.EncodePath(types.PathParams[uint64]{MaxComponentCount: 10, MaxComponentLength: 10, MaxPathLength: 100}, types.Path{{0x01, 0x02}, {0x03, 0x04}}), []byte{0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01, 0x00}...)...),
				pathParams: types.PathParams[uint64]{MaxComponentCount: 10, MaxComponentLength: 10, MaxPathLength: 100},
			},
			want:    123456789,
			want1:   uint64(1),
			want2:   types.Path{{0x01, 0x02}, {0x03, 0x04}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := DecodeKey(tt.args.encodedKey, tt.args.pathParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecodeKey() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("DecodeKey() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("DecodeKey() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_decodeSubspaceId(t *testing.T) {
	type args struct {
		subspaceBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name:    "int8 case",
			args:    args{subspaceBytes: []byte{0x01}},
			want:    1,
			wantErr: false,
		},
		{
			name:    "big.Int case",
			args:    args{subspaceBytes: big.NewInt(123456789).Bytes()},
			want:    123456789,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeSubspaceId[uint64](tt.args.subspaceBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeSubspaceId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeSubspaceId() = %v, want %v", got, tt.want)
			}
		})
	}
}
