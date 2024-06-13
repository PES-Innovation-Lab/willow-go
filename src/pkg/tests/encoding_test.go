package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/src/pkg/utils"
)

func TestBigintToBytes(t *testing.T) {
	testCases := []struct {
		input    uint64
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{18446744073709551615, []byte{255, 255, 255, 255, 255, 255, 255, 255}},
	}

	for _, tc := range testCases {
		result := utils.BigintToBytes(tc.input)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("BigintToBytes(%d) = %v; expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestEncodeDecodeIntMax32(t *testing.T) {
	testCases := []struct {
		num      uint32
		max      uint32
		expected uint32
	}{
		{0, 255, 0},
		{1000, 65535, 1000},
		{459629, 16777215, 459629},
		{16777216, 4294967295, 16777216},
	}

	for _, tc := range testCases {
		encoded := utils.EncodeIntMax32(tc.num, tc.max)
		fmt.Printf("encoded : %v\n", encoded)
		decoded, err := utils.DecodeIntMax32(encoded, tc.max)
		if err != nil {
			t.Errorf("Error decoding: %v", err)
		}
		if decoded != tc.expected {
			t.Errorf("DecodeIntMax32(%d, %d) = %d; expected %d", tc.num, tc.max, decoded, tc.expected)
		}
	}
}

func TestEncodeDecodeIntMax64(t *testing.T) {
	testCases := []struct {
		num      uint64
		max      uint64
		expected uint64
	}{
		{0, 255, 0},
		{1000, 65535, 1000},
		{65536, 4294967295, 65536},
		{4294967296, 18446744073709551615, 4294967296},
	}

	for _, tc := range testCases {
		encoded := utils.EncodeIntMax64(tc.num, tc.max)
		decoded := utils.DecodeIntMax64(encoded, uint32(tc.max))
		if decoded != tc.expected {
			t.Errorf("DecodeIntMax64(%d, %d) = %d; expected %d", tc.num, tc.max, decoded, tc.expected)
		}
	}
}

func BenchmarkEncodeIntMax32(b *testing.B) {
	num := uint32(12345)
	max := uint32(65535)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		utils.EncodeIntMax32(num, max)
	}
}

func BenchmarkDecodeIntMax32(b *testing.B) {
	num := uint32(12345)
	max := uint32(65535)
	encoded := utils.EncodeIntMax32(num, max)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		utils.DecodeIntMax32(encoded, max)
	}
}

func BenchmarkEncodeIntMax64(b *testing.B) {
	num := uint64(123456789012345)
	max := uint64(18446744073709551615)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		utils.EncodeIntMax64(num, max)
	}
}

func BenchmarkDecodeIntMax64(b *testing.B) {
	num := uint64(123456789012345)
	max := uint64(18446744073709551615)
	encoded := utils.EncodeIntMax64(num, max)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		utils.DecodeIntMax64(encoded, uint32(max))
	}
}
