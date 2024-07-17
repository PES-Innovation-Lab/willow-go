package entrydriver

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EntryDriver[PreFingerPrint, FingerPrint constraints.Ordered, T datamodeltypes.KvPart, K constraints.Unsigned] struct {
	MakeStorage             func(namespace types.NamespaceId) datamodeltypes.KDTreeStorage[types.PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	PayloadReferenceCounter datamodeltypes.PayloadReferenceCounter[types.PayloadDigest]
	GetPayloadLength        func(digest types.PayloadDigest) uint64
	Opts                    struct {
		KVDriver          kv_driver.KvDriver[T]
		NamespaceScheme   datamodeltypes.NamespaceScheme[K]
		SubspaceScheme    datamodeltypes.SubspaceScheme[K]
		PayloadScheme     datamodeltypes.PayloadScheme[K]
		PathParams        types.PathParams[K]
		FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint, K]
	}
}