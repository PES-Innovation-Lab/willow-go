package payloadDriver

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

var MockPayloadScheme datamodeltypes.PayloadScheme = datamodeltypes.PayloadScheme{
	EncodingScheme: utils.EncodingScheme[types.PayloadDigest]{
		Encode: func(value types.PayloadDigest) []byte {
			decoded, err := hex.DecodeString(string(value))
			if err != nil {
				return []byte{}
			}
			return decoded
		},
	},
	FromBytes: func(bytes []byte) chan types.PayloadDigest {
		ch := make(chan types.PayloadDigest, 1)
		go func() {
			var hash = sha256.Sum256(bytes)
			ch <- types.PayloadDigest(hex.EncodeToString(hash[:]))
			close(ch)
		}()
		return ch
	},
}
