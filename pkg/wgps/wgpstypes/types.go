package wgpstypes

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
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

type ReadAuthorisation[ReadCapability, SubspaceReadCapability constraints.Ordered] struct {
	Capability ReadCapability
	// SubspaceCapability is optional here
	SubspaceCapability    SubspaceReadCapability
	HasSubspaceCapability bool
}

//will need to check if the type is any or something else

// Transport defines the interface for communication channels
type Transport interface {
	Role() SyncRole
	Send(data []byte) error     // Use byte slice instead of Uint8Array
	Recv() (chan []byte, error) // Returns a receive channel and potential error (PLEASE CHECK IF THIS IS RIGHT)
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

type LogicalChannel int

const (
	/* Logical channel for performing 3d range-based set reconciliation. */
	ReconciliationChannel LogicalChannel = iota
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
	Channel LogicalChannel
}
type MsgControlIssueGuarantee struct {
	Kind MsgKind
	Data ControlIssueGuaranteeData
}

/** Allow the other peer to reduce its total buffer capacity by amount. */
type ControlAbsolveData struct {
	Amount  uint64
	Channel LogicalChannel
}
type MsgControlAbsolve struct {
	Kind MsgKind
	Data ControlAbsolveData
}

/** Ask the other peer to send an ControlAbsolve message such that the receiver remaining guarantees will be target. */
type ControlPleadData struct {
	Target  uint64
	Channel LogicalChannel
}
type MsgControlPlead struct {
	Kind MsgKind
	Data ControlPleadData
}

type ControlAnnounceDroppingData struct {
	Channel LogicalChannel
}
type MsgControlAnnounceDropping struct {
	Kind MsgKind
	Data ControlAnnounceDroppingData
}

/** Notify the other peer that it can stop dropping messages of this logical channel. */
type ControlApologiseData struct {
	Channel LogicalChannel
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
	T    MsgPaiBindFragmentData[PsiGroup]
}

/** Finalise private set intersection for a single item. */
type MsgPaiReplyFragmentData[PsiGroup any] struct {
	Handle      uint64
	GroupMember PsiGroup
}
type MsgPaiReplyFragment[PsiGroup any] struct {
	Kind MsgKind
	T    MsgPaiReplyFragmentData[PsiGroup]
}

/** Request the subspace capability for a given IntersectionHandle (for the least-specific secondary fragment for whose NamespaceId the request is being made). */
type MsgPaiRequestSubspaceCapabilityData struct {
	Handle uint64
}
type MsgPaiRequestSubspaceCapability struct {
	Kind MsgKind
	T    MsgPaiRequestSubspaceCapabilityData
}

/** Send a previously requested SubspaceCapability. */
type MsgPaiReplySubspaceCapabilityData[SubspaceCapability, SyncSubspaceSignature constraints.Ordered] struct {
	Handle     uint64
	Capability SubspaceCapability
	Signature  SyncSubspaceSignature
}
type MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature constraints.Ordered] struct {
	Kind MsgKind
	T    MsgPaiReplySubspaceCapabilityData[SubspaceCapability, SyncSubspaceSignature]
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
	T    MsgSetupBindReadCapabilityData[ReadCapability, SyncSignature]
}

/** Bind an AreaOfInterest to an AreaOfInterestHandle. */
type MsgSetupBindAreaOfInterestData[SubspaceId constraints.Ordered] struct {
	AreaOfInterest types.AreaOfInterest[SubspaceId]
	Authorisation  uint64
}
type MsgSetupBindAreaOfinterest[SubspaceId constraints.Ordered] struct {
	Kind MsgKind
	T    MsgSetupBindAreaOfInterestData[SubspaceId]
}

type MsgSetupBindStaticTokenData[StaticToken any] struct {
	StaticToken StaticToken
}
type MsgSetupBindStaticToken[StaicToken any] struct {
	Kind MsgKind
	T    MsgSetupBindStaticTokenData[StaicToken]
}

/** Send a Fingerprint as part of 3d range-based set reconciliation. */
type MsgReconciliationSendFingerprintData[SubspaceId constraints.Ordered, Fingerprint any] struct {
	Range          types.Range3d[SubspaceId]
	Fingerprint    Fingerprint
	SenderHandle   uint64
	ReceiverHandle uint64
	Covers         uint64
	DoesCover      bool
}
type MsgReconciliationSendFingerprint[SubspaceId constraints.Ordered, Fingerprint any] struct {
	Kind MsgKind
	T    MsgReconciliationSendFingerprintData[SubspaceId, Fingerprint]
}

/** Prepare transmission of the LengthyEntries a peer has in a 3dRange as part of 3d range-based set reconciliation. */
type MsgReconciliationAnnounceEntriesData[SubspaceId constraints.Ordered] struct {
	Range          types.Range3d[SubspaceId]
	Count          uint64
	WantResponse   bool
	WillSort       bool
	SenderHandle   uint64
	ReceiverHandle uint64
	Covers         uint64
	DoesCover      bool
}
type MsgReconciliationAnnounceEntries[SubspaceId constraints.Ordered] struct {
	Kind MsgKind
	T    MsgReconciliationAnnounceEntriesData[SubspaceId]
}

/** Transmit a LengthyEntry as part of 3d range-based set reconciliation. */
type MsgReconciliationSendEntryData[SubspaceId, NamespaceId, PayloadLength constraints.Ordered, DynamicToken any] struct {
	Entry             datamodeltypes.LengthyEntry[SubspaceId, NamespaceId, PayloadLength]
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
}
type MsgReconciliationSendEntry[SubspaceId, NamespaceId, PayloadLength constraints.Ordered, DynamicToken any] struct {
	Kind MsgKind
	T    MsgReconciliationSendEntryData[SubspaceId, NamespaceId, PayloadLength, DynamicToken]
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
type MsgDataSendEntryData[SubspaceId, NamespaceId, PayloadDigest constraints.Ordered, DynamicToken any] struct {
	Entry             types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
	Offset            uint64
}
type MsgDataSendEntry[SubspaceId, NamespaceId, PayloadDigest constraints.Ordered, DynamicToken any] struct {
	Kind MsgKind
	Data MsgDataSendEntryData[SubspaceId, NamespaceId, PayloadDigest, DynamicToken]
}

/** Transmit a Payload to the other peer. */
type MsgDataSendPayloadData struct {
	Amount uint64
	Bytes  []byte
}
type MsgDataSendPayload struct {
	Kind MsgKind
	T    MsgDataSendPayloadData
}

/** Express a preference whether the other peer should eagerly forward Payloads in the intersection of two AreaOfInterests. */
type MsgDataSetMetadataData struct {
	IsEager        bool
	SenderHandle   uint64
	ReceiverHandle uint64
}
type MsgDataSetMetadata struct {
	Kind MsgKind
	T    MsgDataSetMetadataData
}

/** Bind a PayloadRequest to a PayloadRequestHandle. */
type MsgDataBindPayloadRequestData[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	Entry      types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	Offset     uint64
	Capability uint64
}
type MsgDataBindPayloadRequest[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	Kind MsgKind
	T    MsgDataBindPayloadRequestData[NamespaceId, SubspaceId, PayloadDigest]
}

/** Transmit a Payload to the other peer. */
type MsgDataReplyPayloadData struct {
	Handle uint64
}
type MsgDataReplyPayload struct {
	Kind MsgKind
	T    MsgDataReplyPayloadData
}

type SyncMessage interface {
	isSyncMessage()
}

func (MsgControlIssueGuarantee) isSyncMessage()                                                 {}
func (MsgControlAbsolve) isSyncMessage()                                                        {}
func (MsgControlPlead) isSyncMessage()                                                          {}
func (MsgControlAnnounceDropping) isSyncMessage()                                               {}
func (MsgControlApologise) isSyncMessage()                                                      {}
func (MsgControlFree) isSyncMessage()                                                           {}
func (MsgCommitmentReveal) isSyncMessage()                                                      {}
func (MsgPaiBindFragment[PsiGroup]) isSyncMessage()                                             {}
func (MsgPaiReplyFragment[PsiGroup]) isSyncMessage()                                            {}
func (MsgPaiRequestSubspaceCapability) isSyncMessage()                                          {}
func (MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]) isSyncMessage() {}
func (MsgSetupBindReadCapability[ReadCapability, SyncSignature]) isSyncMessage()                {}
func (MsgSetupBindAreaOfinterest[SubspaceId]) isSyncMessage()                                   {}
func (MsgSetupBindStaticToken[StaticToken]) isSyncMessage()                                     {}
func (MsgReconciliationSendFingerprint[SubspaceId, Fingerprint]) isSyncMessage()                {}
func (MsgReconciliationAnnounceEntries[SubspaceId]) isSyncMessage()                             {}
func (MsgReconciliationSendEntry[SubspaceId, NamespaceId, PayloadLength, DynamicToken]) isSyncMessage() {
}
func (MsgReconciliationSendPayload) isSyncMessage()                                      {}
func (MsgReconciliationTerminatePayload) isSyncMessage()                                 {}
func (MsgDataSendPayload) isSyncMessage()                                                {}
func (MsgDataSetMetadata) isSyncMessage()                                                {}
func (MsgDataBindPayloadRequest[NamespaceId, SubspaceId, PayloadDigest]) isSyncMessage() {}
func (MsgDataReplyPayload) isSyncMessage()                                               {}

// Messages categorised by logical channel
type ReconciliationChannelMsg interface {
	isReconciliationChannelMsg()
}

func (MsgReconciliationSendFingerprint[SubspaceId, Fingerprint]) isReconciliationChannelMsg() {}
func (MsgReconciliationAnnounceEntries[SubspaceId]) isReconciliationChannelMsg()              {}
func (MsgReconciliationSendEntry[SubspaceId, NamespaceId, PayloadLength, DynamicToken]) isReconciliationChannelMsg() {
}
func (MsgReconciliationSendPayload) isReconciliationChannelMsg()      {}
func (MsgReconciliationTerminatePayload) isReconciliationChannelMsg() {}

type DataChannelMsg interface {
	isDataChannelMsg()
}

func (MsgDataSendEntry[SubspaceId, NamespaceId, PayloadDigest, DynamicToken]) isDataChannelMsg() {}
func (MsgDataReplyPayload) isDataChannelMsg()                                                    {}
func (MsgDataSendPayload) isDataChannelMsg()                                                     {}

type IntersectionChannelMsg interface {
	isIntersectionChannelMsg()
}

func (MsgPaiBindFragment[PsiGroup]) isIntersectionChannelMsg() {}

type CapabilityChannelMsg interface {
	isCapabilityChannelMsg()
}

func (MsgSetupBindReadCapability[ReadCapability, SyncSignature]) isCapabilityChannelMsg() {}

type AreaOfInterestChannelMsg interface {
	isAreaOfInterestChannelMsg()
}

func (MsgSetupBindAreaOfinterest[SubspaceId]) isAreaOfInterestChannelMsg() {}

type PayloadRequestChannelMsg interface {
	isPayloadRequestChannelMsg()
}

func (MsgDataBindPayloadRequest[NamespaceId, SubspaceId, PayloadDigest]) isPayloadRequestChannelMsg() {
}

type StaticTokenChannelMsg interface {
	isStaticTokenChannelMsg()
}

func (MsgSetupBindStaticToken[StaticToken]) isStaticTokenChannelMsg() {}

/** Messages which belong to no logical channel. */
type NoChannelMsg interface {
	isNoChannelMsg()
}

func (MsgControlIssueGuarantee) isNoChannelMsg()                                                 {}
func (MsgControlAbsolve) isNoChannelMsg()                                                        {}
func (MsgControlPlead) isNoChannelMsg()                                                          {}
func (MsgControlAnnounceDropping) isNoChannelMsg()                                               {}
func (MsgControlApologise) isNoChannelMsg()                                                      {}
func (MsgControlFree) isNoChannelMsg()                                                           {}
func (MsgCommitmentReveal) isNoChannelMsg()                                                      {}
func (MsgPaiReplyFragment[PsiGroup]) isNoChannelMsg()                                            {}
func (MsgPaiRequestSubspaceCapability) isNoChannelMsg()                                          {}
func (MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]) isNoChannelMsg() {}
func (MsgDataSetMetadata) isNoChannelMsg()                                                       {}

// Encodings

type ReadCapPrivy[NamespaceId, SubspaceId constraints.Ordered] struct {
	Outer     types.Area[SubspaceId]
	Namespace NamespaceId
}

// Define the PrivyEncodingScheme type with generics
type PrivyEncodingScheme[ReadCapability any, Privy any] struct {
	ReadCapability ReadCapability
	Privy          Privy
}

// Define the ReadCapEncodingScheme type alias
type ReadCapEncodingScheme[ReadCapability, NamespaceId, SubspaceId constraints.Ordered] struct {
	PrivyEncodingScheme[ReadCapability, ReadCapPrivy[NamespaceId, SubspaceId]]
}

type ReconciliationPrivy[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	PrevSenderHandle      uint64
	PrevReceiverHandle    uint64
	PrevRange             types.Range3d[SubspaceId]
	PrevStaticTokenHandle uint64
	PrevEntry             types.Entry[NamespaceId, SubspaceId, PayloadDigest]
	Announced             struct {
		Range     types.Range3d[SubspaceId]
		Namespace NamespaceId
	}
}

/** The parameter schemes required to instantiate a `WgpsMessenger`. */
type SyncSchemes[ReadCapability, Receiver, SyncSignature, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, AuthorisationOpts, NamespaceId, SubspaceId, PayloadDigest, ReceiverSecretKey, Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey types.OrderableGeneric] struct {
	AccessControl      AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey, NamespaceId, SubspaceId, K]
	SubspaceCap        SubspaceCapScheme[SubspaceCapability, SubspaceReceiver, NamespaceId, SyncSubspaceSignature, SubspaceSecretKey, K]
	Pai                PaiScheme[ReadCapability, PsiScalar, NamespaceId, SubspaceId, K, PsiGroup]
	Namespace          datamodeltypes.NamespaceScheme[NamespaceId, K]
	Subspace           datamodeltypes.SubspaceScheme[NamespaceId, K]
	Path               types.PathParams[K]
	AuhtorisationToken AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken, K]
	Payload            datamodeltypes.PayloadScheme[PayloadDigest, K]
	Fingerprint        datamodeltypes.FingerprintScheme[NamespaceId, SubspaceId, PayloadDigest, Prefingerprint, Fingerprint, K]
}

type AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey, NamespaceId, SubspaceId constraints.Ordered, K constraints.Unsigned] struct {
	GetReceiver         func(cap ReadCapability) Receiver
	GetSecretKey        func(receiver Receiver) ReceiverSecretKey
	GetGrantedArea      func(cap ReadCapability) types.Area[SubspaceId]
	GetGrantedNamespace func(cap ReadCapability) NamespaceId
	Signatures          types.SignatureScheme[Receiver, ReceiverSecretKey, SyncSignature]
	IsValidCap          func(cap ReadCapability) bool
	Encodings           struct {
		ReadCap       ReadCapEncodingScheme[ReadCapability, NamespaceId, SubspaceId]
		SyncSignature utils.EncodingScheme[SyncSignature, K]
	}
}

type SubspaceCapScheme[SubspaceReceiver, SubspaceSecretKey, NamespaceId constraints.Ordered, SubspaceCapability, SyncSubspaceSignature types.OrderableGeneric, K constraints.Unsigned] struct {
	GetSecretKey func(receiver SubspaceReceiver) SubspaceSecretKey
	GetNamespace func(cap SubspaceCapability) NamespaceId
	GetReceiver  func(cap SubspaceCapability) SubspaceReceiver
	IsValidCap   func(cap SubspaceCapability) bool
	Signatures   types.SignatureScheme[SubspaceReceiver, SubspaceSecretKey, SyncSubspaceSignature]
	Encodings    struct {
		SubspaceCapability    utils.EncodingScheme[SubspaceCapability, K]
		SyncSubspaceSignature utils.EncodingScheme[SyncSubspaceSignature, K]
	}
}

type AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken types.OrderableGeneric, K constraints.Unsigned] struct {
	RecomposeAuthToken func(staticToken StaticToken, dynamicToken DynamicToken) AuthorisationToken
	DecomposeAuthToken func(authToken AuthorisationToken) (StaticToken, DynamicToken)
	Encodings          struct {
		StaticToken  utils.EncodingScheme[StaticToken, K]
		DynamicToken utils.EncodingScheme[DynamicToken, K]
	}
}
