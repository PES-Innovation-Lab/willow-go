package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type Opts[DynamicToken constraints.Ordered, ValueType constraints.Unsigned] struct {
	DecodeNamespaceId      func(bytes utils.GrowingBytes) types.NamespaceId
	DecodeSubspaceId       func(bytes utils.GrowingBytes) types.SubspaceId
	DecodeDynamicToken     func(bytes utils.GrowingBytes) DynamicToken
	DecodePayloadDigest    func(bytes utils.GrowingBytes) types.PayloadDigest
	PathScheme             types.PathParams[ValueType]
	CurrentlyReceiverEntry types.Entry
	AoiHandlesToArea       func(senderHandle uint64, receiverHandle uint64) types.Area
	AoiHandlesToNamespace  func(senderHandle uint64, receiverHandle uint64) types.NamespaceId
}

func DecodeDataSendEntry[DynamicToken constraints.Ordered, ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts Opts[DynamicToken, ValueType]) wgpstypes.MsgDataSendEntry[DynamicToken] {
	bytes.NextAbsolute(2)

	FirstByte := bytes.Array[0]
	SecondByte := bytes.Array[1]

	StaticTokenCompactWidth := CompactWidthFromEndOfByte(int(FirstByte))

	IsOffsetEncoded := (SecondByte & 0x80) == 0x80

	IsOffsetPayloadLengthOrZero := !IsOffsetEncoded && (SecondByte&0x20) == 0x20

	var OffsetCompactWidth int

	if IsOffsetEncoded {
		OffsetCompactWidth = CompactWidthFromEndOfByte(int(SecondByte & 0x60 >> 5))
	} else {
		OffsetCompactWidth = 0
	}

	IsEntryEncodedRealative := (SecondByte & 0x10) == 0x10

	var SenderHandleCompactWidth int

	if IsEntryEncodedRealative {
		SenderHandleCompactWidth = 0
	} else {
		SenderHandleCompactWidth = CompactWidthFromEndOfByte(int(SecondByte & 0xc >> 2))
	}

	var ReceiverHandleCompactWidth int

	if IsEntryEncodedRealative {
		ReceiverHandleCompactWidth = 0
	} else {
		ReceiverHandleCompactWidth = CompactWidthFromEndOfByte(int(SecondByte))
	}

	bytes.Prune(2)

	bytes.NextAbsolute(StaticTokenCompactWidth)

	StaticTokenHandle := uint64(DecodeCompactWidth(bytes.Array[:StaticTokenCompactWidth]))

	bytes.Prune(StaticTokenCompactWidth)

	Dynamictoken := opts.DecodeDynamicToken(bytes)

	var Offset uint64

	if IsOffsetEncoded {
		bytes.NextAbsolute(OffsetCompactWidth)

		Offset = uint64(DecodeCompactWidth(bytes.Array[:OffsetCompactWidth]))

		bytes.Prune(OffsetCompactWidth)
	} else {
		Offset = 0
	}

	var Entry types.Entry

	if IsEntryEncodedRealative {
		Entry = utils.DecodeStreamEntryRelativeEntry() //gotta check this out
	} else if !IsntryEncodedRelative && SenderHandleCompactWidth > 0 && ReceiverHandleCompactWidth > 0 {
		bytes.NextAbsolute(SenderHandleCompactWidth + ReceiverHandleCompactWidth)
		SenderHandle := uint64(DecodeCompactWidth(bytes.Array[:SenderHandleCompactWidth]))
		ReceiverHandle := uint64(DecodeCompactWidth(bytes.Array[SenderHandleCompactWidth : SenderHandleCompactWidth+ReceiverHandleCompactWidth]))
		bytes.Prune(SenderHandleCompactWidth + ReceiverHandleCompactWidth)
		Entry = utils.DecodeStreamEntryRelativeEntry() //gotta check this out
	} else {
		//throw an error
	}

	if !IsOffsetEncoded {
		if IsOffsetPayloadLengthOrZero {
			Offset = Entry.Payload_length // Assuming PayloadLength is a field of the entry struct and is of type int64
		} else {
			Offset = 0
		}
	}

	return wgpstypes.MsgDataSendEntry[DynamicToken]{
		Kind: wgpstypes.DataSendEntry,
		Data: wgpstypes.MsgDataSendEntryData[DynamicToken]{
			Entry:             Entry,
			StaticTokenHandle: StaticTokenHandle,
			DynamicToken:      Dynamictoken,
			Offset:            Offset,
		},
	}
}

func DecodeDataSendPayload(bytes *utils.GrowingBytes) wgpstypes.MsgDataSendPayload {
	bytes.NextAbsolute(1)

	Header := bytes.Array

	CompactWidthAmount := CompactWidthFromEndOfByte(int(Header[0]))

	bytes.NextAbsolute(1 + CompactWidthAmount)

	Amount := int(DecodeCompactWidth(bytes.Array[1 : 1+CompactWidthAmount]))

	bytes.Prune(1 + CompactWidthAmount + Amount)

	MsgBytes := bytes.Array[1+CompactWidthAmount : 1+CompactWidthAmount+Amount]

	bytes.Prune(1 + CompactWidthAmount + Amount)

	return wgpstypes.MsgDataSendPayload{
		Kind: wgpstypes.DataSendPayload,
		Data: wgpstypes.MsgDataSendPayloadData{
			Amount: uint64(Amount),
			Bytes:  MsgBytes,
		},
	}
}

func DecodeDataSetEagerness(bytes *utils.GrowingBytes) wgpstypes.MsgDataSetMetadata {
	bytes.NextAbsolute(2)

	FirstByte := bytes.Array[0]
	SecondByte := bytes.Array[1]

	IsEager := (FirstByte & 0x2) == 0x2

	CompactWidthSenderHandle := CompactWidthFromEndOfByte(int(SecondByte >> 6))

	CompactWidthReceiverHandle := CompactWidthFromEndOfByte(int(SecondByte >> 4))

	bytes.NextAbsolute(2 + CompactWidthSenderHandle + CompactWidthReceiverHandle)

	SenderHandle := uint64(DecodeCompactWidth(bytes.Array[2 : 2+CompactWidthSenderHandle]))

	ReceiverHandle := uint64(DecodeCompactWidth(bytes.Array[2+CompactWidthSenderHandle : 2+CompactWidthSenderHandle+CompactWidthReceiverHandle]))

	bytes.Prune(2 + CompactWidthSenderHandle + CompactWidthReceiverHandle)

	return wgpstypes.MsgDataSetMetadata{
		Kind: wgpstypes.DataSetMetadata,
		Data: wgpstypes.MsgDataSetMetadataData{
			IsEager:        IsEager,
			SenderHandle:   SenderHandle,
			ReceiverHandle: ReceiverHandle,
		},
	}
}

type DecodeOpts struct {
	DecodeNamespaceId      func(bytes *utils.GrowingBytes) types.NamespaceId
	DecodeSubspaceId       func(bytes *utils.GrowingBytes) types.SubspaceId
	PathScheme             types.PathParams[uint64] //need to check this out
	CurrentlyReceiverEntry func() types.Entry
	AoiHandlesToArea       func(senderHandle uint64, receiverHandle uint64) types.Area
	AoiHandlesToNamespace  func(senderHandle uint64, receiverHandle uint64) types.NamespaceId
}

func DecodeDataPayloadRequest(bytes *utils.GrowingBytes, opts DecodeOpts) wgpstypes.MsgDataBindPayloadRequest {
	bytes.NextAbsolute(2)

	FirstByte := bytes.Array[0]
	SecondByte := bytes.Array[1]

	CompactWidthCapability := CompactWidthFromEndOfByte(int(FirstByte))

	IsOffsetEncoded := (SecondByte & 0x80) == 0x80

	var CompactWidthOffset int
	if IsOffsetEncoded {
		CompactWidthOffset = CompactWidthFromEndOfByte(int(SecondByte & 0x60 >> 5))
	} else {
		CompactWidthOffset = 0
	}

	IsEncodedRelativeToCurrEntry := (SecondByte & 0x10) == 0x10

	var CompactWidthSenderHandle int
	if IsEncodedRelativeToCurrEntry {
		CompactWidthSenderHandle = 0
	} else {
		CompactWidthSenderHandle = CompactWidthFromEndOfByte(int(SecondByte >> 2))
	}

	var CompactWidthReceiverHandle int
	if IsEncodedRelativeToCurrEntry {
		CompactWidthReceiverHandle = 0
	} else {
		CompactWidthReceiverHandle = CompactWidthFromEndOfByte(int(SecondByte))
	}

	bytes.NextAbsolute(CompactWidthCapability + 2)

	Capability := uint64(DecodeCompactWidth(bytes.Array[2 : 2+CompactWidthCapability]))

	var Offset uint64

	if IsOffsetEncoded {
		Offset = uint64(DecodeCompactWidth(bytes.Array[2+CompactWidthCapability : 2+CompactWidthCapability+CompactWidthOffset]))
		bytes.Prune(2 + CompactWidthCapability + CompactWidthOffset)
	} else {
		Offset = 0
		bytes.Prune(2 + CompactWidthCapability)
	}

	var Entry types.Entry

	if IsEncodedRelativeToCurrEntry {
		Entry = utils.DecodeStreamEntryRelativeEntry() //gotta check this out
	} else if !IsEncodedRelativeToCurrEntry && CompactWidthSenderHandle > 0 && CompactWidthReceiverHandle > 0 {
		bytes.NextAbsolute(CompactWidthSenderHandle + CompactWidthReceiverHandle)
		SenderHandle := uint64(DecodeCompactWidth(bytes.Array[:CompactWidthSenderHandle]))
		ReceiverHandle := uint64(DecodeCompactWidth(bytes.Array[CompactWidthSenderHandle : CompactWidthSenderHandle+CompactWidthReceiverHandle]))
		bytes.Prune(CompactWidthSenderHandle + CompactWidthReceiverHandle)
		Entry = utils.DecodeStreamEntryRelativeEntry() //gotta check this out
	} else {
		//throw an error
	}
	return wgpstypes.MsgDataBindPayloadRequest{
		Kind: wgpstypes.DataBindPayloadRequest,
		Data: wgpstypes.MsgDataBindPayloadRequestData{
			Entry:      Entry,
			Capability: Capability,
			Offset:     Offset,
		},
	}
}

func DecodeDataReplyPayload(bytes *utils.GrowingBytes) wgpstypes.MsgDataReplyPayload {
	bytes.NextAbsolute(1)

	CompactWidthHandle := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(1 + CompactWidthHandle)

	Handle := uint64(DecodeCompactWidth(bytes.Array[1 : 1+CompactWidthHandle]))

	bytes.Prune(1 + CompactWidthHandle)

	return wgpstypes.MsgDataReplyPayload{
		Kind: wgpstypes.DataReplyPayload,
		Data: wgpstypes.MsgDataReplyPayloadData{
			Handle: Handle,
		},
	}
}
