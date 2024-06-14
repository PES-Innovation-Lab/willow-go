package utils

import (
	"fmt"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
	"golang.org/x/exp/constraints"
)

func IsValidPath[T constraints.Unsigned](path types.Path, pathParams types.PathParams[T]) (bool, error) {
	/*
	  This function takes in a pathParams variable defined by the owner of the network and a path variable,
	  checks if the given path satisfies all the constraints given and returns a boolean with an error.
	  We run through each component of the path to check if all three constraints of a path is satisfied.
	*/
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

func IsPathPrefixed(prefix types.Path, path types.Path) bool {
	/*
	   This function, we check if prefix length is smalled than path length and then we run through each component to compare
	   actual path to see if it's equal, if it is we can say that the given prefix prefixes the given path
	*/
	if len(prefix) > len(path) {
		return false
	}
	for index, prefixComponent := range prefix {
		pathComponent := path[index]
		if OrderBytes(prefixComponent, pathComponent) != 0 {
			return false
		}
	}
	return true
}

func CommonPrefix(first types.Path, second types.Path) (types.Path, error) {
	/*
	* In this function we run until the end of one of the paths and check until where they match, if there are no matching
	* prefix, we return nil, if there is we return the slice until the matching prefix.
	 */
	index := 0
	for ; index < len(first) && index < len(second); index++ {
		firstComponent := first[index]
		secondComponent := second[index]
		if OrderBytes(firstComponent, secondComponent) != 0 {
			break
		}
	}
	if index == 0 {
		return nil, fmt.Errorf("There are no common prefiexs!")
	}
	return first[0 : index+1], nil
}

// TO-DO implement Encode and Decode functions for Path
// func EncodePath[T constraints.Unsigned](path types.Path, pathParams types.PathParams[T]) []byte {
// 	componentCountBytes := EncodingIntMax32(T(len(path)), pathParams.MaxPathLength)
// }
