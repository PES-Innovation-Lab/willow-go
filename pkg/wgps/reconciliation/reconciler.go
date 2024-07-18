package reconciliation

import (
	"sync"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type ReconcilerOpts[
	PreFingerPrint, FingerPrint constraints.Ordered,
	AuthorisationToken string,
	NamespaceId types.NamespaceId,
	SubspaceId types.SubspaceId,
	PayloadDigest types.PayloadDigest,
	AuthorisationOpts []byte,
	K constraints.Unsigned] struct {
	Role              wgpstypes.SyncRole
	SubspaceScheme    datamodeltypes.SubspaceScheme
	FingerPrintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
	Namespace         types.NamespaceId
	AoiOurs           types.AreaOfInterest
	AoiTheirs         types.AreaOfInterest
	Store             *store.Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
}

const SEND_ENTRIES_THRESHOLD = 8

type Reconciler[NamespaceId types.NamespaceId,
	SubspaceId types.SubspaceId,
	PayloadDigest types.PayloadDigest,
	K constraints.Unsigned,
	AuthorisationOpts []byte,
	AuthorisationToken string,
	PreFingerprint, Fingerprint constraints.Ordered] struct {
	SubspaceScheme    datamodeltypes.SubspaceScheme
	FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerprint, Fingerprint]
	Store             *store.Store[PreFingerprint, Fingerprint, K, AuthorisationOpts, AuthorisationToken]
	FingerPrintQueue  chan struct {
		Range       types.Range3d
		FingerPrint Fingerprint
		Covers      uint64
	}
	AnnounceQueue chan struct {
		Range        types.Range3d
		Count        int
		WantResponse bool
		Covers       uint64
	}
	Ranges chan types.Range3d
}

func NewReconciler[PreFingerPrint, FingerPrint constraints.Ordered,
	AuthorisationToken string,
	NamespaceId types.NamespaceId,
	SubspaceId types.SubspaceId,
	PayloadDigest types.PayloadDigest,
	AuthorisationOpts []byte,
	K constraints.Unsigned](opts *ReconcilerOpts[PreFingerPrint, FingerPrint, AuthorisationToken, NamespaceId, SubspaceId, PayloadDigest, AuthorisationOpts, K],
) *Reconciler[NamespaceId, SubspaceId, PayloadDigest, K, AuthorisationOpts, AuthorisationToken, PreFingerPrint, FingerPrint] {

	newReconciler := &Reconciler[NamespaceId, SubspaceId, PayloadDigest, K, AuthorisationOpts, AuthorisationToken, PreFingerPrint, FingerPrint]{
		SubspaceScheme:    opts.SubspaceScheme,
		FingerprintScheme: opts.FingerPrintScheme,
		Store:             opts.Store,
		FingerPrintQueue: make(chan struct {
			Range       types.Range3d
			FingerPrint FingerPrint
			Covers      uint64
		}, SEND_ENTRIES_THRESHOLD),
		AnnounceQueue: make(chan struct {
			Range        types.Range3d
			Count        int
			WantResponse bool
			Covers       uint64
		}, SEND_ENTRIES_THRESHOLD),

		Ranges: make(chan types.Range3d, 100),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		newReconciler.DetermineRange(opts.AoiOurs, opts.AoiTheirs)
	}()
	if wgpstypes.IsAlfie(opts.Role) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newReconciler.Initiate()
		}()
		wg.Wait()
	}
	return newReconciler
}

func (r *Reconciler[NamespaceId, SubspaceId, PayloadDigest, K, AuthorisationOpts, AuthorisationToken, PreFingerPrint, FingerPrint]) DetermineRange(
	aoi1, aoi2 types.AreaOfInterest,
) error {
	range1 := r.Store.AreaOfInterestToRange(aoi1)
	range2 := r.Store.AreaOfInterestToRange(aoi2)
}
