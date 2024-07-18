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

type ActiveIngestion struct {
	Kind           string
	Fifo           FIFO
	ReceivedLength int64       // bigint in TypeScript is closest to int64 in Go
	Entry          types.Entry // Assuming Entry is defined elsewhere
}

func (ActiveIngestion) isIngestionState() {}

type PendingIngestion struct {
	Kind  string
	Entry types.Entry
}

func (PendingIngestion) isIngestionState() {}

type UninitialisedIngestion struct {
	Kind string
}

func (UninitialisedIngestion) isIngestionState() {}

// PayloadIngester struct modified to include the currentIngestion field.
type PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts constraints.Ordered] struct {
	currentIngestion       IngestionState
	Events                 []FIFOItem
	ProcessReceivedPayload func(bytes []byte, entryLength uint64) []byte
	// Add a pointer to Entry, which can be nil
	EntryToRequestPayloadFor *types.Entry
	getStore                 GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest]
	processReceivedPayload   func(bytes []byte, entryLength uint64) []byte
}

// NewPayloadIngester creates a new PayloadIngester with the initial state.
func NewPayloadIngesterWithOptions[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest constraints.Ordered](
	getStore GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest],
	processReceivedPayload func(bytes []byte, entryLength uint64) []byte,
) *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts] {
	return &PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]{
		// Assuming the PayloadIngester struct has fields to store these functions
		GetStore:               getStore,
		ProcessReceivedPayload: processReceivedPayload,
		// Initialize other fields as necessary
		currentIngestion:         UninitialisedIngestion{Kind: "uninitialised"},
		Events:                   make([]FIFOItem, 0),
		EntryToRequestPayloadFor: nil,
	}
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Enqueue(entry types.Entry) {
	p.Events = append(p.Events, entry)
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) EnqueueByteArray(bytes []byte) {
	p.Events = append(p.Events, bytes)
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Target(entry types.Entry, requestIfImmediatelyTerminated bool) {
	p.Enqueue(entry)
	if requestIfImmediatelyTerminated {
		p.EntryToRequestPayloadFor = &entry
	}
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Push(bytes []byte, end bool) {
	p.EnqueueByteArray(bytes)
	if end {
		//somehow push CANCELLATION into the queue
	}
	p.EntryToRequestPayloadFor = nil
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Terminate() *types.Entry {
	//somehow push CANCELLATION into the queue
	return p.EntryToRequestPayloadFor
}
