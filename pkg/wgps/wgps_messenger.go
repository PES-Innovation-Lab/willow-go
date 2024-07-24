package wgps

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/data"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/transport"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type WgpsMessengerOpts[
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
	//Transport *wgpstypes.Transport
	/** Sets the [`maximum payload size`](https://willowprotocol.org/specs/sync/index.html#peer_max_payload_size) for this peer, which is 2 to the power of the given number.
	 *
	 * The given power must be a natural number lesser than or equal to 64. */
	//MaxPayloadSizePower int

	/** Sets the [`challenge_length`](https://willowprotocol.org/specs/sync/index.html#challenge_length) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	//ChallengeLength int
	/** Sets the [`challenge_hash_length`](https://willowprotocol.org/specs/sync/index.html#challenge_hash_length) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	//ChallengeHashLength int
	/** Sets the [`challeng_hash`](https://willowprotocol.org/specs/sync/index.html#challenge_hash) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	//ChallengeHash func(bytes []byte) []byte
	/** The parameter schemes used to configure the `WgpsMessenger` for sync. */

	Schemes wgpstypes.SyncSchemes[
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
	//Interests map[*wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest

	GetStore               wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorisationOpts]
	TransformPayload       func(chunk []byte) []byte
	ProcessReceivedPayload func(chunk []byte, entryLength uint64) []byte
}

/** Coordinates an open-ended synchronisation session between two peers using the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).
 */
type WgpsMessenger[
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
	Closed bool
	//Interests map[*wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest
	Transport *transport.QuicTransport
	//Encoder                  encoding.MessageEncoder //TODO: has to be changed to MessageEncoder
	OutChannelReconciliation GuaranteedQueue
	OutChannelData           GuaranteedQueue
	OutChannelIntersection   GuaranteedQueue
	OutChannelCapability     GuaranteedQueue
	OutChannelAreaOfInterest GuaranteedQueue
	OutChannelPayloadRequest GuaranteedQueue
	OutChannelStaticToken    GuaranteedQueue
	InChannelReconciliation  []wgpstypes.ReconciliationChannelMsg
	InChannelData            []wgpstypes.DataChannelMsg
	//InChannelIntersection    []wgpstypes.IntersectionChannelMsg
	//InChannelCapability      []wgpstypes.CapabilityChannelMsg
	//InChannelAreaOfInterest  []wgpstypes.AreaOfInterestChannelMsg
	InChannelPayloadRequest []wgpstypes.PayloadRequestChannelMsg
	InChannelStaticToken    []wgpstypes.StaticTokenChannelMsg
	InChannelNone           []wgpstypes.NoChannelMsg

	// Commitment scheme
	//MaxPayloadSizePower int
	//ChallengeHash            func(bytes []byte) []byte
	//Nonce                    []byte
	//OurChallenge             []byte //Supposed to be async, need to see how this will affect it
	//TheirChallenge           []byte //Supposed to be async, need to see how this will affect it
	Schemes wgpstypes.SyncSchemes[
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
	// Private area intersection
	//HandleIntersectionOurs   handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	//HandleIntersectionTheirs handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	//PaiFinder                pai.PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceCapability, K]

	//Setup
	//HandleCapsOurs   handlestore.HandleStore[ReadCapability]
	//HandleCapsTheirs handlestore.HandleStore[ReadCapability]

	HandlesAoisOurs   handlestore.HandleStore[types.AreaOfInterest]
	HandlesAoisTheirs handlestore.HandleStore[types.AreaOfInterest]

	HandlesStaticTokenOurs   handlestore.HandleStore[StaticToken]
	HandlesStaticTokenTheirs handlestore.HandleStore[StaticToken]

	//Reconciliation
	YourRangeCounter int
	GetStore         wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorisationOpts]
	//ReconcilerMap             reconciliation.ReconcilerMap //TODO: has to be changed to ReconcilerMap
	//AoiIntersectionFinder     reconciliation.AoiIntersectionFinder
	//Announcer                 reconciliation.Announcer
	CurrentlyReceivingEntries struct {
		Namespace types.NamespaceId
		Range     types.Range3d
		Remaining uint64
		//IsReceiving bool
	}
	//ReconciliationPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

	//Data
	//CapFinder               CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]
	CurrentlySentEntry      types.Entry
	CurrentlyReceivedEntry  types.Entry
	CurrentlyReceivedOffset uint64

	HandlesPayloadRequestsOurs   handlestore.HandleStore[data.PayloadRequest]
	HandlesPayloadRequestsTheirs handlestore.HandleStore[data.PayloadRequest]

	DataSender data.DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]

	//DataPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

}

func NewWgpsMessenger[
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
](
	opts WgpsMessengerOpts[
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
	],
) (WgpsMessenger[
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
], error) {

	var newWgpsMessenger WgpsMessenger[
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
	var err error

	newWgpsMessenger.GetStore = opts.GetStore
	newWgpsMessenger.Schemes = opts.Schemes
	newWgpsMessenger.Transport, err = transport.NewQuicTransport("localhost:4242")
	if err != nil {
		return newWgpsMessenger, err

	}
	newWgpsMessenger.OutChannelData = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.OutChannelReconciliation = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.OutChannelPayloadRequest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InChannelData = make([]wgpstypes.DataChannelMsg, 1)
	newWgpsMessenger.InChannelReconciliation = make([]wgpstypes.ReconciliationChannelMsg, 1)
	newWgpsMessenger.InChannelPayloadRequest = make([]wgpstypes.PayloadRequestChannelMsg, 1)
	newWgpsMessenger.InChannelNone = make([]wgpstypes.NoChannelMsg, 1)

	newWgpsMessenger.CurrentlySentEntry = utils.DefaultEntry(
		newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
		newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
		newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
	)
	newWgpsMessenger.CurrentlyReceivedEntry = utils.DefaultEntry(
		newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
		newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
		newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
	)

	newWgpsMessenger.HandlesPayloadRequestsOurs = handlestore.HandleStore[data.PayloadRequest]{
		Map: handlestore.NewMap[data.PayloadRequest](),
	}
	newWgpsMessenger.HandlesPayloadRequestsTheirs = handlestore.HandleStore[data.PayloadRequest]{
		Map: handlestore.NewMap[data.PayloadRequest](),
	}
	newWgpsMessenger.DataSender = data.NewDataSender[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		DynamicToken,
		AuthorisationOpts,
	](data.DataSenderOpts[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		DynamicToken,
		AuthorisationOpts,
	]{
		HandlesPayloadRequestsTheirs: newWgpsMessenger.HandlesPayloadRequestsTheirs,
	})
	return newWgpsMessenger, nil
}
