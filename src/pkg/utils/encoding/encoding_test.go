package utils

import (
	"reflect"
	"testing"
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
		result := BigintToBytes(tc.input)
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
		{65536, 16777215, 65536},
		{16777216, 4294967295, 16777216},
	}

	for _, tc := range testCases {
		encoded := EncodeIntMax32(tc.num, tc.max)
		decoded, err := DecodeIntMax32(encoded, tc.max)
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
		encoded := EncodeIntMax64(tc.num, tc.max)
		decoded := DecodeIntMax64(encoded, tc.max)
		if decoded != tc.expected {
			t.Errorf("DecodeIntMax64(%d, %d) = %d; expected %d", tc.num, tc.max, decoded, tc.expected)
		}
	}
}
