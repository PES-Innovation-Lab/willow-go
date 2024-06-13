package utils

import (
	"encoding/binary"
	"math/big"

	"golang.org/x/exp/constraints"
)

// Encode a bigint
func BigintToBytes(bigint *big.Int) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bigint.Uint64())
	return bytes
}
func GetWidthMax32Int[T constraints.Unsigned](num T) int {
	switch true {
	case int(num) < 1<<8:
		return 1
	case int(num) < 1<<16:
		return 2
	case int(num) < 1<<24:
		return 3
	default:
		return 4
	}
}

func EncodingIntMax32[T constraints.Unsigned](num, max T) []byte {
	width := GetWidthMax32Int(num)
	bytes := make([]byte, width)

	switch width {
	case 1:
		bytes[0] = uint8(num)
	case 2:
		binary.BigEndian.PutUint16(bytes, uint16(num))
	case 3:
		binary.BigEndian.PutUint16(bytes, uint16(num))
		bytes[2] = byte(num & 0xff)
	case 4:
		binary.BigEndian.PutUint32(bytes, uint32(num))
	}

	return bytes
}

func GetWidthMax64Int[T constraints.Unsigned](num T) int {
	switch true {
	case int(num) < 1<<8:
		return 1
	case int(num) < 1<<16:
		return 2
	case int(num) < 1<<32:
		return 4
	default:
		return 8
	}
}

func EncodingIntMax64[T constraints.Unsigned](num, max T) []byte {
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

	return bytes
}

func DecodeCompactWidth(encoded []bytes)
