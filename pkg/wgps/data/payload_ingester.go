package data

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// Define a CANCELLATION type.
type CANCELLATION struct{}

// Define interfaces and structs to represent the union types.
type IngestionState interface {
	isIngestionState()
}

// Define an interface that can be implemented by both []byte and CANCELLATION.
type FIFOItem interface{}

// Modify the FIFO struct to use FIFOItem.
type FIFO struct {
	Items []FIFOItem // Now can hold both []byte and CANCELLATION
}

type ActiveIngestion[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	Kind           string
	Fifo           FIFO
	ReceivedLength int64                                               // bigint in TypeScript is closest to int64 in Go
	Entry          types.Entry[NamespaceId, SubspaceId, PayloadDigest] // Assuming Entry is defined elsewhere
}

func (ActiveIngestion[NamespaceId, SubspaceId, PayloadDigest]) isIngestionState() {}

type PendingIngestion[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	Kind  string
	Entry types.Entry[NamespaceId, SubspaceId, PayloadDigest]
}

func (PendingIngestion[NamespaceId, SubspaceId, PayloadDigest]) isIngestionState() {}

type UninitialisedIngestion struct {
	Kind string
}

func (UninitialisedIngestion) isIngestionState() {}

// PayloadIngester struct modified to include the currentIngestion field.
type PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts constraints.Ordered] struct {
	currentIngestion       IngestionState
	Events                 []FIFOItem
	ProcessReceivedPayload func(bytes []byte, entryLength uint64) []byte
	// Add a pointer to Entry, which can be nil
	EntryToRequestPayloadFor *types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	getStore                 GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest]
	processReceivedPayload   func(bytes []byte, entryLength uint64) []byte
}

// NewPayloadIngester creates a new PayloadIngester with the initial state.
func NewPayloadIngesterWithOptions[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest constraints.Ordered](
	getStore GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest],
	processReceivedPayload func(bytes []byte, entryLength uint64) []byte,
) *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts] {
	return &PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]{
		// Assuming the PayloadIngester struct has fields to store these functions
		GetStore:               getStore,
		ProcessReceivedPayload: processReceivedPayload,
		// Initialize other fields as necessary
		currentIngestion:         UninitialisedIngestion{Kind: "uninitialised"},
		Events:                   make([]FIFOItem, 0),
		EntryToRequestPayloadFor: nil,
	}
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]) Enqueue(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest]) {
	p.Events = append(p.Events, entry)
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]) EnqueueByteArray(bytes []byte) {
	p.Events = append(p.Events, bytes)
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]) Target(entry types.Entry[NamespaceId, SubspaceId, PayloadDigest], requestIfImmediatelyTerminated bool) {
	p.Enqueue(entry)
	if requestIfImmediatelyTerminated {
		p.EntryToRequestPayloadFor = &entry
	}
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]) Push(bytes []byte, end bool) {
	p.EnqueueByteArray(bytes)
	if end {
		//somehow push CANCELLATION into the queue
	}
	p.EntryToRequestPayloadFor = nil
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]) Terminate() *types.Entry[NamespaceId, SubspaceId, PayloadDigest] {
	//somehow push CANCELLATION into the queue
	return p.EntryToRequestPayloadFor
}
