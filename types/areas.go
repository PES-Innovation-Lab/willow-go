package types

type Area struct {
	Subspace_id  SubspaceId
	Path         Path
	Times        Range[uint64]
	Any_subspace bool
}

type AreaOfInterest struct {
	Area     Area
	MaxCount uint64
	MaxSize  uint64
}
