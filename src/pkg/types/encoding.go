package types

import "golang.org/x/exp/constraints"

// TO DO !!!!
// Define Growing Bytes here
// figure out the promise type thing

type EncodingScheme[T any, K constraints.Ordered] interface {
	Encode(value T) []uint8
	Decode(encoded []uint8) (T, error)
	EncodedLength(value T) K
	DecodeStream(encoded []uint8) (T, error)
}
