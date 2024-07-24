package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func EncodeSetupBindReadCapability[ReadCapability, SyncSignature any, ValueType constraints.Unsigned](
	msg wgpstypes.MsgSetupBindReadCapability[ReadCapability, SyncSignature],
	encodeReadCapability wgpstypes.ReadCapEncodingScheme[ReadCapability, ValueType],
	encodeSignature func(value SyncSignature) []byte,
	privy wgpstypes.ReadCapPrivy) []byte {
	HandleWidth := utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Handle))

	Header := byte(CompactWidthOr(0x20, HandleWidth))

	Encoder := encodeReadCapability.Encode(msg.Data.Capability, privy)

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Handle))...)
	Result = append(Result, Encoder...)
	Result = append(Result, encodeSignature(msg.Data.Signature)...)

	return Result
}

func EncodeSetupBindAreaOfInterest[ValueType constraints.Unsigned](
	msg wgpstypes.MsgSetupBindAreaOfInterest,
	opts struct {
		Outer          types.Area
		PathScheme     types.PathParams[ValueType]
		EncodeSubspace func(subspace types.SubspaceId) []byte
		OrderSubspace  types.TotalOrder[types.SubspaceId]
	},
) []byte {
	var Value byte
	if byte(msg.Data.AreaOfInterest.MaxCount) != 0 || byte(msg.Data.AreaOfInterest.MaxSize) != 0 {
		Value = 0x4
	} else {
		Value = 0x0

	}
	Header := byte(CompactWidthOr(0x28, int(byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.Authorisation)))|Value)))

	AuthHandle := utils.EncodeIntMax64[ValueType](ValueType(msg.Data.Authorisation))

	AreaInArea := utils.EncodeAreaInArea(
		utils.EncodeAreaOpts[ValueType]{
			PathScheme:     opts.PathScheme,
			EncodeSubspace: opts.EncodeSubspace,
			OrderSubspace:  opts.OrderSubspace,
		},
		msg.Data.AreaOfInterest.Area,
		opts.Outer,
	)

	if msg.Data.AreaOfInterest.MaxCount == 0 && msg.Data.AreaOfInterest.MaxSize == 0 {
		return append([]byte{Header}, append(AuthHandle, AreaInArea...)...)
	}

	MaxCountMask := byte(CompactWidthOr(0, int(byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.AreaOfInterest.MaxCount))))))

	Shifted := byte(MaxCountMask << 2)

	MaxSizeMask := byte(CompactWidthOr(int(Shifted), int(byte(utils.GetWidthMax64Int[ValueType](ValueType(msg.Data.AreaOfInterest.MaxSize))))))

	LengthBytes := byte(MaxSizeMask << 4)

	var Result []byte
	Result = append(Result, Header)
	Result = append(Result, AuthHandle...)
	Result = append(Result, AreaInArea...)
	var Res []byte
	Res = append(Res, LengthBytes)
	Result = append(Result, Res...)
	Result = append(Result, utils.EncodeIntMax64[ValueType](ValueType(msg.Data.AreaOfInterest.MaxCount))...)
	Result = append(Result, utils.EncodeIntMax64[ValueType](ValueType(msg.Data.AreaOfInterest.MaxSize))...)

	return Result
}

func EncodeSetupBindStaticToken[StaticToken string](
	msg wgpstypes.MsgSetupBindStaticToken[StaticToken],
	encodeStaticToken func(token StaticToken) []byte,
) []byte {
	var Result []byte
	Result = append(Result, 0x30)
	Result = append(Result, encodeStaticToken(msg.Data.StaticToken)...)

	return Result
}
