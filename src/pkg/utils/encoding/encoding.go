package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
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

func EncodeIntMax32[T constraints.Unsigned](num, max T) []byte {
	width := GetWidthMax32Int(max)
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

func DecodeIntMax32[T constraints.Unsigned](bytes []byte, max uint32) (uint32, error) {
	bytesToDecodeLength := GetWidthMax32Int(max)

	if bytesToDecodeLength > 4 {
		return 0, errors.New("Cannot decode non-UintMax bytes")
	}

	if bytesToDecodeLength == 1 {
		return uint32(bytes[0]), nil
	} else if bytesToDecodeLength == 2 {
		return uint32(binary.BigEndian.Uint16(bytes)), nil
	} else if bytesToDecodeLength == 4 {
		return binary.BigEndian.Uint32(bytes), nil
	}

	// Otherwise it's 24 bit.
	a := binary.BigEndian.Uint16(bytes[:2])
	b := uint32(bytes[2])

	return (uint32(a) << 8) + b, nil
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

func EncodeIntMax64[T constraints.Unsigned](num, max T) []byte {
	width := GetWidthMax64Int(max)
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

func DecodeIntMax64[T constraints.Unsigned](encoded []byte, max uint32) any {
	reader := bytes.NewReader(encoded)

	switch len(encoded) {
	case 1:
		var val uint8
		binary.Read(reader, binary.BigEndian, &val)
		return val
	case 2:
		var val uint16
		binary.Read(reader, binary.BigEndian, &val)
		return val
	case 4:
		var val uint32
		binary.Read(reader, binary.BigEndian, &val)
		return val
	default:
		var val uint64
		binary.Read(reader, binary.BigEndian, &val)
		return val
	}

}