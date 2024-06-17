package types

type Area[SubspaceId OrderableGeneric] struct {
	Subspace_id *SubspaceId
	Path        Path
	Times       Range[uint64]
}

type AreaOfInterest[SubspaceId OrderableGeneric] struct {
	Area      Area[SubspaceId]
	Max_count uint64
	Max_size  uint64
}
