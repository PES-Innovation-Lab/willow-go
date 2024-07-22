package decoding

import (
	"math"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type SendOpts[Fingerprint constraints.Ordered, ValueType constraints.Unsigned] struct {
	NeutralFingerprint  Fingerprint
	DecodeFingerprint   func(bytes *utils.GrowingBytes) Fingerprint
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) types.SubspaceId
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

		Covers = uint64(utils.DecodeCompactWidth(bytes.Array[:CoversCompactWidth]))

		bytes.Prune(CoversCompactWidth)
	} else {
		Covers = COVERS_NONE
	}

	if !IsSenderPrevSender {
		bytes.NextAbsolute(SenderCompactWidth)

		SenderHandle = uint64(utils.DecodeCompactWidth(bytes.Array[:SenderCompactWidth]))

		bytes.Prune(SenderCompactWidth)
	} else {
		SenderHandle = Privy.PrevSenderHandle
	}

	if !IsReceiverPrevReceiver {
		bytes.NextAbsolute(ReceiverCompactWidth)

		ReceiverHandle = uint64(utils.DecodeCompactWidth(bytes.Array[:ReceiverCompactWidth]))

		bytes.Prune(ReceiverCompactWidth)
	} else {
		ReceiverHandle = Privy.PrevReceiverHandle
	}

	var fingerprint Fingerprint

	if IsFingerprintNeutral {
		fingerprint = opts.NeutralFingerprint
	} else {
		fingerprint = opts.DecodeFingerprint(bytes)
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

type AnnounceOpts struct {
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) types.SubspaceId
	PathScheme          types.PathParams[uint64]
	GetPrivy            func() wgpstypes.ReconciliationPrivy
	AoiHandlesToRange3d func(senderAoiHandle uint64, receiverAoiHandle uint64) types.Range3d
}

func DecodeReconciliationAnnounceEntries(bytes *utils.GrowingBytes, opts AnnounceOpts) wgpstypes.MsgReconciliationAnnounceEntries {
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

			Covers = uint64(utils.DecodeCompactWidth(bytes.Array[1 : 1+CoversCompactWidth]))

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

		SenderHandle = uint64(utils.DecodeCompactWidth(bytes.Array[:SenderCompactWidth]))

		bytes.Prune(SenderCompactWidth)
	} else {
		SenderHandle = Privy.PrevSenderHandle
	}

	if !IsReceiverPrevReceiver {
		ReceiverCompactWidth := int(math.Pow(2, float64(int((SecondByte>>4)&0x3))))
		bytes.NextAbsolute(ReceiverCompactWidth)

		ReceiverHandle = uint64(utils.DecodeCompactWidth(bytes.Array[:ReceiverCompactWidth]))

		bytes.Prune(ReceiverCompactWidth)
	} else {
		ReceiverHandle = Privy.PrevReceiverHandle
	}

	bytes.NextAbsolute(CountWidth)

	Count := uint64(utils.DecodeCompactWidth(bytes.Array[:CountWidth]))

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

type EntryOpts[DynamicToken any] struct {
	DecodeNamespaceId   func(bytes *utils.GrowingBytes) types.NamespaceId
	DecodeSubspaceId    func(bytes *utils.GrowingBytes) types.SubspaceId
	DecodeDynamicToken  func(bytes *utils.GrowingBytes) DynamicToken
	DecodePayloadDigest func(bytes *utils.GrowingBytes) types.PayloadDigest
	PathScheme          types.PathParams[uint64]
	GetPrivy            func() wgpstypes.ReconciliationPrivy
}

func DecodeReconciliationSendEntry[DynamicToken any](bytes *utils.GrowingBytes, opts EntryOpts[DynamicToken])
