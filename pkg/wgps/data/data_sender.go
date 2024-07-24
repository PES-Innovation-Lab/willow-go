package data

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type DataSendEntry struct {
	Entry             types.Entry
	Offset            uint64
	StaticTokenHandle uint64
	DynamicToken      string
	Payload           datamodeltypes.Payload
}

type DataBindPayloadRequest struct {
	Handle  uint64
	Offset  uint64
	Payload datamodeltypes.Payload
}

type PayloadRequest struct {
	Offset int
	Entry  types.Entry
}

type DataSenderOpts[Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, DynamicToken string, AuthorisationOpts []byte] struct {
	HandlesPayloadRequestsTheirs handlestore.HandleStore
	GetStore                     wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorisationOpts]
	TransformPayload             func(chunk []byte) []byte
}

type DataSender[Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, DynamicToken string, AuthorisationOpts []byte] struct {
	Opts          DataSenderOpts[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]
	InternalQueue []interface{} // Either DataSendEntry or DataBindPayloadRequest
}

func NewDataSender[Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, DynamicToken string, AuthorisationOpts []byte](opts DataSenderOpts[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts] {
	return DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]{
		Opts:          opts,
		InternalQueue: make([]interface{}, 0),
	}
}

/*func (q *DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts]) Delete() QueueItem {
	if len(q.Items) == 0 {
		return nil // or handle underflow
	}
	item := q.Items[0]
	q.Items = q.Items[1:]
	return item
} */

func (q DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueueEntry(entry types.Entry, staticTokenHandle uint64, dynamicToken string, offset uint64) error {
	Store := q.Opts.GetStore(entry.Namespace_id)
	Payload, err := Store.PayloadDriver.Get(entry.Payload_digest)

	if err != nil {
		//throw an error
		return fmt.Errorf("Error getting payload: %v", err)
	}

	q.InternalQueue = append(q.InternalQueue, (DataSendEntry{
		Entry:             entry,
		Offset:            offset,
		StaticTokenHandle: staticTokenHandle,
		DynamicToken:      dynamicToken,
		Payload:           Payload,
	}))
	return nil
}

/* func (q *DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueuePayloadRequest(handle uint64) {
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
*/
//NEED TO FINISH ASYNC MESSAGES
