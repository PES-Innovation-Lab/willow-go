package reconciliation

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

type ReconcileMsgTrackerOpts struct {
	DefaultNamespaceId   types.NamespaceId
	DefaultSubspaceId    types.SubspaceId
	DefaultPayloadDigest types.PayloadDigest

	HandleToNamespaceId func(aoiHandle uint64) types.NamespaceId
	AoiHandlesToRange3d func(senderAoiHandle, receiverAoiHandle uint64) types.Range3d
}

type ReconcileMsgTracker[FingerPrint string, DynamicToken string] struct {
	PrevRange                 types.Range3d
	PrevSenderHandle          uint64
	PrevReceiverHandle        uint64
	PrevEntry                 types.Entry
	PrevToken                 uint64
	AnnouncedRange            types.Range3d
	AnnouncedNamespace        types.NamespaceId
	AnnouncedEntriesRemaining uint64
	HandleToNamespaceId       func(aoiHandle uint64) types.NamespaceId
	IsAwaitingTermination     bool
}

func NewReconcileMsgTracker[FingerPrint string, DynamicToken string](
	opts ReconcileMsgTrackerOpts,
) *ReconcileMsgTracker[FingerPrint, DynamicToken] {
	return &ReconcileMsgTracker[FingerPrint, DynamicToken]{
		PrevRange:           utils.DefaultRange3d(opts.DefaultSubspaceId),
		PrevEntry:           utils.DefaultEntry(opts.DefaultNamespaceId, opts.DefaultSubspaceId, opts.DefaultPayloadDigest),
		AnnouncedRange:      utils.DefaultRange3d(opts.DefaultSubspaceId),
		AnnouncedNamespace:  opts.DefaultNamespaceId,
		HandleToNamespaceId: opts.HandleToNamespaceId,
	}
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) OnSendFingerprint(msg wgpstypes.MsgReconciliationSendFingerprint[FingerPrint]) {

	r.PrevRange = msg.Data.Range
	r.PrevSenderHandle = msg.Data.SenderHandle
	r.PrevReceiverHandle = msg.Data.ReceiverHandle
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) OnAnnounceEntries(msg wgpstypes.MsgReconciliationAnnounceEntries) {

	r.PrevRange = msg.Data.Range
	r.PrevSenderHandle = msg.Data.SenderHandle
	r.PrevReceiverHandle = msg.Data.ReceiverHandle
	r.AnnouncedRange = msg.Data.Range
	r.AnnouncedNamespace = r.HandleToNamespaceId(msg.Data.ReceiverHandle)
	r.AnnouncedEntriesRemaining = msg.Data.Count
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) OnSendEntry(msg wgpstypes.MsgReconciliationSendEntry[DynamicToken]) {

	r.PrevEntry = msg.Data.Entry.Entry
	r.PrevToken = msg.Data.StaticTokenHandle
	r.AnnouncedEntriesRemaining -= 1
	r.IsAwaitingTermination = true

}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) OnTerminatePayload() {
	r.IsAwaitingTermination = false
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) IsExpectingPayloadOrTermination() bool {
	return r.IsAwaitingTermination
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) IsExpectingReconciliationSendEntry() bool {
	if r.AnnouncedEntriesRemaining > 0 {
		return true
	}
	return false
}

func (r *ReconcileMsgTracker[FingerPrint, DynamicToken]) GetPrivy() wgpstypes.ReconciliationPrivy {

	return wgpstypes.ReconciliationPrivy{
		PrevSenderHandle:      r.PrevSenderHandle,
		PrevReceiverHandle:    r.PrevReceiverHandle,
		PrevRange:             r.PrevRange,
		PrevEntry:             r.PrevEntry,
		PrevStaticTokenHandle: r.PrevToken,
		Announced: struct {
			Range     types.Range3d
			Namespace types.NamespaceId
		}{
			Range:     r.AnnouncedRange,
			Namespace: r.AnnouncedNamespace,
		},
	}
}
