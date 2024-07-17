package datamodeltypes

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type CommitType func(isCompletePayload bool)
type RejectType func()

type PayloadDriver[T constraints.Ordered] interface {
	Get(PayloadHash string) Payload
	Set(Payload []byte) (types.PayloadDigest, Payload, uint64)
	Receive(Payload []byte, offset int64, expectedLength uint64, expectedDigest types.PayloadDigest) (types.PayloadDigest, uint64, CommitType, RejectType, error)
	Length(payloadHash types.PayloadDigest) uint64
	Erase(digst types.PayloadDigest) (bool, error)
}
