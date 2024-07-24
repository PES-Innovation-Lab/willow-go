package handlestore

import (
	"fmt"
)

func NewMap() map[uint64]HandleStoreTriple {

	var data = make(map[uint64]HandleStoreTriple)
	return data
}

// Assuming ValueType is a generic type that is ordered, as per your file excerpt.

type HandleStoreTriple struct {
	Value           any
	AskedToFree     bool
	MessageRefCount int
}

type HandleStore struct {
	LeastUnassignedHandle uint64
	/** A map of handles (numeric IDs) to a triple made up of:
	 * - The bound data
	 * - Whether we've asked to free that data (and in doing so committing to no longer using it)
	 * - The number of unprocessed messages which refer to this handle. */
	Map map[uint64]HandleStoreTriple
}

/** Indicates whether this a store of handles we have bound, or a store of handles bound by another peer. */
// private isOurs: boolean;

func (s HandleStore) Get(handle uint64) (any, bool) {
	value, found := s.Map[handle]
	return value.Value, found

}

/** Bind some data to a handle. */
func (s HandleStore) Bind(value any) uint64 {
	handle := s.LeastUnassignedHandle
	s.Map[handle] = HandleStoreTriple{Value: value, AskedToFree: false, MessageRefCount: 0}
	s.LeastUnassignedHandle++
	return handle
}

func (s *HandleStore) Update(handle uint64, value any) error {
	triple, found := s.Map[handle]
	if !found {
		return fmt.Errorf("handle not found")
	}
	s.Map[handle] = HandleStoreTriple{Value: value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount}
	return nil
}

func (s *HandleStore) CanUse(handle uint64) bool {
	triple, found := s.Map[handle]
	return found && !triple.AskedToFree
}

func (s *HandleStore) Free(handle uint64) error {
	triple, found := s.Map[handle]
	if !found {
		return fmt.Errorf("no handle found to free")
	}
	if triple.MessageRefCount == 0 {
		delete(s.Map, handle)

	} else {
		s.Map[handle] = HandleStoreTriple{Value: triple.Value, AskedToFree: true, MessageRefCount: triple.MessageRefCount}

	}
	return nil
}

func (s *HandleStore) IncrementMessageRefCount(handle uint64) error {
	triple, found := s.Map[handle]
	if !found {
		return fmt.Errorf("no handle found to increment")
	}
	s.Map[handle] = HandleStoreTriple{Value: triple.Value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount + 1}
	return nil
}

func (s *HandleStore) DecrementMessageRefCount(handle uint64) error {
	triple, found := s.Map[handle]
	if !found {
		return fmt.Errorf("no handle found to increment")
	}
	if (triple.AskedToFree) && (triple.MessageRefCount-1 == 0) {
		delete(s.Map, handle)
	} else {
		s.Map[handle] = HandleStoreTriple{Value: triple.Value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount - 1}
	}
	return nil
}

/* type EventuallyMap[ValueType any] struct {
	data map[int64]ValueType
}

func NewEventuallyMap[ValueType constraints.Ordered]() *EventuallyMap[ValueType] {
	return &EventuallyMap[ValueType]{
		data: make(map[int64]ValueType),
	}
}
*/
