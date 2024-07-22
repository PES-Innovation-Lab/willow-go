package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func DecodeChannelFromBeginningOfByte(bytes int) wgpstypes.LogicalChannel {
	if (bytes & 0xc0) == 0xc0 {
		return wgpstypes.StaticTokenChannel
	} else if (bytes & 0xa0) == 0xa0 {
		return wgpstypes.PayloadRequestChannel
	} else if (bytes & 0x80) == 0x80 {
		return wgpstypes.AreaOfInterestChannel
	} else if (bytes & 0x60) == 0x60 {
		return wgpstypes.CapabilityChannel
	} else if (bytes & 0x40) == 0x40 {
		return wgpstypes.IntersectionChannel
	} else if (bytes & 0x20) == 0x20 {
		return wgpstypes.DataChannel
	} else {
		return wgpstypes.ReconciliationChannel
	}
}

func DecodeChannelFromEndOfByte(bytes int) wgpstypes.LogicalChannel {
	if (bytes & 0x6) == 0x6 {
		return wgpstypes.StaticTokenChannel
	} else if (bytes & 0x5) == 0x5 {
		return wgpstypes.PayloadRequestChannel
	} else if (bytes & 0x4) == 0x4 {
		return wgpstypes.AreaOfInterestChannel
	} else if (bytes & 0x3) == 0x3 {
		return wgpstypes.CapabilityChannel
	} else if (bytes & 0x2) == 0x2 {
		return wgpstypes.IntersectionChannel
	} else if (bytes & 0x1) == 0x1 {
		return wgpstypes.DataChannel
	} else {
		return wgpstypes.ReconciliationChannel
	}
}

func DecodeHandleTypeFromBeginningOfByte(bytes int) wgpstypes.HandleType {
	if (bytes & 0x80) == 0x80 {
		return wgpstypes.StaticTokenHandle
	} else if (bytes & 0x60) == 0x60 {
		return wgpstypes.PayloadRequestHandle
	} else if (bytes & 0x40) == 0x40 {
		return wgpstypes.AreaOfInterestHandle
	} else if (bytes & 0x20) == 0x20 {
		return wgpstypes.CapabilityHandle
	} else {
		return wgpstypes.IntersectionHandle
	}
}

func DecodeControlIssueGuarantee(bytes *utils.GrowingBytes) wgpstypes.MsgControlIssueGuarantee {
	bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(bytes.Array[1]))

	bytes.NextAbsolute(2 + CompactWidth)
	Amount := types.DecodeCompactWidth(bytes.Array[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlIssueGuarantee{
		Kind: wgpstypes.ControlAbsolve,
		Data: wgpstypes.MsgControlIssueGuaranteeData{
			Channel: Channel,
			Amount:  uint64(Amount),
		},
	}
}

func DecodeControlAbsolve(bytes *utils.GrowingBytes) wgpstypes.MsgControlAbsolve {
	bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(bytes.Array[1]))

	bytes.NextAbsolute(2 + CompactWidth)
	Amount := types.DecodeCompactWidth(bytes.Array[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlAbsolve{
		Kind: wgpstypes.ControlAbsolve,
		Data: wgpstypes.MsgControlAbsolveData{
			Channel: Channel,
			Amount:  uint64(Amount),
		},
	}
}

func DecodeControlPlead(bytes *utils.GrowingBytes) wgpstypes.MsgControlPlead {
	bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(bytes.Array[1]))

	bytes.NextAbsolute(2 + CompactWidth)
	Target := types.DecodeCompactWidth(bytes.Array[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlPlead{
		Kind: wgpstypes.ControlPlead,
		Data: wgpstypes.MsgControlPleadData{
			Channel: Channel,
			Target:  uint64(Target),
		},
	}
}

func DecodeControlAnnounceDropping(bytes *utils.GrowingBytes) wgpstypes.MsgControlAnnounceDropping {
	bytes.NextAbsolute(1)

	Channel := DecodeChannelFromEndOfByte(int(bytes.Array[0]))

	bytes.Prune(1)

	return wgpstypes.MsgControlAnnounceDropping{
		Kind: wgpstypes.ControlAnnounceDropping,
		Data: wgpstypes.MsgControlAnnounceDroppingData{
			Channel: Channel,
		},
	}
}

func DecodeControlApologise(bytes *utils.GrowingBytes) wgpstypes.MsgControlApologise {
	bytes.NextAbsolute(1)

	Channel := DecodeChannelFromEndOfByte(int(bytes.Array[0]))

	bytes.Prune(1)

	return wgpstypes.MsgControlApologise{
		Kind: wgpstypes.ControlApologise,
		Data: wgpstypes.MsgControlApologiseData{
			Channel: Channel,
		},
	}
}

func DecodeControlFree(bytes *utils.GrowingBytes) wgpstypes.MsgControlFree {
	bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(bytes.Array[0]))

	bytes.NextAbsolute(2)

	HandleType := DecodeHandleTypeFromBeginningOfByte(int(bytes.Array[1]))

	bytes.NextAbsolute(2 + CompactWidth)

	Handle := types.DecodeCompactWidth(bytes.Array[2 : 2+CompactWidth])

	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlFree{
		Kind: wgpstypes.ControlFree,
		Data: wgpstypes.MsgControlFreeData{
			HandleType: HandleType,
			Handle:     uint64(Handle),
			Mine:       true,
		},
	}

}
