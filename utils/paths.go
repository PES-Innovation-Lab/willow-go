package utils

import (
	"fmt"
	"log"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

func PrefixesOf(path types.Path) []types.Path {
	prefixes := []types.Path{[][]byte{}}
	for i := range path {
		prefixes = append(prefixes, path[0:i+1])
	}
	return prefixes
}

func IsValidPath[T constraints.Unsigned](path types.Path, pathParams types.PathParams[T]) (bool, error) {
	/*
	  This function takes in a pathParams variable defined by the owner of the network and a path variable,
	  checks if the given path satisfies all the constraints given and returns a boolean with an error.
	  We run through each component of the path to check if all three constraints of a path is satisfied.
	*/
	if len(path) > int(pathParams.MaxComponentCount) {
		return false, fmt.Errorf("the Path exceeds maximum allowed components")
	}

	totalComponentCount := 0
	for _, component := range path {
		if len(component) > int(pathParams.MaxComponentLength) {
			return false, fmt.Errorf("the component: %T exceeds maximum allowed component length", component)
		}
		totalComponentCount += len(component)
	}
	if totalComponentCount > int(pathParams.MaxPathLength) {
		return false, fmt.Errorf("path length exceeds maximum allowed length")
	}
	return true, nil
}

func IsPathPrefixed(prefix types.Path, path types.Path) (bool, error) {
	/*
	   This function, we check if prefix length is smalled than path length and then we run through each component to compare
	   actual path to see if it's equal, if it is we can say that the given prefix prefixes the given path
	*/
	if len(prefix) > len(path) {
		return false, fmt.Errorf("the prefix cannot be greater than the path it prefixes")
	}
	for index, prefixComponent := range prefix {
		pathComponent := path[index]
		if OrderBytes(prefixComponent, pathComponent) != 0 {
			return false, fmt.Errorf("the given prefix is not a prefix for the given path")
		}
	}
	return true, nil
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
		return nil, fmt.Errorf("there are no common prefixes")
	}
	return first[0:index], nil
}

func EncodePath[T constraints.Unsigned](pathParams types.PathParams[T], path types.Path) []byte {
	/*
	   this function takes in a path and a pathParams variable relted to it, we take the path,
	   The way path gets encoded is, the first "MaxComponentCount" width bytes are number of components,
	   the next number of components is the length of the component followed by the respective component.
	*/
	componentCountBytes := EncodeIntMax32(T(len(path)), pathParams.MaxPathLength)
	componentBytes := componentCountBytes
	for _, component := range path {
		lengthBytesComponent := EncodeIntMax32(T(len(component)), pathParams.MaxComponentLength)
		componentBytes = append(componentBytes, lengthBytesComponent...)
		componentBytes = append(componentBytes, component...)
	}
	return componentBytes
}

func DecodePath[T constraints.Unsigned](pathParams types.PathParams[T], encPath []byte) types.Path {
	/*
	   It checks the number of components in the first "MaxComponentCount" width and then interates through each
	   Component, checks it's length and extracts the component based on the length
	*/
	maxCountWidth := GetWidthMax32Int(pathParams.MaxComponentCount)
	componentCountBytes := encPath[0:maxCountWidth]

	componentCount, err := DecodeIntMax32(componentCountBytes, pathParams.MaxComponentCount)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	pos := maxCountWidth

	maxComponentLengthWidth := GetWidthMax32Int(pathParams.MaxComponentLength)
	var path [][]byte

	for i := 0; i < int(componentCount); i++ {
		lengthComponentBytes := encPath[pos : pos+maxComponentLengthWidth]
		lengthComponent, err := DecodeIntMax32(lengthComponentBytes, pathParams.MaxComponentLength)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		pathComponent := encPath[pos+maxComponentLengthWidth : pos+maxComponentLengthWidth+int(lengthComponent)]

		path = append(path, pathComponent)
		pos += maxComponentLengthWidth + int(lengthComponent)

	}
	return path
}

func EncodePathLength[T constraints.Unsigned](pathParams types.PathParams[T], path types.Path) uint64 {
	countWidth := GetWidthMax32Int(pathParams.MaxComponentCount)

	length := countWidth

	compLenWidth := GetWidthMax32Int(pathParams.MaxComponentLength)

	for _, comp := range path {
		length += compLenWidth
		length += len(comp)
	}
	return uint64(length)
}

func EncodeRelativePath[T constraints.Unsigned](pathParams types.PathParams[T], toEncode types.Path, refernce types.Path) []byte {
	longestPrefix, err := CommonPrefix(toEncode, refernce)
	if err != nil {
		log.Fatalf("error in calculating common paths: %s", err)
	}
	longestPrefixLength := len(longestPrefix)
	prefixLengthBytes := EncodeIntMax32(T(longestPrefixLength), pathParams.MaxComponentCount)
	suffix := toEncode[longestPrefixLength:]
	suffixEncoded := EncodePath(pathParams, suffix)

	return append(prefixLengthBytes, suffixEncoded...)
}

func DecodePathStream[T constraints.Unsigned](pathParams types.PathParams[T], bytes *GrowingBytes) types.Path {
	maxCountWidth := GetWidthMax32Int(pathParams.MaxComponentCount)

	accumulatedBytes := bytes.NextAbsolute(maxCountWidth)

	countBytes := accumulatedBytes[0:maxCountWidth]
	componentCount, _ := DecodeIntMax32(countBytes, pathParams.MaxComponentCount)

	bytes.Prune(maxCountWidth)
	componentLengthWidth := GetWidthMax32Int(pathParams.MaxComponentLength)
	var path types.Path

	for i := 0; i < int(componentCount); i++ {
		bytes.NextAbsolute(componentLengthWidth)

		lengthBytes := bytes.Array[0:componentLengthWidth]
		componentLength, _ := DecodeIntMax32(lengthBytes, pathParams.MaxComponentLength)

		bytes.NextAbsolute(componentLengthWidth + componentLengthWidth)

		pathComponent := bytes.Array[componentLengthWidth : componentLengthWidth+int(componentLength)]

		path = append(path, pathComponent)

		bytes.Prune(componentLengthWidth + int(componentLength))
	}
	return path
}

func DecodeRelativePath[T constraints.Unsigned](
	pathParams types.PathParams[T],
	encRelPath []byte,
	refernce types.Path,
) types.Path {
	prefixLengthWidth := GetWidthMax32Int(pathParams.MaxComponentCount)
	prefixLength, err := DecodeIntMax32(encRelPath[0:prefixLengthWidth], pathParams.MaxComponentCount)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	prefix := refernce[0:prefixLength]

	suffix := DecodePath(pathParams, encRelPath[prefixLengthWidth:])

	return append(prefix, suffix...)
}

func DecodeRelPathStream[T constraints.Unsigned](
	pathParams types.PathParams[T],
	bytes *GrowingBytes,
	reference types.Path,
) types.Path {
	prefixLengthWidth := GetWidthMax32Int(pathParams.MaxComponentCount)
	accumulatedBytes := bytes.NextAbsolute(prefixLengthWidth)

	prefixLength, _ := DecodeIntMax32(accumulatedBytes[0:prefixLengthWidth], pathParams.MaxComponentCount)
	prefix := reference[0:prefixLength]
	bytes.Prune(prefixLengthWidth)

	suffix := DecodePathStream(pathParams, bytes)

	return append(prefix, suffix...)
}

func EncodePathRelativeLength[T constraints.Unsigned](pathParams types.PathParams[T], primary types.Path, refernce types.Path) int {
	longestPrefix, err := CommonPrefix(primary, refernce)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	longestPrefixLength := len(longestPrefix)
	prefixLengthLength := GetWidthMax32Int(pathParams.MaxComponentCount)
	suffix := primary[longestPrefixLength:]
	suffixLength := len(suffix)
	return prefixLengthLength + suffixLength
}
