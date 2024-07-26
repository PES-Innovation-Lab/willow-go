package store

// type FPNodes struct {
// 	Range       types.Range3d
// 	Fingerprint string
// 	Covers      uint64 // what the fuck is this supposed to do??
// }

// /* Used for splitting a 3dRange*/
// func SplitRange(Range types.Range3d, size int) {

// }

// func Summarise(Range types.Range3d) struct {
// 	FingerPrint string
// 	size        uint64
// } {
// 	var size uint64
// 	valuesInRange :=

// }

// import (
// 	"log"
// 	"math"

// 	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
// )

// // "fmt"

// // "github.com/PES-Innovation-Lab/willow-go/types"
// // "github.com/PES-Innovation-Lab/willow-go/utils"

// // absolutelty beautiful code
// // TO DO : add routines??

// type Node struct {
// 	Index int
// 	Hash  string
// }

// func BuildFingerprints(entries []kdnode.Key) []Node {
// 	// check for empty entries
// 	if len(entries) == 0 {
// 		log.Fatal(`Empty entry list, line 21, fingerprinting.go`)
// 	}

// 	temp := math.Pow(2, math.Ceil(math.Log2(float64(len(entries)))+1))
// 	fptree := make([]string, int(temp))
// 	mid := len(entries) / 2

// 	fptree[0] = xorStrings(
// 		buildHelper(entries[0:mid], fptree, 1),
// 		buildHelper(entries[mid:], fptree, 2),
// 	)
// 	fptree = ShortenArray(fptree)
// 	nodeTree := make([]Node, int(temp))
// 	for i, fp := range fptree {
// 		newNode := Node{
// 			Index: i,
// 			Hash:  fp,
// 		}
// 		nodeTree[i] = newNode
// 	}

// 	return nodeTree
// }

// func buildHelper(entries []kdnode.Key, fps []string, index int) string {
// 	if len(entries) == 1 {
// 		fps[index] = entries[0].Fingerprint
// 		return entries[0].Fingerprint
// 	}

// 	mid := len(entries) / 2
// 	fingerprint := xorStrings(
// 		buildHelper(entries[0:mid], fps, 2*index+1),
// 		buildHelper(entries[mid:], fps, 2*index+2),
// 	)
// 	fps[index] = fingerprint
// 	return fingerprint
// }

// // ShortenArray shortens the array to include elements up to the last non-empty string.
// func ShortenArray(arr []string) []string {
// 	lastIndex := -1
// 	for i := len(arr) - 1; i >= 0; i-- {
// 		if arr[i] != "" {
// 			lastIndex = i
// 			break
// 		}
// 	}

// 	// if all elements are empty, return an empty slice
// 	if lastIndex == -1 {
// 		return []string{}
// 	}

// 	// return the slice up to the last non-empty string
// 	return arr[:lastIndex+1]
// }

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
