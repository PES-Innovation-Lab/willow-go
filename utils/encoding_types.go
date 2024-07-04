package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// TO DO !!!!
// Define Growing Bytes here
// figure out the promise type thing

type EncodingScheme[ValueType types.OrderableGeneric, K constraints.Unsigned] struct {
	Encode        func(value ValueType) []byte
	Decode        func(encoded []byte) (ValueType, error)
	EncodedLength func(value ValueType) K
	DecodeStream  func(value *GrowingBytes) (ValueType, error)
}

type PrivyEncodingScheme[ValueType types.OrderableGeneric, PrivyType any, K constraints.Unsigned] struct {
	// e.g. Value type here is a SetupBindReadCapability
	// the privy type is what both sides know - in this case,
	// the outer area and the namespace.
	Encode        func(value ValueType, privy PrivyType) []byte
	Decode        func(encoded []byte, privy PrivyType) (ValueType, error)
	EncodedLength func(value ValueType, privy PrivyType) K
	DecodeStream  func(value *GrowingBytes) (ValueType, error)
	// Although it would seem natural to put the privy type in the params,
	// Before calling this function we cannot know what this message is privy to.
	// e.g. if this is an encoded SetupBindReadCapability
	// then we can only know what we are privy to once we have the intersection handle
	// encoded in the message.
	// we then need to dereference that handle to get the privy info
	// e.g. the outer area of the intersection handle.
	// sometimes we have to dereference something in the message,
	// sometimes we don't (e.g. in reconciliation messages.)
}
