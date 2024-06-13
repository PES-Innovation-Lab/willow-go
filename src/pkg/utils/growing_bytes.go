package utils

import "sync"

// Defer processing of bytestrings until a certain chunk size (length, in bytes) has been reached.
// Put the array of bytes into the resolver
type DeferredUntilLength struct {
	length   int
	resolver chan []byte
}

// GrowingBytes objects allows us to process bytestreams in a nonblocking fashion with buffered channels and
// also provdes us with useful helper functions.

type GrowingBytes struct {
	Array                  []byte        //Output array of bytes
	Incoming               <-chan []byte // Buffered channel
	HasUnfulfilledRequests chan struct{} //
	DeferredUntilLength    *DeferredUntilLength
	Mu                     sync.Mutex
}

func NewGrowingBytes(incoming <-chan []byte) *GrowingBytes {
	gb := &GrowingBytes{
		Incoming:               incoming,
		Array:                  []byte{},
		HasUnfulfilledRequests: make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-gb.HasUnfulfilledRequests:

			case chunk, ok := <-gb.Incoming:
				if !ok {
					return
				}
				gb.Mu.Lock()
				defer gb.Mu.Unlock()
				gb.Array = append(gb.Array, chunk...)
				if gb.DeferredUntilLength != nil && len(gb.Array) >= gb.DeferredUntilLength.length {
					gb.DeferredUntilLength.resolver <- gb.Array
					gb.DeferredUntilLength = nil
					gb.HasUnfulfilledRequests <- struct{}{}

				}

			}
		}

	}()
	return gb

}

// NextRelative pulls bytes until the accumulated bytestring has grown by the given amount
func (gb *GrowingBytes) NextRelative(length int) []byte {
	target := len(gb.Array) + length
	return gb.NextAbsolute(target)
}

// NextAbsolute pulls bytes until the accumulated bytestring has grown to the given size
func (gb *GrowingBytes) NextAbsolute(length int) []byte {
	gb.Mu.Lock()
	defer gb.Mu.Unlock()
	if len(gb.Array) >= length {
		return gb.Array
	}
	// If there's already a deferred request for the same length, return the resolved result
	if gb.DeferredUntilLength != nil && gb.DeferredUntilLength.length == length {
		resolver := gb.DeferredUntilLength.resolver
		return <-resolver
	}

	resolver := make(chan []byte, 1)
	gb.DeferredUntilLength = &DeferredUntilLength{
		length:   length,
		resolver: resolver,
	}

	gb.HasUnfulfilledRequests <- struct{}{}
	return <-resolver
}

// Prune the array by the given byte length
func (gb *GrowingBytes) Prune(length int) {
	gb.Mu.Lock()
	defer gb.Mu.Unlock()
	if length >= len(gb.Array) {
		gb.Array = []byte{}
	} else {
		gb.Array = gb.Array[length:]
	}
}
