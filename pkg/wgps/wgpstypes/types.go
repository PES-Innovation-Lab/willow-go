package wgpstypes

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

// Define constants for peer roles
const (
	SyncRoleAlfie SyncRole = "alfie"
	SyncRoleBetty SyncRole = "betty"
)

// Define type SyncRole as an alias for string
type SyncRole string

// Function to check if role is Alfie (optional)
func IsAlfie(role SyncRole) bool {
	return role == SyncRoleAlfie
}

// Function to check if role is Betty (optional)
func IsBetty(role SyncRole) bool {
	return role == SyncRoleBetty
}

type GetStoreFn[PreFingerPrint,
	Fingerprint constraints.Ordered,
	K constraints.Unsigned,
	AuthorisationToken string,
	AuthorisationOpts []byte] func(namespace types.NamespaceId) store.Store[PreFingerPrint, Fingerprint, K, AuthorisationOpts, AuthorisationToken]

type ReadAuthorisation[ReadCapability, SubspaceReadCapability any] struct {
	Capability ReadCapability
	// SubspaceCapability is optional here
	SubspaceCapability    SubspaceReadCapability
	HasSubspaceCapability bool
}

//will need to check if the type is any or something else

// Transport defines the interface for communication channels
type Transport interface {
	Send(data []byte, channel Channel) error // Use byte slice instead of Uint8Array
	Recv(channel Channel) ([]byte, error)    // Returns a receive channel and potential error (PLEASE CHECK IF THIS IS RIGHT)
	Close() error
	IsClosed() bool
}

type HandleType int

const (
	/* Resource handle for the private set intersection part of private area intersection. More precisely, an IntersectionHandle stores a PsiGroup member together with one of two possible states:
	- pending (waiting for the other peer to perform scalar multiplication),
	 - completed (both peers performed scalar multiplication). */
	IntersectionHandle HandleType = iota
	/* Logical channel for controlling the binding of new CapabilityHandles. */
	CapabilityHandle
	/* Resource handle for AreaOfInterests that peers wish to sync. */
	AreaOfInterestHandle
	/* Resource handle that controls the matching from Payload transmissions to Payload requests. */
	PayloadRequestHandle
	/* Resource handle for StaticTokens that peers need to transmit. */
	StaticTokenHandle
)

type Channel int

const (
	ControlChannel Channel = iota
	/* Logical channel for performing 3d range-based set reconciliation. */
	ReconciliationChannel
	/* Logical channel for transmitting Entries and Payloads outside of 3d range-based set reconciliation. */
	DataChannel
	/* Logical channel for controlling the binding of new IntersectionHandles. */
	IntersectionChannel
	/* Logical channel for controlling the binding of new CapabilityHandles. */
	CapabilityChannel
	/* Logical channel for controlling the binding of new AreaOfInterestHandles. */
	AreaOfInterestChannel
	/* Logical channel for controlling the binding of new PayloadRequestHandles. */
	PayloadRequestChannel
	/* Logical channel for controlling the binding of new StaticTokenHandles. */
	StaticTokenChannel
)

type MsgKind int

const (
	CommitmentReveal MsgKind = iota
	PaiBindFragment
	PaiReplyFragment
	PaiRequestSubspaceCapability
	PaiReplySubspaceCapability
	SetupBindReadCapability
	SetupBindAreaOfInterest
	SetupBindStaticToken
	ReconciliationSendFingerprint
	ReconciliationAnnounceEntries
	ReconciliationSendEntry
	ReconciliationSendPayload
	ReconciliationTerminatePayload
	DataSendEntry
	DataSendPayload
	DataSetMetadata
	DataBindPayloadRequest
	DataReplyPayload
	ControlIssueGuarantee
	ControlAbsolve
	ControlPlead
	ControlAnnounceDropping
	ControlApologise
	ControlFree
)

// 1. Control messages

/** Make a binding promise of available buffer capacity to the other peer. */
type ControlIssueGuaranteeData struct {
	Amount  uint64
	Channel Channel
}
type MsgControlIssueGuarantee struct {
	Kind MsgKind
	Data ControlIssueGuaranteeData
}

/** Allow the other peer to reduce its total buffer capacity by amount. */
type ControlAbsolveData struct {
	Amount  uint64
	Channel Channel
}
type MsgControlAbsolve struct {
	Kind MsgKind
	Data ControlAbsolveData
}

/** Ask the other peer to send an ControlAbsolve message such that the receiver remaining guarantees will be target. */
type ControlPleadData struct {
	Target  uint64
	Channel Channel
}
type MsgControlPlead struct {
	Kind MsgKind
	Data ControlPleadData
}

type ControlAnnounceDroppingData struct {
	Channel Channel
}
type MsgControlAnnounceDropping struct {
	Kind MsgKind
	Data ControlAnnounceDroppingData
}

/** Notify the other peer that it can stop dropping messages of this logical channel. */
type ControlApologiseData struct {
	Channel Channel
}
type MsgControlApologise struct {
	Kind MsgKind
	Data ControlApologiseData
}

type MsgControlFreeData struct {
	Handle uint64
	/** Indicates whether the peer sending this message is the one who created the handle (true) or not (false). */
	Mine       bool
	HandleType HandleType
}
type MsgControlFree struct {
	Kind MsgKind
	Data MsgControlFreeData
}

/** Complete the commitment scheme to determine the challenge for read authentication. */
type MsgCommitmentRevealData struct {
	Nonce []byte
}
type MsgCommitmentReveal struct {
	Kind MsgKind
	Data MsgCommitmentRevealData
}

// 2. Intersection messages

/** Bind data to an IntersectionHandle for performing private area intersection. */
type MsgPaiBindFragmentData[PsiGroup any] struct {
	GroupMember PsiGroup
	IsSecondary bool
}
type MsgPaiBindFragment[PsiGroup any] struct {
	Kind MsgKind
	Data MsgPaiBindFragmentData[PsiGroup]
}

/** Finalise private set intersection for a single item. */
type MsgPaiReplyFragmentData[PsiGroup any] struct {
	Handle      uint64
	GroupMember PsiGroup
}
type MsgPaiReplyFragment[PsiGroup any] struct {
	Kind MsgKind
	Data MsgPaiReplyFragmentData[PsiGroup]
}

/** Request the subspace capability for a given IntersectionHandle (for the least-specific secondary fragment for whose NamespaceId the request is being made). */
type MsgPaiRequestSubspaceCapabilityData struct {
	Handle uint64
}
type MsgPaiRequestSubspaceCapability struct {
	Kind MsgKind
	Data MsgPaiRequestSubspaceCapabilityData
}

/** Send a previously requested SubspaceCapability. */
type MsgPaiReplySubspaceCapabilityData[SubspaceCapability, SyncSubspaceSignature constraints.Ordered] struct {
	Handle     uint64
	Capability SubspaceCapability
	Signature  SyncSubspaceSignature
}
type MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature constraints.Ordered] struct {
	Kind MsgKind
	Data MsgPaiReplySubspaceCapabilityData[SubspaceCapability, SyncSubspaceSignature]
}

// 3. Setup messages

/** Bind a ReadCapability to a CapabilityHandle. */
type MsgSetupBindReadCapabilityData[ReadCapability, SyncSignature constraints.Ordered] struct {
	Capability ReadCapability
	Handle     uint64
	Signature  SyncSignature
}
type MsgSetupBindReadCapability[ReadCapability, SyncSignature constraints.Ordered] struct {
	Kind MsgKind
	Data MsgSetupBindReadCapabilityData[ReadCapability, SyncSignature]
}

/** Bind an AreaOfInterest to an AreaOfInterestHandle. */
type MsgSetupBindAreaOfInterestData struct {
	AreaOfInterest types.AreaOfInterest
	Authorisation  uint64
}
type MsgSetupBindAreaOfinterest struct {
	Kind MsgKind
	Data MsgSetupBindAreaOfInterestData
}

type MsgSetupBindStaticTokenData[StaticToken constraints.Ordered] struct {
	StaticToken StaticToken
}
type MsgSetupBindStaticToken[StaticToken constraints.Ordered] struct {
	Kind MsgKind
	Data MsgSetupBindStaticTokenData[StaticToken]
}

/** Send a Fingerprint as part of 3d range-based set reconciliation. */
type MsgReconciliationSendFingerprintData[Fingerprint constraints.Ordered] struct {
	Range          types.Range3d
	Fingerprint    Fingerprint
	SenderHandle   uint64
	ReceiverHandle uint64
	Covers         uint64
	DoesCover      bool
}
type MsgReconciliationSendFingerprint[Fingerprint constraints.Ordered] struct {
	Kind MsgKind
	Data MsgReconciliationSendFingerprintData[Fingerprint]
}

/** Prepare transmission of the LengthyEntries a peer has in a 3dRange as part of 3d range-based set reconciliation. */
type MsgReconciliationAnnounceEntriesData struct {
	Range          types.Range3d
	Count          uint64
	WantResponse   bool
	WillSort       bool
	SenderHandle   uint64
	ReceiverHandle uint64
	Covers         uint64
	DoesCover      bool
}
type MsgReconciliationAnnounceEntries struct {
	Kind MsgKind
	Data MsgReconciliationAnnounceEntriesData
}

/** Transmit a LengthyEntry as part of 3d range-based set reconciliation. */
type MsgReconciliationSendEntryData[DynamicToken constraints.Ordered] struct {
	Entry             datamodeltypes.LengthyEntry
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
}
type MsgReconciliationSendEntry[DynamicToken constraints.Ordered] struct {
	Kind MsgKind
	Data MsgReconciliationSendEntryData[DynamicToken]
}

/** Transmit a Payload as part of 3d range-based set reconciliation. */
type MsgReconciliationSendPayloadData struct {
	Amount uint64
	Bytes  []byte
}
type MsgReconciliationSendPayload struct {
	Kind MsgKind
	Data MsgReconciliationSendPayloadData
}

/** Notify the other peer that the payload transmission is complete. */
type MsgReconciliationTerminatePayload struct {
	Kind MsgKind
}

// 4. Data messages

/** Transmit an AuthorisedEntry to the other peer, and optionally prepare transmission of its Payload. */
type MsgDataSendEntryData[DynamicToken constraints.Ordered] struct {
	Entry             types.Entry
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
	Offset            uint64
}
type MsgDataSendEntry[DynamicToken constraints.Ordered] struct {
	Kind MsgKind
	Data MsgDataSendEntryData[DynamicToken]
}

/** Transmit a Payload to the other peer. */
type MsgDataSendPayloadData struct {
	Amount uint64
	Bytes  []byte
}
type MsgDataSendPayload struct {
	Kind MsgKind
	Data MsgDataSendPayloadData
}

/** Express a preference whether the other peer should eagerly forward Payloads in the intersection of two AreaOfInterests. */
type MsgDataSetMetadataData struct {
	IsEager        bool
	SenderHandle   uint64
	ReceiverHandle uint64
}
type MsgDataSetMetadata struct {
	Kind MsgKind
	Data MsgDataSetMetadataData
}

/** Bind a PayloadRequest to a PayloadRequestHandle. */
type MsgDataBindPayloadRequestData struct {
	Entry      types.Entry
	Offset     uint64
	Capability uint64
}
type MsgDataBindPayloadRequest struct {
	Kind MsgKind
	Data MsgDataBindPayloadRequestData
}

/** Transmit a Payload to the other peer. */
type MsgDataReplyPayloadData struct {
	Handle uint64
}
type MsgDataReplyPayload struct {
	Kind MsgKind
	Data MsgDataReplyPayloadData
}

type SyncMessage interface {
	IsSyncMessage()
}

func (MsgControlIssueGuarantee) IsSyncMessage()                                                 {}
func (MsgControlAbsolve) IsSyncMessage()                                                        {}
func (MsgControlPlead) IsSyncMessage()                                                          {}
func (MsgControlAnnounceDropping) IsSyncMessage()                                               {}
func (MsgControlApologise) IsSyncMessage()                                                      {}
func (MsgControlFree) IsSyncMessage()                                                           {}
func (MsgCommitmentReveal) IsSyncMessage()                                                      {}
func (MsgPaiBindFragment[PsiGroup]) IsSyncMessage()                                             {}
func (MsgPaiReplyFragment[PsiGroup]) IsSyncMessage()                                            {}
func (MsgPaiRequestSubspaceCapability) IsSyncMessage()                                          {}
func (MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]) IsSyncMessage() {}
func (MsgSetupBindReadCapability[ReadCapability, SyncSignature]) IsSyncMessage()                {}
func (MsgSetupBindAreaOfinterest) IsSyncMessage()                                               {}
func (MsgSetupBindStaticToken[StaticToken]) IsSyncMessage()                                     {}
func (MsgReconciliationSendFingerprint[Fingerprint]) IsSyncMessage()                            {}
func (MsgReconciliationAnnounceEntries) IsSyncMessage()                                         {}
func (MsgReconciliationSendEntry[DynamicToken]) IsSyncMessage() {
}
func (MsgReconciliationSendPayload) IsSyncMessage()      {}
func (MsgReconciliationTerminatePayload) IsSyncMessage() {}
func (MsgDataSendPayload) IsSyncMessage()                {}
func (MsgDataSetMetadata) IsSyncMessage()                {}
func (MsgDataBindPayloadRequest) IsSyncMessage()         {}
func (MsgDataReplyPayload) IsSyncMessage()               {}

// Messages categorised by logical channel
type ReconciliationChannelMsg interface {
	IsReconciliationChannelMsg()
}

func (MsgReconciliationSendFingerprint[Fingerprint]) IsReconciliationChannelMsg() {}
func (MsgReconciliationAnnounceEntries) IsReconciliationChannelMsg()              {}
func (MsgReconciliationSendEntry[DynamicToken]) IsReconciliationChannelMsg() {
}
func (MsgReconciliationSendPayload) IsReconciliationChannelMsg()      {}
func (MsgReconciliationTerminatePayload) IsReconciliationChannelMsg() {}

type DataChannelMsg interface {
	IsDataChannelMsg()
}

func (MsgDataSendEntry[DynamicToken]) IsDataChannelMsg() {}
func (MsgDataReplyPayload) IsDataChannelMsg()            {}
func (MsgDataSendPayload) IsDataChannelMsg()             {}

type IntersectionChannelMsg interface {
	IsIntersectionChannelMsg()
}

func (MsgPaiBindFragment[PsiGroup]) IsIntersectionChannelMsg() {}

type CapabilityChannelMsg interface {
	IsCapabilityChannelMsg()
}

func (MsgSetupBindReadCapability[ReadCapability, SyncSignature]) isCapabilityChannelMsg() {}

type AreaOfInterestChannelMsg interface {
	IsAreaOfInterestChannelMsg()
}

func (MsgSetupBindAreaOfinterest) IsAreaOfInterestChannelMsg() {}

type PayloadRequestChannelMsg interface {
	IsPayloadRequestChannelMsg()
}

func (MsgDataBindPayloadRequest) IsPayloadRequestChannelMsg() {
}

type StaticTokenChannelMsg interface {
	IsStaticTokenChannelMsg()
}

func (MsgSetupBindStaticToken[StaticToken]) IsStaticTokenChannelMsg() {}

/** Messages which belong to no logical channel. */
type NoChannelMsg interface {
	IsNoChannelMsg()
}

func (MsgControlIssueGuarantee) IsNoChannelMsg()                                                 {}
func (MsgControlAbsolve) IsNoChannelMsg()                                                        {}
func (MsgControlPlead) IsNoChannelMsg()                                                          {}
func (MsgControlAnnounceDropping) IsNoChannelMsg()                                               {}
func (MsgControlApologise) IsNoChannelMsg()                                                      {}
func (MsgControlFree) IsNoChannelMsg()                                                           {}
func (MsgCommitmentReveal) IsNoChannelMsg()                                                      {}
func (MsgPaiReplyFragment[PsiGroup]) IsNoChannelMsg()                                            {}
func (MsgPaiRequestSubspaceCapability) IsNoChannelMsg()                                          {}
func (MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]) IsNoChannelMsg() {}
func (MsgDataSetMetadata) IsNoChannelMsg()                                                       {}

// Encodings

type ReadCapPrivy struct {
	Outer     types.Area
	Namespace types.NamespaceId
}

// Define the PrivyEncodingScheme type with generics
type PrivyEncodingScheme[ReadCapability any, Privy any] struct {
	ReadCapability ReadCapability
	Privy          Privy
}

// Define the ReadCapEncodingScheme type alias
type ReadCapEncodingScheme[ReadCapability any] struct {
	PrivyEncodingScheme[ReadCapability, ReadCapPrivy]
}

type ReconciliationPrivy struct {
	PrevSenderHandle      uint64
	PrevReceiverHandle    uint64
	PrevRange             types.Range3d
	PrevStaticTokenHandle uint64
	PrevEntry             types.Entry
	Announced             struct {
		Range     types.Range3d
		Namespace types.NamespaceId
	}
}

/** The parameter schemes required to instantiate a `WgpsMessenger`. */
type SyncSchemes[ReadCapability any,
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
	K constraints.Unsigned] struct {
	AccessControl      AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey, K]
	SubspaceCap        SubspaceCapScheme[SubspaceReceiver, SubspaceSecretKey, SubspaceCapability, SyncSubspaceSignature, K]
	Pai                PaiScheme[ReadCapability, PsiGroup, PsiScalar, K]
	NamespaceScheme    datamodeltypes.NamespaceScheme
	SubspaceScheme     datamodeltypes.SubspaceScheme
	PathParams         types.PathParams[K]
	AuhtorisationToken AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken]
	Payload            datamodeltypes.PayloadScheme
	Fingerprint        datamodeltypes.FingerprintScheme[Prefingerprint, Fingerprint]
}

type AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey any, K constraints.Unsigned] struct {
	GetReceiver         func(cap ReadCapability) Receiver
	GetSecretKey        func(receiver Receiver) ReceiverSecretKey
	GetGrantedArea      func(cap ReadCapability) types.Area
	GetGrantedNamespace func(cap ReadCapability) types.NamespaceId
	Signatures          types.SignatureScheme[Receiver, ReceiverSecretKey, SyncSignature]
	IsValidCap          func(cap ReadCapability) bool
	Encodings           struct {
		ReadCap       ReadCapEncodingScheme[ReadCapability]
		SyncSignature utils.EncodingScheme[K]
	}
}

type SubspaceCapScheme[SubspaceReceiver types.SubspaceId, SubspaceSecretKey, SubspaceCapability, SyncSubspaceSignature any, K constraints.Unsigned] struct {
	GetSecretKey func(receiver SubspaceReceiver) SubspaceSecretKey
	GetNamespace func(cap SubspaceCapability) types.NamespaceId
	GetReceiver  func(cap SubspaceCapability) SubspaceReceiver
	IsValidCap   func(cap SubspaceCapability) bool
	Signatures   types.SignatureScheme[SubspaceReceiver, SubspaceSecretKey, SyncSubspaceSignature]
	Encodings    struct {
		SubspaceCapability    utils.EncodingScheme[SubspaceCapability]
		SyncSubspaceSignature utils.EncodingScheme[SyncSubspaceSignature]
	}
}

type AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken string] struct {
	RecomposeAuthToken func(staticToken StaticToken, dynamicToken DynamicToken) AuthorisationToken
	DecomposeAuthToken func(authToken AuthorisationToken) (StaticToken, DynamicToken)
	Encodings          struct {
		StaticToken  utils.EncodingScheme[StaticToken]
		DynamicToken utils.EncodingScheme[DynamicToken]
	}
}
