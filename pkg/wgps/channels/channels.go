package channels

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
)

var MsgLogicalChannels = map[wgpstypes.MsgKind]wgpstypes.Channel{
	wgpstypes.PaiBindFragment:                wgpstypes.IntersectionChannel,
	wgpstypes.PaiReplyFragment:               wgpstypes.IntersectionChannel,
	wgpstypes.SetupBindReadCapability:        wgpstypes.CapabilityChannel,
	wgpstypes.SetupBindAreaOfInterest:        wgpstypes.AreaOfInterestChannel,
	wgpstypes.SetupBindStaticToken:           wgpstypes.StaticTokenChannel,
	wgpstypes.ReconciliationSendFingerprint:  wgpstypes.ReconciliationChannel,
	wgpstypes.ReconciliationAnnounceEntries:  wgpstypes.ReconciliationChannel,
	wgpstypes.ReconciliationSendEntry:        wgpstypes.ReconciliationChannel,
	wgpstypes.ReconciliationSendPayload:      wgpstypes.ReconciliationChannel,
	wgpstypes.ReconciliationTerminatePayload: wgpstypes.ReconciliationChannel,
	wgpstypes.DataSendEntry:                  wgpstypes.DataChannel,
	wgpstypes.DataSendPayload:                wgpstypes.DataChannel,
	wgpstypes.DataReplyPayload:               wgpstypes.DataChannel,
	wgpstypes.DataBindPayloadRequest:         wgpstypes.PayloadRequestChannel,
	wgpstypes.CommitmentReveal:               0,
	wgpstypes.ControlAbsolve:                 0,
	wgpstypes.ControlAnnounceDropping:        0,
	wgpstypes.ControlApologise:               0,
	wgpstypes.ControlFree:                    0,
	wgpstypes.ControlPlead:                   0,
	wgpstypes.ControlIssueGuarantee:          0,
	wgpstypes.PaiRequestSubspaceCapability:   0,
	wgpstypes.PaiReplySubspaceCapability:     0,
	wgpstypes.DataSetMetadata:                0,
}
