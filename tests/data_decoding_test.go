package tests

import (
	"reflect"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/decoding"
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

func TestDecodeDataSendEntry[DynamicToken string, ValueType constraints.Unsigned](t *testing.T) {
	type args struct {
		bytes *utils.GrowingBytes
		opts  decoding.Opts[DynamicToken, ValueType]
	}
	tests := []struct {
		name string
		args args
		want wgpstypes.MsgDataSendEntry[DynamicToken]
	}{
		{
			name: "Valid input case 1",
			args: args{
				bytes: utils.NewGrowingBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05}),
				opts: decoding.Opts[DynamicToken, ValueType]{
					DecodeDynamicToken: func(bytes *utils.GrowingBytes) DynamicToken {
						return "dynamicToken1"
					},
					DecodeNamespaceId: func(bytes *utils.GrowingBytes) chan types.NamespaceId {
						ch := make(chan types.NamespaceId, 1)
						ch <- types.NamespaceId{1}
						close(ch)
						return ch
					},
					DecodeSubspaceId: func(bytes *utils.GrowingBytes) chan types.SubspaceId {
						ch := make(chan types.SubspaceId, 1)
						ch <- types.SubspaceId{1}
						close(ch)
						return ch
					},
					DecodePayloadDigest: func(bytes *utils.GrowingBytes) chan types.PayloadDigest {
						ch := make(chan types.PayloadDigest, 1)
						ch <- types.PayloadDigest{Digest: []byte{0x01, 0x02}}
						close(ch)
						return ch
					},
					PathScheme: types.PathParams[ValueType]{},
					CurrentlyReceivedEntry: types.Entry{
						Payload_length: 10,
					},
					AoiHandlesToArea: func(senderHandle, receiverHandle uint64) types.Area {
						return types.Area{1}
					},
					AoiHandlesToNamespace: func(senderHandle, receiverHandle uint64) types.Namespace {
						return types.NamespaceId{1}
					},
				},
			},
			want: wgpstypes.MsgDataSendEntry[DynamicToken]{
				Kind: wgpstypes.DataSendEntry,
				Data: wgpstypes.MsgDataSendEntryData[DynamicToken]{
					Entry: types.Entry{
						Payload_length: 10,
					},
					StaticTokenHandle: 1,
					DynamicToken:      "dynamicToken1",
					Offset:            10,
				},
			},
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a channel and send the bytes into it
			byteChan := make(chan *utils.GrowingBytes, 1)
			byteChan <- tt.args.bytes
			close(byteChan)

			// Read from the channel and pass to DecodeDataSendEntry
			bytes := <-byteChan
			if got := decoding.DecodeDataSendEntry(bytes, tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeDataSendEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
