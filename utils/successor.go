package utils

import (
	"log"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

/** Returns the successor to a path given a `Path` and `PathScheme`.  */
func SuccessorPath[T constraints.Unsigned](rawpath types.Path, scheme types.PathParams[T]) types.Path {
	path := make(types.Path, len(rawpath))
	copy(path, rawpath)
	if len(path) == 0 {
		return types.Path{[]byte{}}
	}
	if T(len(path)) < scheme.MaxComponentCount {
		return PathAppend(path, []byte{}, scheme)
	}

	for i := len(path) - 1; i >= 0; i-- {
		newComponent := TryAppendZeroByte(path[i], scheme)
		if newComponent != nil {
			return PathAppend(path[:i], newComponent, scheme)
		}
		newComponent = SuccessorBytesFixedWidth(path[i])
		if newComponent != nil {
			return PathAppend(path, newComponent, scheme)
		}
	}
	return nil
}

/** Return a successor to a prefix, that is, the next element that is not prefixed by the given path. */
func SuccessorPrefix[T constraints.Unsigned](rawpath types.Path, pathParams types.PathParams[T]) types.Path {
	path := make(types.Path, len(rawpath))
	copy(path, rawpath)
	for i := len(path) - 1; i >= 0; i-- {
		successorComp := TryAppendZeroByte(path[i], pathParams)
		if successorComp != nil {
			return PathAppend(path[:i], successorComp, pathParams)
		}
		prefixSuccessor := PrefixSuccessor(path[i])
		if prefixSuccessor != nil {
			return PathAppend(path[:i], prefixSuccessor, pathParams)
		}
	}
	return nil
}

// Does all path checks before appending a component to a path!
func PathAppend[T constraints.Unsigned](path types.Path, component []byte, scheme types.PathParams[T]) types.Path {
	if T(len(path)+1) > scheme.MaxComponentCount {
		log.Fatal("Too many Components! The path components exceeds max component count")
	}
	pathLength := len(component)
	for _, component := range path {
		pathLength += len(component)
	}
	if T(pathLength) > scheme.MaxPathLength {
		log.Fatal("Path length too long, consider decreasing the length of the components")
	}

	return append(path, component)
}

// If the component length does not exceed max component length, it appends a zero to the byte, else returns nil.
func TryAppendZeroByte[T constraints.Unsigned](component []byte, scheme types.PathParams[T]) []byte {
	if T(len(component)) == scheme.MaxComponentLength {
		return nil
	}
	newComponent := append(component, 0)
	return newComponent
}

// This function goes through each byte of a component adds one if it is not greater than 255 and retrn the slice until that specific byte!
func PrefixSuccessor(component []byte) []byte {
	for i := len(component) - 1; i >= 0; i-- {
		if component[i] != 255 {
			component = append(component[:i], component[i]+1)
			return component
		}
	}
	return nil
}

/** Return the succeeding bytestring of the given bytestring without increasing that bytestring's length.  */
func SuccessorBytesFixedWidth(bytes []byte) []byte {
	newBytes := make([]byte, len(bytes))
	copy(newBytes, bytes)

	didIncrement := false

	for i := len(newBytes) - 1; i >= 0; i-- {
		byteVal := newBytes[i]

		if byteVal >= 255 {
			newBytes[i] = 0
			continue
		}

		if !didIncrement {
			newBytes[i] = byteVal + 1
			didIncrement = true
			break
		}
	}

	if !didIncrement {
		return nil
	}

	return newBytes
}
