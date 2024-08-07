package wgps

import (
	"encoding/base64"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type Options[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey any, K constraints.Unsigned] struct {
	HandleStoreOurs handlestore.HandleStore[ReadCapability]
	Schemes         struct {
		Namespace     datamodeltypes.NamespaceScheme
		Subspace      datamodeltypes.SubspaceScheme
		AccessControl wgpstypes.AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey, K]
	}
}

type CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey any, K constraints.Unsigned] struct {
	NamespaceMap map[string]map[uint64]struct{}
	Opts         Options[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]
}

func isEmpty[T any](value T) bool {
	switch v := any(value).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return v == 0
	case float32, float64:
		return v == 0.0
	case string:
		return v == ""
	default:
		// This case should not happen if T is truly constraints.Ordered,
		// but it's here for completeness. Handling complex types or
		// defining "empty" for them would require more context.
		return false
	}
}

// NewCapFinder creates a new instance of CapFinder with initialized NamespaceMap.
func NewCapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey any, K constraints.Unsigned](opts Options[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]) *CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K] {
	return &CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]{
		NamespaceMap: make(map[string]map[uint64]struct{}),
		Opts:         opts,
	}
}

func (c *CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]) GetNamespaceKey(namespace types.NamespaceId) (string, error) {
	encoded := c.Opts.Schemes.Namespace.EncodingScheme.Encode(namespace)
	return base64.StdEncoding.EncodeToString(encoded), nil
}

func (c *CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]) AddCap(handle uint64) {
	cap, _ := c.Opts.HandleStoreOurs.Get(handle)
	empty := isEmpty(cap)
	if empty {
		//WillowError (TODO)
	}

	namespace := c.Opts.Schemes.AccessControl.GetGrantedNamespace(cap)
	key, _ := c.GetNamespaceKey(namespace)
	res := c.NamespaceMap[key]
	if _, ok := res[handle]; !ok {
		res[handle] = struct{}{}
	}
	// Check if the key exists in NamespaceMap
	if _, exists := c.NamespaceMap[key]; !exists {
		// If the key doesn't exist, initialize a new set and add the handle
		c.NamespaceMap[key] = make(map[uint64]struct{})
	}
	// Add the handle to the set for the key
	c.NamespaceMap[key][handle] = struct{}{}
}

func (c *CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]) FindCapHandle(entry types.Entry) uint64 {
	key, _ := c.GetNamespaceKey(entry.Namespace_id)
	set := c.NamespaceMap[key]
	if set == nil {
		return 0
	}
	entryPos := utils.EntryPosition(entry)

	for handle := range set {
		cap, _ := c.Opts.HandleStoreOurs.Get(handle)
		empty := isEmpty(cap)
		if empty {
			//WillowError (TODO)
		}

		grantedArea := c.Opts.Schemes.AccessControl.GetGrantedArea(cap)

		isInArea := utils.IsIncludedArea(c.Opts.Schemes.Subspace.Order, grantedArea, entryPos)

		if isInArea {
			return handle
		}
	}
	return 0
}
