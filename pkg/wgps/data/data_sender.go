package data

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type DataSendEntryType interface {
	types.Entry
	uint64
	uint64
	string
	datamodeltypes.Payload
}
type DataSendEntryPack[DynamicToken string] struct {
	Entry             types.Entry
	Offset            uint64
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
	Payload           datamodeltypes.Payload
}

type DataBindPayloadRequestPack struct {
	Handle  uint64
	Offset  uint64
	Payload datamodeltypes.Payload
}

type PayloadRequest struct {
	Offset uint64
	Entry  types.Entry
}

type DataSenderOpts[Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, DynamicToken string, AuthorisationOpts []byte] struct {
	HandlesPayloadRequestsTheirs handlestore.HandleStore[PayloadRequest]
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
		InternalQueue: make([]interface{}, 1),
	}
}

func (q *DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueueEntry(entry types.Entry, staticTokenHandle uint64, dynamicToken DynamicToken, offset uint64) error {
	Store := q.Opts.GetStore(entry.Namespace_id)
	Payload, err := Store.PayloadDriver.Get(entry.Payload_digest)

	if err != nil {
		//throw an error
		return fmt.Errorf("error getting payload: %v", err)
	}

	q.InternalQueue = append(q.InternalQueue, (DataSendEntryPack[DynamicToken]{
		Entry:             entry,
		Offset:            offset,
		StaticTokenHandle: staticTokenHandle,
		DynamicToken:      dynamicToken,
		Payload:           Payload,
	}))
	return nil
}

func (q *DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) QueuePayloadRequest(handle uint64) error {
	payloadRequest, found := q.Opts.HandlesPayloadRequestsTheirs.Get(handle) //This is actually supposed to be getEventually, need to see if writing it this way affects the code in any way
	if !found {
		return fmt.Errorf("handle not found")
	}
	store := q.Opts.GetStore(payloadRequest.Entry.Namespace_id)
	payload, err := store.PayloadDriver.Get(payloadRequest.Entry.Payload_digest)
	if err != nil {
		//throw an error
		return fmt.Errorf("error getting payload: %v", err)
	}
	q.InternalQueue = append(q.InternalQueue, (DataBindPayloadRequestPack{
		Handle:  handle,
		Offset:  payloadRequest.Offset,
		Payload: payload,
	}))
	return nil
}

func (q *DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]) Messages() ([]interface{}, error) {

	var messages []interface{}

	for _, msg := range q.InternalQueue {
		switch msg := msg.(type) {
		case DataSendEntryPack[DynamicToken]:
			messages = append(messages, wgpstypes.MsgDataSendEntry[DynamicToken]{
				Kind: wgpstypes.DataSendEntry,
				Data: wgpstypes.MsgDataSendEntryData[DynamicToken]{
					Entry:             msg.Entry,
					DynamicToken:      msg.DynamicToken,
					Offset:            msg.Offset,
					StaticTokenHandle: msg.StaticTokenHandle,
				},
			})
			payloadIterator, err := msg.Payload.BytesWithOffset(int(msg.Offset))
			if err != nil {
				return messages, err

			}
			for _, chunk := range payloadIterator {
				transformed := q.Opts.TransformPayload([]byte{chunk})
				messages = append(messages, wgpstypes.MsgDataSendPayload{
					Kind: wgpstypes.DataSendPayload,
					Data: wgpstypes.MsgDataSendPayloadData{
						Amount: uint64(len(transformed)),
						Bytes:  transformed,
					},
				})

			}

		case DataBindPayloadRequestPack:
			messages = append(messages, wgpstypes.MsgDataReplyPayload{
				Kind: wgpstypes.DataReplyPayload,
				Data: wgpstypes.MsgDataReplyPayloadData{
					Handle: msg.Handle,
				},
			})

			payloadIterator, err := msg.Payload.BytesWithOffset(int(msg.Offset))
			if err != nil {
				return messages, err

			}
			for _, chunk := range payloadIterator {
				transformed := q.Opts.TransformPayload([]byte{chunk})
				messages = append(messages, wgpstypes.MsgDataSendPayload{
					Kind: wgpstypes.DataSendPayload,
					Data: wgpstypes.MsgDataSendPayloadData{
						Amount: uint64(len(transformed)),
						Bytes:  transformed,
					},
				})

			}

		}

	}

	return messages, nil
}
