package wgps

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"golang.org/x/exp/constraints"
)

func IsSubspaceReadAuthorisation[ReadCapability, SubspaceReadCapability constraints.Ordered](authorisation wgpstypes.ReadAuthorisation[ReadCapability, SubspaceReadCapability]) bool {
	if authorisation.HasSubspaceCapability {
		return true
	}
	return false
}
