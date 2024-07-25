package wgps

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/data"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/encoding"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/reconciliation"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/syncutils"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/transport"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type NewMessengerReturn[

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
	NewMessenger *WgpsMessenger[
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
	Error error
}
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

	InitiatorEncoder *encoding.MessageEncoder[
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
	AcceptedEncoder *encoding.MessageEncoder[
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
	InitiatorOutChannelReconciliation GuaranteedQueue
	InitiatorOutChannelData           GuaranteedQueue
	InitiatorOutChannelIntersection   GuaranteedQueue
	InitiatorOutChannelCapability     GuaranteedQueue
	InitiatorOutChannelAreaOfInterest GuaranteedQueue
	InitiatorOutChannelPayloadRequest GuaranteedQueue
	InitiatorOutChannelStaticToken    GuaranteedQueue
	AcceptedOutChannelReconciliation  GuaranteedQueue
	AcceptedOutChannelData            GuaranteedQueue
	AcceptedOutChannelIntersection    GuaranteedQueue
	AcceptedOutChannelCapability      GuaranteedQueue
	AcceptedOutChannelAreaOfInterest  GuaranteedQueue
	AcceptedOutChannelPayloadRequest  GuaranteedQueue
	AcceptedOutChannelStaticToken     GuaranteedQueue
	InitiatorInChannelReconciliation  chan wgpstypes.ReconciliationChannelMsg
	InitiatorInChannelData            chan wgpstypes.DataChannelMsg
	//InChannelIntersection    []wgpstypes.IntersectionChannelMsg
	//InChannelCapability      []wgpstypes.CapabilityChannelMsg
	//InChannelAreaOfInterest  []wgpstypes.AreaOfInterestChannelMsg
	InitiatorInChannelPayloadRequest chan wgpstypes.PayloadRequestChannelMsg
	InitiatorInChannelStaticToken    chan wgpstypes.StaticTokenChannelMsg
	InitiatorInChannelNone           chan wgpstypes.NoChannelMsg
	AcceptedInChannelReconciliation  chan wgpstypes.ReconciliationChannelMsg
	AcceptedInChannelData            chan wgpstypes.DataChannelMsg
	//InChannelIntersection    []wgpstypes.IntersectionChannelMsg
	//InChannelCapability      []wgpstypes.CapabilityChannelMsg
	//InChannelAreaOfInterest  []wgpstypes.AreaOfInterestChannelMsg
	AcceptedInChannelPayloadRequest chan wgpstypes.PayloadRequestChannelMsg
	AcceptedInChannelStaticToken    chan wgpstypes.StaticTokenChannelMsg
	AcceptedInChannelNone           chan wgpstypes.NoChannelMsg

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
	ReconciliationPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]

	//Data
	//CapFinder               CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]
	CurrentlySentEntry      types.Entry
	CurrentlyReceivedEntry  types.Entry
	CurrentlyReceivedOffset uint64

	HandlesPayloadRequestsOurs   handlestore.HandleStore[data.PayloadRequest]
	HandlesPayloadRequestsTheirs handlestore.HandleStore[data.PayloadRequest]

	DataSender data.DataSender[Prefingerprint, Fingerprint, K, AuthorisationToken, DynamicToken, AuthorisationOpts]

	DataPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]
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
	newMessengerChan chan NewMessengerReturn[
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
) {

	var newWgpsMessenger *WgpsMessenger[
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

	newWgpsMessenger.InitiatorOutChannelData = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.InitiatorOutChannelReconciliation = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.InitiatorOutChannelPayloadRequest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelData = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelReconciliation = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelPayloadRequest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorInChannelData = make(chan wgpstypes.DataChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelReconciliation = make(chan wgpstypes.ReconciliationChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelPayloadRequest = make(chan wgpstypes.PayloadRequestChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelNone = make(chan wgpstypes.NoChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelData = make(chan wgpstypes.DataChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelReconciliation = make(chan wgpstypes.ReconciliationChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelPayloadRequest = make(chan wgpstypes.PayloadRequestChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelNone = make(chan wgpstypes.NoChannelMsg, 32)

	newWgpsMessenger.ReconciliationPayloadIngester = data.NewPayloadIngester[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		AuthorisationOpts,
	](data.PayloadIngesterOpts[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		AuthorisationOpts,
	]{
		GetStore:               opts.GetStore,
		ProcessReceivedPayload: opts.ProcessReceivedPayload,
	})

	newWgpsMessenger.AcceptedEncoder = encoding.NewMessageEncoder[
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
	](newWgpsMessenger.Schemes,
		struct {
			reconciliation.ReconcileMsgTrackerOpts
			//GetIntersectionPrivy  func(handle uint64) wgpstypes.ReadCapPrivy
			//GetCap                func(handle uint64) ReadCapability
			GetCurrentlySentEntry func() types.Entry
		}{
			reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
			},
			func() types.Entry {

				return newWgpsMessenger.CurrentlySentEntry

			}},
	)

	newWgpsMessenger.InitiatorEncoder = encoding.NewMessageEncoder[
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
	](newWgpsMessenger.Schemes,
		struct {
			reconciliation.ReconcileMsgTrackerOpts
			//GetIntersectionPrivy  func(handle uint64) wgpstypes.ReadCapPrivy
			//GetCap                func(handle uint64) ReadCapability
			GetCurrentlySentEntry func() types.Entry
		}{
			reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
			},
			func() types.Entry {

				return newWgpsMessenger.CurrentlySentEntry

			}},
	)

	newWgpsMessenger.HandlesStaticTokenOurs = handlestore.HandleStore[StaticToken]{
		Map: handlestore.NewMap[StaticToken](),
	}

	newWgpsMessenger.HandlesStaticTokenTheirs = handlestore.HandleStore[StaticToken]{
		Map: handlestore.NewMap[StaticToken](),
	}
	newWgpsMessenger.HandlesAoisOurs = handlestore.HandleStore[types.AreaOfInterest]{
		Map: handlestore.NewMap[types.AreaOfInterest](),
	}

	newWgpsMessenger.HandlesAoisTheirs = handlestore.HandleStore[types.AreaOfInterest]{
		Map: handlestore.NewMap[types.AreaOfInterest](),
	}

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

	newWgpsMessenger.DataPayloadIngester = data.NewPayloadIngester[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		AuthorisationOpts,
	](data.PayloadIngesterOpts[
		Prefingerprint,
		Fingerprint,
		K,
		AuthorisationToken,
		AuthorisationOpts,
	]{
		GetStore:               opts.GetStore,
		ProcessReceivedPayload: opts.ProcessReceivedPayload,
	})

	go syncutils.AsyncReceive[encoding.EncodedSyncMessage](newWgpsMessenger.InitiatorEncoder.MessageChannel, func(msg encoding.EncodedSyncMessage) error {
		switch msg.Channel {
		case wgpstypes.ReconciliationChannel:
			newWgpsMessenger.InitiatorOutChannelReconciliation.Push(msg.Message)

		}
		return nil
	}, nil)

	go syncutils.AsyncReceive[encoding.EncodedSyncMessage](newWgpsMessenger.AcceptedEncoder.MessageChannel, func(msg encoding.EncodedSyncMessage) error {
		switch msg.Channel {
		case wgpstypes.ReconciliationChannel:
			newWgpsMessenger.AcceptedOutChannelReconciliation.Push(msg.Message)

		}
		return nil
	}, nil)

	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelData.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelReconciliation.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.ReconciliationChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelPayloadRequest.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.PayloadRequestChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelStaticToken.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.StaticTokenChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelCapability.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.CapabilityChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelIntersection.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.IntersectionChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.AcceptedOutChannelAreaOfInterest.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.AreaOfInterestChannel, wgpstypes.SyncRoleBetty)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelData.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelReconciliation.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.ReconciliationChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelPayloadRequest.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.PayloadRequestChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelStaticToken.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.StaticTokenChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelCapability.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.CapabilityChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelIntersection.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.IntersectionChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)
	go syncutils.AsyncReceive[[]byte](newWgpsMessenger.InitiatorOutChannelAreaOfInterest.Queue, func(value []byte) error {
		err := newWgpsMessenger.Transport.Send(value, wgpstypes.AreaOfInterestChannel, wgpstypes.SyncRoleAlfie)
		return err
	}, nil)

	newWgpsMessenger.Transport, err = transport.NewQuicTransport("localhost:4242")
	if err != nil {

		newMessengerChan <- NewMessengerReturn[
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
		]{
			NewMessenger: newWgpsMessenger,
			Error:        err,
		}
		return

	}

	newMessengerChan <- NewMessengerReturn[
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
	]{
		NewMessenger: newWgpsMessenger,
		Error:        nil,
	}
	select {}
}

func (w *WgpsMessenger[
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
]) Initiate(addr string, areasofinterests []types.AreaOfInterest) error {

	err := w.Transport.Initiate(addr)
	return err

}

func (w *WgpsMessenger[
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
]) HandleMsgData(msg wgpstypes.DataChannelMsg) error {
	switch msg := msg.(type) {
	case wgpstypes.MsgDataSendEntry[DynamicToken]:
		staticToken, found := w.HandlesStaticTokenTheirs.Get(msg.Data.StaticTokenHandle)
		if !found {
			return fmt.Errorf("static token not found")
		}
		authToken := w.Schemes.AuthorisationToken.RecomposeAuthToken(staticToken, msg.Data.DynamicToken)
		store := w.GetStore(msg.Data.Entry.Namespace_id)
		_, err := store.IngestEntry(msg.Data.Entry, authToken)
		if err != nil {
			return fmt.Errorf("could not ingest entry")
		}
		w.DataPayloadIngester.Target(msg.Data.Entry, false)
	}
	return nil
}

func (w *WgpsMessenger[
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
]) HandleMsgPayloadRequest(
	msg wgpstypes.PayloadRequestChannelMsg) {

	switch msg := msg.(type) {
	case wgpstypes.MsgDataBindPayloadRequest:
		handle := w.HandlesPayloadRequestsTheirs.Bind(data.PayloadRequest{
			Offset: msg.Data.Offset,
			Entry:  msg.Data.Entry,
		})
		w.DataSender.QueuePayloadRequest(handle)
	}

}

func (w *WgpsMessenger[
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
]) HandleMsgStaticToken(
	msg wgpstypes.StaticTokenChannelMsg) {
	switch msg := msg.(type) {
	case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
		w.HandlesStaticTokenTheirs.Bind(msg.Data.StaticToken)
	}
}

func (w *WgpsMessenger[
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
]) Close() error {
	w.Closed = true
	err := w.Transport.Close()
	return err
}
