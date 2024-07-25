package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func DecodeSetupBindReadCapability[ReadCapability, SyncSigntature any, ValueType constraints.Unsigned](bytes *utils.GrowingBytes, readCapScheme wgpstypes.ReadCapEncodingScheme[ReadCapability, ValueType], gerPrivy func(handle uint64) wgpstypes.ReadCapPrivy, decodeSignature func(bytes *utils.GrowingBytes) SyncSigntature) wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSigntature] { //need to fix the type definition
	bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(1 + CompactWidth)

	Handle, _ := utils.DecodeIntMax64(bytes.Array[1 : 1+CompactWidth])

	bytes.Prune(1 + CompactWidth)

	Capability, _ := readCapScheme.DecodeStream(bytes) //where is this defined?

	Signature := decodeSignature(bytes)

	return wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSigntature]{
		Kind: wgpstypes.SetupBindReadCapability,
		Data: wgpstypes.MsgSetupBindReadCapabilityData[ReadCapability, SyncSigntature]{
			Handle:     Handle,
			Capability: Capability,
			Signature:  Signature,
		},
	}
}

func DecodeSetupBindAreaOfInterest[ValueType constraints.Unsigned](bytes *utils.GrowingBytes, getPrivy func(handle uint64) types.Area, decodeStreamSubspace utils.EncodingScheme[ValueType], pathScheme types.PathParams[ValueType]) wgpstypes.MsgSetupBindAreaOfInterest { //need to check the types out once
	bytes.NextAbsolute(1)
	HasALimit := (0x4 & bytes.Array[0]) == 0x4

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(1 + CompactWidth)

	AuthHandle, _ := utils.DecodeIntMax64(bytes.Array[1 : 1+CompactWidth])

	bytes.Prune(1 + CompactWidth)

	Outer := getPrivy(uint64(AuthHandle))

	Area, _ := utils.DecodeStreamAreaInArea[ValueType](utils.DecodeStreamAreaInAreaOptions[ValueType]{
		PathScheme:           pathScheme,
		DecodeStreamSubspace: utils.EncodingScheme[types.SubspaceId]{},
	}, bytes, Outer)

	if !HasALimit {
		return wgpstypes.MsgSetupBindAreaOfInterest{
			Kind: wgpstypes.SetupBindAreaOfInterest,
			Data: wgpstypes.MsgSetupBindAreaOfInterestData{
				AreaOfInterest: types.AreaOfInterest{
					Area:     Area,
					MaxCount: 0,
					MaxSize:  0,
				}, //just check this out once later
				Authorisation: uint64(AuthHandle),
			},
		}
	}

	bytes.NextAbsolute(1)

	Maxes := bytes.Array[0]

	CompactWidthCount := CompactWidthFromEndOfByte(int(Maxes >> 6))
	CompactWidthSize := CompactWidthFromEndOfByte(int(Maxes >> 4))

	bytes.NextAbsolute(1 + CompactWidthCount + CompactWidthSize)

	MaxCount, _ := utils.DecodeIntMax64(bytes.Array[1 : 1+CompactWidthCount])

	MaxSize, _ := utils.DecodeIntMax64(bytes.Array[1+CompactWidthCount : 1+CompactWidthCount+CompactWidthSize])

	bytes.Prune(1 + CompactWidthCount + CompactWidthSize)

	return wgpstypes.MsgSetupBindAreaOfInterest{
		Kind: wgpstypes.SetupBindAreaOfInterest,
		Data: wgpstypes.MsgSetupBindAreaOfInterestData{
			AreaOfInterest: types.AreaOfInterest{
				Area:     Area,
				MaxCount: MaxCount,
				MaxSize:  uint64(MaxSize),
			}, //just check this out once later
			Authorisation: uint64(AuthHandle),
		},
	}
}

func DecodeSetupBindStaticToken[StaticToken string](bytes *utils.GrowingBytes, decodeStaticToken func(bytes *utils.GrowingBytes) StaticToken) wgpstypes.MsgSetupBindStaticToken[StaticToken] {
	bytes.NextAbsolute(1)

	bytes.Prune(1)

	staticToken := decodeStaticToken(bytes)

	return wgpstypes.MsgSetupBindStaticToken[StaticToken]{
		Kind: wgpstypes.SetupBindStaticToken,
		Data: wgpstypes.MsgSetupBindStaticTokenData[StaticToken]{
			StaticToken: staticToken,
		},
	}
}
