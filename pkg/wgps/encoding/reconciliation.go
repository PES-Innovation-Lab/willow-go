package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func EncodeReconciliationSendFingerprint[Fingerprint string, ValueType constraints.Unsigned](msg wgpstypes.MsgReconciliationSendFingerprint[Fingerprint], opts struct {
	OrderSubspace        types.TotalOrder[types.SubspaceId]
	EncodeSubspaceId     func(subspace types.SubspaceId) []byte
	PathScheme           types.PathParams[ValueType]
	IsFingerprintNeutral func(fingerprint Fingerprint) bool
	EncodeFingerprint    func(fingerprint Fingerprint) []byte
	Privy                wgpstypes.ReconciliationPrivy
}) []byte {
	MessageTypeMask := byte(0x40)

	var NeutralMask byte
	if opts.IsFingerprintNeutral(msg.Data.Fingerprint) {
		NeutralMask = 0x8
	} else {
		NeutralMask = 0x0
	}

	COVERS_NONE := 1

	EncodedRelativeToPrevRange := byte(0x4)

	SenderHandleIsSame := msg.Data.SenderHandle == opts.Privy.PrevSenderHandle
	ReceiverHandleIsSame := msg.Data.ReceiverHandle == opts.Privy.PrevReceiverHandle

	var UsingPrevSenderhandleMask byte
	if SenderHandleIsSame {
		UsingPrevSenderhandleMask = 0x2
	} else {
		UsingPrevSenderhandleMask = 0x0
	}

	var UsingPrevReceiverHandleMask byte
	if ReceiverHandleIsSame {
		UsingPrevReceiverHandleMask = 0x1
	} else {
		UsingPrevReceiverHandleMask = 0x0
	}

	var HeaderByte byte = MessageTypeMask | NeutralMask |
		EncodedRelativeToPrevRange | UsingPrevSenderhandleMask |
		UsingPrevReceiverHandleMask

	HandleLengthNumber := byte(0x0)

	CompactWidthSender := utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.SenderHandle))
	CompactWidthReceiver := utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.ReceiverHandle))

	if !SenderHandleIsSame {
		Unshifted := CompactWidthOr(0x0, CompactWidthReceiver)
		Shifted := byte(Unshifted << 6)
		HandleLengthNumber = HandleLengthNumber | Shifted
	}

	if !ReceiverHandleIsSame {
		Unshifted := CompactWidthOr(0x0, CompactWidthSender)
		Shifted := byte(Unshifted << 4)
		HandleLengthNumber = HandleLengthNumber | Shifted
	}

	if msg.Data.Covers != uint64(COVERS_NONE) {
		HandleLengthNumber = HandleLengthNumber | 0x8
		HandleLengthNumber = byte(CompactWidthOr(int(HandleLengthNumber), utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Covers))))
	}

	HandleLengthByte := []byte{HandleLengthNumber}

	var EncodedCovers []byte
	if msg.Data.Covers == uint64(COVERS_NONE) {
		EncodedCovers = []byte{}
	} else {
		EncodedCovers = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Covers))
	}

	var EncodedSenderhandle []byte
	if SenderHandleIsSame {
		EncodedSenderhandle = []byte{}
	} else {
		EncodedSenderhandle = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.SenderHandle))
	}

	var EncodedReceiverHandle []byte
	if ReceiverHandleIsSame {
		EncodedReceiverHandle = []byte{}
	} else {
		EncodedReceiverHandle = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.ReceiverHandle))
	}

	var EncodedFingerprint []byte
	if opts.IsFingerprintNeutral(msg.Data.Fingerprint) {
		EncodedFingerprint = []byte{}
	} else {
		EncodedFingerprint = opts.EncodeFingerprint(msg.Data.Fingerprint)
	}

	EncodedRelativeRange := utils.EncodeRange3dRelative[ValueType](struct {
		OrderSubspace    types.TotalOrder[types.SubspaceId]
		EncodeSubspaceId func(subspace types.SubspaceId) []byte
		PathScheme       types.PathParams[ValueType]
	}{
		OrderSubspace:    opts.OrderSubspace,
		EncodeSubspaceId: opts.EncodeSubspaceId,
		PathScheme:       opts.PathScheme,
	}, msg.Data.Range, opts.Privy.PrevRange)

	var Result []byte
	Result = append(Result, MessageTypeMask|HeaderByte)
	Result = append(Result, HandleLengthByte...)
	Result = append(Result, EncodedCovers...)
	Result = append(Result, EncodedSenderhandle...)
	Result = append(Result, EncodedReceiverHandle...)
	Result = append(Result, EncodedFingerprint...)
	Result = append(Result, EncodedRelativeRange...)

	return Result
}

func EncodeReconciliationAnnounceEntries[ValueType constraints.Unsigned](msg wgpstypes.MsgReconciliationAnnounceEntries, opts struct {
	Privy            wgpstypes.ReconciliationPrivy
	OrderSubspace    types.TotalOrder[types.SubspaceId]
	EncodeSubspaceId func(subspace types.SubspaceId) []byte
	PathScheme       types.PathParams[ValueType]
}) []byte {

	COVERS_NONE := 1

	MessageTyoeMask := byte(0x50)

	var WantResponseBit byte
	if msg.Data.WantResponse {
		WantResponseBit = 0x8
	} else {
		WantResponseBit = 0x0
	}

	EncodedRelativebit := byte(0x4)

	SenderHandleIsSame := msg.Data.SenderHandle == opts.Privy.PrevSenderHandle
	ReceiverHandleIsSame := msg.Data.ReceiverHandle == opts.Privy.PrevReceiverHandle

	var UsingPrevSenderHandleMask byte
	if SenderHandleIsSame {
		UsingPrevSenderHandleMask = 0x2
	} else {
		UsingPrevSenderHandleMask = 0x0
	}

	var UsingPrevReceiverHandleMask byte
	if ReceiverHandleIsSame {
		UsingPrevReceiverHandleMask = 0x1
	} else {
		UsingPrevReceiverHandleMask = 0x0
	}

	FirstByte := MessageTyoeMask | WantResponseBit | EncodedRelativebit | UsingPrevSenderHandleMask | UsingPrevReceiverHandleMask

	CompactWidthSender := byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.SenderHandle)))
	CompactWidthReceiver := byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.ReceiverHandle)))

	var SenderReceiverWidthFlags byte

	if !SenderHandleIsSame && !ReceiverHandleIsSame {
		Unshifted := byte(CompactWidthOr(0x0, int(CompactWidthSender)))
		Shifted2 := byte(Unshifted << 2)
		Unshifted2 := byte(CompactWidthOr(int(Shifted2), int(CompactWidthReceiver)))
		SenderReceiverWidthFlags = byte(Unshifted2 << 4)
	} else if !SenderHandleIsSame && ReceiverHandleIsSame {
		Unshifted := byte(CompactWidthOr(0x0, int(CompactWidthSender)))
		SenderReceiverWidthFlags = byte(Unshifted << 6)
	} else if SenderHandleIsSame && !ReceiverHandleIsSame {
		Unshifted := byte(CompactWidthOr(0x0, int(CompactWidthReceiver)))
		SenderReceiverWidthFlags = byte(Unshifted << 4)
	} else {
		SenderReceiverWidthFlags = 0x0
	}

	CountCompactWidth := byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Count)))

	CountCompactWidthFlags := byte(CompactWidthOr(0, int(CountCompactWidth))) << 2

	var WillSortFlag byte
	if msg.Data.WillSort {
		WillSortFlag = 0x2
	} else {
		WillSortFlag = 0x0
	}

	var CoversNotNone byte

	if msg.Data.Covers != uint64(COVERS_NONE) {
		CoversNotNone = 0x1
	} else {
		CoversNotNone = 0x0
	}

	SecondByte := SenderReceiverWidthFlags | CountCompactWidthFlags | WillSortFlag | CoversNotNone

	var CoversCompactWidth []byte
	var CoversEncoded []byte

	if msg.Data.Covers == uint64(COVERS_NONE) {
		CoversCompactWidth = []byte{}
		CoversEncoded = []byte{}
	} else if msg.Data.Covers >= 252 {
		CoversCompactWidth = []byte{byte(CompactWidthOr(0xfc, utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Covers))))}
		CoversEncoded = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Covers))
	} else {
		CoversCompactWidth = []byte{byte(msg.Data.Covers)}
		CoversEncoded = []byte{}
	}

	var EncodedSenderHandle []byte
	if !SenderHandleIsSame {
		EncodedSenderHandle = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.SenderHandle))
	} else {
		EncodedSenderHandle = []byte{}
	}

	var EncodedReceiverHandle []byte
	if !ReceiverHandleIsSame {
		EncodedReceiverHandle = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.ReceiverHandle))
	} else {
		EncodedReceiverHandle = []byte{}
	}

	EncodedCount := utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Count))

	EncodedRelativeRange := utils.EncodeRange3dRelative[ValueType](struct {
		OrderSubspace    types.TotalOrder[types.SubspaceId]
		EncodeSubspaceId func(subspace types.SubspaceId) []byte
		PathScheme       types.PathParams[ValueType]
	}{
		OrderSubspace:    opts.OrderSubspace,
		EncodeSubspaceId: opts.EncodeSubspaceId,
		PathScheme:       opts.PathScheme,
	}, msg.Data.Range, opts.Privy.PrevRange)

	var Result []byte
	Result = append(Result, FirstByte)
	Result = append(Result, SecondByte)
	Result = append(Result, CoversCompactWidth...)
	Result = append(Result, CoversEncoded...)
	Result = append(Result, EncodedSenderHandle...)
	Result = append(Result, EncodedReceiverHandle...)
	Result = append(Result, EncodedCount...)
	Result = append(Result, EncodedRelativeRange...)

	return Result
}

func EncodeReconciliationSendEntry[DynamicToken string, ValueType constraints.Unsigned](
	msg wgpstypes.MsgReconciliationSendEntry[DynamicToken],
	opts struct {
		Privy               wgpstypes.ReconciliationPrivy
		IsEqualNamespace    func(a, b types.NamespaceId) bool
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		EncodeNamespaceId   func(namespace types.NamespaceId) []byte
		EncodeSubspaceId    func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(digest types.PayloadDigest) []byte
		EncodeDynamicToken  func(token DynamicToken) []byte
		PathScheme          types.PathParams[ValueType]
	},
) []byte {

	MessageTypeMask := byte(0x50)

	IsPrevTokenEqual := msg.Data.StaticTokenHandle == opts.Privy.PrevStaticTokenHandle

	var IsPrevStaticTokenFlag byte
	if IsPrevTokenEqual {
		IsPrevStaticTokenFlag = 0x8
	} else {
		IsPrevStaticTokenFlag = 0x0
	}

	IsEncodedRelativeToPrevEntryFlag := byte(0x4)

	CompactWidthAvailableFlag := byte(CompactWidthOr(0, utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Entry.Available))))

	Header := MessageTypeMask | IsPrevStaticTokenFlag | IsEncodedRelativeToPrevEntryFlag | CompactWidthAvailableFlag

	var EncodedStaticTokenWidth []byte

	CompactWidthStaticToken := utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.StaticTokenHandle))

	if IsPrevTokenEqual {
		EncodedStaticTokenWidth = []byte{}
	} else if msg.Data.StaticTokenHandle < uint64(63) {
		EncodedStaticTokenWidth = []byte{byte(msg.Data.StaticTokenHandle)}
	} else if CompactWidthStaticToken == 1 {
		EncodedStaticTokenWidth = []byte{0x3f}
	} else if CompactWidthStaticToken == 2 {
		EncodedStaticTokenWidth = []byte{0x7f}
	} else if CompactWidthStaticToken == 4 {
		EncodedStaticTokenWidth = []byte{0xbf}
	} else {
		EncodedStaticTokenWidth = []byte{0xff}
	}

	var EncodedStaticToken []byte

	if !IsPrevTokenEqual && msg.Data.StaticTokenHandle > uint64(63) {
		EncodedStaticToken = utils.EncodeIntMax64[ValueType](ValueType(msg.Data.StaticTokenHandle))
	} else {
		EncodedStaticToken = []byte{}
	}

	EncodedAvailable := utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Entry.Available))

	EncodedDynamicToken := opts.EncodeDynamicToken(msg.Data.DynamicToken)

	EncodedRelativeEntry := utils.EncodeEntryRelativeEntry[ValueType](struct {
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(digest types.PayloadDigest) []byte
		IsEqualNamespace    func(a types.NamespaceId, b types.NamespaceId) bool
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		PathScheme          types.PathParams[ValueType]
	}{
		EncodeNamespace:     opts.EncodeNamespaceId,
		EncodeSubspace:      opts.EncodeSubspaceId,
		EncodePayloadDigest: opts.EncodePayloadDigest,
		IsEqualNamespace:    opts.IsEqualNamespace,
		OrderSubspace:       opts.OrderSubspace,
		PathScheme:          opts.PathScheme,
	}, msg.Data.Entry.Entry, opts.Privy.PrevEntry)

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, EncodedStaticTokenWidth...)
	Result = append(Result, EncodedStaticToken...)
	Result = append(Result, EncodedAvailable...)
	Result = append(Result, EncodedDynamicToken...)
	Result = append(Result, EncodedRelativeEntry...)

	return Result

}

func EncodeReconciliationSendPayload[ValueType constraints.Unsigned](msg wgpstypes.MsgReconciliationSendPayload) []byte {
	Header := byte(CompactWidthOr(0x50, utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Amount))))
	AmountEncoded := utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Amount))

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, AmountEncoded...)
	Result = append(Result, msg.Data.Bytes...)

	return Result
}

func EncodeReconciliationTerminatePayload() []byte {
	return []byte{0x50}
}
