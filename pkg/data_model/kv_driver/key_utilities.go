package kv_driver

import (
	"bytes"
	"encoding/binary"
	"fmt"

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
func EncodeKey[T constraints.Unsigned](timestamp uint64, subspaceId T, pathParams types.PathParams[T], path types.Path) ([]byte, error) {
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
func encodeSubspaceId[T constraints.Unsigned](subspace T) ([]byte, error) {
	var subspaceBytes []byte

	switch any(subspace).(type) {
	case uint8:
		subspaceBytes = []byte{byte(subspace)}
	case uint16:
		subspaceBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(subspaceBytes, uint16(subspace))
	case uint32:
		subspaceBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(subspaceBytes, uint32(subspace))
	case uint64:
		subspaceBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(subspaceBytes, uint64(subspace))
	default:
		return nil, fmt.Errorf("unsupported subspace type: %T", subspace)
	}

	return subspaceBytes, nil
}

/* Decodes the key from the kv store into the timestamp, subspaceId, and path */
func DecodeKey[T constraints.Unsigned](encodedKey []byte, pathParams types.PathParams[T]) (uint64, T, types.Path, error) {
	var timestamp uint64
	var subspaceId T

	// Read timestamp from the encoded key
	timestampBytes := encodedKey[:8]
	timestamp = binary.BigEndian.Uint64(timestampBytes)

	// Decode path
	pathEndIndex, decodedPath, err := utils.DecodePath(pathParams, encodedKey[8:])
	if err != nil {
		return 0, *new(T), nil, fmt.Errorf("failed to decode path: %w", err)
	}

	// Decode subspaceId
	subspaceBytes := encodedKey[8+pathEndIndex:]
	_, err = decodeSubspaceId[T](subspaceBytes)
	if err != nil {
		return 0, *new(T), nil, fmt.Errorf("failed to decode subspaceId: %w", err)
	}

	return timestamp, subspaceId, decodedPath, nil
}

/* decodeSubspaceId decodes the subspaceId from []byte */
func decodeSubspaceId[T constraints.Unsigned](subspaceBytes []byte) (T, error) {
	var subspaceId T
	encodedPayloadDigest := bytes.NewReader(subspaceBytes)

	// Determine the length of the subspaceId
	length := len(subspaceBytes)

	switch length {
	case 1:
		var value uint8
		value = uint8(subspaceBytes[0])
		subspaceId = T(value)
	case 2:
		var value uint16
		err := binary.Read(encodedPayloadDigest, binary.BigEndian, &value)
		if err != nil {
			return 0, fmt.Errorf("failed to decode int16: %v", err)
		}
		subspaceId = T(value)
	case 4:
		var value uint32
		err := binary.Read(encodedPayloadDigest, binary.BigEndian, &value)
		if err != nil {
			return 0, fmt.Errorf("failed to decode int32: %v", err)
		}
		subspaceId = T(value)
	case 8:
		var value uint64
		err := binary.Read(encodedPayloadDigest, binary.BigEndian, &value)
		if err != nil {
			return 0, fmt.Errorf("failed to decode int64: %v", err)
		}
		subspaceId = T(value)
	}

	return subspaceId, nil
}

// Takes in payload digest, payload length and returns a flattened array of both
// First 8 bytes is payload length and the rest of the bytes are payload digest
func EncodeEntryValues(payloadDigest string, payloadLength uint64) []byte {
	var encodedValues []byte
	encodedPayloadLength := encodePayloadLength(payloadLength)
	encodedValues = append(encodedValues, encodedPayloadLength...)
	encodedPayloadDigest := encodePayloadDigest(payloadDigest)

	encodedValues = append(encodedValues, encodedPayloadLength...)
	encodedValues = append(encodedValues, encodedPayloadDigest...)

	return encodedValues
}

// Takes in encoded value and returns decoded payloadDigest and payload length
// First 8 bytes are payload length, so takes amd comverts it into uint64 and then returns payload digest next
func DecodeEntryValues[T constraints.Ordered](encodedValue []byte) (uint64, string) {
	encodedPayloadLength := encodedValue[0:8]
	encodedPayloadDigest := encodedValue[8:]
	payloadLength := decodePayloadLength(encodedPayloadLength)
	payloadDigest := decodePayloadDigest(encodedPayloadDigest)

	return payloadLength, payloadDigest
}

func encodePayloadLength(payloadLength uint64) []byte {
	var encoded []byte
	binary.BigEndian.PutUint64(encoded, payloadLength)
	return encoded
}

func decodePayloadLength(encodedPayloadLength []byte) uint64 {
	payloadLength := binary.BigEndian.Uint64(encodedPayloadLength)
	return payloadLength
}

func encodePayloadDigest(payloadDigest string) []byte {
	return []byte(payloadDigest)
}

func decodePayloadDigest(buf []byte) string {
	return string(buf)
}
