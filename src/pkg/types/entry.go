// Types related to Entry in the data model
package types

import "golang.org/x/exp/constraints"

type Entry[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered] struct {
	// ID of the namespace the Entry is a part of
	namespace_id NamespaceId
	// ID of the subspace to which the Entry belongs to
	subspace_id SubspaceId
	// The path which the entry has
	path Path
	// The hashed payload
	payload_digest PayloadDigest
	// The time at which the entry was created in microseconds
	timestamp uint64
	// The length of the payload
	payload_length uint64
}
