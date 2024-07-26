package wgps

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
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
	Fingerprint string,
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
	Fingerprint string,
	AuthorisationToken,
	StaticToken,
	DynamicToken string,
	AuthorisationOpts []byte,
	K constraints.Unsigned,
] struct {
	//Transport *wgpstypes.Transport
	/** Sets the [`maximum payload size`](https://willowprotocol.org/specs/sync/index.html#peer_max_payload_size) for newWgpsMessenger peer, which is 2 to the power of the given number.
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
	Fingerprint string,
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
	// Initiator side
	InitiatorOutChannelReconciliation GuaranteedQueue
	InitiatorOutChannelData           GuaranteedQueue
	InitiatorOutChannelIntersection   GuaranteedQueue
	InitiatorOutChannelCapability     GuaranteedQueue
	InitiatorOutChannelAreaOfInterest GuaranteedQueue
	InitiatorOutChannelPayloadRequest GuaranteedQueue
	InitiatorOutChannelStaticToken    GuaranteedQueue

	// Accepted side
	AcceptedOutChannelReconciliation GuaranteedQueue
	AcceptedOutChannelData           GuaranteedQueue
	AcceptedOutChannelIntersection   GuaranteedQueue
	AcceptedOutChannelCapability     GuaranteedQueue
	AcceptedOutChannelAreaOfInterest GuaranteedQueue
	AcceptedOutChannelPayloadRequest GuaranteedQueue
	AcceptedOutChannelStaticToken    GuaranteedQueue

	// Initiator side
	InitiatorInChannelReconciliation chan wgpstypes.ReconciliationChannelMsg
	InitiatorInChannelData           chan wgpstypes.DataChannelMsg
	InitiatorInChannelIntersection   chan wgpstypes.IntersectionChannelMsg
	InitiatorInChannelCapability     chan wgpstypes.CapabilityChannelMsg
	InitiatorInChannelAreaOfInterest chan wgpstypes.AreaOfInterestChannelMsg
	InitiatorInChannelPayloadRequest chan wgpstypes.PayloadRequestChannelMsg
	InitiatorInChannelStaticToken    chan wgpstypes.StaticTokenChannelMsg
	InitiatorInChannelNone           chan wgpstypes.SyncMessage

	// Accepted side
	AcceptedInChannelReconciliation chan wgpstypes.ReconciliationChannelMsg
	AcceptedInChannelData           chan wgpstypes.DataChannelMsg
	AcceptedInChannelIntersection   chan wgpstypes.IntersectionChannelMsg
	AcceptedInChannelCapability     chan wgpstypes.CapabilityChannelMsg
	AcceptedInChannelAreaOfInterest chan wgpstypes.AreaOfInterestChannelMsg
	AcceptedInChannelPayloadRequest chan wgpstypes.PayloadRequestChannelMsg
	AcceptedInChannelStaticToken    chan wgpstypes.StaticTokenChannelMsg
	AcceptedInChannelNone           chan wgpstypes.SyncMessage

	// Commitment scheme
	//MaxPayloadSizePower int
	//ChallengeHash            func(bytes []byte) []byte
	//Nonce                    []byte
	//OurChallenge             []byte //Supposed to be async, need to see how newWgpsMessenger will affect it
	//TheirChallenge           []byte //Supposed to be async, need to see how newWgpsMessenger will affect it
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
	// GetStore         wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, K, AuthorisationToken, AuthorisationOpts]
	Store store.Store[Prefingerprint, Fingerprint, K, AuthorisationOpts, AuthorisationToken]
	// ReconcilerMap    reconciliation.ReconcilerMap[K, Prefingerprint, Fingerprint, AuthorisationOpts, AuthorisationToken] //TODO: has to be changed to ReconcilerMap
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

	ReconcilerMap         reconciliation.ReconcilerMap[K, Prefingerprint, Fingerprint, AuthorisationOpts, AuthorisationToken]
	AoiIntersectionFinder reconciliation.AoiIntersectionFinder
	DataPayloadIngester   data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorisationOpts]
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
	Fingerprint string,
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
	], addr string, // ONLY FOR TESTING!!!!
) {

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

	newWgpsMessenger.InitiatorOutChannelReconciliation = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorOutChannelData = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorOutChannelIntersection = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorOutChannelCapability = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorOutChannelAreaOfInterest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorOutChannelPayloadRequest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.InitiatorOutChannelStaticToken = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.AcceptedOutChannelReconciliation = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelData = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelIntersection = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelCapability = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelAreaOfInterest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.AcceptedOutChannelPayloadRequest = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}
	newWgpsMessenger.AcceptedOutChannelStaticToken = GuaranteedQueue{
		Queue:         make(chan []byte, 32),
		ReceivedBytes: make([]byte, 1),
		OutGoingBytes: make([]byte, 1),
	}

	newWgpsMessenger.InitiatorInChannelData = make(chan wgpstypes.DataChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelReconciliation = make(chan wgpstypes.ReconciliationChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelPayloadRequest = make(chan wgpstypes.PayloadRequestChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelNone = make(chan wgpstypes.SyncMessage, 32)
	newWgpsMessenger.InitiatorInChannelIntersection = make(chan wgpstypes.IntersectionChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelCapability = make(chan wgpstypes.CapabilityChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelStaticToken = make(chan wgpstypes.StaticTokenChannelMsg, 32)
	newWgpsMessenger.InitiatorInChannelAreaOfInterest = make(chan wgpstypes.AreaOfInterestChannelMsg, 32)

	newWgpsMessenger.AcceptedInChannelData = make(chan wgpstypes.DataChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelReconciliation = make(chan wgpstypes.ReconciliationChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelPayloadRequest = make(chan wgpstypes.PayloadRequestChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelNone = make(chan wgpstypes.SyncMessage, 32)
	newWgpsMessenger.AcceptedInChannelIntersection = make(chan wgpstypes.IntersectionChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelCapability = make(chan wgpstypes.CapabilityChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelStaticToken = make(chan wgpstypes.StaticTokenChannelMsg, 32)
	newWgpsMessenger.AcceptedInChannelAreaOfInterest = make(chan wgpstypes.AreaOfInterestChannelMsg, 32)

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

	// Reconciliation helpers

	newWgpsMessenger.ReconcilerMap = reconciliation.ReconcilerMap[K, Prefingerprint, Fingerprint, AuthorisationOpts, AuthorisationToken]{
		Map: make(map[uint64]map[uint64]reconciliation.Reconciler[K, Prefingerprint, Fingerprint, AuthorisationOpts, AuthorisationToken]),
	}
	newWgpsMessenger.AoiIntersectionFinder = *reconciliation.NewAoiIntersectionFinder(reconciliation.AoiIntersectionFinderOpts{
		NamespaceScheme: newWgpsMessenger.Schemes.NamespaceScheme,
		SubspaceScheme:  newWgpsMessenger.Schemes.SubspaceScheme,
		HandlesOurs:     newWgpsMessenger.HandlesAoisOurs,
		HandlesTheirs:   newWgpsMessenger.HandlesAoisTheirs,
	})
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
		case wgpstypes.DataChannel:
			newWgpsMessenger.InitiatorOutChannelData.Push(msg.Message)
		case wgpstypes.IntersectionChannel:
			newWgpsMessenger.InitiatorOutChannelIntersection.Push(msg.Message)
		case wgpstypes.CapabilityChannel:
			newWgpsMessenger.InitiatorOutChannelCapability.Push(msg.Message)
		case wgpstypes.AreaOfInterestChannel:
			newWgpsMessenger.InitiatorOutChannelAreaOfInterest.Push(msg.Message)
		case wgpstypes.StaticTokenChannel:
			newWgpsMessenger.InitiatorOutChannelStaticToken.Push(msg.Message)
		case wgpstypes.PayloadRequestChannel:
			newWgpsMessenger.InitiatorOutChannelPayloadRequest.Push(msg.Message)
		default:
			newWgpsMessenger.Transport.Send(msg.Message, wgpstypes.ControlChannel, wgpstypes.SyncRoleAlfie)
		}
		return nil
	}, nil)

	go syncutils.AsyncReceive[encoding.EncodedSyncMessage](newWgpsMessenger.AcceptedEncoder.MessageChannel, func(msg encoding.EncodedSyncMessage) error {
		switch msg.Channel {
		case wgpstypes.ReconciliationChannel:
			newWgpsMessenger.AcceptedOutChannelReconciliation.Push(msg.Message)
		case wgpstypes.DataChannel:
			newWgpsMessenger.AcceptedOutChannelData.Push(msg.Message)
		case wgpstypes.IntersectionChannel:
			newWgpsMessenger.AcceptedOutChannelIntersection.Push(msg.Message)
		case wgpstypes.CapabilityChannel:
			newWgpsMessenger.AcceptedOutChannelCapability.Push(msg.Message)
		case wgpstypes.AreaOfInterestChannel:
			newWgpsMessenger.AcceptedOutChannelAreaOfInterest.Push(msg.Message)
		case wgpstypes.StaticTokenChannel:
			newWgpsMessenger.AcceptedOutChannelStaticToken.Push(msg.Message)
		case wgpstypes.PayloadRequestChannel:
			newWgpsMessenger.AcceptedOutChannelPayloadRequest.Push(msg.Message)
		default:
			newWgpsMessenger.Transport.Send(msg.Message, wgpstypes.ControlChannel, wgpstypes.SyncRoleBetty)

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

	initiatorControlChannelListener := make(chan []byte, 32)
	initiatorReconciliationChannelListener := make(chan []byte, 32)
	initiatorDataChannelListener := make(chan []byte, 32)
	initiatorIntersectionChannelListener := make(chan []byte, 32)
	initiatorCapabilityChannelListener := make(chan []byte, 32)
	initiatorAreaOfInterestChannelListener := make(chan []byte, 32)
	initiatorPayloadRequestChannelListener := make(chan []byte, 32)
	initiatorStaticTokenChannelListener := make(chan []byte, 32)
	acceptedControlChannelListener := make(chan []byte, 32)
	acceptedReconciliationChannelListener := make(chan []byte, 32)
	acceptedDataChannelListener := make(chan []byte, 32)
	acceptedIntersectionChannelListener := make(chan []byte, 32)
	acceptedCapabilityChannelListener := make(chan []byte, 32)
	acceptedAreaOfInterestChannelListener := make(chan []byte, 32)
	acceptedPayloadRequestChannelListener := make(chan []byte, 32)
	acceptedStaticTokenChannelListener := make(chan []byte, 32)

	newWgpsMessenger.Transport, err = transport.NewQuicTransport(addr)
	fmt.Println("Listening Now!!")
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
			NewMessenger: &newWgpsMessenger,
			Error:        err,
		}
		return

	}

	/*go syncutils.AsyncReceive[[]byte](initiatorReconciliationChannelListener, func(msg []byte) error {

	}, nil) */

	go syncutils.AsyncReceive[[]byte](acceptedReconciliationChannelListener, func(msg []byte) error {
		fmt.Println(string(msg))
		return nil

	}, nil)

	/*go syncutils.AsyncReceive[[]byte](initiatorDataChannelListener, func(msg []byte) error {

	}, nil) */

	/*go syncutils.AsyncReceive[[]byte](acceptedDataChannelListener, func(msg []byte) error {

	}, nil) */

	go newWgpsMessenger.Transport.Recv(initiatorControlChannelListener, wgpstypes.ControlChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorReconciliationChannelListener, wgpstypes.ReconciliationChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorDataChannelListener, wgpstypes.DataChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorIntersectionChannelListener, wgpstypes.IntersectionChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorCapabilityChannelListener, wgpstypes.CapabilityChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorAreaOfInterestChannelListener, wgpstypes.AreaOfInterestChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorPayloadRequestChannelListener, wgpstypes.PayloadRequestChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(initiatorStaticTokenChannelListener, wgpstypes.StaticTokenChannel, wgpstypes.SyncRoleAlfie)
	go newWgpsMessenger.Transport.Recv(acceptedControlChannelListener, wgpstypes.ControlChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedReconciliationChannelListener, wgpstypes.ReconciliationChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedDataChannelListener, wgpstypes.DataChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedIntersectionChannelListener, wgpstypes.IntersectionChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedCapabilityChannelListener, wgpstypes.CapabilityChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedAreaOfInterestChannelListener, wgpstypes.AreaOfInterestChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedPayloadRequestChannelListener, wgpstypes.PayloadRequestChannel, wgpstypes.SyncRoleBetty)
	go newWgpsMessenger.Transport.Recv(acceptedStaticTokenChannelListener, wgpstypes.StaticTokenChannel, wgpstypes.SyncRoleBetty)
	/*
		decodedInitiatorControlChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorReconciliationChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorDataChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorIntersectionChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorCapabilityChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorAreaOfInterestChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorPayloadRequestChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedInitiatorStaticTokenChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedControlChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedReconciliationChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedDataChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedIntersectionChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedCapabilityChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedAreaOfInterestChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedPayloadRequestChannelListener := make(chan wgpstypes.SyncMessage, 32)
		decodedAcceptedStaticTokenChannelListener := make(chan wgpstypes.SyncMessage, 32)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorControlChannelListener, decodedInitiatorControlChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorReconciliationChannelListener, decodedInitiatorReconciliationChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorDataChannelListener, decodedInitiatorDataChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorIntersectionChannelListener, decodedInitiatorIntersectionChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorCapabilityChannelListener, decodedInitiatorCapabilityChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorAreaOfInterestChannelListener, decodedInitiatorAreaOfInterestChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorPayloadRequestChannelListener, decodedInitiatorPayloadRequestChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			initiatorStaticTokenChannelListener, decodedInitiatorStaticTokenChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedControlChannelListener, decodedAcceptedControlChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedDataChannelListener, decodedAcceptedDataChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedReconciliationChannelListener, decodedAcceptedReconciliationChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedIntersectionChannelListener, decodedAcceptedIntersectionChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedPayloadRequestChannelListener, decodedAcceptedPayloadRequestChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedAreaOfInterestChannelListener, decodedAcceptedAreaOfInterestChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedStaticTokenChannelListener, decodedAcceptedStaticTokenChannelListener)

		go decoding.DecodeMessages[
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
		](decoding.DecodeMessageOpts[ReadCapability, Receiver, SyncSignature, ReceiverSecretKey, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, Prefingerprint, Fingerprint, AuthorisationToken, StaticToken, DynamicToken, AuthorisationOpts, K]{
			Schemes: newWgpsMessenger.Schemes,
			Reconcile: reconciliation.ReconcileMsgTrackerOpts{
				DefaultNamespaceId:   newWgpsMessenger.Schemes.NamespaceScheme.DefaultNamespaceId,
				DefaultSubspaceId:    newWgpsMessenger.Schemes.SubspaceScheme.MinimalSubspaceId,
				DefaultPayloadDigest: newWgpsMessenger.Schemes.Payload.DefaultPayloadDigest,
				HandleToNamespaceId: func(handle uint64) types.NamespaceId {
					return newWgpsMessenger.AoiIntersectionFinder.HandleToNamespaceId(handle, false)
				},
				AoiHandlesToRange3d: func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d {
					reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(senderAoiHandle, receiverAoiHandle)
					return <-reconciler.Ranges
				},
			},
			AoiHandlesToArea: func(senderHandle uint64, receiverHandle uint64) types.Area {
				senderAoi, _ := newWgpsMessenger.HandlesAoisTheirs.Get(senderHandle)
				receiverAoi, _ := newWgpsMessenger.HandlesAoisOurs.Get(receiverHandle)

				intersectionArea := utils.IntersectArea(newWgpsMessenger.Schemes.SubspaceScheme.Order, senderAoi.Area, receiverAoi.Area)

				return *intersectionArea
			},
			GetCurrentlyReceivedEntry: func() types.Entry {
				return newWgpsMessenger.CurrentlyReceivedEntry
			},
			AoiHandlesToNamespace: func(senderHandle uint64, receiverHandle uint64) types.NamespaceId {

				reconciler, _ := newWgpsMessenger.ReconcilerMap.GetReconciler(
					receiverHandle, senderHandle,
				)
				return reconciler.Store.NameSpaceId
			},
		},
			acceptedCapabilityChannelListener, decodedAcceptedCapabilityChannelListener)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorControlChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorDataChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorIntersectionChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorPayloadRequestChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorCapabilityChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorStaticTokenChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorReconciliationChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedInitiatorAreaOfInterestChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedControlChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedDataChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedIntersectionChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedPayloadRequestChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedCapabilityChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedReconciliationChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedStaticTokenChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](decodedAcceptedAreaOfInterestChannelListener, func(msg wgpstypes.SyncMessage) error {
			switch msg := msg.(type) {
			case wgpstypes.MsgDataSendEntry[DynamicToken]:
				newWgpsMessenger.CurrentlyReceivedEntry = msg.Data.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = msg.Data.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgDataReplyPayload:
				request, found := newWgpsMessenger.HandlesPayloadRequestsOurs.Get(msg.Data.Handle)
				if !found {
					return fmt.Errorf("No Payload handler")
				}

				newWgpsMessenger.CurrentlyReceivedEntry = request.Entry
				newWgpsMessenger.CurrentlyReceivedOffset = request.Offset
				newWgpsMessenger.InitiatorInChannelData <- msg
			case wgpstypes.MsgPaiBindFragment[PsiGroup]:
				newWgpsMessenger.InitiatorInChannelIntersection <- msg

			case wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature]:
				newWgpsMessenger.InitiatorInChannelCapability <- msg

			case wgpstypes.MsgSetupBindAreaOfInterest:
				newWgpsMessenger.InitiatorInChannelAreaOfInterest <- msg

			case wgpstypes.MsgSetupBindStaticToken[StaticToken]:
				newWgpsMessenger.InitiatorInChannelStaticToken <- msg

			case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationAnnounceEntries:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendEntry[DynamicToken]:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationSendPayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgReconciliationTerminatePayload:
				newWgpsMessenger.InitiatorInChannelReconciliation <- msg
			case wgpstypes.MsgDataSendPayload:
				newWgpsMessenger.InitiatorInChannelData <- msg

			case wgpstypes.MsgDataBindPayloadRequest:
				newWgpsMessenger.InitiatorInChannelPayloadRequest <- msg

			default:
				newWgpsMessenger.InitiatorInChannelNone <- msg

			}
			return nil
		}, nil)

		// Handle received messages

		go syncutils.AsyncReceive[wgpstypes.ReconciliationChannelMsg](newWgpsMessenger.InitiatorInChannelReconciliation, func(msg wgpstypes.ReconciliationChannelMsg) error {
			//newWgpsMessenger.HandleMsgReconciliation(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.StaticTokenChannelMsg](newWgpsMessenger.InitiatorInChannelStaticToken, func(msg wgpstypes.StaticTokenChannelMsg) error {
			//newWgpsMessenger.HandleMsgStaticToken(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.DataChannelMsg](newWgpsMessenger.InitiatorInChannelData, func(msg wgpstypes.DataChannelMsg) error {
			newWgpsMessenger.HandleMsgData(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.IntersectionChannelMsg](newWgpsMessenger.InitiatorInChannelIntersection, func(msg wgpstypes.IntersectionChannelMsg) error {
			//newWgpsMessenger.HandleMsgIntersection(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.PayloadRequestChannelMsg](newWgpsMessenger.InitiatorInChannelPayloadRequest, func(msg wgpstypes.PayloadRequestChannelMsg) error {
			newWgpsMessenger.HandleMsgPayloadRequest(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.CapabilityChannelMsg](newWgpsMessenger.InitiatorInChannelCapability, func(msg wgpstypes.CapabilityChannelMsg) error {
			//newWgpsMessenger.HandleMsgCapability(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.AreaOfInterestChannelMsg](newWgpsMessenger.InitiatorInChannelAreaOfInterest, func(msg wgpstypes.AreaOfInterestChannelMsg) error {
			//newWgpsMessenger.HandleMsgAreaOfInterest(msg)
			return nil

		}, nil)

		go syncutils.AsyncReceive[wgpstypes.SyncMessage](newWgpsMessenger.InitiatorInChannelNone, func(msg wgpstypes.SyncMessage) error {
			//newWgpsMessenger.HandleMsg(msg)
			return nil

		}, nil) */

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
		NewMessenger: &newWgpsMessenger,
		Error:        nil,
	}
	fmt.Println("Everything looks good!")
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
]) Initiate(addr string) error {

	err := w.Transport.Initiate(addr)
	time.Sleep(time.Second * 1)
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
		_, err := w.Store.IngestEntry(msg.Data.Entry, authToken)
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

func (
	w *WgpsMessenger[
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
	]) HandleMsgReconciliationAcceptor(
	msg wgpstypes.ReconciliationChannelMsg,
	role wgpstypes.SyncRole,
) {
	switch msg := msg.(type) {
	case wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]:
		reconciler, err := w.ReconcilerMap.GetReconciler(msg.Data.ReceiverHandle, msg.Data.SenderHandle)
		if err != nil {
			log.Fatal(err)
		}
		left, right, response := reconciler.Respond(msg.Data.Range, msg.Data.Fingerprint, w.YourRangeCounter)

		if response.WantResponse {
			extendedEntries, err := w.Store.EntryDriver.Query(response.Range)
			if err != nil {
				log.Fatal(err)
			}
			for _, extendedEntry := range extendedEntries {
				payload, _ := w.Store.GetPayload(types.Position3d{
					Path:     extendedEntry.Entry.Path,
					Subspace: extendedEntry.Entry.Subspace_id,
					Time:     extendedEntry.Entry.Timestamp,
				})
				encodedEntryPayload := EncodeEntryPayload(extendedEntry.Entry, payload)
				var finalEncoded []byte
				binary.BigEndian.PutUint64(finalEncoded, uint64(len(encodedEntryPayload)))
				finalEncoded = append(finalEncoded, encodedEntryPayload...)
				w.Transport.Send(finalEncoded, wgpstypes.DataChannel, role)
			}
		} else {
			if reflect.DeepEqual(left, struct {
				Range       types.Range3d
				FingerPrint Fingerprint
				Covers      uint64
			}{}) && reflect.DeepEqual(right, struct {
				Range       types.Range3d
				FingerPrint Fingerprint
				Covers      uint64
			}{}) {
				return
			}
			var leftEncodedExtendedRange []byte
			binary.BigEndian.PutUint64(leftEncodedExtendedRange, uint64(len(EncodeExtendedRange(left))))
			leftEncodedExtendedRange = append(leftEncodedExtendedRange, EncodeExtendedRange(left)...)
			w.Transport.Send(leftEncodedExtendedRange, wgpstypes.ReconciliationChannel, role)

			var rightEncodedExtendedRange []byte
			binary.BigEndian.PutUint64(rightEncodedExtendedRange, uint64(len(EncodeExtendedRange(right))))
			rightEncodedExtendedRange = append(rightEncodedExtendedRange, EncodeExtendedRange(right)...)
			w.Transport.Send(rightEncodedExtendedRange, wgpstypes.ReconciliationChannel, role)
		}
	}

}

func (
	w *WgpsMessenger[
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
	]) HandleMsgReconciliationInitiator(
	left struct {
		Range       types.Range3d
		FingerPrint Fingerprint
		Covers      uint64
	},
	right struct {
		Range       types.Range3d
		FingerPrint Fingerprint
		Covers      uint64
	},
	role wgpstypes.SyncRole,
) {
	summaryLeft := w.Store.EntryDriver.Storage.Summarise(left.Range)
	summaryRight := w.Store.EntryDriver.Storage.Summarise(right.Range)

	w.Transport.Send(EncodeSummary(summaryLeft, left.Range), wgpstypes.ReconciliationChannel, role)
	w.Transport.Send(EncodeSummary(summaryRight, right.Range), wgpstypes.ReconciliationChannel, role)
}

func (
	w *WgpsMessenger[
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
	]) HandleDataInitiator(
	entry types.Entry,
	payload []byte,
	role wgpstypes.SyncRole,
) {
	var entryInput datamodeltypes.EntryInput
	entryInput.Path = entry.Path
	entryInput.Subspace = entry.Subspace_id
	entryInput.Timestamp = entry.Timestamp
	entryInput.Payload = payload
	w.Store.Set(entryInput, []byte(entry.Subspace_id))
}

func EncodeEntry(entry types.Entry) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	// Register the types with gob
	gob.Register(types.NamespaceId(nil))
	gob.Register(types.SubspaceId(nil))
	gob.Register(types.PayloadDigest(""))
	gob.Register(types.Path(nil))
	gob.Register(uint64(0))
	gob.Register(uint64(0))

	// Encode the entry
	encoder.Encode(entry)
	return buffer.Bytes()
}

func DecodeEntry(data []byte) (types.Entry, error) {
	var entry types.Entry
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	// Register the types with gob
	gob.Register(types.NamespaceId(nil))
	gob.Register(types.SubspaceId(nil))
	gob.Register(types.PayloadDigest(""))
	gob.Register(types.Path(nil))
	gob.Register(uint64(0))
	gob.Register(uint64(0))

	// Decode the entry
	err := decoder.Decode(&entry)
	if err != nil {
		return entry, fmt.Errorf("failed to decode entry: %w", err)
	}

	return entry, nil
}

func EncodeExtendedRange[Fingerprint string](value struct {
	Range       types.Range3d
	FingerPrint Fingerprint
	Covers      uint64
}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	gob.Register(types.Range3d{})
	gob.Register(Fingerprint(""))
	gob.Register(uint64(0))
	encoder.Encode(value)
	return buffer.Bytes()
}

func DecodeExtendedRange[Fingerprint string](value []byte) struct {
	Range       types.Range3d
	FingerPrint Fingerprint
	Covers      uint64
} {
	var decoded struct {
		Range       types.Range3d
		FingerPrint Fingerprint
		Covers      uint64
	}
	buffer := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buffer)
	gob.Register(types.Range3d{})
	gob.Register(Fingerprint(""))
	gob.Register(uint64(0))
	decoder.Decode(&decoded)
	return decoded
}

func EncodeSummary[Fingerprint string](value struct {
	FingerPrint string
	Size        uint64
}, ourRange types.Range3d) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	gob.Register(Fingerprint(""))
	gob.Register(uint64(0))
	gob.Register(types.Range3d{})
	encoder.Encode(value)
	return buffer.Bytes()
}
func DecodeSummary[Fingerprint string](value []byte) (struct {
	FingerPrint string
	Size        uint64
}, types.Range3d) {
	var decoded struct {
		FingerPrint string
		Size        uint64
		ourRange    types.Range3d
	}
	buffer := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buffer)
	gob.Register(Fingerprint(""))
	gob.Register(uint64(0))
	gob.Register(types.Range3d{})
	decoder.Decode(&decoded)
	return struct {
		FingerPrint string
		Size        uint64
	}{
		decoded.FingerPrint, decoded.Size,
	}, decoded.ourRange
}

func EncodeEntryPayload(entry types.Entry, payload []byte) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	gob.Register(types.Entry{})
	gob.Register([]byte(nil))
	encoder.Encode(entry)
	encoder.Encode(payload)
	return buffer.Bytes()
}
func DecodeEntryPayload(value []byte) (entry types.Entry, payload []byte) {
	var decoded struct {
		Entry   types.Entry
		Payload []byte
	}
	buffer := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buffer)
	gob.Register(types.Entry{})
	gob.Register([]byte(nil))
	decoder.Decode(&decoded)
	return decoded.Entry, decoded.Payload
}
