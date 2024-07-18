package tests

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
)

func TestHandStoreBind(t *testing.T) {
	handles := wgps.HandleStore[string]{
		Map: wgps.NewMap[string](),
	}
	handleA := handles.Bind("a")
	if !reflect.DeepEqual(handleA, uint64(0)) {
		t.Fatalf("handle %v and %v are not equal", handleA, uint64(0))
	}
	handleB := handles.Bind("b")
	if !reflect.DeepEqual(handleB, uint64(1)) {
		t.Fatalf("handle %v and %v are not equal", handleB, uint64(1))
	}
	handleC := handles.Bind("c")
	if !reflect.DeepEqual(handleC, uint64(2)) {
		t.Fatalf("handle %v and %v are not equal", handleC, uint64(2))
	}

}
