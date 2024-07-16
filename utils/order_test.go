package utils

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/types"
)

func TestOrderSubspace(t *testing.T) {
	type args struct {
		a []byte
		b []byte
	}
	tests := []struct {
		name string
		args args
		want types.Rel
	}{
		{
			name: "Equal slices",
			args: args{
				a: []byte{1, 2, 3},
				b: []byte{1, 2, 3},
			},
			want: types.Equal,
		},
		{
			name: "a is less than b",
			args: args{
				a: []byte{1, 2, 3},
				b: []byte{4, 5, 6},
			},
			want: types.Less,
		},
		{
			name: "a is greater than b",
			args: args{
				a: []byte{4, 5, 6},
				b: []byte{1, 2, 3},
			},
			want: types.Greater,
		},
		{
			name: "",
			args: args{
				a: []byte{1, 2, 6},
				b: []byte{1, 2, 3},
			},
			want: types.Greater,
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OrderSubspace(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("OrderSubspace() = %v, want %v", got, tt.want)
			}
		})
	}
}
