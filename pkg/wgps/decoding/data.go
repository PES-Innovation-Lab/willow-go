package decoding

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type Opts[DynamicToken string, ValueType constraints.Unsigned] struct {
	DecodeNamespaceId      func(bytes *utils.GrowingBytes) chan types.NamespaceId
	DecodeSubspaceId       func(bytes *utils.GrowingBytes) chan types.SubspaceId
	DecodeDynamicToken     func(bytes *utils.GrowingBytes) DynamicToken
	DecodePayloadDigest    func(bytes *utils.GrowingBytes) chan types.PayloadDigest
	PathScheme             types.PathParams[ValueType]
	CurrentlyReceivedEntry types.Entry
	AoiHandlesToArea       func(senderHandle uint64, receiverHandle uint64) types.Area
	AoiHandlesToNamespace  func(senderHandle uint64, receiverHandle uint64) types.NamespaceId
}

func DecodeDataSendEntry[DynamicToken string, ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts Opts[DynamicToken, ValueType]) wgpstypes.MsgDataSendEntry[DynamicToken] {
	received := bytes.NextAbsolute(2)

	FirstByte := received[0]
	SecondByte := received[1]

	StaticTokenCompactWidth := CompactWidthFromEndOfByte(int(FirstByte))

	IsOffsetEncoded := (SecondByte & 0x80) == 0x80

	IsOffsetPayloadLengthOrZero := !IsOffsetEncoded && (SecondByte&0x20) == 0x20

	var OffsetCompactWidth int

	if IsOffsetEncoded {
		OffsetCompactWidth = CompactWidthFromEndOfByte(int(SecondByte & 0x60 >> 5))
	} else {
		OffsetCompactWidth = 0
	}

	IsEntryEncodedRelative := (SecondByte & 0x10) == 0x10

	var SenderHandleCompactWidth int

	if IsEntryEncodedRelative {
		SenderHandleCompactWidth = 0
	} else {
		SenderHandleCompactWidth = CompactWidthFromEndOfByte(int(SecondByte & 0xc >> 2))
	}

	var ReceiverHandleCompactWidth int

	if IsEntryEncodedRelative {
		ReceiverHandleCompactWidth = 0
	} else {
		ReceiverHandleCompactWidth = CompactWidthFromEndOfByte(int(SecondByte))
	}

	bytes.Prune(2)

	received = bytes.NextAbsolute(StaticTokenCompactWidth)

	StaticTokenHandle, _ := utils.DecodeIntMax64(received[:StaticTokenCompactWidth])

	bytes.Prune(StaticTokenCompactWidth)

	Dynamictoken := opts.DecodeDynamicToken(bytes)

	var Offset uint64

	if IsOffsetEncoded {
		received = bytes.NextAbsolute(OffsetCompactWidth)

		Offset, _ = utils.DecodeIntMax64(received[:OffsetCompactWidth])

		bytes.Prune(OffsetCompactWidth)
	} else {
		Offset = 0
	}

	var decodeResultChan chan utils.DecodeResult
	var Entry types.Entry

	if IsEntryEncodedRelative {
		decodeResultChan = utils.DecodeStreamEntryRelativeEntry[ValueType](
			struct {
				DecodeStreamNamespace     func(bytes *utils.GrowingBytes) chan types.NamespaceId
				DecodeStreamSubspace      func(bytes *utils.GrowingBytes) chan types.SubspaceId
				DecodeStreamPayloadDigest func(bytes *utils.GrowingBytes) chan types.PayloadDigest
				PathScheme                types.PathParams[ValueType]
			}{
				DecodeStreamNamespace:     opts.DecodeNamespaceId,
				DecodeStreamSubspace:      opts.DecodeSubspaceId,
				DecodeStreamPayloadDigest: opts.DecodePayloadDigest,
				PathScheme:                opts.PathScheme,
			}, bytes, opts.CurrentlyReceivedEntry)
		result := <-decodeResultChan
		Entry = result.Entry
	} else if !IsEntryEncodedRelative && SenderHandleCompactWidth > 0 && ReceiverHandleCompactWidth > 0 {
		received = bytes.NextAbsolute(SenderHandleCompactWidth + ReceiverHandleCompactWidth)
		SenderHandle, _ := utils.DecodeIntMax64(received[:SenderHandleCompactWidth])
		ReceiverHandle, _ := utils.DecodeIntMax64(received[SenderHandleCompactWidth : SenderHandleCompactWidth+ReceiverHandleCompactWidth])
		bytes.Prune(SenderHandleCompactWidth + ReceiverHandleCompactWidth)
		Entry, _ = utils.DecodeStreamEntryInNamespaceArea[ValueType](
			utils.EntryOpts[ValueType]{
				DecodeStreamSubspace:      opts.DecodeSubspaceId,
				DecodeStreamPayloadDigest: opts.DecodePayloadDigest,
				PathScheme:                opts.PathScheme,
			}, bytes, opts.AoiHandlesToArea(SenderHandle, ReceiverHandle), opts.AoiHandlesToNamespace(SenderHandle, ReceiverHandle),
		) //gotta check this out
	} else {
		fmt.Errorf("could not decode entry encoded relative to area when no handles are provided")
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
	received := bytes.NextAbsolute(1)

	Header := received

	CompactWidthAmount := CompactWidthFromEndOfByte(int(Header[0]))

	received = bytes.NextAbsolute(1 + CompactWidthAmount)

	Amount, _ := utils.DecodeIntMax64(received[1 : 1+CompactWidthAmount])

	bytes.Prune(1 + CompactWidthAmount + int(Amount))

	MsgBytes := received[1+CompactWidthAmount : 1+CompactWidthAmount+int(Amount)]

	bytes.Prune(1 + CompactWidthAmount + int(Amount))

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

	SenderHandle, _ := utils.DecodeIntMax64(bytes.Array[2 : 2+CompactWidthSenderHandle])

	ReceiverHandle, _ := utils.DecodeIntMax64(bytes.Array[2+CompactWidthSenderHandle : 2+CompactWidthSenderHandle+CompactWidthReceiverHandle])

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

type DecodeOpts[ValueType constraints.Unsigned] struct {
	DecodeNamespaceId         func(bytes *utils.GrowingBytes) chan types.NamespaceId
	DecodeSubspaceId          func(bytes *utils.GrowingBytes) chan types.SubspaceId
	DecodePayloadDigest       func(bytes *utils.GrowingBytes) chan types.PayloadDigest
	PathScheme                types.PathParams[ValueType] //need to check this out
	GetCurrentlyReceivedEntry types.Entry
	AoiHandlesToArea          func(senderHandle uint64, receiverHandle uint64) types.Area
	AoiHandlesToNamespace     func(senderHandle uint64, receiverHandle uint64) types.NamespaceId
}

func DecodeDataBindPayloadRequest[ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts DecodeOpts[ValueType]) wgpstypes.MsgDataBindPayloadRequest {
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

	Capability, _ := utils.DecodeIntMax64(bytes.Array[2 : 2+CompactWidthCapability])

	var Offset uint64

	if IsOffsetEncoded {
		Offset, _ = utils.DecodeIntMax64(bytes.Array[2+CompactWidthCapability : 2+CompactWidthCapability+CompactWidthOffset])
		bytes.Prune(2 + CompactWidthCapability + CompactWidthOffset)
	} else {
		Offset = 0
		bytes.Prune(2 + CompactWidthCapability)
	}

	var decodeResultChan chan utils.DecodeResult
	var Entry types.Entry

	if IsEncodedRelativeToCurrEntry {
		decodeResultChan = utils.DecodeStreamEntryRelativeEntry[ValueType](
			struct {
				DecodeStreamNamespace     func(bytes *utils.GrowingBytes) chan types.NamespaceId
				DecodeStreamSubspace      func(bytes *utils.GrowingBytes) chan types.SubspaceId
				DecodeStreamPayloadDigest func(bytes *utils.GrowingBytes) chan types.PayloadDigest
				PathScheme                types.PathParams[ValueType]
			}{
				DecodeStreamNamespace:     opts.DecodeNamespaceId,
				DecodeStreamSubspace:      opts.DecodeSubspaceId,
				DecodeStreamPayloadDigest: opts.DecodePayloadDigest,
				PathScheme:                opts.PathScheme,
			}, bytes, opts.GetCurrentlyReceivedEntry)
		result := <-decodeResultChan
		Entry = result.Entry //gotta check this out
	} else if !IsEncodedRelativeToCurrEntry && CompactWidthSenderHandle > 0 && CompactWidthReceiverHandle > 0 {
		bytes.NextAbsolute(CompactWidthSenderHandle + CompactWidthReceiverHandle)
		SenderHandle, _ := utils.DecodeIntMax64(bytes.Array[:CompactWidthSenderHandle])
		ReceiverHandle, _ := utils.DecodeIntMax64(bytes.Array[CompactWidthSenderHandle : CompactWidthSenderHandle+CompactWidthReceiverHandle])
		bytes.Prune(CompactWidthSenderHandle + CompactWidthReceiverHandle)
		Entry, _ = utils.DecodeStreamEntryInNamespaceArea[ValueType](
			utils.EntryOpts[ValueType]{
				DecodeStreamSubspace:      opts.DecodeSubspaceId,
				DecodeStreamPayloadDigest: opts.DecodePayloadDigest,
				PathScheme:                opts.PathScheme,
			}, bytes, opts.AoiHandlesToArea(SenderHandle, ReceiverHandle), opts.AoiHandlesToNamespace(SenderHandle, ReceiverHandle),
		)
	} else {
		fmt.Errorf("Could not decode")
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

	Handle, _ := utils.DecodeIntMax64(bytes.Array[1 : 1+CompactWidthHandle])

	bytes.Prune(1 + CompactWidthHandle)

	return wgpstypes.MsgDataReplyPayload{
		Kind: wgpstypes.DataReplyPayload,
		Data: wgpstypes.MsgDataReplyPayloadData{
			Handle: Handle,
		},
	}
}
