package utils

import (
	"encoding/binary"
	"math/big"
)

// Encode a bigint
func BigintToBytes(bigint *big.Int) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bigint.Uint64())
	return bytes
}

func DecodeCompactWidth(encoded []bytes)
