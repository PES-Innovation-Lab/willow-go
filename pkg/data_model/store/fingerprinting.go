package store

import (
	"log"
	"math"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
)

// "fmt"

// "github.com/PES-Innovation-Lab/willow-go/types"
// "github.com/PES-Innovation-Lab/willow-go/utils"

// absolutelty beautiful code
// TO DO : add routines??

func BuildFingerprints(entries []Kdtree.KDNodeKey) []string {
	// check for empty entries
	if len(entries) == 0 {
		log.Fatal(`Empty entry list, line 21, fingerprinting.go`)
	}

	temp := math.Pow(2, math.Ceil(math.Log2(float64(len(entries)))+1))
	fptree := make([]string, int(temp))
	mid := len(entries) / 2

	fptree[0] = xorStrings(
		buildHelper(entries[0:mid], fptree, 1),
		buildHelper(entries[mid:], fptree, 2),
	)

	return fptree
}

func buildHelper(entries []Kdtree.KDNodeKey, fps []string, index int) string {
	if len(entries) == 1 {
		fps[index] = entries[0].Fingerprint
		return entries[0].Fingerprint
	}

	mid := len(entries) / 2
	fingerprint := xorStrings(
		buildHelper(entries[0:mid], fps, 2*index+1),
		buildHelper(entries[mid:], fps, 2*index+2),
	)
	fps[index] = fingerprint
	return fingerprint
}

// ACTUAL DOG SHIT CODE BRO. how the fuck have i written something this stupid

// func BuildFingerprints(entries []Kdtree.KDNodeKey) []string {
// 	sort.Slice(entries, func(i, j int) bool {
// 		return entries[i].Timestamp < entries[j].Timestamp
// 	})

// 	// calculate the number of layers required in the tree
// 	level_log := math.Log2(float64(len(entries)))
// 	level_log = math.Ceil(level_log) // number of levels in the tree in float64
// 	level_count := uint64(level_log) // number of levels in the tree in uint64

// 	// going for a tree stored as an array
// 	// length of the array will be 2**(level_count - 1) + len(entries)
// 	max_len := int(math.Pow(2, level_log-1)) + len(entries) - 1
// 	fptree := make([]string, max_len)
// 	leaf_start_index := max_len - len(entries)

// 	// setting the leaf node values
// 	for _, entry := range entries {
// 		fptree[leaf_start_index] = entry.Fingerprint
// 		leaf_start_index++
// 	}

// 	// now need to calculate the internal nodes 🥵
// 	// each internal node will be the xor of it's children 💪

// 	// outermost loop decrements through the (max depth - 1) to root level of tree
// 	for i := level_count; i > 0; i-- {
// 		index := math.Pow(2, float64(i-1))
// 		for j := int(index) - 1; j < int(math.Pow(2, float64(i))-1); j++ {
// 			if j < len(entries) {
// 				fptree[j] = xorStrings(fptree[2*j+1], fptree[2*j+2])
// 			}
// 		}
// 	}
// 	return fptree
// }

// Placeholder for xorStrings function
func xorStrings(a, b string) string {
	// Implement the actual XOR logic for simplicity
	return a + b
}

// func xorStrings(a, b string) string {
// 	// ensure both strings have the same length
// 	// they should always be the same length 🙄
// 	// if len(a) > len(b) {
// 	// 	b += string(make([]byte, len(a)-len(b)))
// 	// } else if len(b) > len(a) {
// 	// 	a += string(make([]byte, len(b)-len(a)))
// 	// }

// 	if len(a) != len(b) {
// 		log.Fatal("Hashes of payloads are of different length 😨, fingerprinting.go, line 63")
// 	}

// 	result := make([]byte, len(a))
// 	for i := range a {
// 		result[i] = a[i] ^ b[i]
// 	}
// 	return string(result)
// }
