package payloadDriver

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

var MockPayloadScheme datamodeltypes.PayloadScheme[string, uint64] = datamodeltypes.PayloadScheme[string, uint64]{
	EncodingScheme: utils.EncodingScheme[string, uint64]{
		Encode: func(value string) []byte {
			decoded, err := hex.DecodeString(value)
			if err != nil {
				return []byte{}
			}
			return decoded
		},
	},
	FromBytes: func(bytes []byte) chan string {
		ch := make(chan string, 1)
		go func() {
			hash := sha256.Sum256(bytes)
			ch <- hex.EncodeToString(hash[:])
			close(ch)
		}()
		return ch
	},
}
