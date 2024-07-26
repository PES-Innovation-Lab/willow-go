package reconciliation

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/handlestore"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

type AoiIntersectionFinderOpts struct {
	NamespaceScheme datamodeltypes.NamespaceScheme
	SubspaceScheme  datamodeltypes.SubspaceScheme
	HandlesOurs     handlestore.HandleStore[types.AreaOfInterest]
	HandlesTheirs   handlestore.HandleStore[types.AreaOfInterest]
}

// Define the AoiIntersectionFinder struct with generic types
type AoiIntersectionFinder struct {
	NamespaceScheme datamodeltypes.NamespaceScheme
	SubspaceScheme  datamodeltypes.SubspaceScheme
	HandlesOurs     handlestore.HandleStore[types.AreaOfInterest]
	HandlesTheirs   handlestore.HandleStore[types.AreaOfInterest]

	HandlesOursNamespaceMap   map[uint64]types.NamespaceId
	HandlesTheirsNamespaceMap map[uint64]types.NamespaceId

	IntersectingAoiQueue []struct {
		Namespace types.NamespaceId
		Ours      uint64
		Theirs    uint64
	}
}

// NewAoiIntersectionFinder is the constructor function for AoiIntersectionFinder
func NewAoiIntersectionFinder(opts AoiIntersectionFinderOpts) *AoiIntersectionFinder {
	return &AoiIntersectionFinder{
		NamespaceScheme:           opts.NamespaceScheme,
		SubspaceScheme:            opts.SubspaceScheme,
		HandlesOurs:               opts.HandlesOurs,
		HandlesTheirs:             opts.HandlesTheirs,
		HandlesOursNamespaceMap:   make(map[uint64]types.NamespaceId),
		HandlesTheirsNamespaceMap: make(map[uint64]types.NamespaceId),
		IntersectingAoiQueue: make([]struct {
			Namespace types.NamespaceId
			Ours      uint64
			Theirs    uint64
		}, 0),
	}
}

func (a *AoiIntersectionFinder) AddAoiHandleToNamespace(handle uint64, namespace types.NamespaceId, ours bool) {
	var HandleNamespaceMap map[uint64]types.NamespaceId
	if ours {
		HandleNamespaceMap = a.HandlesOursNamespaceMap
	} else {
		HandleNamespaceMap = a.HandlesTheirsNamespaceMap
	}

	var OtherHandleNamespaceMap map[uint64]types.NamespaceId
	if ours {
		OtherHandleNamespaceMap = a.HandlesTheirsNamespaceMap
	} else {
		OtherHandleNamespaceMap = a.HandlesOursNamespaceMap
	}

	var HandleStore handlestore.HandleStore[types.AreaOfInterest]
	if ours {
		HandleStore = a.HandlesOurs
	} else {
		HandleStore = a.HandlesTheirs
	}

	var OtherHandleStore handlestore.HandleStore[types.AreaOfInterest]
	if ours {
		OtherHandleStore = a.HandlesTheirs
	} else {
		OtherHandleStore = a.HandlesOurs
	}

	HandleNamespaceMap[handle] = namespace

	// Now check for all other AOIs with the same namespace.
	for OtherHandle, OtherNamespace := range OtherHandleNamespaceMap {
		if !a.NamespaceScheme.IsEqual(namespace, OtherNamespace) {
			continue
		}

		Aoi, found := HandleStore.Get(handle)
		if !found {
			fmt.Errorf("Could not dereference an AOI handle")
		}

		AoiOther, foundother := OtherHandleStore.Get(OtherHandle)
		if !foundother {
			fmt.Errorf("Could not dereference an AOI handle")
		}

		Intersection := utils.IntersectArea(a.SubspaceScheme.Order, Aoi.Area, AoiOther.Area)

		if Intersection == nil {
			continue
		}
		var Ours uint64
		if ours {
			Ours = handle
		} else {
			Ours = OtherHandle
		}

		var Theirs uint64
		if ours {
			Theirs = OtherHandle
		} else {
			Theirs = handle
		}
		a.IntersectingAoiQueue = append(a.IntersectingAoiQueue, struct {
			Namespace types.NamespaceId
			Ours      uint64
			Theirs    uint64
		}{
			Namespace: namespace,
			Ours:      Ours,
			Theirs:    Theirs,
		})

	}
}

func (a *AoiIntersectionFinder) HandleToNamespaceId(handle uint64, ours bool) types.NamespaceId {
	var HandleNamespaceMap map[uint64]types.NamespaceId
	if ours {
		HandleNamespaceMap = a.HandlesOursNamespaceMap
	} else {
		HandleNamespaceMap = a.HandlesTheirsNamespaceMap
	}

	return HandleNamespaceMap[handle]
}

// Intersections returns a channel through which intersections can be received.
func (a *AoiIntersectionFinder) Intersections() <-chan struct {
	Namespace types.NamespaceId
	Ours      uint64
	Theirs    uint64
} {
	// Create a channel for sending Intersection values.
	ch := make(chan struct {
		Namespace types.NamespaceId
		Ours      uint64
		Theirs    uint64
	})

	// Start a goroutine to send values to the channel.
	go func() {
		// Ensure the channel is closed when all values have been sent.
		defer close(ch)

		// Iterate over the slice of intersections.
		for _, intersection := range a.IntersectingAoiQueue {
			// Send each intersection through the channel.
			ch <- intersection
		}
	}()

	// Return the channel to the caller.
	return ch
}
