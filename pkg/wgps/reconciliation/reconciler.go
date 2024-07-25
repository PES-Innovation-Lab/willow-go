package reconciliation

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type ReconcilerOpts[
	PreFingerPrint, FingerPrint constraints.Ordered,
	K constraints.Unsigned,
	AuthorisationOpts []byte, AuthorisationToken string] struct {
	Role              wgpstypes.SyncRole
	SubspaceScheme    datamodeltypes.SubspaceScheme
	FingerPrintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
	Namespace         types.NamespaceId
	AoiOurs           types.AreaOfInterest
	AoiTheirs         types.AreaOfInterest
	Store             *store.Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
}

const SEND_ENTRIES_THRESHOLD = 8

type Reconciler[
	K constraints.Unsigned,
	PreFingerprint, Fingerprint constraints.Ordered, AuthorisationOpts []byte, AuthorisationToken string] struct {
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

/*func NewReconciler[PreFingerPrint, FingerPrint constraints.Ordered,
	K constraints.Unsigned, AuthorisationOpts []byte, AuthorisationToken string](opts *ReconcilerOpts[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken],
) *Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken] {

	newReconciler := &Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]{
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
		newReconciler.DetermineRange(opts.AoiOurs, opts.AoiTheirs, opts.Role)
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
} */

func (r Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]) DetermineRange(
	aoi1, aoi2 types.AreaOfInterest, role wgpstypes.SyncRole,
) error {
	// Remove the interest from both.
	range1, _ := (*r.Store).AreaOfInterestToRange(aoi1)
	range2, _ := r.Store.AreaOfInterestToRange(aoi2)

	isIntersecting, intersection := utils.IntersectRange3d(
		r.SubspaceScheme.Order,
		range1,
		range2,
	)

	if !isIntersecting {
		return fmt.Errorf("There was no intersection between two range-ified AOIs. That shouldn't happen...")
	}
	if wgpstypes.IsAlfie(role) {
		r.Ranges <- intersection
	}
	return nil
}

/*func (r *Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]) Initiate() {
// Initialize the reconciliation process.
intersection := <-r.Ranges */
// TODO : Implement Summarise function in store
/*preFingerprint := r.Store.Summarise(intersection)
	finalised := r.FingerprintScheme.FingerPrintFinalise(preFingerprint)
	r.FingerPrintQueue <- struct {
		Range       types.Range3d
		FingerPrint FingerPrint
		Covers      uint64
	}{
		Range:       intersection,
		FingerPrint: finalised,
	}
}

func (r *Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]) Respond(
	yourRange types.Range3d,
	fingerprint FingerPrint,
	yourRangeCounter int,

) {
	// TODO Implement Summarise function in store
	ourFingerprint, size := r.Store.Summarise(yourRange)

	fingerprintOursFinal := r.FingerprintScheme.FingerPrintFinalise(ourFingerprint)
	if r.FingerprintScheme.IsEqual(fingerprint, fingerprintOursFinal) {
		r.AnnounceQueue <- struct {
			Range        types.Range3d
			Count        int
			WantResponse bool
			Covers       uint64
		}{
			Range:        yourRange,
			Count:        0,
			WantResponse: false,
			Covers:       uint64(yourRangeCounter),
		}
	} else if size <= SEND_ENTRIES_THRESHOLD {
		r.AnnounceQueue <- struct {
			Range        types.Range3d
			Count        int
			WantResponse bool
			Covers       uint64
		}{
			Range:        yourRange,
			Count:        size,
			WantResponse: true,
			Covers:       uint64(yourRangeCounter),
		}
		return
	} else {
		// TODO: Implement Store Split Range
		left, right := r.Store.SplitRange(yourRange, size)
		fingerprintLeftFinal := r.FingerprintScheme.FingerPrintFinalise(r.Store.Summarise(left))
		fingerprintRightFinal := r.FingerprintScheme.FingerPrintFinalise(r.Store.Summarise(right))
		r.FingerPrintQueue <- struct {
			Range       types.Range3d
			FingerPrint FingerPrint
			Covers      uint64
		}{
			Range:       left,
			FingerPrint: fingerprintLeftFinal,
		}
		r.FingerPrintQueue <- struct {
			Range       types.Range3d
			FingerPrint FingerPrint
			Covers      uint64
		}{
			Range:       right,
			FingerPrint: fingerprintRightFinal,
			Covers:      uint64(yourRangeCounter),
		}

	}

}
*/
