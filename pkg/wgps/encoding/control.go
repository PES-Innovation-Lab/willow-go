package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func ChannelMaskStart(mask int, channel wgpstypes.Channel) int {
	switch channel {
	case wgpstypes.ReconciliationChannel:
		return mask
	case wgpstypes.DataChannel:
		return mask | 0x20
	case wgpstypes.IntersectionChannel:
		return mask | 0x40
	case wgpstypes.CapabilityChannel:
		return mask | 0x60
	case wgpstypes.AreaOfInterestChannel:
		return mask | 0x80
	case wgpstypes.PayloadRequestChannel:
		return mask | 0xa0
	case wgpstypes.StaticTokenChannel:
		return mask | 0xc0
	default:
		return mask
	}
}

func ChannelMaskEnd(mask int, channel wgpstypes.Channel) int {
	switch channel {
	case wgpstypes.ReconciliationChannel:
		return mask
	case wgpstypes.DataChannel:
		return mask | 0x1
	case wgpstypes.IntersectionChannel:
		return mask | 0x2
	case wgpstypes.CapabilityChannel:
		return mask | 0x3
	case wgpstypes.AreaOfInterestChannel:
		return mask | 0x4
	case wgpstypes.PayloadRequestChannel:
		return mask | 0x5
	case wgpstypes.StaticTokenChannel:
		return mask | 0x6
	default:
		return mask
	}

}
func HandleMask(mask int, handleType wgpstypes.HandleType) int {
	switch handleType {
	case wgpstypes.IntersectionHandle:
		return mask
	case wgpstypes.CapabilityHandle:
		return mask | 0x20
	case wgpstypes.AreaOfInterestHandle:
		return mask | 0x40
	case wgpstypes.PayloadRequestHandle:
		return mask | 0x60
	case wgpstypes.StaticTokenHandle:
		return mask | 0x80
	default:
		return mask

	}
}

func EncodeControlIssueGuarantee(msg wgpstypes.MsgControlIssueGuarantee) []byte {

	amountWidth := utils.GetWidthMax64Int(msg.Data.Amount)
	var header uint

	switch amountWidth {
	case 1:
		header = 0x80
	case 2:
		header = 0x81
	case 4:
		header = 0x82
	default:
		header = 0x83
	}
	return append([]byte{byte(header), byte(ChannelMaskStart(0, msg.Data.Channel))},
		utils.EncodeIntMax64[uint64](msg.Data.Amount)...)

}

func EncodeControlAbsolve(msg wgpstypes.MsgControlAbsolve) []byte {
	amountWidth := utils.GetWidthMax64Int(msg.Data.Amount)
	var header uint

	switch amountWidth {
	case 1:
		header = 0x84
	case 2:
		header = 0x85
	case 4:
		header = 0x86
	default:
		header = 0x87
	}
	return append([]byte{byte(header), byte(ChannelMaskStart(0, msg.Data.Channel))},
		utils.EncodeIntMax64[uint64](msg.Data.Amount)...)

}

func EncodeControlPlead(msg wgpstypes.MsgControlPlead) []byte {
	targetWidth := utils.GetWidthMax64Int(msg.Data.Target)
	var header uint

	switch targetWidth {
	case 1:
		header = 0x88
	case 2:
		header = 0x89
	case 4:
		header = 0x8a
	default:
		header = 0x8b
	}
	return append([]byte{byte(header), byte(ChannelMaskStart(0, msg.Data.Channel))},
		utils.EncodeIntMax64[uint64](msg.Data.Target)...)
}

func EncodeControlAnnounceDropping(msg wgpstypes.MsgControlAnnounceDropping) []byte {
	return []byte{byte(ChannelMaskEnd(0x90, msg.Data.Channel))}
}

func EncodeControlApologise(msg wgpstypes.MsgControlApologise) []byte {
	return []byte{byte(ChannelMaskEnd(0x98, msg.Data.Channel))}
}

func EncodeControlFree(msg wgpstypes.MsgControlFree) []byte {

	handleWidth := utils.GetWidthMax64Int(msg.Data.Handle)
	var header uint
	switch handleWidth {
	case 1:
		header = 0x8c
	case 2:
		header = 0x8d
	case 4:
		header = 0x8e
	default:
		header = 0x8f
	}

	var handleTypeByte int
	if msg.Data.Mine {
		handleTypeByte = HandleMask(0, msg.Data.HandleType) | 0x10

	} else {
		handleTypeByte = HandleMask(0, msg.Data.HandleType)
	}
	return append([]byte{byte(header), byte(handleTypeByte)}, utils.EncodeIntMax64(msg.Data.Handle)...)

}
