package wgps

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/data"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
	"crypto/rand"
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
	//Transport transport.Transport
	/** Sets the [`maximum payload size`](https://willowprotocol.org/specs/sync/index.html#peer_max_payload_size) for this peer, which is 2 to the power of the given number.
	 *
	 * The given power must be a natural number lesser than or equal to 64. */
	MaxPayloadSizePower int

	/** Sets the [`challeng_length`](https://willowprotocol.org/specs/sync/index.html#challenge_length) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	ChallengeLength int
	/** Sets the [`challeng_hash_length`](https://willowprotocol.org/specs/sync/index.html#challenge_hash_length) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	ChallengeHashLength int
	/** Sets the [`challeng_hash`](https://willowprotocol.org/specs/sync/index.html#challenge_hash) for the [Willow General Purpose Sync Protocol](https://willowprotocol.org/specs/sync/index.html#sync).*/
	ChallengeHash func(bytes []byte) []byte
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
	Interests map[*wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest

	GetStore               wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts]
	TransformPayload       func(chunk []byte) []byte
	ProcessReceivedPayload func(chunk []byte, entryLength uint64) []byte
}
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
	//Interests    [][]types.AreaOfInterest
	Interests map[*wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]][]types.AreaOfInterest
	//Capabilities []wgpstypes.ReadAuthorisation[ReadCapability, SubspaceCapability]
	//Transport    transport.ReadyTransport
	//Encoder                  Encoder //TODO: has to be changed to MessageEncoder
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
	/*HandleIntersectionOurs   handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	HandleIntersectionTheirs handlestore.HandleStore[wgpstypes.Intersection[PsiGroup]]
	PaiFinder                pai.PaiFinder[ReadCapability, PsiGroup, PsiScalar, SubspaceCapability, K]
	*/
	//Setup
	HandleCapsOurs   handlestore.HandleStore[ReadCapability]
	HandleCapsTheirs handlestore.HandleStore[ReadCapability]

	HandlesAoisOurs   handlestore.HandleStore[types.AreaOfInterest]
	HandlesAoisTheirs handlestore.HandleStore[types.AreaOfInterest]

	HandlesStaticTokenOurs   handlestore.HandleStore[StaticToken]
	HandlesStaticTokenTheirs handlestore.HandleStore[StaticToken]

	//Reconciliation
	YourRangeCounter int
	GetStore         wgpstypes.GetStoreFn[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts]
	//ReconcilerMap             reconciliation.ReconcilerMap //TODO: has to be changed to ReconcilerMap
	//AoiIntersectionFinder     reconciliation.AoiIntersectionFinder
	//Announcer                 reconciliation.Announcer
	CurrentlyReceivingEntries struct {
		Namespace   types.NamespaceId
		Range       types.Range3d
		Remaining   uint64
		IsReceiving bool
	}
	ReconciliationPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

	//Data
	//CapFinder               CapFinder[ReadCapability, SyncSignature, Receiver, ReceiverSecretKey, K]
	CurrentlySentEntry      types.Entry
	CurrentlyReceivedEntry  types.Entry
	CurrentlyReceivedOffset uint64

	HandlesPayloadRequestsOurs   handlestore.HandleStore[handlestore.HandleStore[types.AreaOfInterest]] //types.AreaOfInterest is just placehoder
	HandlesPayloadRequestsTheirs handlestore.HandleStore[handlestore.HandleStore[types.AreaOfInterest]] //types.AreaOfInterest is just placehoder

	DataSender data.DataSender[Prefingerprint, Fingerprint, AuthorisationToken, DynamicToken, AuthorisationOpts] //Need to change the type definition

	DataPayloadIngester data.PayloadIngester[Prefingerprint, Fingerprint, AuthorisationToken, AuthorsationOpts] //will have to change the type definition

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

	if opts.MaxPayloadSizePower < 0 || opts.MaxPayloadSizePower > 64 {
		return newWgpsMessenger, fmt.Errorf("MaxPayloadSizePower must be a natural number lesser than or equal to 64")
	}

	for authorisation, areas := range opts.Interests {
		if len(areas) == 0 {
			return newWgpsMessenger, fmt.Errorf("No Area of Interest given")
		}

		// Get granted area of authorisation
		grantedArea := opts.Schemes.AccessControl.GetGrantedArea(
			(*authorisation).Capability,
		)
		for _, aoi := range areas {
			isWithin := utils.AreaIsIncluded(
				opts.Schemes.SubspaceScheme.Order,
				aoi.Area,
				grantedArea,
			)
			if !isWithin {
				return newWgpsMessenger, fmt.Errorf("Given authorisation is not within authorisation's granted area")
			}
		}
	}

 newWgpsMessenger{
	GetStore : opts.GetStore,
	Interests: opts.Interests,
	Schemes : opts.Schemes,
	Nonce: rand.Read(make([]byte, opts.ChallengeLength)),
    	


}

