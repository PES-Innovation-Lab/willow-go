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
	IntersectionHandle = iota
	CapabilityHandle
	AreaOfInterestHandle
	PayloadRequestHandle
	StaticTokenHandle
)

type LogicalChannel int

const (
	ReconciliationChannel = iota
	DataChannel
	InteresectionChannel
	CapabilityChannel
	AreaOfInterestChannel
	PayloadRequestChannel
	StaticTokenChannel
)

type MsgKind int

const (
	CommitmentReveal = iota
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
	T    ControlIssueGuaranteeData
}

/** Allow the other peer to reduce its total buffer capacity by amount. */
type ControlAbsolveData struct {
	Amount  uint64
	Channel LogicalChannel
}
type MsgControlAbsolve struct {
	Kind MsgKind
	T    ControlAbsolveData
}

/** Ask the other peer to send an ControlAbsolve message such that the receiver remaining guarantees will be target. */
type ControlPleadData struct {
	Target  uint64
	Channel LogicalChannel
}
type MsgControlPlead struct {
	Kind MsgKind
	T    ControlPleadData
}

type ControlAnnounceDroppingData struct {
	Channel LogicalChannel
}
type MsgControlAnnounceDropping struct {
	Kind MsgKind
	T    ControlAnnounceDroppingData
}

/** Notify the other peer that it can stop dropping messages of this logical channel. */
type ControlApologiseData struct {
	Channel LogicalChannel
}
type MsgControlApologise struct {
	Kind MsgKind
	T    ControlApologiseData
}

type MsgControlFreeData struct {
	Handle uint64
	/** Indicates whether the peer sending this message is the one who created the handle (true) or not (false). */
	Mine       bool
	HandleType HandleType
}
type MsgControlFree struct {
	Kind MsgKind
	T    MsgControlFreeData
}

/** Complete the commitment scheme to determine the challenge for read authentication. */
type MsgCommitmentRevealData struct {
	Nonce []byte
}
type MsgCommitmentReveal struct {
	Kind MsgKind
	T    MsgCommitmentRevealData
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
type MsgSetupBindAreaOfInterestData struct {
	AreaOfInterest types.AreaOfInterest
	Authorisation  uint64
}
type MsgSetupBindAreaOfinterest struct {
	Kind MsgKind
	T    MsgSetupBindAreaOfInterestData
}

type MsgSetupBindStaticTokenData[StaticToken any] struct {
	StaticToken StaticToken
}
type MsgSetupBindStaticToken[StaicToken any] struct {
	Kind MsgKind
	T    MsgSetupBindStaticTokenData[StaicToken]
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
	T    MsgReconciliationSendFingerprintData[Fingerprint]
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
	T    MsgReconciliationAnnounceEntriesData
}

/** Transmit a LengthyEntry as part of 3d range-based set reconciliation. */
type MsgReconciliationSendEntryData[PayloadDigest, DynamicToken constraints.Ordered] struct {
	Entry             datamodeltypes.LengthyEntry[PayloadDigest]
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
}
type MsgReconciliationSendEntry[PayloadDigest, DynamicToken constraints.Ordered] struct {
	Kind MsgKind
	T    MsgReconciliationSendEntryData[PayloadDigest, DynamicToken]
}

/** Transmit a Payload as part of 3d range-based set reconciliation. */
type MsgReconciliationSendPayloadData struct {
	Amount uint64
	Bytes  []byte
}
type MsgReconciliationSendPayload struct {
	Kind MsgKind
	T    MsgReconciliationSendPayloadData
}

/** Notify the other peer that the payload transmission is complete. */
type MsgReconciliationTerminatePayload struct {
	Kind MsgKind
}

// 4. Data messages

/** Transmit an AuthorisedEntry to the other peer, and optionally prepare transmission of its Payload. */
type MsgDataSendEntryData[PayloadDigest, DynamicToken constraints.Ordered] struct {
	Entry             types.Entry[PayloadDigest]
	StaticTokenHandle uint64
	DynamicToken      DynamicToken
	Offset            uint64
}
type MsgDataSendEntry[PayloadDigest, DynamicToken constraints.Ordered] struct {
	Kind MsgKind
	T    MsgDataSendEntryData[PayloadDigest, DynamicToken]
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
type MsgDataBindPayloadRequestData[PayloadDigest constraints.Ordered] struct {
	Entry      types.Entry[PayloadDigest]
	Offset     uint64
	Capability uint64
}
type MsgDataBindPayloadRequest[PayloadDigest constraints.Ordered] struct {
	Kind MsgKind
	T    MsgDataBindPayloadRequestData[PayloadDigest]
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
func (MsgSetupBindAreaOfinterest) isSyncMessage()                                               {}
func (MsgSetupBindStaticToken[StaticToken]) isSyncMessage()                                     {}
func (MsgReconciliationSendFingerprint[Fingerprint]) isSyncMessage()                            {}
func (MsgReconciliationAnnounceEntries) isSyncMessage()                                         {}
func (MsgReconciliationSendEntry[PayloadLength, DynamicToken]) isSyncMessage() {
}
func (MsgReconciliationSendPayload) isSyncMessage()             {}
func (MsgReconciliationTerminatePayload) isSyncMessage()        {}
func (MsgDataSendPayload) isSyncMessage()                       {}
func (MsgDataSetMetadata) isSyncMessage()                       {}
func (MsgDataBindPayloadRequest[PayloadDigest]) isSyncMessage() {}
func (MsgDataReplyPayload) isSyncMessage()                      {}

// Messages categorised by logical channel
type ReconciliationChannelMsg interface {
	isReconciliationChannelMsg()
}

func (MsgReconciliationSendFingerprint[Fingerprint]) isReconciliationChannelMsg() {}
func (MsgReconciliationAnnounceEntries) isReconciliationChannelMsg()              {}
func (MsgReconciliationSendEntry[PayloadLength, DynamicToken]) isReconciliationChannelMsg() {
}
func (MsgReconciliationSendPayload) isReconciliationChannelMsg()      {}
func (MsgReconciliationTerminatePayload) isReconciliationChannelMsg() {}

type DataChannelMsg interface {
	isDataChannelMsg()
}

func (MsgDataSendEntry[PayloadDigest, DynamicToken]) isDataChannelMsg() {}
func (MsgDataReplyPayload) isDataChannelMsg()                           {}
func (MsgDataSendPayload) isDataChannelMsg()                            {}

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

func (MsgSetupBindAreaOfinterest) isAreaOfInterestChannelMsg() {}

type PayloadRequestChannelMsg interface {
	isPayloadRequestChannelMsg()
}

func (MsgDataBindPayloadRequest[PayloadDigest]) isPayloadRequestChannelMsg() {
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
type ReadCapEncodingScheme[ReadCapability constraints.Ordered] struct {
	PrivyEncodingScheme[ReadCapability, ReadCapPrivy]
}

type ReconciliationPrivy[PayloadDigest constraints.Ordered] struct {
	PrevSenderHandle      uint64
	PrevReceiverHandle    uint64
	prevRange             types.Range3d
	PrevStaticTokenHandle uint64
	PrevEntry             types.Entry[PayloadDigest]
	Announced             struct {
		Range     types.Range3d
		Namespace types.NamespaceId
	}
}

/** The parameter schemes required to instantiate a `WgpsMessenger`. */
type SyncSchemes[ReadCapability, Receiver, SyncSignature, PsiGroup, PsiScalar, SubspaceCapability, SubspaceReceiver, AuthorisationOpts, PayloadDigest, ReceiverSecretKey, Prefingerprint, Fingerprint constraints.Ordered, K constraints.Unsigned, AuthorisationToken, StaticToken, DynamicToken, SyncSubspaceSignature, SubspaceSecretKey types.OrderableGeneric] struct {
	AccessControl      AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey, K]
	SubspaceCap        SubspaceCapScheme[SubspaceCapability, SubspaceReceiver, SyncSubspaceSignature, SubspaceSecretKey, K]
	Pai                PaiScheme[ReadCapability, PsiGroup, PsiScalar, K]
	Namespace          datamodeltypes.NamespaceScheme[K]
	Subspace           datamodeltypes.SubspaceScheme[K]
	Path               types.PathParams[K]
	AuhtorisationToken AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken, K]
	Payload            datamodeltypes.PayloadScheme[PayloadDigest, K]
	Fingerprint        datamodeltypes.FingerprintScheme[PayloadDigest, Prefingerprint, Fingerprint, K]
}

type AccessControlScheme[SyncSignature, ReadCapability, Receiver, ReceiverSecretKey constraints.Ordered, K constraints.Unsigned] struct {
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

type SubspaceCapScheme[SubspaceReceiver, SubspaceSecretKey constraints.Ordered, SubspaceCapability, SyncSubspaceSignature types.OrderableGeneric, K constraints.Unsigned] struct {
	GetSecretKey func(receiver SubspaceReceiver) SubspaceSecretKey
	GetNamespace func(cap SubspaceCapability) types.NamespaceId
	GetReceiver  func(cap SubspaceCapability) SubspaceReceiver
	IsValidCap   func(cap SubspaceCapability) bool
	Signatures   types.SignatureScheme[SubspaceReceiver, SubspaceSecretKey, SyncSubspaceSignature]
	Encodings    struct {
		SubspaceCapability    utils.EncodingScheme[SyncSubspaceSignature]
		SyncSubspaceSignature utils.EncodingScheme[SyncSubspaceSignature]
	}
}

type AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken types.OrderableGeneric, K constraints.Unsigned] struct {
	RecomposeAuthToken func(staticToken StaticToken, dynamicToken DynamicToken) AuthorisationToken
	DecomposeAuthToken func(authToken AuthorisationToken) (StaticToken, DynamicToken)
	Encodings          struct {
		StaticToken  utils.EncodingScheme[StaticToken]
		DynamicToken utils.EncodingScheme[DynamicToken]
	}
}
