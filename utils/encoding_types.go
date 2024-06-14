package utils

import "golang.org/x/exp/constraints"

// TO DO !!!!
// Define Growing Bytes here
// figure out the promise type thing

type StreamDecoder[ValueType any] func(value GrowingBytes) (chan ValueType, error)

type EncodingScheme[ValueType any, K constraints.Unsigned] interface {
	Encode(value ValueType) []byte
	Decode(encoded []byte) (TValueType error)
	EncodedLength(value ValueType) K
	DecodeStream() StreamDecoder[ValueType]
}

type PrivyEncodingScheme[ValueType, PrivyType any, K constraints.Unsigned] interface {
	// e.g. Value type here is a SetupBindReadCapability
	// the privy type is what both sides know - in this case,
	// the outer area and the namespace.
	Encode(value ValueType, privy PrivyType) []byte
	Decode(encoded []byte, privy PrivyType) (ValueType, error)
	EncodedLength(value ValueType, privy PrivyType) K
	DecodeStream() StreamDecoder[ValueType]
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
