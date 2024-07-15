package datamodeltypes

import (
	"golang.org/x/exp/constraints"
)

type CommitType func(isCompletePayload bool)
type RejectType func()

type PayloadDriver[PayloadDigest, T constraints.Ordered] interface {
	Get(PayloadHash PayloadDigest) Payload
	Set(Payload []byte) (PayloadDigest, Payload, uint64)
	Receive(Payload Payload, offset T, expectedLength uint64, expectedDigest PayloadDigest) (PayloadDigest, uint64, CommitType, RejectType)
	Length(payloadHash PayloadDigest) uint64
	Erase(digst PayloadDigest) (bool, error)
}
