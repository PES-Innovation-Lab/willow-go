package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"golang.org/x/exp/constraints"
)

/** Returns the successor to a path given a `Path` and `PathScheme`.  */
func SuccessorPath[T constraints.Signed](path types.Path, scheme types.PathParams[T]) types.Path {
	if len(path) == 0 {
		nextPath := types.Path{make([]byte, 1)}

		valid, _ := IsValidPath(nextPath, scheme)

		if valid {
			return nextPath
		}
		return nil

	}

	workingPath := make(types.Path, len(path))
	copy(workingPath, path)

	for i := len(path) - 1; i >= 0; i-- {
		component := workingPath[i]

		simplestNextComponent := append(component, 0)

		simplestNextPath := make(types.Path, len(path))
		copy(simplestNextPath, path)
		simplestNextPath[i] = simplestNextComponent

		valid, _ := IsValidPath(simplestNextPath, scheme)

		if valid {
			return simplestNextPath
		}

		//Otherwise

		incrementedComponent := SuccessorBytesFixedWidth(component)

		if incrementedComponent != nil {
			nextPath := append(path[:i], incrementedComponent)
			return nextPath
		}

		//In the case of an overflow

		workingPath = workingPath[:len(workingPath)-1]
	}

	if len(workingPath) == 0 {
		return nil
	}

	return workingPath
}

/** Return a successor to a prefix, that is, the next element that is not a prefix of the given path. */
func SuccessorPrefix(path types.Path) types.Path {
	if len(path) == 0 {
		return nil
	}

	workingPath := make(types.Path, len(path))
	copy(workingPath, path)

	for i := len(path) - 1; i >= 0; i-- {
		component := workingPath[i]

		incrementedComponent := SuccessorBytesFixedWidth(component)

		if incrementedComponent != nil {
			nextPath := append(path[:i], incrementedComponent)
			return nextPath
		}

		if len(component) == 0 {
			nextPath := append(path[:i], []byte{0})
			return nextPath
		}

		workingPath = workingPath[:len(workingPath)-1]
	}

	if len(workingPath) == 0 {
		return nil
	}

	return workingPath
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
