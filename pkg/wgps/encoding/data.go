package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func EncodeDataSendEntry[DynamicToken constraints.Ordered, K constraints.Unsigned](
	msg wgpstypes.MsgDataSendEntry[DynamicToken],
	opts struct {
		EncodeDynamicToken  func(token DynamicToken) []byte
		CurrentlySentEntry  types.Entry
		IsEqualNamespace    func(a, b types.NamespaceId) bool
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(payloadDigest types.PayloadDigest) []byte
		PathParams          types.PathParams[K]
	},
) []byte {
	var messageTypeMask = 0x60
	var compactWidthStaticTokenFlag = CompactWidthOr(0, int(msg.Data.StaticTokenHandle))
	var firstByte = byte(messageTypeMask | compactWidthStaticTokenFlag)
	var encodeOffsetFlag int
	if msg.Data.Offset != 0 && msg.Data.Offset != msg.Data.Entry.Payload_length {
		encodeOffsetFlag = 0x80
	} else {
		encodeOffsetFlag = 0x0
	}
	var compactWidthOffsetFlag int
	if msg.Data.Offset == 0 {
		compactWidthOffsetFlag = 0x0
	} else {
		if msg.Data.Offset == msg.Data.Entry.Payload_length {
			compactWidthOffsetFlag = 0x20
		} else {
			compactWidthOffsetFlag = CompactWidthOr(0, utils.GetWidthMax64Int(msg.Data.Offset)<<5)
		}
	}
	// This is always flagged to true
	var encodedRelativeToCurrent = 0x10

	// which means this is always 0x0
	var compactWidthSenderHandle = 0x0
	// and this is always 0x0
	var compactWidthReceiverHandle = 0x0

	var secondByte = byte(encodeOffsetFlag | compactWidthOffsetFlag |
		encodedRelativeToCurrent | compactWidthSenderHandle |
		compactWidthReceiverHandle)

	var encodedStaticToken = utils.EncodeIntMax64(msg.Data.StaticTokenHandle)
	var encodedDynamicToken = opts.EncodeDynamicToken(msg.Data.DynamicToken)
	var encodedOffset []byte
	var encodedEntry = utils.EncodeEntryRelativeEntry[K](
		struct {
			EncodeNamespace     func(namespace types.NamespaceId) []byte
			EncodeSubspace      func(subspace types.SubspaceId) []byte
			EncodePayloadDigest func(digest types.PayloadDigest) []byte
			IsEqualNamespace    func(a types.NamespaceId, b types.NamespaceId) bool
			OrderSubspace       types.TotalOrder[types.SubspaceId]
			PathScheme          types.PathParams[K]
		}{
			EncodeNamespace:     opts.EncodeNamespace,
			EncodeSubspace:      opts.EncodeSubspace,
			EncodePayloadDigest: opts.EncodePayloadDigest,
			IsEqualNamespace:    opts.IsEqualNamespace,
			OrderSubspace:       opts.OrderSubspace,
			PathScheme:          opts.PathParams,
		},
		msg.Data.Entry,
		opts.CurrentlySentEntry,
	)

	return append(append(append(append([]byte{firstByte, secondByte}, encodedStaticToken...), encodedDynamicToken...), encodedOffset...), encodedEntry...)
}

func EncodeDataSendPayload(msg wgpstypes.MsgDataSendPayload) []byte {
	var messageKindFlag = 0x64
	var header = byte(CompactWidthOr(messageKindFlag, utils.GetWidthMax64Int(msg.Data.Amount)))
	var encodedAmount = utils.EncodeIntMax64(msg.Data.Amount)
	return append(append([]byte{header}, encodedAmount...), msg.Data.Bytes...)
}

func EncodeDataSetEagerness(msg wgpstypes.MsgDataSetMetadata) []byte {
	var messageKind = 0x68
	var eagernessFlag int
	if msg.Data.IsEager {
		eagernessFlag = 0x2
	} else {
		eagernessFlag = 0x0
	}
	var firstByte = byte(messageKind | eagernessFlag)

	var secondByte = 0x0

	secondByte = CompactWidthOr(secondByte, utils.GetWidthMax64Int(msg.Data.SenderHandle)) << 2
	secondByte = CompactWidthOr(secondByte, utils.GetWidthMax64Int(msg.Data.ReceiverHandle)) << 4

	return append(append([]byte{firstByte, byte(secondByte)}, utils.EncodeIntMax64(msg.Data.SenderHandle)...), utils.EncodeIntMax64(msg.Data.ReceiverHandle)...)
}

func EncodeDataBindPayloadRequest[K constraints.Unsigned](
	msg wgpstypes.MsgDataBindPayloadRequest,
	opts struct {
		CurrentlySentEntry  types.Entry
		IsEqualNamespace    func(a, b types.NamespaceId) bool
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(payloadDigest types.PayloadDigest) []byte
		PathParams          types.PathParams[K]
	},
) []byte {
	var messageKind = 0x6c
	var firstByte = byte(CompactWidthOr(messageKind, utils.GetWidthMax64Int(msg.Data.Capability)))
	var encodedOffsetFlag int
	if msg.Data.Offset != 0 {
		encodedOffsetFlag = 0x80
	} else {
		encodedOffsetFlag = 0x0
	}

	var compactWidthOffset int
	if encodedOffsetFlag == 0x0 {
		compactWidthOffset = 0x0

	} else {
		compactWidthOffset = CompactWidthOr(0x0, utils.GetWidthMax64Int(msg.Data.Offset)) << 5
	}

	var encodedRelativeFlag = 0x10

	// Don't encode sender and receiver handle widths.

	var secondByte = byte(encodedOffsetFlag | compactWidthOffset | encodedRelativeFlag)

	var encodedCapability = utils.EncodeIntMax64(msg.Data.Capability)

	var encodedOffset []byte

	if encodedOffsetFlag == 0x0 {
		encodedOffset = []byte{}
	} else {
		encodedOffset = utils.EncodeIntMax64(msg.Data.Offset)
	}

	var encodedEntry = utils.EncodeEntryRelativeEntry[K](
		struct {
			EncodeNamespace     func(namespace types.NamespaceId) []byte
			EncodeSubspace      func(subspace types.SubspaceId) []byte
			EncodePayloadDigest func(digest types.PayloadDigest) []byte
			IsEqualNamespace    func(a types.NamespaceId, b types.NamespaceId) bool
			OrderSubspace       types.TotalOrder[types.SubspaceId]
			PathScheme          types.PathParams[K]
		}{
			EncodeNamespace:     opts.EncodeNamespace,
			EncodeSubspace:      opts.EncodeSubspace,
			EncodePayloadDigest: opts.EncodePayloadDigest,
			IsEqualNamespace:    opts.IsEqualNamespace,
			OrderSubspace:       opts.OrderSubspace,
			PathScheme:          opts.PathParams,
		},
		msg.Data.Entry,
		opts.CurrentlySentEntry,
	)

	return append(
		append(
			append(
				[]byte{firstByte, secondByte},
				encodedCapability...,
			), encodedOffset...,
		), encodedEntry...,
	)
}

func EncodeDataReplyPayload(msg wgpstypes.MsgDataReplyPayload) []byte {
	var messageKind = 0x70
	var header = CompactWidthOr(messageKind, utils.GetWidthMax64Int(msg.Data.Handle))

	var encodedHandle = utils.EncodeIntMax64(msg.Data.Handle)

	return append([]byte{byte(header)}, encodedHandle...)
}
