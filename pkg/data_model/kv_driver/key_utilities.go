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
func EncodeKey[Params constraints.Unsigned](timestamp uint64, subspaceId []byte, pathParams types.PathParams[Params], path types.Path) ([]byte, error) {
	// Convert timestamp to byte slice
	timestampBytes := utils.BigIntToBytes(timestamp)

	// Convert path to byte slice
	pathBytes := utils.EncodePath(pathParams, path)

	// Combine all byte slices
	encodedKey := append(timestampBytes, pathBytes...)
	encodedKey = append(encodedKey, subspaceId...)

	return encodedKey, nil
}

/* Decodes the key from the kv store into the timestamp, subspaceId, and path */
func DecodeKey(encodedKey []byte, pathParams types.PathParams[uint64]) (uint64, []byte, types.Path, error) {
	var timestamp uint64

	// Read timestamp from the encoded key
	timestampBytes := encodedKey[:8]
	timestamp = binary.BigEndian.Uint64(timestampBytes)

	// Decode path
	pathEndIndex, decodedPath, err := utils.DecodePath(pathParams, encodedKey[8:])
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to decode path: %w", err)
	}

	// Extract subspaceId
	subspaceId := encodedKey[8+pathEndIndex:]

	return timestamp, subspaceId, decodedPath, nil
}

// payloaddigest payloadlength and authdigest
func EncodeValues(PayloadLength uint64, PayloadDigest string, AuthDigest string) []byte {
	var buffer bytes.Buffer

	// Encode PayloadLength
	encodedLength := utils.BigIntToBytes(PayloadLength)
	buffer.Write(encodedLength)

	// Encode PayloadDigest
	encodedPayloadDigest := stringToBytes(PayloadDigest)
	buffer.Write(encodedPayloadDigest)

	// Encode AuthDigest
	encodedAuthDigest := stringToBytes(AuthDigest)
	buffer.Write(encodedAuthDigest)

	return buffer.Bytes()
}

// Takes a byte array and returns payloadLength, payloadDigest, authDigest
func DecodeValues(encoded []byte) (uint64, string, string) {
	// Decode PayloadLength
	payloadLength := binary.BigEndian.Uint64(encoded[:8])
	remaining := encoded[8:]

	// Decode PayloadDigest
	payloadDigest, remaining := bytesToString(remaining)

	// Decode AuthDigest
	authDigest, _ := bytesToString(remaining)

	return payloadLength, payloadDigest, authDigest
}

// stringToBytes converts a string to a byte slice with a length prefix.
func stringToBytes(str string) []byte {
	length := uint64(len(str))
	lengthBytes := utils.BigIntToBytes(length)
	return append(lengthBytes, []byte(str)...)
}

// BytesToString converts a byte slice with a length prefix to a string.
func bytesToString(b []byte) (string, []byte) {
	length := binary.BigEndian.Uint64(b[:8])
	strBytes := b[8 : 8+length]
	return string(strBytes), b[8+length:]
}
