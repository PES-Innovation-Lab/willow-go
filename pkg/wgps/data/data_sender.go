package data

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type QueueItem interface {
	// Define common methods here, if any
}

type DataSenderEntry[DynamicToken constraints.Ordered] struct {
	Entry             types.Entry
	Offset            int
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
	Payload           datamodeltypes.Payload
}

type DataBindPayloadRequestPack struct {
	Handle  uint64
	Offset  int
	Payload datamodeltypes.Payload
}

type PayloadRequest struct {
	Offset int
	Entry  types.Entry
}

type DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts constraints.Ordered] struct {
	HandlesPayloadRequestsTheirs handlestore.HandleStore[PayloadRequest]
	GetStore                     GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]
	TransformPayload             func(chunk []byte) []byte
}

type DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts constraints.Ordered] struct {
	Opts  DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]
	Items []QueueItem
}

func NewDataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts constraints.Ordered](opts DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts] {
	return &DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]{
		Opts:  opts,
		Items: make([]QueueItem, 0),
	}
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) Push(item QueueItem) {
	q.Items = append(q.Items, item)
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) Delete() QueueItem {
	if len(q.Items) == 0 {
		return nil // or handle underflow
	}
	item := q.Items[0]
	q.Items = q.Items[1:]
	return item
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueueEntry(entry types.Entry, staticTokenHandle uint64, dynamicToken DynamicToken, offset int) {
	Store := q.Opts.GetStore(PayloadRequest[PayloadDigest].Entry.NamespaceId)
	Payload := Store.GetPayload(entry)

	if Payload == nil {
		//throw an error
	}

	q.Push(DataSenderEntry[DynamicToken]{
		Entry:             entry,
		Offset:            offset,
		StaticTokenHandle: staticTokenHandle,
		DynamicToken:      dynamicToken,
		Payload:           Payload,
	})
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueuePayloadRequest(handle uint64) {
	payloadRequest, _ := q.Opts.HandlesPayloadRequestsTheirs.Get(handle) //This is actually supposed to be getEventually, need to see if writing it this way affects the code in any way
	store := q.Opts.GetStore(payloadRequest.Entry.NamespaceId)
	payload := store.GetPayload(payloadRequest.Entry)
	if payload == nil {
		//throw an error
	}
	q.Push(DataBindPayloadRequestPack{
		Handle:  handle,
		Offset:  payloadRequest.Offset,
		Payload: payload,
	})
}

//NEED TO FINISH ASYNC MESSAGES
