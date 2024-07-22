package reconciliation

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type AnnouncerOpts[AuthorisationToken string, StaticToken, DynamicToken, ValueType constraints.Ordered] struct {
	AuthorisationTokenScheme   wgpstypes.AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken]
	PayloadScheme              datamodeltypes.PayloadScheme
	StaticTokenHandleStoreOurs wgps.HandleStore[ValueType]
}

type AnnouncementPack[StaticToken, DynamicToken constraints.Ordered] struct {
	StaticTokenBinds []StaticToken

	// Then send a ReconciliationAnnounceEntries
	Announcement struct {
		Range          types.Range3d
		Count          int
		WantResponse   bool
		SenderHandle   uint64
		ReceiverHandle uint64
		Covers         uint64
	}
	// Then send many ReconciliationSendEntry
	Entries []struct {
		LengthyEntry      datamodeltypes.LengthyEntry
		StaticTokenHandle uint64
		DynamicToken      DynamicToken
	}
}

type Announcer[PreFingerPrint, FingerPrint, StaticToken, DynamicToken, ValueType constraints.Ordered, Authorisationopts []byte, AuthorisationToken string, K constraints.Unsigned] struct {
	AuthorisationTokenScheme   wgpstypes.AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken]
	PayloadScheme              datamodeltypes.PayloadScheme
	StaticTokenHandleStoreOurs wgps.HandleStore[ValueType]
	StaticTokenHandleMap       map[string]uint64
	AnnouncementPackQueue      chan AnnouncementPack[StaticToken, DynamicToken]
}

func NewAnnouncer[AuthorisationOpts []byte,
	AuthorisationToken string, PreFingerPrint,
	FingerPrint, StaticToken, DynamicToken,
	ValueType constraints.Ordered, K constraints.Unsigned](
	opts AnnouncerOpts[AuthorisationToken, StaticToken,
		DynamicToken, ValueType]) *Announcer[PreFingerPrint,
	FingerPrint, StaticToken, DynamicToken, ValueType,
	AuthorisationOpts, AuthorisationToken, K] {

	return &Announcer[PreFingerPrint, FingerPrint, StaticToken, DynamicToken, ValueType, AuthorisationOpts, AuthorisationToken, K]{
		AuthorisationTokenScheme:   opts.AuthorisationTokenScheme,
		PayloadScheme:              opts.PayloadScheme,
		StaticTokenHandleStoreOurs: opts.StaticTokenHandleStoreOurs,
		StaticTokenHandleMap:       make(map[string]uint64),
		AnnouncementPackQueue:      make(chan AnnouncementPack[StaticToken, DynamicToken]),
	}

}

func (a *Announcer[PreFingerPrint, FingerPrint, StaticToken, DynamicToken, ValueType, AuthorisationOpts, AuthorisationToken, K]) GetStaticTokenHandle(
	staticToken StaticToken,
) (struct {
	Handle         uint64
	AlreadyExisted bool
}, error) {
	type returnStruct struct {
		Handle         uint64
		AlreadyExisted bool
	}
	encoded := a.AuthorisationTokenScheme.Encodings.StaticToken.Encode(staticToken)
	existingHandle, ok := a.StaticTokenHandleMap[string(encoded)]
	if ok {
		canUse := a.StaticTokenHandleStoreOurs.CanUse(existingHandle)
		if !canUse {
			return returnStruct{}, fmt.Errorf("Could not use a static token handle")
		}
		return returnStruct{
			Handle:         existingHandle,
			AlreadyExisted: true,
		}, nil
	}
	newHandle := a.StaticTokenHandleStoreOurs.Bind(staticToken, encoded)
	a.StaticTokenHandleMap[string(encoded)] = newHandle
	return returnStruct{
		Handle:         newHandle,
		AlreadyExisted: false,
	}, nil
}

func (a *Announcer[PreFingerPrint, FingerPrint, StaticToken, DynamicToken, ValueType, AuthorisationOpts, AuthorisationToken, K]) QueueAnnounce(
	announcement struct {
		SenderHandle   uint64
		ReceiverHandle uint64
		Store          *store.Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
		Namespace      types.NamespaceId
		Range          types.Range3d
		WantResponse   bool
		Covers         uint64
	},
) {
	// Queue announcement message.
	staticTokenBinds := []StaticToken{}
	entries := []struct {
		LengthyEntry      datamodeltypes.LengthyEntry
		StaticTokenHandle uint64
		DynamicToken      DynamicToken
	}{}

}
