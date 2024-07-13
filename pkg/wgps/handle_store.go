package wgps

import "golang.org/x/exp/constraints"

var leastUnassignedHandle uint64 = 0

// Assuming ValueType is a generic type that is ordered, as per your file excerpt.
type HandleData[ValueType constraints.Ordered] struct {
	Value    ValueType
	Free     bool
	Messages int
}

type Map[ValueType constraints.Ordered] struct {
	data map[uint64]HandleData[ValueType]
}

func NewMap[ValueType constraints.Ordered]() *Map[ValueType] {
	return &Map[ValueType]{
		data: make(map[uint64]HandleData[ValueType]),
	}
}

type EventuallyMap[ValueType constraints.Ordered] struct {
	data map[int64]ValueType
}

func NewEventuallyMap[ValueType constraints.Ordered]() *EventuallyMap[ValueType] {
	return &EventuallyMap[ValueType]{
		data: make(map[int64]ValueType),
	}
}

func (em *Map[ValueType]) Get(handle uint64) (ValueType, bool) {
	value, found := em.data[handle]
	return value.Value, found
}
