package wgps

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/data"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/pai"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type WgpsMessengerOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubsapceReceiver, Prefingerprint, Fingerprint constraints.Ordered, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey types.OrderableGeneric, AuthorisationOpts []byte, AuthorisationToken string, K constraints.Unsigned] struct {
	Transport              Transport
	MaxPayloadSizePower    int
	CHallengeHashLength    int
	ChallengeHash          func(bytes []byte) []byte
	Schemes                wgpstypes.SyncSchemes[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubsapceReceiver, Prefingerprint, Fingerprint, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey, AuthorisationOpts, AuthorisationToken, K]
	Interests              map[wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest
	GetStore               GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts]
	TransformPayload       func(chunk []byte) []byte
	ProcessReceivedPayload func(chunk []byte, entryLength uint64) []byte
}

type WgpsMessenger[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubsapceReceiver, Prefingerprint, Fingerprint constraints.Ordered, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey types.OrderableGeneric, AuthorisationOpts []byte, AuthorisationToken string, K constraints.Unsigned] struct {
	Closed                   bool
	Interests                map[wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest
	Transport                ReadyTransport
	Encoder                  Encoder //TODO: has to be changed to MessageEncoder
	OutChannelReconciliation GuaranteedQueue
	OutChannelData           GuaranteedQueue
	OutChannelIntersection   GuaranteedQueue
	OutChannelCapability     GuaranteedQueue
	OutChannelAreaOfInterest GuaranteedQueue
	OutChannelPayloadRequest GuaranteedQueue
	OutChannelStaticToken    GuaranteedQueue
	InChannelReconciliation  []wgpstypes.ReconciliationChannelMsg
	InChannelData            []wgpstypes.DataChannelMsg
	InChannelIntersection    []wgpstypes.IntersectionChannelMsg
	InChannelCapability      []wgpstypes.CapabilityChannelMsg
	InChannelAreaOfInterest  []wgpstypes.AreaOfInterestChannelMsg
	InChannelPayloadRequest  []wgpstypes.PayloadRequestChannelMsg
	InChannelStaticToken     []wgpstypes.StaticTokenChannelMsg
	InChannelNone            []wgpstypes.NoChannelMsg

	// Commitment scheme
	MaxPayloadSizePower int
	ChallengeHash       func(bytes []byte) []byte
	Nonce               []byte
	OurChallenge        []byte //Supposed to be async, need to see how this will affect it
	TheirChallenge      []byte //Supposed to be async, need to see how this will affect it
	Schemes             wgpstypes.SyncSchemes[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubsapceReceiver, Prefingerprint, Fingerprint, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey, AuthorisationOpts, AuthorisationToken, K]

	// Private area intersection
	HandleIntersectionOurs   HandleStore[wgpstypes.Intersection[PsiGroup]]
	HandleIntersectionTheirs HandleStore[wgpstypes.Intersection[PsiGroup]]
	PaiFinder                pai.PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceCapability, K]

	//Setup
	HandleCapsOurs   HandleStore[ReadCapability]
	HandleCapsTheirs HandleStore[ReadCapability]

	HandlesAoisOurs   HandleStore[types.AreaOfInterest]
	HandlesAoisTheirs HandleStore[types.AreaOfInterest]

	HandlesStaticTokenOurs   HandleStore[StaticToken]
	HandlesStaticTokenTheirs HandleStore[StaticToken]

	//Reconciliation
	YourRangeCounter          int
	GetStore                  GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts]
	ReconcilerMap             ReconcilerMap //TODO: has to be changed to ReconcilerMap
	AoiIntersectionFinder     AoiIntersectionFinder
	Announcer                 Announcer
	CurrentlyReceivingEntries struct {
		Namespace   types.NamespaceId
		Range       types.Range3d
		Remaining   uint64
		IsReceiving bool
	}
	ReconciliationPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

	//Data
	CapFinder               CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]
	CurrentlySentEntry      types.Entry
	CurrentlyReceivedEntry  types.Entry
	CurrentlyReceivedOffset uint64

	HandlesPayloadRequestsOurs   HandleStore[HandleStore[types.AreaOfInterest]] //types.AreaOfInterest is just placehoder
	HandlesPayloadRequestsTheirs HandleStore[HandleStore[types.AreaOfInterest]] //types.AreaOfInterest is just placehoder

	DataSender data.DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts] //Need to change the type definition

	DataPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

}
