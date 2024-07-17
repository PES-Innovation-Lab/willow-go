package entrydriver

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// All the necesarry functions and options requires along with the EntryDriver struct!!!
type EntryDriver[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	PayloadReferenceCounter payloadDriver.PayloadReferenceCounter
	// GetPayloadLength        func(digest types.PayloadDigest) uint64 why do we need this again????
	Opts struct {
		KVDriver          kv_driver.KvDriver
		NamespaceScheme   datamodeltypes.NamespaceScheme
		SubspaceScheme    datamodeltypes.SubspaceScheme
		PayloadScheme     datamodeltypes.PayloadScheme
		PathParams        types.PathParams[K]
		FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
	}
}

/*
Instantiates a new KD tree and then returns the KD tree
Thos function will be used when we want to instantiate a new KD tree at the start of the application
*/
func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) MakeStorage(nameSpaceId types.NamespaceId) datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K] {
	storage := datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]{
		KDTree: Kdtree.NewKDTreeWithValues[Kdtree.KDNodeKey](3, []Kdtree.KDNodeKey{}),
		Opts: struct {
			Namespace         types.NamespaceId
			SubspaceScheme    datamodeltypes.SubspaceScheme
			PayloadScheme     datamodeltypes.PayloadScheme
			PathParams        types.PathParams[K]
			FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
		}{
			Namespace:         nameSpaceId,
			SubspaceScheme:    e.Opts.SubspaceScheme,
			PayloadScheme:     e.Opts.PayloadScheme,
			PathParams:        e.Opts.PathParams,
			FingerprintScheme: e.Opts.FingerprintScheme,
		},
	}
	return storage
}
