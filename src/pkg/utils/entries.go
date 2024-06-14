package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/src/pkg/types"
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
		EncodeNamespace func(namespace NamespaceKey) []byte
		EncodeSubspace  func(subspace SubspaceKey) []byte
		encodePayload   func(digest PayloadDigest) []byte
		PathParams      types.PathParams[ValueType]
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
		opts.encodePayload(entry.Payload_digest)...)

	return result

}
