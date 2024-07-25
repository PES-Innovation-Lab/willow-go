package tests

import (
	"fmt"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/syncutils"
)

func TestAsyncReceive(t *testing.T) {
	type args struct {
		receiver chan int
		callback func(int) error
		onEnd    func()
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Integer test",
			args: args{
				receiver: make(chan int, 6),
				callback: func(i int) error {
					fmt.Printf("Async Received: %v", i)
					fmt.Println("Async is still running")
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.receiver <- 1
			go syncutils.AsyncReceive(tt.args.receiver, tt.args.callback, tt.args.onEnd)
			tt.args.receiver <- 2
			tt.args.receiver <- 3
			tt.args.receiver <- 89
			select {}
		})
	}
}
