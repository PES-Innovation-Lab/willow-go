package datamodeltypes

import (
	"golang.org/x/exp/constraints"
)

type commitType func(isCompletePayload bool)
type rejectType func()

type PayloadDriver[PayloadDigest, T constraints.Ordered] interface {
	get(PayloadHash PayloadDigest) Payload
	set(Payload Payload) (PayloadDigest, uint64, Payload)
	receive(Payload Payload, offset T, expectedLength uint64, expectedDigest PayloadDigest) (PayloadDigest, uint64, commitType, rejectType)
	length(payloadHash PayloadDigest) uint64
	erase(digst PayloadDigest) (bool, error)
}
