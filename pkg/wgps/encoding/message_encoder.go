package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type EncodedSyncMessage struct {
	Channel wgpstypes.LogicalChannel
	message []byte
}

type MessageEncoder[
	ReadCapability,
	Receiver,
	SyncSignature,
	PsiGroup,
	PsiScalar,
	SubspaceCapability,
	SubspaceReceiver,
	SyncSubspaceSignature,
	SubspaceSecretKey,
	PreFingerPrint,
	FingerPrint constraints.Ordered,
	AuthorisationToken string,
	StaticToken any,
	DynamicToken constraints.Ordered,
	NamespaceId types.NamespaceId,
	SubsapceId types.SubspaceId,
	PayloadDigest types.PayloadDigest,
	AuthorisationOpts []byte,

] struct {
	MessageChannel      chan EncodedSyncMessage
	ReconcileMsgTracker ReconcileMsgTracker[FingerPrint, DynamicToken, NamespaceId, SubsapceId, PayloadDigest]
}
