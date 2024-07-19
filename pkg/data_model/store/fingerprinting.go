package store

import (
	"math"
	"sort"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
)

// "fmt"

// "github.com/PES-Innovation-Lab/willow-go/types"
// "github.com/PES-Innovation-Lab/willow-go/utils"

func BuildFingerprints(entries []Kdtree.KDNodeKey) []string {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})

	// calculate the number of layers required in the tree
	level_log := math.Log2(float64(len(entries)))
	level_log = math.Ceil(level_log) // number of levels in the tree in float64
	level_count := uint64(level_log) // number of levels in the tree in uint64

	// going for a tree stored as an array
	// length of the array will be 2**(level_count - 1) + len(entries)
	max_len := int(math.Pow(2, level_log-1)) + len(entries) - 1
	fptree := make([]string, max_len)
	leaf_start_index := max_len - len(entries)

	// setting the leaf node values
	for _, entry := range entries {
		fptree[leaf_start_index] = entry.Fingerprint
		leaf_start_index++
	}

	// now need to calculate the internal nodes 🥵
	// each internal node will be the xor of it's children 💪

	// outermost loop decrements through the (max depth - 1) to root level of tree
	for i := level_count; i > 0; i-- {
		index := math.Pow(2, float64(i-1))
		for j := int(index) - 1; j < int(math.Pow(2, float64(i))-1); j++ {
			if j < len(entries) {
				fptree[j] = xorStrings(fptree[2*j+1], fptree[2*j+2])
			}
		}
	}
	return fptree
}

func xorStrings(a, b string) string {
	// ensure both strings have the same length
	if len(a) > len(b) {
		b += string(make([]byte, len(a)-len(b)))
	} else if len(b) > len(a) {
		a += string(make([]byte, len(b)-len(a)))
	}

	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return string(result)
}
