package tests

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type AnnouncerOpts[AuthorisationToken, StaticToken, DynamicToken string, ValueType any] struct {
	AuthorisationTokenScheme   wgpstypes.AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken]
	PayloadScheme              datamodeltypes.PayloadScheme
	StaticTokenHandleStoreOurs handlestore.HandleStore[StaticToken] //need to check this out
}

type AnnouncementPack[StaticToken, DynamicToken string] struct {
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

type Announcer[PreFingerPrint, FingerPrint string, ValueType any, StaticToken, DynamicToken string, AuthorisationOpts []byte, AuthorisationToken string, K constraints.Unsigned] struct {
	AuthorisationTokenScheme   wgpstypes.AuthorisationTokenScheme[AuthorisationToken, StaticToken, DynamicToken]
	PayloadScheme              datamodeltypes.PayloadScheme
	StaticTokenHandleStoreOurs handlestore.HandleStore[StaticToken]
	StaticTokenHandleMap       map[string]uint64
	AnnouncementPackQueue      chan AnnouncementPack[StaticToken, DynamicToken]
}

func NewAnnouncer[PreFingerPrint, FingerPrint string, ValueType any, StaticToken, DynamicToken string, AuthorisationOpts []byte, AuthorisationToken string, K constraints.Unsigned](
	opts AnnouncerOpts[AuthorisationToken, StaticToken,
		DynamicToken, ValueType]) *Announcer[PreFingerPrint, FingerPrint, ValueType, StaticToken, DynamicToken, AuthorisationOpts, AuthorisationToken, K] {

	return &Announcer[PreFingerPrint, FingerPrint, ValueType, StaticToken, DynamicToken, AuthorisationOpts, AuthorisationToken, K]{
		AuthorisationTokenScheme:   opts.AuthorisationTokenScheme,
		PayloadScheme:              opts.PayloadScheme,
		StaticTokenHandleStoreOurs: opts.StaticTokenHandleStoreOurs,
		StaticTokenHandleMap:       make(map[string]uint64),
		AnnouncementPackQueue:      make(chan AnnouncementPack[StaticToken, DynamicToken]),
	}

}

func (a *Announcer[PreFingerPrint, FingerPrint, ValueType, StaticToken, DynamicToken, AuthorisationOpts, AuthorisationToken, K]) GetStaticTokenHandle(
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
	newHandle := a.StaticTokenHandleStoreOurs.Bind(staticToken)
	a.StaticTokenHandleMap[string(encoded)] = newHandle
	return returnStruct{
		Handle:         newHandle,
		AlreadyExisted: false,
	}, nil
}

func (a *Announcer[PreFingerPrint, FingerPrint, ValueType, StaticToken, DynamicToken, AuthorisationOpts, AuthorisationToken, K]) QueueAnnounce(
	announcement struct {
		SenderHandle   uint64
		ReceiverHandle uint64
		KDStore        *datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]
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

	//TODO: Implement QueryRange in Store
	results := announcement.KDStore.Query(announcement.Range)

	for _, result := range results {

		timestamp := result.Timestamp
		path := result.Path
		subspace := result.Subspace

		SubspaceName := string(subspace)

		staticToken, dynamicToken := a.AuthorisationTokenScheme.DecomposeAuthToken(AuthorisationToken(SubspaceName))

		TokenHandle, _ := a.GetStaticTokenHandle(staticToken)

		staticTokenHandle := TokenHandle.Handle
		staticTokenHandleAlreadyExisted := TokenHandle.AlreadyExisted

		if !staticTokenHandleAlreadyExisted {
			staticTokenBinds = append(staticTokenBinds, staticToken)
		}

		entries = append(entries, struct {
			LengthyEntry      datamodeltypes.LengthyEntry
			StaticTokenHandle uint64
			DynamicToken      DynamicToken
		}{
			LengthyEntry:      entries, //need to change this later
			StaticTokenHandle: staticTokenHandle,
			DynamicToken:      dynamicToken,
		})
	}

	a.AnnouncementPackQueue <- AnnouncementPack[StaticToken, DynamicToken]{
		StaticTokenBinds: staticTokenBinds,
		Announcement: struct {
			Range          types.Range3d
			Count          int
			WantResponse   bool
			SenderHandle   uint64
			ReceiverHandle uint64
			Covers         uint64
		}{
			Range:          announcement.Range,
			Count:          len(entries),
			WantResponse:   announcement.WantResponse,
			SenderHandle:   announcement.SenderHandle,
			ReceiverHandle: announcement.ReceiverHandle,
			Covers:         announcement.Covers,
		},
		Entries: entries,
	}
}

func (a *Announcer[PreFingerPrint, FingerPrint, ValueType, StaticToken, DynamicToken, AuthorisationOpts, AuthorisationToken, K]) announcementPacks() <-chan AnnouncementPack[StaticToken, DynamicToken] {
	out := make(chan AnnouncementPack[StaticToken, DynamicToken])
	go func() {
		defer close(out)
		for pack := range a.AnnouncementPackQueue {
			out <- pack
		}
	}()
	return out
}
