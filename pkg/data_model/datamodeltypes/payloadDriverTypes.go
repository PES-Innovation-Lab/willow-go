package datamodeltypes

import (
	"golang.org/x/exp/constraints"
)

type CommitType func(isCompletePayload bool)
type RejectType func()

type PayloadDriver[PayloadDigest, T constraints.Ordered] interface {
	Get(PayloadHash string) Payload
	Set(Payload []byte) (PayloadDigest, Payload, uint64)
	Receive(Payload []byte, offset int64, expectedLength uint64, expectedDigest PayloadDigest) (PayloadDigest, uint64, CommitType, RejectType, error)
	Length(payloadHash PayloadDigest) uint64
	Erase(digst PayloadDigest) (bool, error)
}
