package kv_driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

/*
	The following encode and decode functions are ordered as such - Timestamp, Path, Subspace.
	Timestamp has a known size, and size of Path can be calculated using Path Params.
	This leaves only SubspaceId with an uncertain type (this is a parameter of Willow).
	To combat it's unknown size, we've kept it at the end.
*/

/* Encodes the time, subspace and path from the kd tree into a key usable by the entries kv store */
func EncodeKey[K constraints.Unsigned](timestamp uint64, subspaceId types.SubspaceId, pathParams types.PathParams[K], path types.Path) ([]byte, error) {
	// Convert timestamp to byte slice
	timestampBytes := utils.BigIntToBytes(timestamp)

	// Convert path to byte slice
	pathBytes := utils.EncodePath(pathParams, path)

	// Convert subspaceId to byte slice
	subspaceBytes, err := encodeSubspaceId(subspaceId)
	if err != nil {
		return nil, fmt.Errorf("here is the error you L bozo : %w", err)
	}

	// Combine all byte slices
	encodedKey := append(timestampBytes, pathBytes...)
	encodedKey = append(encodedKey, subspaceBytes...)

	return encodedKey, nil
}

/* EncodeSubspaceId encodes the subspaceId into []byte */
func encodeSubspaceId(subspace types.SubspaceId) ([]byte, error) {
	var subspaceBytes []byte

	switch any(subspace).(type) {
	case uint8:
		subspaceBytes = []byte{byte(reflect.ValueOf(subspace).Uint())}
	case uint16:
		subspaceBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(subspaceBytes, uint16(reflect.ValueOf(subspace).Uint()))
	case uint32:
		subspaceBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(subspaceBytes, uint32(reflect.ValueOf(subspace).Uint()))
	case uint64:
		subspaceBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(subspaceBytes, uint64(reflect.ValueOf(subspace).Uint()))
	case int8:
		subspaceBytes = []byte{byte(reflect.ValueOf(subspace).Uint())}
	case int16:
		subspaceBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(subspaceBytes, uint16(reflect.ValueOf(subspace).Int()))
	case int32:
		subspaceBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(subspaceBytes, uint32(reflect.ValueOf(subspace).Int()))
	case int64:
		subspaceBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(subspaceBytes, uint64(reflect.ValueOf(subspace).Int()))
	case string:
		subspaceBytes = []byte(reflect.ValueOf(subspace).String())
	default:
		return nil, fmt.Errorf("unsupported subspace type: %T", subspace)
	}
	fmt.Println(subspaceBytes)
	return subspaceBytes, nil
}

/* Decodes the key from the kv store into the timestamp, subspaceId, and path */
func DecodeKey[K constraints.Unsigned](encodedKey []byte, pathParams types.PathParams[K]) (uint64, uint64, types.Path, error) {
	var timestamp uint64
	var subspaceId uint64

	// Read timestamp from the encoded key
	timestampBytes := encodedKey[:8]
	timestamp = binary.BigEndian.Uint64(timestampBytes)

	// Decode path
	pathEndIndex, decodedPath, err := utils.DecodePath(pathParams, encodedKey[8:])
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to decode path: %w", err)
	}

	// Decode subspaceId
	subspaceBytes := encodedKey[8+pathEndIndex:]
	subspaceId, err = decodeSubspaceId(subspaceBytes)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to decode subspaceId: %w", err)
	}

	return timestamp, subspaceId, decodedPath, nil
}

/* decodeSubspaceId decodes the subspaceId from []byte */
func decodeSubspaceId(subspaceBytes []byte) (uint64, error) {
	var subspaceId uint64
	length := len(subspaceBytes)

	if length > 8 {
		return 0, fmt.Errorf("subspaceId byte array too long: %d bytes", length)
	}

	// Create an 8-byte array and copy subspaceBytes into the rightmost part
	paddedSubspaceBytes := make([]byte, 8)
	copy(paddedSubspaceBytes[8-length:], subspaceBytes)

	buf := bytes.NewReader(paddedSubspaceBytes)
	err := binary.Read(buf, binary.BigEndian, &subspaceId)
	if err != nil {
		return 0, fmt.Errorf("failed to decode uint64: %v", err)
	}

	return subspaceId, nil
}

// /* Decodes the key from the kv store into the timestamp, subspaceId, and path */
// func DecodeKey[T constraints.Unsigned](encodedKey []byte, pathParams types.PathParams[T]) (uint64, T, types.Path, error) {
// 	var timestamp uint64
// 	var subspaceId T

// 	// Read timestamp from the encoded key
// 	timestampBytes := encodedKey[:8]
// 	timestamp = binary.BigEndian.Uint64(timestampBytes)

// 	// Decode path
// 	pathEndIndex, decodedPath, err := utils.DecodePath(pathParams, encodedKey[8:])
// 	if err != nil {
// 		return 0, *new(T), nil, fmt.Errorf("failed to decode path: %w", err)
// 	}

// 	// Decode subspaceId
// 	subspaceBytes := encodedKey[8+pathEndIndex:]
// 	subspaceId, err = decodeSubspaceId[T](subspaceBytes)
// 	if err != nil {
// 		return 0, *new(T), nil, fmt.Errorf("failed to decode subspaceId: %w", err)
// 	}

// 	return timestamp, subspaceId, decodedPath, nil
// }

// /* decodeSubspaceId decodes the subspaceId from []byte */
// func decodeSubspaceId[T constraints.Unsigned](subspaceBytes []byte) (T, error) {
// 	var subspaceId T
// 	buf := bytes.NewReader(subspaceBytes)

// 	// Determine the length of the subspaceId
// 	length := len(subspaceBytes)

// 	switch length {
// 	case 1:
// 		var value uint8
// 		value = uint8(subspaceBytes[0])
// 		subspaceId = T(value)
// 	case 2:
// 		var value uint16
// 		err := binary.Read(buf, binary.BigEndian, &value)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed to decode int16: %v", err)
// 		}
// 		subspaceId = T(value)
// 	case 4:
// 		var value uint32
// 		err := binary.Read(buf, binary.BigEndian, &value)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed to decode int32: %v", err)
// 		}
// 		subspaceId = T(value)
// 	case 8:
// 		var value uint64
// 		err := binary.Read(buf, binary.BigEndian, &value)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed to decode int64: %v", err)
// 		}
// 		subspaceId = T(value)
// 	}

// 	return subspaceId, nil
// }
