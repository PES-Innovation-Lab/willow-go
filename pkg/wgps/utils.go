package wgps

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
)

func IsSubspaceReadAuthorisation[ReadCapability, SubspaceReadCapability any](authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) bool {
	if authorisation.HasSubspaceCapability {
		return true
	}
	return false
}
