package utils

import (
	"encoding/binary"
	"fmt"
	"math"

	types "github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

/** Returns the `Position3d` of an `Entry`. */
func EntryPosition(entry types.Entry) types.Position3d {
	return types.Position3d{
		Time:     entry.Timestamp,
		Path:     entry.Path,
		Subspace: entry.Subspace_id,
	}
}

/* Encode an `Entry`.

https://willowprotocol.org/specs/encodings/index.html#enc_entry
*/

func EncodeEntry[ValueType constraints.Unsigned](
	opts struct {
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(digest types.PayloadDigest) []byte
		PathParams          types.PathParams[ValueType]
	},
	entry types.Entry,
) []byte {
	result := append(
		append(
			append(
				append(
					append(
						opts.EncodeNamespace(entry.Namespace_id),
						opts.EncodeSubspace(entry.Subspace_id)...),
					EncodePath(opts.PathParams, entry.Path)...), //EncodePath to be defined
				BigIntToBytes(entry.Timestamp)...),
			BigIntToBytes(entry.Payload_length)...),
		opts.EncodePayloadDigest(entry.Payload_digest)...)

	return result

}

/** Decode bytes to an `Entry`.
 *
 * https://willowprotocol.org/specs/encodings/index.html#enc_entry
 */
func DecodeEntry[ValueType constraints.Unsigned](
	opts struct {
		NameSpaceScheme EncodingScheme[types.NamespaceId]
		SubSpaceScheme  EncodingScheme[types.SubspaceId]
		PayloadScheme   EncodingScheme[types.PayloadDigest]
		PathScheme      types.PathParams[ValueType]
	},
	encEntry []byte,
) (types.Entry, error) {
	// first get the namespace.
	namespaceId, err := opts.NameSpaceScheme.Decode(encEntry)
	if err != nil {
		return types.Entry{}, fmt.Errorf("failed to decode namespace: %w", err)
	}
	// now get the the subSpace after finding starting position
	subspacePos := opts.NameSpaceScheme.EncodedLength(namespaceId)
	subspaceId, err := opts.SubSpaceScheme.Decode(encEntry[subspacePos:])
	if err != nil {
		return types.Entry{}, fmt.Errorf("failed to decode subspace: %w", err)
	}
	// Similar approach for Path but decoded as a stream instead
	pathPos := subspacePos + opts.SubSpaceScheme.EncodedLength(subspaceId)

	pathStream := make(chan []byte, 10)

	encPath := encEntry[pathPos:]

	pathBytes := NewGrowingBytes(pathStream)

	go func() {
		for _, encByte := range encPath {
			pathStream <- []byte{encByte}
		}
	}()
	path := DecodePathStream(opts.PathScheme, pathBytes)
	// now get the timestamp

	timestampPos := pathPos + EncodePathLength(opts.PathScheme, path)
	timestamp := binary.BigEndian.Uint64(encEntry[timestampPos:])

	// timestamp takes up 8 bytes

	payloadLengthPos := timestampPos + 8
	payloadLength := binary.BigEndian.Uint64(encEntry[payloadLengthPos:])
	// payload digest takes up 8 bytes
	payloadDigestPos := payloadLengthPos + 8

	payloadDigest, err := opts.PayloadScheme.Decode(encEntry[payloadDigestPos:])
	if err != nil {
		return types.Entry{}, fmt.Errorf("failed to decode payloaddigest: %w", err)
	}
	// decoded entry
	return types.Entry{
		Namespace_id:   namespaceId,
		Subspace_id:    subspaceId,
		Path:           path,
		Timestamp:      timestamp,
		Payload_length: payloadLength,
		Payload_digest: payloadDigest,
	}, nil
}

/* Encode entry relative to another entry */

func EncodeEntryRelativeEntry[ValueType constraints.Unsigned](
	opts struct {
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(digest types.PayloadDigest) []byte
		IsEqualNamespace    func(a types.NamespaceId, b types.NamespaceId) bool
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		PathScheme          types.PathParams[ValueType]
	}, entry types.Entry,
	ref types.Entry,
) []byte {

	// Time difference
	timeDiff := AbsDiffuint64(entry.Timestamp, ref.Timestamp)

	var encodedNamespaceFlag int
	/* Are namespaces equal or not? Does it need to be encoded? */
	if !opts.IsEqualNamespace(entry.Namespace_id, ref.Namespace_id) {
		encodedNamespaceFlag = 0x80
	} else {
		encodedNamespaceFlag = 0x00
	}
	var encodedSubspaceFlag int
	/* Does subspace need to be encoded */
	if opts.OrderSubspace(entry.Subspace_id, ref.Subspace_id) != 0 {
		encodedSubspaceFlag = 0x40
	} else {
		encodedSubspaceFlag = 0x0
	}
	var addOrSubtractTimeDiff int
	// Add or subtract

	if entry.Timestamp > ref.Timestamp {
		addOrSubtractTimeDiff = 0x20
	} else {
		addOrSubtractTimeDiff = 0x0
	}
	fmt.Println(addOrSubtractTimeDiff)
	// 2-bit integer n such that 2^n gives compact_width(time_diff)
	compactWidthTimeDiffFlag := CompactWidthEndMasks[GetWidthMax64Int(timeDiff)] << 2
	// 2-bit integer n such that 2^n gives compact_width(e.payload_length)
	compactWidthPayloadLengthFlag := CompactWidthEndMasks[GetWidthMax64Int(entry.Payload_length)]

	var encodedNamespace []byte
	// Encoded namespace
	if encodedNamespaceFlag == 0x0 {
		encodedNamespace = []byte{}
	} else {
		encodedNamespace = opts.EncodeNamespace(entry.Namespace_id)
	}

	var encodedSubspace []byte
	// Encoded subspace
	if encodedSubspaceFlag == 0x0 {
		encodedSubspace = []byte{}
	} else {
		encodedSubspace = opts.EncodeSubspace(entry.Subspace_id)
	}

	encodedPath := EncodeRelativePath[ValueType](opts.PathScheme, entry.Path, ref.Path)
	// time_diff, encoded as an unsigned, big-endian compact_width(time_diff)-byte integer
	encodedTimeDiff := EncodeIntMax64(timeDiff)
	//e.payload_length, encoded as an unsigned, big-endian compact_width(e.payload_length)-byte integer
	encodedPayloadLength := EncodeIntMax64(entry.Payload_length)
	// e.payload_digest, encoded as a payload_digest
	encodedDigest := opts.EncodePayloadDigest(entry.Payload_digest)

	header := encodedNamespaceFlag | encodedSubspaceFlag | addOrSubtractTimeDiff | compactWidthPayloadLengthFlag | compactWidthTimeDiffFlag

	return append(append(
		append(
			append(
				append(
					append(
						[]byte{byte(header)},
						encodedNamespace...),
					encodedSubspace...),
				encodedPath...),
			encodedTimeDiff...),
		encodedPayloadLength...),
		encodedDigest...)
}

var CompactWidthEndMasks = map[int]int{
	1: 0x0,
	2: 0x1,
	4: 0x2,
	8: 0x3,
}

type DecodeResult struct {
	Entry types.Entry
	Err   error
}

// Decode an entry encoded relative to another `Entry` from GrowingBytes.

func DecodeStreamEntryRelativeEntry[ValueType constraints.Unsigned](
	opts struct {
		DecodeStreamNamespace     func(bytes *GrowingBytes) chan types.NamespaceId
		DecodeStreamSubspace      func(bytes *GrowingBytes) chan types.SubspaceId
		DecodeStreamPayloadDigest func(bytes *GrowingBytes) chan types.PayloadDigest
		PathScheme                types.PathParams[ValueType]
	}, bytes *GrowingBytes, ref types.Entry,
) chan DecodeResult {

	resultChan := make(chan DecodeResult, 1)
	go func() {
		firstByte := bytes.NextAbsolute(1)
		header := firstByte[0]

		isNamespaceEncoded := (header & 0x80) == 0x80
		isSubspaceEncoded := (header & 0x40) == 0x40
		addTimeDiff := (header & 0x20) == 0x20

		compactWidthTimeDiff := math.Pow(2, float64((header&0xc)>>2))
		compactWidthPayloadLength := math.Pow(2, float64(header&0x3))

		bytes.Prune(1)
		var namespaceId types.NamespaceId
		if isNamespaceEncoded {

			namespaceStream := opts.DecodeStreamNamespace(bytes)

			namespaceId = <-namespaceStream

		} else {
			namespaceId = ref.Namespace_id

		}

		var subspaceId types.SubspaceId
		if isSubspaceEncoded {

			subspaceStream := opts.DecodeStreamSubspace(bytes)

			subspaceId = <-subspaceStream

		} else {
			subspaceId = ref.Subspace_id

		}

		path := DecodeRelPathStream[ValueType](opts.PathScheme, bytes, ref.Path)

		timeDiffBytes := bytes.NextAbsolute(int(compactWidthTimeDiff))
		timeDiff, err := (DecodeIntMax64(timeDiffBytes[:int(compactWidthTimeDiff)]))
		if err != nil {
			resultChan <- DecodeResult{Err: fmt.Errorf("failed to decode time diff: %w", err)}
			return
		}

		bytes.Prune(int(compactWidthTimeDiff))

		payloadLengthBytes := bytes.NextAbsolute(int(compactWidthPayloadLength))
		payloadLength, err := (DecodeIntMax64(payloadLengthBytes[:int(compactWidthPayloadLength)]))
		if err != nil {
			resultChan <- DecodeResult{Err: fmt.Errorf("failed to decode time diff: %w", err)}
			return
		}

		bytes.Prune(int(compactWidthPayloadLength))

		payloadDigestChan := opts.DecodeStreamPayloadDigest(bytes)

		payloadDigest := <-payloadDigestChan

		var timestamp uint64
		if addTimeDiff {
			timestamp = ref.Timestamp + uint64(timeDiff)
		} else {
			timestamp = ref.Timestamp - uint64(timeDiff)
		}
		resultChan <- DecodeResult{types.Entry{
			Namespace_id:   namespaceId,
			Subspace_id:    subspaceId,
			Path:           path,
			Timestamp:      timestamp,
			Payload_length: payloadLength,
			Payload_digest: payloadDigest,
		}, nil}

	}()

	return resultChan

}

func EncodeEntryRelativeRange3d[ValueType constraints.Unsigned](
	opts struct {
		EncodeNamespace     func(namespace types.NamespaceId) []byte
		EncodeSubspace      func(subspace types.SubspaceId) []byte
		EncodePayloadDigest func(digest types.PayloadDigest) []byte
		OrderSubspace       types.TotalOrder[types.SubspaceId]
		PathScheme          types.PathParams[ValueType]
	}, entry types.Entry,
	outer types.Range3d,
) ([]byte, error) {
	var timeDiff uint64
	if !outer.TimeRange.OpenEnd {
		timeDiff = min(AbsDiffuint64(entry.Timestamp, outer.TimeRange.Start), AbsDiffuint64(entry.Timestamp, outer.TimeRange.End))
	} else {
		timeDiff = AbsDiffuint64(entry.Timestamp, outer.TimeRange.Start)
	}

	var encodedSubspaceIdFlag int
	var encodedSubspace []byte

	if opts.OrderSubspace(entry.Subspace_id, outer.SubspaceRange.Start) != 0 {
		encodedSubspaceIdFlag = 0x80
		encodedSubspace = opts.EncodeSubspace(entry.Subspace_id)
	} else {
		encodedSubspaceIdFlag = 0x00
		encodedSubspace = []byte{}
	}

	var encodePathRelativeToStartFlag int
	var encodedPath []byte

	if !outer.PathRange.OpenEnd {
		commonPrefixStart, err := CommonPrefix(entry.Path, outer.PathRange.Start)
		if err != nil {
			return []byte{}, err
		}
		commonPrefixEnd, err := CommonPrefix(entry.Path, outer.PathRange.End)
		if err != nil {
			return []byte{}, err
		}

		if len(commonPrefixStart) >= len(commonPrefixEnd) {
			encodePathRelativeToStartFlag = 0x40
			encodedPath = EncodeRelativePath[ValueType](opts.PathScheme, entry.Path, outer.PathRange.Start)

		} else {
			encodePathRelativeToStartFlag = 0x0
			encodedPath = EncodeRelativePath[ValueType](opts.PathScheme, entry.Path, outer.PathRange.End)
		}
	} else {
		encodePathRelativeToStartFlag = 0x40
		encodedPath = EncodeRelativePath[ValueType](opts.PathScheme, entry.Path, outer.PathRange.Start)
	}
	var applyTimeDiffWithStartOrEnd int
	if timeDiff == AbsDiffuint64(entry.Timestamp, outer.TimeRange.Start) {
		applyTimeDiffWithStartOrEnd = 0x20

	} else {
		applyTimeDiffWithStartOrEnd = 0x0
	}

	var addOrSubtractTimeDiffFlag int
	if !outer.TimeRange.OpenEnd {
		if (applyTimeDiffWithStartOrEnd == 0x20 && entry.Timestamp >= outer.TimeRange.Start) ||
			(applyTimeDiffWithStartOrEnd == 0x0 &&
				entry.Timestamp >= outer.TimeRange.End) {
			addOrSubtractTimeDiffFlag = 0x10
		} else {
			addOrSubtractTimeDiffFlag = 0x0
		}
	} else {
		if applyTimeDiffWithStartOrEnd == 0x20 &&
			entry.Timestamp >= outer.TimeRange.Start {
			addOrSubtractTimeDiffFlag = 0x10
		} else {
			addOrSubtractTimeDiffFlag = 0x0
		}
	}

	var timeDiffCompactWidthFlag = CompactWidthEndMasks[GetWidthMax64Int(timeDiff)] << 2
	var payloadLengthFlag = CompactWidthEndMasks[GetWidthMax64Int(entry.Payload_length)]

	var header = encodedSubspaceIdFlag | encodePathRelativeToStartFlag |
		applyTimeDiffWithStartOrEnd | addOrSubtractTimeDiffFlag |
		timeDiffCompactWidthFlag | payloadLengthFlag

	var encodedTimeDiff = EncodeIntMax64(timeDiff)

	var encodedPayloadLength = EncodeIntMax64(entry.Payload_length)

	var encodedPayloadDigest = opts.EncodePayloadDigest(entry.Payload_digest)

	return append(
		append(
			append(
				append(
					append(
						[]byte{byte(header)},
						encodedSubspace...),
					encodedPath...),
				encodedTimeDiff...),
			encodedPayloadLength...),
		encodedPayloadDigest...), nil
}

func DecodeStreamEntryRelativeRange3d[ValueType constraints.Unsigned](
	opts struct {
		DecodeStreamSubspace      func(bytes *GrowingBytes) chan types.SubspaceId
		DecodeStreamPayloadDigest func(bytes *GrowingBytes) chan types.PayloadDigest
		PathScheme                types.PathParams[ValueType]
	},
	bytes *GrowingBytes,
	outer types.Range3d,
	namespaceId types.NamespaceId,
) chan DecodeResult {
	resultChan := make(chan DecodeResult, 1)

	go func() {

		firstByte := bytes.NextAbsolute(1)
		header := firstByte[0]

		isSubspaceEncoded := (header & 0x80) == 0x80
		isPathEncodedRelativeToStart := (header & 0x40) == 0x40
		isTimeDiffCombinedWithStart := (header & 0x20) == 0x20
		addOrSubtractTimedDiff := (header & 0x10) == 0x10
		timeDiffCompactWidth := math.Pow(2, float64((header&0xc)>>2))
		payloadLengthCompactWidth := math.Pow(2, float64(header&0x3))

		var subspaceId types.SubspaceId

		bytes.Prune(1)

		if isSubspaceEncoded {
			subspaceId = <-opts.DecodeStreamSubspace(bytes)
		} else {
			subspaceId = outer.SubspaceRange.Start
		}

		var path types.Path

		if !isPathEncodedRelativeToStart {
			if outer.PathRange.OpenEnd {
				panic("The path cannot be encoded relative to an open end.")
			}
			path = DecodeRelPathStream[ValueType](opts.PathScheme, bytes, outer.PathRange.End)
		} else {
			path = DecodeRelPathStream[ValueType](opts.PathScheme, bytes, outer.PathRange.Start)
		}

		timeDiffBytes := bytes.NextAbsolute(int(timeDiffCompactWidth))
		timeDiff, err := DecodeIntMax64(timeDiffBytes[:int(timeDiffCompactWidth)])
		if err != nil {
			resultChan <- DecodeResult{Err: fmt.Errorf("failed to decode time diff: %w", err)}
			return

		}

		bytes.Prune(int(timeDiffCompactWidth))

		var timestamp uint64

		if isTimeDiffCombinedWithStart && addOrSubtractTimedDiff {
			timestamp = outer.TimeRange.Start + timeDiff
		} else if isTimeDiffCombinedWithStart && !addOrSubtractTimedDiff {
			timestamp = outer.TimeRange.Start - timeDiff
		} else if !isTimeDiffCombinedWithStart && addOrSubtractTimedDiff {
			if outer.TimeRange.OpenEnd {
				resultChan <- DecodeResult{Err: fmt.Errorf("cannot apply time diff to an open end")}
				return
			}

			timestamp = outer.TimeRange.End + timeDiff
		} else {
			if outer.TimeRange.OpenEnd {
				resultChan <- DecodeResult{Err: fmt.Errorf("cannot apply time diff to an open end")}
				return
			}

			timestamp = outer.TimeRange.End - timeDiff
		}

		payloadLengthBytes := bytes.NextAbsolute(int(payloadLengthCompactWidth))
		payloadLength, err := DecodeIntMax64(payloadLengthBytes[:int(payloadLengthCompactWidth)])
		if err != nil {
			resultChan <- DecodeResult{Err: fmt.Errorf("failed to decode payload %w", err)}
			return
		}

		bytes.Prune(int(payloadLengthCompactWidth))

		payloadDigest := <-opts.DecodeStreamPayloadDigest(bytes)

		resultChan <- DecodeResult{types.Entry{
			Namespace_id:   namespaceId,
			Subspace_id:    subspaceId,
			Path:           path,
			Timestamp:      timestamp,
			Payload_length: payloadLength,
			Payload_digest: payloadDigest,
		}, nil}

	}()

	return resultChan

}

func DefaultEntry(
	defaultNamespace types.NamespaceId,
	defaultSubspace types.SubspaceId,
	defaultPayloadDigest types.PayloadDigest,
) types.Entry {
	return types.Entry{
		Namespace_id:   defaultNamespace,
		Subspace_id:    defaultSubspace,
		Path:           types.Path{},
		Timestamp:      0,
		Payload_length: 0,
		Payload_digest: defaultPayloadDigest,
	}
}
