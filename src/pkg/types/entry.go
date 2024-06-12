// Types related to Entry in the data model
package types

type Entry[Timestamp uint64, NamespaceId, SubspaceId, PayloadDigest any] struct {
	// ID of the namespace the Entry is a part of
	namespace_id NamespaceId
	// ID of the subspace to which the Entry belongs to
	subspace_id SubspaceId
	// The path which the entry has
	path Path
	// The time at which the entry was created in microseconds
	timestamp Timestamp
	// The length of the payload
	payload_length uint64
	// The hashed payload
	payload_digest PayloadDigest
}
