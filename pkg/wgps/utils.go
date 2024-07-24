package wgps

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"golang.org/x/exp/constraints"
)

func IsSubspaceReadAuthorisation[ReadCapability, SubspaceReadCapability constraints.Ordered](authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) bool {
	if authorisation.HasSubspaceCapability {
		return true
	}
	return false
}

func AsyncReceive(receiver chan any, callback func(interface{}) error, onEnd func()) {
	go func() {
		for {
			value := <-receiver
			err := callback(value)
			if err != nil {
				fmt.Println("Error in callback:", err)
				return

			}
			if onEnd != nil {
				onEnd()
				break
			}
		}
	}()
}
