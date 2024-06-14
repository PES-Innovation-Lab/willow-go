package utils

import (
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

// Encode a bigint
func BigintToBytes(bigint uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bigint)
	return bytes
}

func GetWidthMax32Int[T constraints.Unsigned](num T) int {
	switch true {
	case uint(num) < 1<<8:
		return 1
	case uint(num) < 1<<16:
		return 2
	case uint(num) < 1<<24:
		return 3
	default:
		return 4
	}
}

func EncodeIntMax32[T constraints.Unsigned](num, max T) []byte {
	width := GetWidthMax32Int(max)

	bytes := make([]byte, width)

	switch width {
	case 1:
		bytes[0] = uint8(num)
	case 2:
		binary.BigEndian.PutUint16(bytes, uint16(num))
	case 3:
		bytes[0] = byte(num >> 16)
		binary.BigEndian.PutUint16(bytes[1:], uint16(num))
	case 4:
		binary.BigEndian.PutUint32(bytes, uint32(num))
	}

	return bytes
}

func DecodeIntMax32[T constraints.Unsigned](bytes []byte, max T) (uint32, error) {
	bytesToDecodeLength := GetWidthMax32Int(max)

	if len(bytes) != bytesToDecodeLength {
		return 0, errors.New("invalid byte slice length")
	}
	if bytesToDecodeLength > 4 {
		return 0, errors.New("cannot decode non-UintMax bytes")
	}

	switch bytesToDecodeLength {
	case 1:
		return uint32(bytes[0]), nil
	case 2:
		return uint32(binary.BigEndian.Uint16(bytes)), nil
	case 4:
		return binary.BigEndian.Uint32(bytes), nil
	}

	// Otherwise it's 24 bit.
	a := binary.BigEndian.Uint16(bytes[:2])
	b := uint32(bytes[2])

	return (uint32(a) << 8) + b, nil
}

func GetWidthMax64Int[T constraints.Unsigned](num T) int {
	switch true {
	case uint(num) < 1<<8:
		return 1
	case uint(num) < 1<<16:
		return 2
	case uint(num) < 1<<32:
		return 4
	default:
		return 8
	}
}

func EncodeIntMax64[T constraints.Unsigned](num T) []byte {
	width := GetWidthMax64Int(num)
	bytes := make([]byte, width)

	switch width {
	case 1:
		bytes[0] = uint8(num)
	case 2:
		binary.BigEndian.PutUint16(bytes, uint16(num))
	case 4:
		binary.BigEndian.PutUint32(bytes, uint32(num))
	case 8:
		binary.BigEndian.PutUint64(bytes, uint64(num))
	}

	fmt.Println(bytes, width, num)
	return bytes
}

func DecodeIntMax64(encoded []byte) any {
	// reader := bytes.NewReader(encoded)
	switch len(encoded) {
	case 1:
		return uint8(encoded[0])

	case 2:
		return binary.BigEndian.Uint16(encoded)

	case 4:
		return binary.BigEndian.Uint32(encoded)

	case 8:
		return binary.BigEndian.Uint64(encoded)
	default:
		panic("invalid length")
	}
}
