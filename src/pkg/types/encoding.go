package types

//TO DO !!!!
// Define Growing Bytes here
// figure out the promise type thing

type EncodingScheme[T interface{}] interface {
	Encode(value T) []uint8
	Decode(encoded []uint8) (T, error)
	EncodedLength(value T)
	DecodeStream(encoded []uint8) (T, error)
}
