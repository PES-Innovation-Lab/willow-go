package data

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type QueueItem interface {
	// Define common methods here, if any
}

type DataSenderEntry[DynamicToken, PayloadDigest constraints.Ordered] struct {
	Entry             types.Entry[PayloadDigest]
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

type PayloadRequest[PayloadDigest constraints.Ordered] struct {
	Offset int
	Entry  types.Entry[PayloadDigest]
}

type DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts constraints.Ordered] struct {
	HandlesPayloadRequestsTheirs wgps.HandleStore[PayloadRequest[PayloadDigest]]
	GetStore                     GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts]
	TransformPayload             func(chunk []byte) []byte
}

type DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts constraints.Ordered] struct {
	Opts  DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]
	Items []QueueItem
}

func NewDataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts constraints.Ordered](opts DataSenderOpts[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]) *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts] {
	return &DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]{
		Opts:  opts,
		Items: make([]QueueItem, 0),
	}
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]) Push(item QueueItem) {
	q.Items = append(q.Items, item)
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]) Delete() QueueItem {
	if len(q.Items) == 0 {
		return nil // or handle underflow
	}
	item := q.Items[0]
	q.Items = q.Items[1:]
	return item
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]) QueueEntry(entry types.Entry[PayloadDigest], staticTokenHandle uint64, dynamicToken DynamicToken, offset int) {
	Store := q.Opts.GetStore(PayloadRequest[PayloadDigest].Entry.NamespaceId)
	Payload := Store.GetPayload(entry)

	if Payload == nil {
		//throw an error
	}

	q.Push(DataSenderEntry[DynamicToken, PayloadDigest]{
		Entry:             entry,
		Offset:            offset,
		StaticTokenHandle: staticTokenHandle,
		DynamicToken:      dynamicToken,
		Payload:           Payload,
	})
}

func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, PayloadDigest, AuthorisationOpts]) QueuePayloadRequest(handle uint64) {
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
