package tests

// import (
// 	"encoding/binary"
// 	"reflect"
// 	"testing"

// 	"github.com/PES-Innovation-Lab/willow-go/types"
// 	"github.com/PES-Innovation-Lab/willow-go/utils"
// )

// func TestEncodeDecodeEntry(t *testing.T) {
// 	type PayloadDigest uint64
// 	type ValueType uint64

// 	// Sample encode and decode functions
// 	encodeNamespace := func(namespace types.NamespaceId) []byte {
// 		return utils.BigIntToBytes(uint64(namespace))
// 	}
// 	encodeSubspace := func(subspace SubspaceKey) []byte {
// 		return utils.BigIntToBytes(uint64(subspace))
// 	}
// 	encodePayloadDigest := func(digest PayloadDigest) []byte {

// 		return utils.BigIntToBytes(uint64(digest))
// 	}
// 	decodeNamespace := func(data []byte) (NamespaceKey, error) {
// 		return NamespaceKey(binary.BigEndian.Uint64(data)), nil
// 	}
// 	decodeSubspace := func(data []byte) (SubspaceKey, error) {
// 		return SubspaceKey(binary.BigEndian.Uint64(data)), nil
// 	}
// 	decodePayloadDigest := func(data []byte) (PayloadDigest, error) {

// 		return PayloadDigest(binary.BigEndian.Uint64(data)), nil
// 	}

// 	namespaceScheme := utils.EncodingScheme[NamespaceKey, ValueType]{
// 		Decode: decodeNamespace,
// 		EncodedLength: func(_ NamespaceKey) ValueType {
// 			return 8
// 		},
// 	}
// 	subspaceScheme := utils.EncodingScheme[SubspaceKey, ValueType]{
// 		Decode:        decodeSubspace,
// 		EncodedLength: func(_ SubspaceKey) ValueType { return 8 },
// 	}
// 	payloadScheme := utils.EncodingScheme[PayloadDigest, ValueType]{
// 		Decode:        decodePayloadDigest,
// 		EncodedLength: func(_ PayloadDigest) ValueType { return 8 },
// 	}

// 	pathParams := types.PathParams[ValueType]{}

// 	entry := types.Entry[NamespaceKey, SubspaceKey, PayloadDigest]{
// 		Namespace_id:   1,
// 		Subspace_id:    2,
// 		Path:           types.Path{{0x01, 0x02, 0x03}},
// 		Timestamp:      1234567890,
// 		Payload_length: 3,
// 		Payload_digest: 4,
// 	}

// 	encodedEntry := utils.EncodeEntry(
// 		struct {
// 			EncodeNamespace     func(namespace types.NamespaceId) []byte
// 			EncodeSubspace      func(subspace types.SubspaceId) []byte
// 			EncodePayloadDigest func(digest PayloadDigest) []byte
// 			PathParams          types.PathParams[ValueType]
// 		}{
// 			EncodeNamespace:     encodeNamespace,
// 			EncodeSubspace:      encodeSubspace,
// 			EncodePayloadDigest: encodePayloadDigest,
// 			PathParams:          pathParams,
// 		},
// 		entry,
// 	)

// 	decodedEntry, err := utils.DecodeEntry(
// 		struct {
// 			NameSpaceScheme utils.EncodingScheme[NamespaceKey, ValueType]
// 			SubSpaceScheme  utils.EncodingScheme[SubspaceKey, ValueType]
// 			PayloadScheme   utils.EncodingScheme[PayloadDigest, ValueType]
// 			PathScheme      types.PathParams[ValueType]
// 		}{
// 			NameSpaceScheme: namespaceScheme,
// 			SubSpaceScheme:  subspaceScheme,
// 			PayloadScheme:   payloadScheme,
// 			PathScheme:      pathParams,
// 		},
// 		encodedEntry,
// 	)

// 	if err != nil {
// 		t.Fatalf("Failed to decode entry: %v", err)
// 	}

// 	if !reflect.DeepEqual(entry, decodedEntry) {
// 		t.Fatalf("Original and decoded entries do not match. Original: %+v, Decoded: %+v", entry, decodedEntry)
// 	}

// }

// func TestEntryRelativeEntryEncodeDecode(t *testing.T) {
// 	type NamespaceKey uint
// 	type SubspaceKey uint
// 	type PayloadDigest uint
// 	type ValueType uint64

// 	// Sample encode and decode functions
// 	encodeNamespace := func(namespace NamespaceKey) []byte {
// 		return utils.BigIntToBytes(uint64(namespace))
// 	}
// 	encodeSubspace := func(subspace SubspaceKey) []byte {
// 		return utils.BigIntToBytes(uint64(subspace))
// 	}
// 	encodePayloadDigest := func(digest PayloadDigest) []byte {

// 		return utils.BigIntToBytes(uint64(digest))
// 	}

// 	isEqualNamespace := func(a NamespaceKey, b NamespaceKey) bool {
// 		return a == b
// 	}

// 	orderSubspace := func(a, b SubspaceKey) types.Rel {
// 		if a < b {
// 			return types.Less
// 		} else if a > b {
// 			return types.Greater
// 		}
// 		return types.Equal
// 	}

// 	pathParams := types.PathParams[ValueType]{}

// 	namespaceA := NamespaceKey(1)
// 	namespaceB := NamespaceKey(2)
// 	subspaceA := SubspaceKey(10)
// 	pathA := types.Path{[]byte{1}, []byte{2}}
// 	pathB := types.Path{[]byte{1}}
// 	payloadLengthA := uint64(1024)
// 	payloadLengthB := uint64(1024)
// 	payloadDigestA := PayloadDigest(875)// Replace with actual digest value
// 	payloadDigestB := PayloadDigest(875) // Replace with actual digest value
// 	timestampA := uint64(1000)
// 	timestampB := uint64(500)

// 	entryA := types.Entry[NamespaceKey, SubspaceKey, PayloadDigest]{
// 		Namespace_id:   namespaceA,
// 		Subspace_id:    subspaceA,
// 		Path:           pathA,
// 		Timestamp:      timestampA,
// 		Payload_length: payloadLengthA,
// 		Payload_digest: payloadDigestA,
// 	}

// 	entryB := types.Entry[NamespaceKey, SubspaceKey, PayloadDigest]{
// 		Namespace_id:   namespaceB,
// 		Subspace_id:    subspaceA,
// 		Path:           pathB,
// 		Timestamp:      timestampB,
// 		Payload_length: payloadLengthB,
// 		Payload_digest: payloadDigestB,
// 	}

// 	encodeOpts := struct {
// 		EncodeNamespace     func(namespace NamespaceKey) []byte
// 		EncodeSubspace      func(subspace SubspaceKey) []byte
// 		EncodePayloadDigest func(digest PayloadDigest) []byte
// 		IsEqualNamespace    func(a, b NamespaceKey) bool
// 		OrderSubspace       types.TotalOrder[SubspaceKey]
// 		PathScheme          types.PathParams[ValueType]
// 	}{
// 		EncodeNamespace:     encodeNamespace,
// 		EncodeSubspace:      encodeSubspace,
// 		EncodePayloadDigest: encodePayloadDigest,
// 		IsEqualNamespace:    isEqualNamespace,
// 		OrderSubspace:       orderSubspace,
// 		PathScheme:          pathParams,
// 	}

// 	ref := entryA
// 	entry := entryB

// 	encodedEntry := utils.EncodeEntryRelativeEntry(
// 		encodeOpts, entry, ref,
// 	)

// 	stream := make(chan []byte, 64)
// 	bytes := utils.NewGrowingBytes(stream)

// 	go func() {
// 		for _, encByte := range encodedEntry {
// 			stream <- []byte{encByte}

// 		}
// 	}()

// 	decodeStreamNamespace := func(bytes *utils.GrowingBytes) chan NamespaceKey {
// 		namespaceChan := make(chan NamespaceKey, 1)
// 		accumaltedBytes := bytes.NextAbsolute(8)
// 		namespace := accumaltedBytes[0:8]

// 		bytes.Prune(8)
// 		namespaceChan <- NamespaceKey(binary.BigEndian.Uint64(namespace))
// 		return namespaceChan
// 	}
// 	decodeStreamSubspace := func(bytes *utils.GrowingBytes) chan SubspaceKey {
// 		subspaceChan := make(chan SubspaceKey, 1)
// 		accumaltedBytes := bytes.NextAbsolute(8)
// 		subspace := accumaltedBytes[0:8]

// 		bytes.Prune(8)
// 		subspaceChan <- SubspaceKey(binary.BigEndian.Uint64(subspace))
// 		return subspaceChan
// 	}
// 	decodeStreamPayloadDigest := func(bytes *utils.GrowingBytes) chan PayloadDigest {
// 		payloadDigestChan := make(chan PayloadDigest, 1)
// 		accumulatedBytes := bytes.NextAbsolute(8)

// 		payloadDigest := accumulatedBytes[0:8]

// 		bytes.Prune(8)
// 		payloadDigestChan <- PayloadDigest(binary.BigEndian.Uint64(payloadDigest))
// 		return payloadDigestChan
// 	}

// 	decodeOpts := struct {
// 		DecodeStreamNamespace     func(bytes *utils.GrowingBytes) chan NamespaceKey
// 		DecodeStreamSubspace      func(bytes *utils.GrowingBytes) chan SubspaceKey
// 		DecodeStreamPayloadDigest func(bytes *utils.GrowingBytes) chan PayloadDigest
// 		PathScheme                types.PathParams[ValueType]
// 	}{
// 		DecodeStreamNamespace:     decodeStreamNamespace,
// 		DecodeStreamSubspace:      decodeStreamSubspace,
// 		DecodeStreamPayloadDigest: decodeStreamPayloadDigest,
// 		PathScheme:                pathParams,
// 	}

// 	decodedEntryChan := utils.DecodeStreamEntryRelativeEntry(
// 		decodeOpts, bytes, ref,
// 	)
// 	result := <-decodedEntryChan
// 	if result.Err != nil {
// 		t.Fatalf("Failed to decode entry: %v", result.Err)

// 	}

// 	if !reflect.DeepEqual(entry, result.Entry) {
// 		t.Fatalf("Original and decoded entries do not match. Original: %+v, Decoded: %+v", entry, result.Entry)
// 	}

// }
