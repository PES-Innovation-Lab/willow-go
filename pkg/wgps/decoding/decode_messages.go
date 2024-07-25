package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/reconciliation"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type DecodeMessageOpts[
	ReadCapability any,
	Receiver types.SubspaceId,
	SyncSignature,
	ReceiverSecretKey,
	PsiGroup any,
	PsiScalar int,
	SubspaceCapability any,
	SubspaceReceiver types.SubspaceId,
	SyncSubspaceSignature,
	SubspaceSecretKey any,
	Prefingerprint,
	Fingerprint constraints.Ordered,
	AuthorisationToken,
	StaticToken,
	DynamicToken string,
	AuthorisationOpts []byte,
	K constraints.Unsigned,
] struct {
	Reconcile reconciliation.ReconcileMsgTrackerOpts
	Schemes   wgpstypes.SyncSchemes[
		ReadCapability,
		Receiver,
		SyncSignature,
		ReceiverSecretKey,
		PsiGroup,
		PsiScalar,
		SubspaceCapability,
		SubspaceReceiver,
		SyncSubspaceSignature,
		SubspaceSecretKey,
		Prefingerprint,
		Fingerprint,
		AuthorisationToken,
		StaticToken,
		DynamicToken,
		AuthorisationOpts,
		K,
	]
	Transport                 transport.Transport
	ChallengeLength           int
	GetIntersectionPrivy      func(handle uint64) wgpstypes.ReadCapPrivy
	GetTheirCap               func(handle uint64) ReadCapability
	GetCurrentlyReceivedEntry types.Entry
	AoiHandlesToNamespace     func(senderHandle uint64, receiverHandle uint64) types.NamespaceId
	AoiHandlesToArea          func(senderHandle uint64, receiverHandle uint64) types.Area
} //need to see what can be done about the ampersand

func DecodeMessgaes[
	ReadCapability any,
	Receiver types.SubspaceId,
	SyncSignature,
	ReceiverSecretKey,
	PsiGroup any,
	PsiScalar int,
	SubspaceCapability any,
	SubspaceReceiver types.SubspaceId,
	SyncSubspaceSignature,
	SubspaceSecretKey any,
	Prefingerprint,
	Fingerprint constraints.Ordered,
	AuthorisationToken,
	StaticToken,
	DynamicToken string,
	AuthorisationOpts []byte,
	K constraints.Unsigned,
](opts DecodeMessageOpts[
	ReadCapability,
	Receiver,
	SyncSignature,
	ReceiverSecretKey,
	PsiGroup,
	PsiScalar,
	SubspaceCapability,
	SubspaceReceiver,
	SyncSubspaceSignature,
	SubspaceSecretKey,
	Prefingerprint,
	Fingerprint,
	AuthorisationToken,
	StaticToken,
	DynamicToken,
	AuthorisationOpts,
	K,
]) wgpstypes.SyncMessage {
	reconcilerMsgTracker := reconciliation.NewReconcileMsgTracker[Fingerprint, DynamicToken](opts.Reconcile)

	bytes := *utils.GrowingBytes(opts.Transport)

	for !opts.Transport.IsClosed {
		bytes.NextAbsolute(1)

		FirstByte := bytes.Array[0]

		if FirstByte == 0x0 {
			return DecodeCommitmentReveal(bytes, opts.ChallengeLength)
		} else if (FirstByte & 0x98) == 0x98 {
			// Control aplogise
			return DecodeControlApologise(bytes)
		} else if (FirstByte & 0x90) == 0x90 {
			// Control announce dropping
			return DecodeControlAnnounceDropping(bytes)
		} else if (FirstByte & 0x8c) == 0x8c {
			// Control free
			return DecodeControlFree(bytes)
		} else if (FirstByte & 0x88) == 0x88 {
			// Control plead
			return DecodeControlPlead(bytes)
		} else if (FirstByte & 0x84) == 0x84 {
			// Control Absolve
			return DecodeControlAbsolve(bytes)
		} else if (FirstByte & 0x80) == 0x80 {
			// Control Issue Guarantee.
			return DecodeControlIssueGuarantee(bytes)
		} else if (FirstByte & 0x70) == 0x70 {
			// Data Reply Payload
			return DecodeDataReplyPayload(bytes)
		} else if (FirstByte & 0x6c) == 0x6c {
			// Data Bind Payload request
			return DecodeDataBindPayloadRequest(bytes, DecodeOpts[K]{
				DecodeNamespaceId:         opts.Schemes.NamespaceScheme.EncodingScheme.DecodeStream,
				DecodeSubspaceId:          opts.Schemes.SubspaceScheme.EncodingScheme.DecodeStream,
				PathScheme:                opts.Schemes.PathParams,
				GetCurrentlyReceivedEntry: opts.GetCurrentlyReceivedEntry,
				AoiHandlesToNamespace:     opts.AoiHandlesToNamespace,
				AoiHandlesToArea:          opts.AoiHandlesToArea,
			})
		} else if (FirstByte & 0x50) == 0x50 {
			if reconcilerMsgTracker.IsExpectingPayloadOrTermination() {
				// Reconciliation Send Entry
				if (FirstByte & 0x58) == 0x58 {
					bytes.Prune(1)

					return wgpstypes.MsgReconciliationSendPayload{
						Kind: wgpstypes.ReconciliationSendPayload,
					}
				} else {
					return DecodeReconciliationSendPayload(bytes)
				}
			} else if reconcilerMsgTracker.IsExpectingReconciliationSendEntry() {
				var tracker reconciliation.ReconcileMsgTracker[Fingerprint, DynamicToken]
				Message := DecodeReconciliationSendEntry[DynamicToken](bytes, EntryOpts[DynamicToken, K]{
					DecodeNamespaceId:   opts.Schemes.NamespaceScheme.EncodingScheme.DecodeStream,
					DecodeSubspaceId:    opts.Schemes.SubspaceScheme.EncodingScheme.DecodeStream,
					PathScheme:          opts.Schemes.PathParams,
					DecodeDynamicToken:  opts.Schemes.AuthorisationToken.Encodings.DynamicToken.DecodeStream,
					DecodePayloadDigest: opts.Schemes.Payload.EncodingScheme.DecodeStream,
					GetPrivy:            tracker.GetPrivy,
				})
				reconcilerMsgTracker.OnSendEntry(Message)
				return Message
			} else {
				Message := DecodeReconciliationAnnounceEntries(bytes, AnnounceOpts[K]{ //NEED TO CHECK ANNOUNCEOPTS ONCE, THERE MIGHT BE A FIELD MISMATCH
					DecodeSubspaceId: opts.Schemes.SubspaceScheme.EncodingScheme.DecodeStream,
					PathScheme:       opts.Schemes.PathParams,
				})
				reconcilerMsgTracker.OnAnnounceEntries(Message)
				return Message
			}
		} else if (FirstByte & 0x40) == 0x40 {
			// Reconciliation Send Fingerprint
			var tracker reconciliation.ReconcileMsgTracker[Fingerprint, DynamicToken]
			Message := DecodeReconciliationSendFingerprint(bytes, SendOpts[Fingerprint, K]{
				NeutralFingerprint:  opts.Schemes.Fingerprint.NeutralFinalised,
				DecodeFingerprint:   opts.Schemes.Fingerprint.Encoding.DecodeStream,
				DecodeSubspaceId:    opts.Schemes.SubspaceScheme.EncodingScheme.DecodeStream,
				PathScheme:          opts.Schemes.PathParams,
				GetPrivy:            tracker.GetPrivy,
				AoiHandlesToRange3d: opts.Reconcile.AoiHandlesToRange3d,
			})
			reconcilerMsgTracker.OnSendFingerprint(Message)
			return Message
		} else if (FirstByte & 0x30) == 0x30 {
			// Setup Bind Static Token
			return DecodeSetupBindStaticToken(bytes, opts.Schemes.AuthorisationToken.Encodings.StaticToken.DecodeStream)
		} else if (FirstByte & 0x28) == 0x28 {
			// Setup Bind Area of Interest
			return DecodeSetupBindAreaOfInterest(bytes, func(authHandle uint64) types.Area {
				Cap := opts.GetTheirCap(authHandle)
				return opts.Schemes.AccessControl.GetGrantedArea(Cap)
			}, opts.Schemes.SubspaceScheme.EncodingScheme.DecodeStream, opts.Schemes.PathParams)
		} else if (FirstByte & 0x20) == 0x20 {
			// Setup Bind Read Capability
			return DecodeSetupBindReadCapability(bytes, opts.Schemes.AccessControl.Encodings.ReadCap, opts.GetIntersectionPrivy, opts.Schemes.AccessControl.Encodings.SyncSignature.DecodeStream)
		} else if (FirstByte & 0x10) == 0x10 {
			// PAI Reply Subspace Capability
			return DecodePaiReplySubspaceCapability(bytes, opts.Schemes.SubspaceCap.Encodings.SubspaceCapability.DecodeStream, opts.Schemes.SubspaceCap.Encodings.SyncSubspaceSignature.DecodeStream)
		} else if (FirstByte & 0xc) == 0xc {
			// PAI Request Subspace Capability
			return DecodePaiRequestSubspaceCapability(bytes)
		} else if (FirstByte & 0x8) == 0x8 {
			// PAI Reply Fragment
			return DecodePaiReplyFragment(bytes, opts.Schemes.Pai.GroupMemberEncoding.DecodeStream)
		} else if (FirstByte & 0x4) == 0x4 {
			// PAI Bind Fragment
			return DecodePaiBindFragment(bytes, opts.Schemes.Pai.GroupMemberEncoding.DecodeStream)
		} else {
			//throw an error
		}
	}
	return nil
}
