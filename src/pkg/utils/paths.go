package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"golang.org/x/exp/constraints"
)

func IsValidPath[T constraints.Signed](path types.Path, pathParams types.PathParams[T]) (bool, error) {
	if len(path) > int(pathParams.MaxComponentcount) {
		return false, fmt.Errorf("The Path exceeds maximum allowed components.")
	}

	totalComponentCount := 0
	for _, component := range path {
		if len(component) > int(pathParams.MaxComponentLength) {
			return false, fmt.Errorf("The component: %T exceeds maximum allowed component length.", component)
		}
		totalComponentCount += len(component)
	}
	if totalComponentCount > int(pathParams.MaxPathLength) {
		return false, fmt.Errorf("Path length exceeds maximum allowed length.")
	}
	return true, nil
}

func IsPathPrefixed(prefix types.Path, path types.Path) (bool, error) {
	if len(prefix) > len(path) {
		return false, fmt.Errorf("The prefix cannot be greater than the path it prefixes.")
	}
	for index, prefixComponent := range prefix {
		pathComponent := path[index]
		if OrderBytes(prefixComponent, pathComponent) != 0 {
			return false, fmt.Errorf("The given prefix is not a prefix for the given path.")
		}
	}
	return true, nil
}

func CommonPrefix(first types.Path, second types.Path) types.Path {
	index := 0
	for ; index < len(first) && index < len(second); index++ {
		firstComponent := first[index]
		secondComponent := second[index]
		if OrderBytes(firstComponent, secondComponent) != 0 {
			break
		}
	}
	if index == 0 {
		return nil
	}
	return first[0 : index+1]
}

// TO-DO implement Encode and Decode functions for Path
func EncodePath[T constraints.Signed](path types.Path, pathParams types.PathParams[T]) []byte {
	return nil
}
