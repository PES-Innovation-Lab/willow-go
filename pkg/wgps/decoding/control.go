package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func DecodeChannelFromBeginningOfByte(bytes int) wgpstypes.Channel {
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

func DecodeChannelFromEndOfByte(bytes int) wgpstypes.Channel {
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
	width1 := bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(width1[0]))

	width2 := bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(width2[1]))

	width3 := bytes.NextAbsolute(2 + CompactWidth)
	Amount, _ := utils.DecodeIntMax64(width3[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlIssueGuarantee{
		Kind: wgpstypes.ControlAbsolve,
		Data: wgpstypes.ControlIssueGuaranteeData{
			Channel: Channel,
			Amount:  uint64(Amount),
		},
	}
}

func DecodeControlAbsolve(bytes *utils.GrowingBytes) wgpstypes.MsgControlAbsolve {
	width1 := bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(width1[0]))

	width2 := bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(width2[1]))

	width3 := bytes.NextAbsolute(2 + CompactWidth)
	Amount, _ := utils.DecodeIntMax64(width3[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlAbsolve{
		Kind: wgpstypes.ControlAbsolve,
		Data: wgpstypes.ControlAbsolveData{
			Channel: Channel,
			Amount:  uint64(Amount),
		},
	}
}

func DecodeControlPlead(bytes *utils.GrowingBytes) wgpstypes.MsgControlPlead {
	width1 := bytes.NextAbsolute(1)
	CompactWidth := CompactWidthFromEndOfByte(int(width1[0]))

	width2 := bytes.NextAbsolute(2)
	Channel := DecodeChannelFromBeginningOfByte(int(width2[1]))

	width3 := bytes.NextAbsolute(2 + CompactWidth)
	Target, _ := utils.DecodeIntMax64(width3[2 : 2+CompactWidth]) //TODO: Need to see why this is not defined anywhere
	bytes.Prune(2 + CompactWidth)

	return wgpstypes.MsgControlPlead{
		Kind: wgpstypes.ControlPlead,
		Data: wgpstypes.ControlPleadData{
			Channel: Channel,
			Target:  uint64(Target),
		},
	}
}

func DecodeControlAnnounceDropping(bytes *utils.GrowingBytes) wgpstypes.MsgControlAnnounceDropping {
	width1 := bytes.NextAbsolute(1)

	Channel := DecodeChannelFromEndOfByte(int(width1[0]))

	bytes.Prune(1)

	return wgpstypes.MsgControlAnnounceDropping{
		Kind: wgpstypes.ControlAnnounceDropping,
		Data: wgpstypes.ControlAnnounceDroppingData{
			Channel: Channel,
		},
	}
}

func DecodeControlApologise(bytes *utils.GrowingBytes) wgpstypes.MsgControlApologise {
	width1 := bytes.NextAbsolute(1)

	Channel := DecodeChannelFromEndOfByte(int(width1[0]))

	bytes.Prune(1)

	return wgpstypes.MsgControlApologise{
		Kind: wgpstypes.ControlApologise,
		Data: wgpstypes.ControlApologiseData{
			Channel: Channel,
		},
	}
}

func DecodeControlFree(bytes *utils.GrowingBytes) wgpstypes.MsgControlFree {
	width1 := bytes.NextAbsolute(1)

	CompactWidth := CompactWidthFromEndOfByte(int(width1[0]))

	width2 := bytes.NextAbsolute(2)

	HandleType := DecodeHandleTypeFromBeginningOfByte(int(width2[1]))

	width3 := bytes.NextAbsolute(2 + CompactWidth)

	Handle, _ := utils.DecodeIntMax64(width3[2 : 2+CompactWidth])

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
