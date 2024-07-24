package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func EncodeCommitmentReveal(msg wgpstypes.MsgCommitmentReveal) []byte {
	var Result []byte
	Result = append(Result, msg.Data.Nonce...)
	return Result
}

func EncodePaiBindFragment[PsiGroup any](msg wgpstypes.MsgPaiBindFragment[PsiGroup], encodeGroupMember func(group PsiGroup) []byte) []byte {
	var Result []byte
	if msg.Data.IsSecondary {
		Result = append(Result, 6)
	} else {
		Result = append(Result, 4)
	}

	Result = append(Result, encodeGroupMember(msg.Data.GroupMember)...)
	return Result
}

func EncodePaiReplyFragment[PsiGroup any](msg wgpstypes.MsgPaiReplyFragment[PsiGroup], encodeGroupMember func(group PsiGroup) []byte) []byte {
	HandleWidth := byte(utils.GetWidthMax64Int(msg.Data.Handle))

	var Header byte

	if HandleWidth == 1 {
		Header = 0x8
	} else if HandleWidth == 2 {
		Header = 0x9
	} else if HandleWidth == 4 {
		Header = 0xa
	} else {
		Header = 0xb
	}

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, byte(utils.GetWidthMax64Int(msg.Data.Handle)))
	Result = append(Result, (encodeGroupMember(msg.Data.GroupMember))...)

	return Result
}

func EncodePaiRequestSubspaceCapability(msg wgpstypes.MsgPaiRequestSubspaceCapability) []byte {
	HandleWidth := byte(utils.GetWidthMax64Int(msg.Data.Handle))

	var Header byte
	if HandleWidth == 1 {
		Header = 0xc
	} else if HandleWidth == 2 {
		Header = 0xd
	} else if HandleWidth == 4 {
		Header = 0xe
	} else {
		Header = 0xf
	}

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, byte(utils.GetWidthMax64Int(msg.Data.Handle)))

	return Result
}

func EncodePaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature any](
	msg wgpstypes.MsgPaiReplySubspaceCapability[SubspaceCapability, SyncSubspaceSignature],
	encodeSubspaceCapability func(capability SubspaceCapability) []byte,
	encodeSubspaceSignature func(signature SyncSubspaceSignature) []byte,
) []byte {
	HandleWidth := byte(utils.GetWidthMax64Int(msg.Data.Handle))

	var Header byte
	if HandleWidth == 1 {
		Header = 0x10
	} else if HandleWidth == 2 {
		Header = 0x11
	} else if HandleWidth == 4 {
		Header = 0x12
	} else {
		Header = 0x13
	}

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, byte(utils.GetWidthMax64Int(msg.Data.Handle)))
	Result = append(Result, encodeSubspaceCapability(msg.Data.Capability)...)
	Result = append(Result, encodeSubspaceSignature(msg.Data.Signature)...)

	return Result
}
