package syncutils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
)

func IsSubspaceReadAuthorisation[ReadCapability, SubspaceReadCapability any](authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) bool {
	if authorisation.HasSubspaceCapability {
		return true
	}
	return false
}

func AsyncReceive[ValueType any](receiver chan ValueType, callback func(ValueType) error, onEnd func()) {

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

}
