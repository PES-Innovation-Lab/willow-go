package store

import (
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"golang.org/x/exp/constraints"
)

func Set[T constraints.Ordered](input datamodeltypes.EntryInput[T]) {
	var timestamp uint64
	if input.Timestamp == 0 {
		timestamp = uint64(time.Now().UnixMicro())
	}
    digest, payload, length := 
}
