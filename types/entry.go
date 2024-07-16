// Types related to Entry in the data model
package types

import "golang.org/x/exp/constraints"

type Entry[PayloadDigest constraints.Ordered] struct {
	// ID of the namespace the Entry is a part of
	Namespace_id NamespaceId
	// ID of the subspace to which the Entry belongs to
	Subspace_id SubspaceId
	// The hashed payload
	Payload_digest PayloadDigest
	// The path which the entry has
	Path Path
	// The time at which the entry was created in microseconds
	Timestamp uint64
	// The length of the payload
	Payload_length uint64
}
