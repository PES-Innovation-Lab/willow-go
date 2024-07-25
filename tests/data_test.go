package tests

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/decoding"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func TestDecodeDataReplyPayload(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected wgpstypes.MsgDataReplyPayload
	}{
		{
			name:  "Valid input",
			input: []byte{0x02, 0x01, 0x02},
			expected: wgpstypes.MsgDataReplyPayload{
				Kind: wgpstypes.DataReplyPayload,
				Data: wgpstypes.MsgDataReplyPayloadData{
					Handle: 258, // Example handle value
				},
			},
		},
		{
			name:  "Another valid input",
			input: []byte{0x01, 0x01},
			expected: wgpstypes.MsgDataReplyPayload{
				Kind: wgpstypes.DataReplyPayload,
				Data: wgpstypes.MsgDataReplyPayloadData{
					Handle: 1, // Example handle value
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a channel of type []byte
			ch := make(chan []byte, 1)

			// Send the byte slice to the channel
			ch <- tt.input

			// Receive the byte slice from the channel
			bytes := utils.NewGrowingBytes(ch)

			// Call the function to test
			result := decoding.DecodeDataReplyPayload(bytes)

			// Check if the result matches the expected output
			if result != tt.expected {
				t.Errorf("DecodeDataReplyPayload() = %v, want %v", result, tt.expected)
			}
		})
	}
}
