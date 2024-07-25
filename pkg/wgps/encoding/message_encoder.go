package encoding

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/channels"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/reconciliation"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EncodedSyncMessage struct {
	Channel wgpstypes.Channel
	Message []byte
}

type MessageEncoder[
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
	Fingerprint string,
	AuthorisationToken,
	StaticToken,
	DynamicToken string,
	AuthorisationOpts []byte,
	K constraints.Unsigned,

] struct {
	MessageChannel      chan EncodedSyncMessage
	ReconcileMsgTracker *reconciliation.ReconcileMsgTracker[Fingerprint, DynamicToken]
	Schemes             wgpstypes.SyncSchemes[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]
	Opts                struct {
		reconciliation.ReconcileMsgTrackerOpts
		//GetIntersectionPrivy  func(handle uint64) wgpstypes.ReadCapPrivy
		//GetCap                func(handle uint64) ReadCapability
		GetCurrentlySentEntry func() types.Entry
	}
}

func NewMessageEncoder[ReadCapability any,
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
	Fingerprint string,
	AuthorisationToken,
	StaticToken,
	DynamicToken string,
	AuthorisationOpts []byte,
	K constraints.Unsigned](schemes wgpstypes.SyncSchemes[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K], opts struct {
	reconciliation.ReconcileMsgTrackerOpts
	//GetIntersectionPrivy  func(handle uint64) wgpstypes.ReadCapPrivy
	//GetCap                func(handle uint64) ReadCapability
	GetCurrentlySentEntry func() types.Entry
}) *MessageEncoder[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K] {

	var newMessageEncoder *MessageEncoder[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]
	newMessageEncoder.Schemes = schemes
	newMessageEncoder.Opts = opts
	newMessageEncoder.MessageChannel = make(chan EncodedSyncMessage, 32)
	newMessageEncoder.ReconcileMsgTracker = reconciliation.NewReconcileMsgTracker[Fingerprint, DynamicToken](reconciliation.ReconcileMsgTrackerOpts{
		DefaultNamespaceId:   opts.DefaultNamespaceId,
		DefaultSubspaceId:    opts.DefaultSubspaceId,
		DefaultPayloadDigest: opts.DefaultPayloadDigest,
		HandleToNamespaceId:  opts.HandleToNamespaceId,
		AoiHandlesToRange3d:  opts.AoiHandlesToRange3d,
	})

	return newMessageEncoder
}

func (me *MessageEncoder[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]) Encode(message wgpstypes.SyncMessage) error {
	Push := func(channel wgpstypes.Channel, message []byte) {
		me.MessageChannel <- EncodedSyncMessage{Channel: channel, Message: message}
	}

	var bytes []byte

	switch msg := message.(type) {
	case wgpstypes.MsgControlIssueGuarantee:
		bytes = EncodeControlIssueGuarantee(msg)
		break
	case wgpstypes.MsgControlAbsolve:
		bytes = EncodeControlAbsolve(msg)
		break
	case wgpstypes.MsgControlPlead:
		bytes = EncodeControlPlead(msg)
		break
	case wgpstypes.MsgControlAnnounceDropping:
		bytes = EncodeControlAnnounceDropping(msg)
		break
	case wgpstypes.MsgControlApologise:
		bytes = EncodeControlApologise(msg)
		break
	case wgpstypes.MsgControlFree:
		bytes = EncodeControlFree(msg)
		break

	// Commitment scheme and PAI
	case wgpstypes.MsgCommitmentReveal:
		bytes = EncodeCommitmentReveal(msg)
		break
	case wgpstypes.MsgPaiBindFragment[PsiGroup]:
		bytes = EncodePaiBindFragment[PsiGroup](msg, me.Schemes.Pai.GroupMemberEncoding.Encode)
		break
	case wgpstypes.MsgPaiReplyFragment[PsiGroup]:
		bytes = EncodePaiReplyFragment[PsiGroup](msg, me.Schemes.Pai.GroupMemberEncoding.Encode)
		break
	case wgpstypes.MsgPaiRequestSubspaceCapability:
		bytes = EncodePaiRequestSubspaceCapability(msg)
		break
	case wgpstypes.MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]:
		bytes = EncodePaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature](msg, me.Schemes.SubspaceCap.Encodings.SubspaceCapability.Encode, me.Schemes.SubspaceCap.Encodings.SyncSubspaceSignature.Encode)
		break

	// Setup
	case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
		//Privy := me.Opts.GetIntersectionPrivy(msg.Data.Handle)
		//bytes = EncodeSetupBindReadCapability[ReadCapability, SyncSignature](msg, me.Schemes.AccessControl.Encodings.ReadCap, me.Schemes.AccessControl.Encodings.SyncSignature.Encode, Privy)
		break
	case wgpstypes.MsgSetupBindAreaOfInterest:
		//Cap := me.Opts.GetCap(msg.Data.Authorisation)
		//Outer := me.Schemes.AccessControl.GetGrantedArea(Cap)
		/*bytes = EncodeSetupBindAreaOfInterest[K](msg, struct {
			Outer          types.Area
			PathScheme     types.PathParams[K]
			EncodeSubspace func(subspace types.SubspaceId) []byte
			OrderSubspace  types.TotalOrder[types.SubspaceId]
		}{
			Outer:          Outer,
			PathScheme:     me.Schemes.PathParams,
			EncodeSubspace: me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			OrderSubspace:  me.Schemes.SubspaceScheme.Order,
		}) */
		break
	case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
		bytes = EncodeSetupBindStaticToken[StaticToken](msg, me.Schemes.AuthorisationToken.Encodings.StaticToken.Encode)
		break
	case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
		bytes = EncodeReconciliationSendFingerprint[Fingerprint](msg, struct {
			OrderSubspace        types.TotalOrder[types.SubspaceId]
			EncodeSubspaceId     func(subspace types.SubspaceId) []byte
			PathScheme           types.PathParams[K]
			IsFingerprintNeutral func(fingerprint Fingerprint) bool
			EncodeFingerprint    func(fingerprint Fingerprint) []byte
			Privy                wgpstypes.ReconciliationPrivy
		}{
			IsFingerprintNeutral: func(fingerprint Fingerprint) bool {
				return me.Schemes.Fingerprint.IsEqual(
					fingerprint,
					me.Schemes.Fingerprint.NeutralFinalised,
				)
			},
			EncodeSubspaceId:  me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			OrderSubspace:     me.Schemes.SubspaceScheme.Order,
			PathScheme:        me.Schemes.PathParams,
			Privy:             me.ReconcileMsgTracker.GetPrivy(),
			EncodeFingerprint: me.Schemes.Fingerprint.Encoding.Encode,
		})
		me.ReconcileMsgTracker.OnSendFingerprint(msg)
		break
	case wgpstypes.MsgReconciliationAnnounceEntries:
		bytes = EncodeReconciliationAnnounceEntries(msg, struct {
			Privy            wgpstypes.ReconciliationPrivy
			OrderSubspace    types.TotalOrder[types.SubspaceId]
			EncodeSubspaceId func(subspace types.SubspaceId) []byte
			PathScheme       types.PathParams[K]
		}{
			Privy:            me.ReconcileMsgTracker.GetPrivy(),
			OrderSubspace:    me.Schemes.SubspaceScheme.Order,
			EncodeSubspaceId: me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			PathScheme:       me.Schemes.PathParams,
		})
		me.ReconcileMsgTracker.OnAnnounceEntries(msg)
		break
	case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
		bytes = EncodeReconciliationSendEntry[DynamicToken](msg, struct {
			Privy               wgpstypes.ReconciliationPrivy
			IsEqualNamespace    func(a, b types.NamespaceId) bool
			OrderSubspace       types.TotalOrder[types.SubspaceId]
			EncodeNamespaceId   func(namespace types.NamespaceId) []byte
			EncodeSubspaceId    func(subspace types.SubspaceId) []byte
			EncodePayloadDigest func(digest types.PayloadDigest) []byte
			EncodeDynamicToken  func(token DynamicToken) []byte
			PathScheme          types.PathParams[K]
		}{
			Privy:               me.ReconcileMsgTracker.GetPrivy(),
			IsEqualNamespace:    me.Schemes.NamespaceScheme.IsEqual,
			OrderSubspace:       me.Schemes.SubspaceScheme.Order,
			EncodeNamespaceId:   me.Schemes.NamespaceScheme.EncodingScheme.Encode,
			EncodeSubspaceId:    me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			EncodePayloadDigest: me.Schemes.Payload.EncodingScheme.Encode,
			EncodeDynamicToken:  me.Schemes.AuthorisationToken.Encodings.DynamicToken.Encode,
			PathScheme:          me.Schemes.PathParams,
		})
		me.ReconcileMsgTracker.OnSendEntry(msg)
		break
	case wgpstypes.MsgReconciliationSendPayload:
		bytes = EncodeReconciliationSendPayload[K](msg)
		break
	case wgpstypes.MsgReconciliationTerminatePayload:
		bytes = EncodeReconciliationTerminatePayload()
		break
	case wgpstypes.MsgDataSendEntry[DynamicToken]:
		bytes = EncodeDataSendEntry[DynamicToken](msg, struct {
			EncodeDynamicToken  func(token DynamicToken) []byte
			CurrentlySentEntry  types.Entry
			IsEqualNamespace    func(a, b types.NamespaceId) bool
			OrderSubspace       types.TotalOrder[types.SubspaceId]
			EncodeNamespace     func(namespace types.NamespaceId) []byte
			EncodeSubspace      func(subspace types.SubspaceId) []byte
			EncodePayloadDigest func(payloadDigest types.PayloadDigest) []byte
			PathParams          types.PathParams[K]
		}{
			EncodeDynamicToken:  me.Schemes.AuthorisationToken.Encodings.DynamicToken.Encode,
			CurrentlySentEntry:  me.Opts.GetCurrentlySentEntry(),
			IsEqualNamespace:    me.Schemes.NamespaceScheme.IsEqual,
			OrderSubspace:       me.Schemes.SubspaceScheme.Order,
			EncodeNamespace:     me.Schemes.NamespaceScheme.EncodingScheme.Encode,
			EncodeSubspace:      me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			EncodePayloadDigest: me.Schemes.Payload.EncodingScheme.Encode,
			PathParams:          me.Schemes.PathParams,
		})
		break
	case wgpstypes.MsgDataSendPayload:
		bytes = EncodeDataSendPayload(msg)
		break
	case wgpstypes.MsgDataSetMetadata:
		bytes = EncodeDataSetEagerness(msg)
		break
	case wgpstypes.MsgDataBindPayloadRequest:
		bytes = EncodeDataBindPayloadRequest[K](msg, struct {
			CurrentlySentEntry  types.Entry
			IsEqualNamespace    func(a, b types.NamespaceId) bool
			OrderSubspace       types.TotalOrder[types.SubspaceId]
			EncodeNamespace     func(namespace types.NamespaceId) []byte
			EncodeSubspace      func(subspace types.SubspaceId) []byte
			EncodePayloadDigest func(payloadDigest types.PayloadDigest) []byte
			PathParams          types.PathParams[K]
		}{
			CurrentlySentEntry:  me.Opts.GetCurrentlySentEntry(),
			IsEqualNamespace:    me.Schemes.NamespaceScheme.IsEqual,
			OrderSubspace:       me.Schemes.SubspaceScheme.Order,
			EncodeNamespace:     me.Schemes.NamespaceScheme.EncodingScheme.Encode,
			EncodeSubspace:      me.Schemes.SubspaceScheme.EncodingScheme.Encode,
			EncodePayloadDigest: me.Schemes.Payload.EncodingScheme.Encode,
			PathParams:          me.Schemes.PathParams,
		})
		break
	case wgpstypes.MsgDataReplyPayload:
		bytes = EncodeDataReplyPayload(msg)
		break
	default:
		return fmt.Errorf("did not know how to encode message")
	}
	Push(channels.MsgLogicalChannels[message.GetKind()], bytes)
	return nil
}
