package decoding

import (
	"math"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type SendOpts[Fingerprint constraints.Ordered, ValueType constraints.Unsigned] struct {
	NeutralFingerprint  Fingerprint
	DecodeFingerprint   func(bytes *utils.GrowingBytes) chan Fingerprint
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) chan types.SubspaceId
	PathScheme          types.PathParams[ValueType] //need to check if this is right
	GetPrivy            func() wgpstypes.ReconciliationPrivy
	AoiHandlesToRange3d func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d
}

func DecodeReconciliationSendFingerprint[Fingerprint constraints.Ordered, ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts SendOpts[Fingerprint, ValueType]) wgpstypes.MsgReconciliationSendFingerprint[Fingerprint] {
	Privy := opts.GetPrivy()

	bytes.NextAbsolute(2)

	FirstByte := bytes.Array[0]
	SecondByte := bytes.Array[1]

	IsFingerprintNeutral := (FirstByte & 0x8) == 0x8

	EncodeRelativeToPrevRange := (FirstByte & 0x4) == 0x4

	IsSenderPrevSender := (FirstByte & 0x2) == 0x2
	IsReceiverPrevReceiver := (FirstByte & 0x1) == 0x1

	SenderCompactWidth := int(math.Pow(2, float64(int(SecondByte>>6))))
	ReceiverCompactWidth := int(math.Pow(2, float64(int((SecondByte>>4)&0x3))))

	CoversNotNone := (SecondByte & 0x8) == 0x8
	CoversCompactWidth := int(math.Pow(2, float64(int(SecondByte&0x3))))

	COVERS_NONE := uint64(0)

	var Covers uint64
	var SenderHandle uint64
	var ReceiverHandle uint64

	bytes.Prune(2)

	if CoversNotNone {
		bytes.NextAbsolute(CoversCompactWidth)

		Covers, _ = utils.DecodeIntMax64(bytes.Array[:CoversCompactWidth])

		bytes.Prune(CoversCompactWidth)
	} else {
		Covers = COVERS_NONE
	}

	if !IsSenderPrevSender {
		bytes.NextAbsolute(SenderCompactWidth)

		SenderHandle, _ = utils.DecodeIntMax64(bytes.Array[:SenderCompactWidth])

		bytes.Prune(SenderCompactWidth)
	} else {
		SenderHandle = Privy.PrevSenderHandle
	}

	if !IsReceiverPrevReceiver {
		bytes.NextAbsolute(ReceiverCompactWidth)

		ReceiverHandle, _ = utils.DecodeIntMax64(bytes.Array[:ReceiverCompactWidth])

		bytes.Prune(ReceiverCompactWidth)
	} else {
		ReceiverHandle = Privy.PrevReceiverHandle
	}

	var fingerprint Fingerprint

	if IsFingerprintNeutral {
		fingerprint = opts.NeutralFingerprint
	} else {
		fingerprint = <-opts.DecodeFingerprint(bytes)
	}

	var Outer types.Range3d

	if EncodeRelativeToPrevRange {
		Outer = Privy.PrevRange
	} else {
		Outer = opts.AoiHandlesToRange3d(SenderHandle, ReceiverHandle)
	}

	Range, _ := utils.DecodeStreamRange3dRelative(opts.DecodeSubspaceId, opts.PathScheme, bytes, Outer)

	return wgpstypes.MsgReconciliationSendFingerprint[Fingerprint]{
		Kind: wgpstypes.ReconciliationSendFingerprint,
		Data: wgpstypes.MsgReconciliationSendFingerprintData[Fingerprint]{
			Fingerprint:    fingerprint,
			Range:          Range,
			SenderHandle:   SenderHandle,
			ReceiverHandle: ReceiverHandle,
			Covers:         Covers,
			DoesCover:      true,
		},
	}
}

type AnnounceOpts[ValueType constraints.Unsigned] struct {
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) chan types.SubspaceId
	PathScheme          types.PathParams[ValueType]
	GetPrivy            func() wgpstypes.ReconciliationPrivy
	AoiHandlesToRange3d func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d
}

func DecodeReconciliationAnnounceEntries[ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts AnnounceOpts[ValueType]) wgpstypes.MsgReconciliationAnnounceEntries {
	Privy := opts.GetPrivy()

	bytes.NextAbsolute(2)

	FirstByte, SecondByte := bytes.Array[0], bytes.Array[1]

	WantResponse := (FirstByte & 0x8) == 0x8

	EncodeRelativeToPrevRange := (FirstByte & 0x4) == 0x4

	IsSenderPrevSender := (FirstByte & 0x2) == 0x2
	IsReceiverPrevReceiver := (FirstByte & 0x1) == 0x1

	CountWidth := int(math.Pow(2, float64(int((SecondByte&0xc)>>2))))

	WillSort := (SecondByte & 0x2) == 0x2

	CoversNotNone := (SecondByte & 0x1) == 0x1

	COVERS_NONE := uint64(0)

	var Covers uint64

	bytes.Prune(2)

	if CoversNotNone {
		bytes.NextAbsolute(1)

		CoversLength := bytes.Array[0]

		if (CoversLength & 0xfc) == 0xfc {
			CoversCompactWidth := int(math.Pow(2, float64(int((SecondByte & 0x3)))))

			bytes.NextAbsolute(CoversCompactWidth)

			Covers, _ = utils.DecodeIntMax64(bytes.Array[1 : 1+CoversCompactWidth])

			bytes.Prune(1 + CoversCompactWidth)
		} else {
			Covers = uint64(CoversLength)
			bytes.Prune(1)
		}
	} else {
		Covers = COVERS_NONE
	}

	var SenderHandle uint64
	var ReceiverHandle uint64

	if !IsSenderPrevSender {
		SenderCompactWidth := int(math.Pow(2, float64(int((SecondByte)>>6))))
		bytes.NextAbsolute(SenderCompactWidth)

		SenderHandle, _ = utils.DecodeIntMax64(bytes.Array[:SenderCompactWidth])

		bytes.Prune(SenderCompactWidth)
	} else {
		SenderHandle = Privy.PrevSenderHandle
	}

	if !IsReceiverPrevReceiver {
		ReceiverCompactWidth := int(math.Pow(2, float64(int((SecondByte>>4)&0x3))))
		bytes.NextAbsolute(ReceiverCompactWidth)

		ReceiverHandle, _ = utils.DecodeIntMax64(bytes.Array[:ReceiverCompactWidth])

		bytes.Prune(ReceiverCompactWidth)
	} else {
		ReceiverHandle = Privy.PrevReceiverHandle
	}

	bytes.NextAbsolute(CountWidth)

	Count, _ := utils.DecodeIntMax64(bytes.Array[:CountWidth])

	bytes.Prune(CountWidth)

	var Outer types.Range3d

	if EncodeRelativeToPrevRange {
		Outer = Privy.PrevRange
	} else {
		Outer = opts.AoiHandlesToRange3d(SenderHandle, ReceiverHandle)
	}

	Range, _ := utils.DecodeStreamRange3dRelative(opts.DecodeSubspaceId, opts.PathScheme, bytes, Outer)

	return wgpstypes.MsgReconciliationAnnounceEntries{
		Kind: wgpstypes.ReconciliationAnnounceEntries,
		Data: wgpstypes.MsgReconciliationAnnounceEntriesData{
			Range:          Range,
			Count:          Count,
			WantResponse:   WantResponse,
			WillSort:       WillSort,
			SenderHandle:   SenderHandle,
			ReceiverHandle: ReceiverHandle,
			Covers:         Covers,
			DoesCover:      true,
		},
	}
}

type EntryOpts[DynamicToken string, ValueType constraints.Unsigned] struct {
	DecodeNamespaceId   func(bytes *utils.GrowingBytes) chan types.NamespaceId
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) chan types.SubspaceId
	DecodeDynamicToken  func(bytes *utils.GrowingBytes) chan DynamicToken
	DecodePayloadDigest func(bytes *utils.GrowingBytes) chan types.PayloadDigest
	PathScheme          types.PathParams[ValueType]
	GetPrivy            func() wgpstypes.ReconciliationPrivy
}

func DecodeReconciliationSendEntry[DynamicToken string, ValueType constraints.Unsigned](bytes *utils.GrowingBytes, opts EntryOpts[DynamicToken, ValueType]) wgpstypes.MsgReconciliationSendEntry[DynamicToken] {
	Privy := opts.GetPrivy()

	bytes.NextAbsolute(1)

	Header := bytes.Array[0]

	IsPrevStaticToken := (Header & 0x8) == 0x8
	IsEncodedRelativeToPrev := (Header & 0x4) == 0x4
	CompactWidthAvailable := int(math.Pow(2, float64(int((Header & 0x3)))))

	var StaticTokenHandle uint64

	if IsPrevStaticToken {
		StaticTokenHandle = Privy.PrevStaticTokenHandle
		bytes.Prune(1)
	} else {
		bytes.NextAbsolute(2)
		StaticTokensizeByte := bytes.Array[1]

		var CompactWidth int = 0

		if (StaticTokensizeByte & 0xff) == 0xff {
			CompactWidth = 8
		} else if (StaticTokensizeByte & 0xbf) == 0xbf {
			CompactWidth = 4
		} else if (StaticTokensizeByte & 0x7f) == 0x7f {
			CompactWidth = 2
		} else {
			CompactWidth = 1
		}

		bytes.NextAbsolute(2 + CompactWidth)

		if StaticTokensizeByte < 63 {
			StaticTokenHandle = uint64(StaticTokensizeByte)
		} else {
			StaticTokenHandle, _ = utils.DecodeIntMax64(bytes.Array[2 : 2+CompactWidth])
		}
		bytes.Prune(2 + CompactWidth)
	}

	bytes.NextAbsolute(CompactWidthAvailable)

	Available, _ := utils.DecodeIntMax64(bytes.Array[:CompactWidthAvailable])

	bytes.Prune(CompactWidthAvailable)

	dynamicToken := opts.DecodeDynamicToken(bytes)

	var decodeResultChan chan utils.DecodeResult
	var Entry types.Entry

	if IsEncodedRelativeToPrev {
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
			}, bytes, Privy.PrevEntry)
		result := <-decodeResultChan
		Entry = result.Entry //gotta check this out
	} else {
		decodeResultChan = utils.DecodeStreamEntryRelativeRange3d[ValueType](
			struct {
				DecodeStreamSubspace      func(bytes *utils.GrowingBytes) chan types.SubspaceId
				DecodeStreamPayloadDigest func(bytes *utils.GrowingBytes) chan types.PayloadDigest
				PathScheme                types.PathParams[ValueType]
			}{
				DecodeStreamSubspace:      opts.DecodeSubspaceId,
				DecodeStreamPayloadDigest: opts.DecodePayloadDigest,
				PathScheme:                opts.PathScheme,
			}, bytes, Privy.Announced.Range, Privy.Announced.Namespace,
		)
	}

	return wgpstypes.MsgReconciliationSendEntry[DynamicToken]{ //need to verify the type of DynamicToken
		Kind: wgpstypes.ReconciliationSendEntry,
		Data: wgpstypes.MsgReconciliationSendEntryData[DynamicToken]{
			Entry: datamodeltypes.LengthyEntry{
				Available: Available,
				Entry:     Entry,
			},
			DynamicToken:      <-dynamicToken,
			StaticTokenHandle: StaticTokenHandle,
		},
	}
}

func DecodeReconciliationSendPayload(bytes *utils.GrowingBytes) wgpstypes.MsgReconciliationSendPayload {
	bytes.NextAbsolute(1)

	AmountCompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(1 + AmountCompactWidth)

	Amount, _ := utils.DecodeIntMax64(bytes.Array[1 : 1+AmountCompactWidth])

	bytes.Prune(1 + AmountCompactWidth)

	bytes.NextAbsolute(int(Amount))

	MessageBytes := bytes.Array[:int(Amount)]

	bytes.Prune(int(Amount))

	return wgpstypes.MsgReconciliationSendPayload{
		Kind: wgpstypes.ReconciliationSendPayload,
		Data: wgpstypes.MsgReconciliationSendPayloadData{
			Amount: uint64(Amount),
			Bytes:  MessageBytes,
		},
	}
}
