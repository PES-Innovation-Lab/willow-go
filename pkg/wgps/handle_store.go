package wgps

import (
	"fmt"
)

func NewMap[ValueType any]() *map[uint64]HandleStoreTriple[ValueType] {

	var data = make(map[uint64]HandleStoreTriple[ValueType])
	return &data
}

// Assuming ValueType is a generic type that is ordered, as per your file excerpt.

type HandleStoreTriple[ValueType any] struct {
	Value           ValueType
	AskedToFree     bool
	MessageRefCount int
}

type HandleStore[ValueType any] struct {
	LeastUnassignedHandle uint64
	/** A map of handles (numeric IDs) to a triple made up of:
	 * - The bound data
	 * - Whether we've asked to free that data (and in doing so committing to no longer using it)
	 * - The number of unprocessed messages which refer to this handle. */
	Map *map[uint64]HandleStoreTriple[ValueType]
}

/** Indicates whether this a store of handles we have bound, or a store of handles bound by another peer. */
// private isOurs: boolean;

func (s *HandleStore[ValueType]) Get(handle uint64) (ValueType, bool) {
	value, found := (*s.Map)[handle]
	return value.Value, found

}

/** Bind some data to a handle. */
func (s *HandleStore[ValueType]) Bind(value ValueType) uint64 {
	handle := s.LeastUnassignedHandle
	(*s.Map)[handle] = HandleStoreTriple[ValueType]{Value: value, AskedToFree: false, MessageRefCount: 0}
	s.LeastUnassignedHandle++
	return handle
}

func (s *HandleStore[ValueType]) Update(handle uint64, value ValueType) error {
	triple, found := (*s.Map)[handle]
	if !found {
		return fmt.Errorf("handle not found")
	}
	(*s.Map)[handle] = HandleStoreTriple[ValueType]{Value: value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount}
	return nil
}

func (s *HandleStore[ValueType]) CanUse(handle uint64) bool {
	triple, found := (*s.Map)[handle]
	return found && !triple.AskedToFree
}

func (s *HandleStore[ValueType]) Free(handle uint64) error {
	triple, found := (*s.Map)[handle]
	if !found {
		return fmt.Errorf("no handle found to free")
	}
	if triple.MessageRefCount == 0 {
		delete((*s.Map), handle)

	} else {
		(*s.Map)[handle] = HandleStoreTriple[ValueType]{Value: triple.Value, AskedToFree: true, MessageRefCount: triple.MessageRefCount}

	}
	return nil
}

func (s *HandleStore[ValueType]) IncrementMessageRefCount(handle uint64) error {
	triple, found := (*s.Map)[handle]
	if !found {
		return fmt.Errorf("no handle found to increment")
	}
	(*s.Map)[handle] = HandleStoreTriple[ValueType]{Value: triple.Value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount + 1}
	return nil
}

func (s *HandleStore[ValueType]) DecrementMessageRefCount(handle uint64) error {
	triple, found := (*s.Map)[handle]
	if !found {
		return fmt.Errorf("no handle found to increment")
	}
	if (triple.AskedToFree) && (triple.MessageRefCount-1 == 0) {
		delete((*s.Map), handle)
	} else {
		(*s.Map)[handle] = HandleStoreTriple[ValueType]{Value: triple.Value, AskedToFree: triple.AskedToFree, MessageRefCount: triple.MessageRefCount - 1}
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
