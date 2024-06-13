package utils

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
)

func TestOrderPath(t *testing.T) {
	type args struct {
		a types.Path
		b types.Path
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Equal paths",
			args: args{
				a: types.Path{[]byte("a"), []byte("b")},
				b: types.Path{[]byte("a"), []byte("b")},
			},
			want: 0,
		},
		{
			name: "First path less than second path",
			args: args{
				a: types.Path{[]byte("a"), []byte("aa")},
				b: types.Path{[]byte("a"), []byte("b")},
			},
			want: -1,
		},
		{
			name: "First path greater than second path",
			args: args{
				a: types.Path{[]byte("a"), []byte("c")},
				b: types.Path{[]byte("a"), []byte("b")},
			},
			want: 1,
		},
		{
			name: "First path shorter but identical prefix",
			args: args{
				a: types.Path{[]byte("a")},
				b: types.Path{[]byte("a"), []byte("b")},
			},
			want: -1,
		},
		{
			name: "First path longer but identical prefix",
			args: args{
				a: types.Path{[]byte("a"), []byte("b")},
				b: types.Path{[]byte("a")},
			},
			want: 1,
		},
		{
			name: "First path completely different",
			args: args{
				a: types.Path{[]byte("x")},
				b: types.Path{[]byte("y")},
			},
			want: -1,
		},
		{
			name: "Second path completely different",
			args: args{
				a: types.Path{[]byte("2")},
				b: types.Path{[]byte("3")},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OrderPath(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("OrderPath(%s, %s) = %v, want %v", tt.args.a, tt.args.b, got, tt.want)
			}
		})
	}
}
