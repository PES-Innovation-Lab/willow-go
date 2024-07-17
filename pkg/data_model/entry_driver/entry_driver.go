package entrydriver

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EntryDriver[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	MakeStorage             func(namespace types.NamespaceId) datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]
	PayloadReferenceCounter payloadDriver.PayloadReferenceCounter
	GetPayloadLength        func(digest types.PayloadDigest) uint64
	Opts                    struct {
		KVDriver          kv_driver.KvDriver
		NamespaceScheme   datamodeltypes.NamespaceScheme
		SubspaceScheme    datamodeltypes.SubspaceScheme
		PayloadScheme     datamodeltypes.PayloadScheme
		PathParams        types.PathParams[K]
		FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
	}
}