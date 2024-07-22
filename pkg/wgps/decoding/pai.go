package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func DecodePaiFragment[PsiGroup types.OrderableGeneric](bytes *utils.GrowingBytes, groupDecoder utils.StreamDecoder[PsiGroup]) wgpstypes.MsgPaiBindFragment[PsiGroup] {
	bytes.NextAbsolute(1)

	IsSecondary := bytes.Array[0] == 0x6

	bytes.Prune(1)

	GroupMember := groupDecoder(bytes)

	return wgpstypes.MsgPaiBindFragment[PsiGroup]{
		Kind: wgpstypes.PaiBindFragment,
		Data: wgpstypes.MsgPaiBindFragmentData[PsiGroup]{
			IsSecondary: IsSecondary,
			GroupMember: GroupMember,
		},
	}
}

func DecodePaiReplyFragment[PsiGroup types.OrderableGeneric](bytes *utils.GrowingBytes, groupDecoder utils.StreamDecoder[PsiGroup]) wgpstypes.MsgPaiReplyFragment[PsiGroup] {
	bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	Handle := utils.DecodeCompactWidth(bytes.Array[1 : 1+CompactWidth])

	bytes.Prune(1 + CompactWidth)

	GroupMember := groupDecoder(bytes)

	return wgpstypes.MsgPaiReplyFragment[PsiGroup]{
		Kind: wgpstypes.PaiReplyFragment,
		Data: wgpstypes.MsgPaiReplyFragmentData[PsiGroup]{
			Handle:      uint64(Handle),
			GroupMember: GroupMember,
		},
	}
}

func DecodePaiRequestSubspaceCapability(bytes *utils.GrowingBytes) wgpstypes.MsgPaiRequestSubspaceCapability {
	bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	Handle := utils.DecodeCompactWidth(bytes.Array[1 : 1+CompactWidth])

	bytes.Prune(1 + CompactWidth)

	return wgpstypes.MsgPaiRequestSubspaceCapability{
		Kind: wgpstypes.PaiRequestSubspaceCapability,
		Data: wgpstypes.MsgPaiRequestSubspaceCapabilityData{
			Handle: uint64(Handle),
		},
	}
}

func DecodePaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature any](bytes utils.GrowingBytes, decodeCap utils.StreamDecoder[SubspaceCapability], decodeSig utils.StreamDecoder[SyncSubspaceSignature]) wgpstypes.MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature] { //NEED TO CHANGE THESE TYPE DEFINITIONS
	bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	Handle := utils.DecodeCompactWidth(bytes.Array[1 : 1+CompactWidth])

	bytes.Prune(1 + CompactWidth)

	Capability := decodeCap(&bytes)
	Signature := decodeSig(&bytes)

	return wgpstypes.MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature]{ //need to check this out
		Kind: wgpstypes.PaiReplySubspaceCapability,
		Data: wgpstypes.MsgPaiReplySubspaceCapabilityData[SubspaceCapability, SyncSubspaceSignature]{
			Handle:     uint64(Handle),
			Capability: Capability,
			Signature:  Signature,
		},
	}
}
