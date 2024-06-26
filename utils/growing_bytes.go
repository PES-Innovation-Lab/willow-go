package utils

import (
	"sync"
)

// Defer processing of bytestrings until a certain chunk size (length, in bytes) has been reached.
// Put the array of bytes into the resolver
type DeferredUntilLength struct {
	resolver chan []byte
	length   int
}

// GrowingBytes objects allows us to process bytestreams in a nonblocking fashion with buffered channels and
// also provdes us with useful helper functions.

type GrowingBytes struct {
	Incoming               chan []byte
	HasUnfulfilledRequests chan struct{}
	DeferredUntilLength    *DeferredUntilLength
	Array                  []byte
	Mu                     sync.Mutex
}

// Construct a new new Growing Bytes instance and return a pointer to it
func NewGrowingBytes(incoming chan []byte) *GrowingBytes {
	gb := &GrowingBytes{
		Incoming:               incoming,
		Array:                  []byte{},
		HasUnfulfilledRequests: make(chan struct{}, 1),
	}
	// Non blocking goroutine to take in byte chunks, synchronize and append to array buffer.
	go func() {
		defer close(gb.HasUnfulfilledRequests)
		for {
			select {
			case <-gb.HasUnfulfilledRequests:

			case chunk, ok := <-gb.Incoming:

				if !ok {
					return
				}

				gb.Mu.Lock()

				gb.Array = append(gb.Array, chunk...)
				gb.Mu.Unlock()

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

	if len(gb.Array) >= length {
		return gb.Array
	}

	gb.Mu.Lock()

	// If there's already a deferred request for the same length, return the resolved result
	if gb.DeferredUntilLength != nil && gb.DeferredUntilLength.length == length {
		resolver := gb.DeferredUntilLength.resolver
		return <-resolver
	}

	resolver := make(chan []byte, 10)
	gb.DeferredUntilLength = &DeferredUntilLength{
		length:   length,
		resolver: resolver,
	}

	gb.HasUnfulfilledRequests <- struct{}{}
	gb.Mu.Unlock()

	return <-resolver
}

// Prune the array by the given byte length
func (gb *GrowingBytes) Prune(length int) {
	gb.Mu.Lock()

	if length >= len(gb.Array) {
		gb.Array = []byte{}
	} else {
		gb.Array = gb.Array[length:]
	}
	gb.Mu.Unlock()
}
