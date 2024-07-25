package data

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/syncutils"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// Define interfaces and structs to represent the union types.

// Define an interface that can be implemented by both []byte and CANCELLATION.
//type FIFOItem interface{}

// Modify the FIFO struct to use FIFOItem.
//type FIFO struct {
//	Items []FIFOItem // Now can hold both []byte and CANCELLATION
//}

type CurrentIngestion struct {
	Kind           string // Active or Cancelled
	FifoQueue      chan Event
	ReceivedLength uint64       // bigint in TypeScript is closest to int64 in Go
	Entry          *types.Entry // Assuming Entry is defined elsewhere

}

type Event struct {
	Entry  *types.Entry
	Data   []byte
	Cancel bool
}

type PayloadIngesterOpts[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationToken string, AuthorisationOpts []byte] struct {
	GetStore               wgpstypes.GetStoreFn[PreFingerPrint, FingerPrint, K, AuthorisationToken, AuthorisationOpts]
	ProcessReceivedPayload func(bytes []byte, entryLength uint64) []byte
}

// PayloadIngester struct modified to include the currentIngestion field.
// This class can handle both the payload sending procedures for payloads sent via reconciliation AND data channels. It would probably be better to split them up.
type PayloadIngester[Prefingerprint, Fingerprint constraints.Ordered, AuthorisationToken string, AuthorisationOpts []byte] struct {
	CurrentIngestion       CurrentIngestion
	Events                 chan Event
	ProcessReceivedPayload func(bytes []byte, entryLength uint64) []byte
	// Add a pointer to Entry, which can be nil
	EntryToRequestPayloadFor *types.Entry
	//getStore                 GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest]
	//processReceivedPayload   func(bytes []byte, entryLength uint64) []byte
}

// NewPayloadIngester creates a new PayloadIngester with the initial state.
func NewPayloadIngester[Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken string, AuthorisationOpts []byte](
	opts PayloadIngesterOpts[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorisationOpts],
) PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts] {

	var newPayloadIngester PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]
	newPayloadIngester.ProcessReceivedPayload = opts.ProcessReceivedPayload
	newPayloadIngester.Events = make(chan Event, 32)
	newPayloadIngester.CurrentIngestion = CurrentIngestion{
		Kind: "Uninitialised",
	}

	go syncutils.AsyncReceive[Event](newPayloadIngester.Events, func(value Event) error {

		if value.Cancel {
			if newPayloadIngester.CurrentIngestion.Kind == "Active" {
				newPayloadIngester.CurrentIngestion.FifoQueue <- Event{
					Cancel: true,
				}
				newPayloadIngester.CurrentIngestion.Kind = "Cancelled"
			} else if value.Entry != nil {
				if newPayloadIngester.CurrentIngestion.Kind == "Active" {
					newPayloadIngester.CurrentIngestion.FifoQueue <- Event{
						Cancel: true,
					}
					newPayloadIngester.CurrentIngestion = CurrentIngestion{
						Kind:  "Pending",
						Entry: value.Entry,
					}

				}
			} else {
				if newPayloadIngester.CurrentIngestion.Kind == "Active" {
					transformed := newPayloadIngester.ProcessReceivedPayload(value.Data, newPayloadIngester.CurrentIngestion.Entry.Payload_length)
					newPayloadIngester.CurrentIngestion.ReceivedLength += uint64(len(transformed))
					newPayloadIngester.CurrentIngestion.FifoQueue <- Event{
						Entry:  newPayloadIngester.CurrentIngestion.Entry,
						Data:   transformed,
						Cancel: false,
					}

				} else if newPayloadIngester.CurrentIngestion.Kind == "Pending" {

					store := opts.GetStore(newPayloadIngester.CurrentIngestion.Entry.Namespace_id)
					fifoQueue := make(chan Event, 32)
					transformed := newPayloadIngester.ProcessReceivedPayload(value.Data, newPayloadIngester.CurrentIngestion.Entry.Payload_length)
					fifoQueue <- Event{
						Entry:  newPayloadIngester.CurrentIngestion.Entry,
						Data:   transformed,
						Cancel: false,
					}
					store.IngestPayload(
						types.Position3d{
							Subspace: newPayloadIngester.CurrentIngestion.Entry.Subspace_id,
							Path:     newPayloadIngester.CurrentIngestion.Entry.Path,
							Time:     newPayloadIngester.CurrentIngestion.Entry.Timestamp,
						}, transformed, false, 0,
					)
					entry := newPayloadIngester.CurrentIngestion.Entry
					newPayloadIngester.CurrentIngestion = CurrentIngestion{
						Kind:           "Active",
						FifoQueue:      fifoQueue,
						ReceivedLength: uint64(len(transformed)),
						Entry:          entry,
					}
				}

			}
		}

		return nil
	}, nil)

	return newPayloadIngester

}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Target(entry types.Entry, requestifImmediatelyTerminated bool) {
	p.Events <- Event{
		Entry:  &entry,
		Cancel: false,
	}
	if requestifImmediatelyTerminated {
		p.EntryToRequestPayloadFor = &entry //
	}
}

func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Push(bytes []byte, end bool) {
	p.Events <- Event{
		Data: bytes,
	}
	if end {
		p.Events <- Event{
			Cancel: true,
		}
	}
	p.EntryToRequestPayloadFor = nil
}

// Returns the entry to request a payload for or null
func (p *PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]) Terminate() *types.Entry {
	p.Events <- Event{
		Cancel: true,
	}
	return p.EntryToRequestPayloadFor
}
