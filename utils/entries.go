package utils

import (
	"encoding/binary"

	types "github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

/** Returns the `Position3d` of an `Entry`. */
func EntryPosition[NamespaceKey, SubspaceKey, PayloadDigest constraints.Ordered](entry types.Entry[NamespaceKey, SubspaceKey, PayloadDigest]) types.Position3d[SubspaceKey] {
	return types.Position3d[SubspaceKey]{
		Time:     entry.Timestamp,
		Path:     entry.Path,
		Subspace: entry.Subspace_id,
	}
}

/* Encode an `Entry`.

https://willowprotocol.org/specs/encodings/index.html#enc_entry
*/

func EncodeEntry[NamespaceKey, SubspaceKey, PayloadDigest constraints.Ordered, ValueType constraints.Unsigned](
	opts struct {
		EncodeNamespace     func(namespace NamespaceKey) []byte
		EncodeSubspace      func(subspace SubspaceKey) []byte
		EncodePayloadDigest func(digest PayloadDigest) []byte
		PathParams          types.PathParams[ValueType]
	},
	entry types.Entry[NamespaceKey, SubspaceKey, PayloadDigest],
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

func DecodeEntry[NamespaceKey, SubspaceKey, PayloadDigest constraints.Ordered, ValueType constraints.Unsigned](
	opts struct {
		NameSpaceScheme EncodingScheme[NamespaceKey, ValueType]
		SubSpaceScheme  EncodingScheme[SubspaceKey, ValueType]
		PayloadScheme   EncodingScheme[PayloadDigest, ValueType]
		PathScheme      types.PathParams[ValueType]
	},
	encEntry []byte,
) types.Entry[NamespaceKey, SubspaceKey, PayloadDigest] {

	namespaceId, err := opts.NameSpaceScheme.Decode(encEntry)
	if err != nil {
		panic(err)
	}

	subspacePos := opts.NameSpaceScheme.EncodedLength(namespaceId)
	subspaceId, err := opts.SubSpaceScheme.Decode(encEntry[subspacePos:])
	if err != nil {
		panic(err)
	}

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

	timestampPos := pathPos + EncodePathLength(opts.PathScheme, path)
	timestamp := binary.BigEndian.Uint64(encEntry[timestampPos:])

	payloadLengthPos := timestampPos + 8
	payloadLength := binary.BigEndian.Uint64(encEntry[payloadLengthPos:])
	payloadDigestPos := payloadLength + 8

	payloadDigest, err := opts.PayloadScheme.Decode(encEntry[payloadDigestPos:])
	if err != nil {
		panic(err)
	}

	return types.Entry[NamespaceKey, SubspaceKey, PayloadDigest]{
		Namespace_id:   namespaceId,
		Subspace_id:    subspaceId,
		Path:           path,
		Timestamp:      timestamp,
		Payload_length: payloadLength,
		Payload_digest: payloadDigest,
	}
}

func EncodeEntryRelativeEntry[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered, ValueType constraints.Unsigned](
	opts struct {
		EncodeNamespace     func(namespace NamespaceId) []byte
		EncodeSubspace      func(subspace SubspaceId) []byte
		EncodePayloadDigest func(digest PayloadDigest) []byte
		IsEqualNamespace    func(a NamespaceId, b NamespaceId) bool
		OrderSubspace       types.TotalOrder[SubspaceId]
		PathScheme          types.PathParams[ValueType]
	}, entry types.Entry[NamespaceId, SubspaceId, PayloadDigest],
	ref types.Entry[NamespaceId, SubspaceId, PayloadDigest],
) []byte {
	timeDiff := AbsDiffuint64(entry.Timestamp, ref.Timestamp)
	var encodeNamespaceFlag int
	if !opts.IsEqualNamespace(entry.Namespace_id, ref.Namespace_id) {
		encodeNamespaceFlag = 0x80
	} else {
		encodeNamespaceFlag = 0x00
	}
	var encodeSubspaceFlag int

	if opts.OrderSubspace(entry.Subspace_id, ref.Subspace_id) != 0 {
		encodeSubspaceFlag = 0x40
	} else {
		encodeSubspaceFlag = 0x0
	}

}
