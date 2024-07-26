package reconciliation

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type ReconcilerMap[K constraints.Unsigned, PreFingerPrint, FingerPrint string, AuthorisationOpts []byte, AuthorisationToken string] struct {
	Map map[uint64]map[uint64]Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]
}

func (r *ReconcilerMap[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]) AddReconciler(
	aoiHandleOurs, aoiHandleTheirs uint64, reconciler Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken],
) {
	existingInnerMap, ok := r.Map[aoiHandleOurs]
	if ok {
		existingInnerMap[aoiHandleTheirs] = reconciler
	}
	newInnerMap := make(map[uint64]Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken])
	newInnerMap[aoiHandleTheirs] = reconciler
	r.Map[aoiHandleOurs] = newInnerMap

}

func (r *ReconcilerMap[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]) GetReconciler(
	aoiHandleOurs, aoiHandleTheirs uint64,
) (re Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken], err error) {
	innerMap, ok := r.Map[aoiHandleOurs]
	var reconciler Reconciler[K, PreFingerPrint, FingerPrint, AuthorisationOpts, AuthorisationToken]
	if !ok {

		return reconciler, fmt.Errorf("Could not dereference one of our AOI handles to a reconciler")
	}
	reconciler, ok = innerMap[aoiHandleTheirs]
	if !ok {
		return reconciler, fmt.Errorf("Could not dereference one of their AOI handles to a reconciler")
	}
	return reconciler, nil
}
