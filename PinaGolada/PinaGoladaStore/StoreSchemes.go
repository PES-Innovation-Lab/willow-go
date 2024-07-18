package PinaGoladaStore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

// Test schemes defined!
var NameSpaceEncoding utils.EncodingScheme[types.NamespaceId] = utils.EncodingScheme[types.NamespaceId]{
	Encode: func(id types.NamespaceId) []byte {
		return []byte(id)
	},
	Decode: func(bytes []byte) (types.NamespaceId, error) {
		return types.NamespaceId(bytes), nil
	},
	EncodedLength: func(id types.NamespaceId) uint64 {
		return uint64(len(id))
	},
	DecodeStream: func(value *utils.GrowingBytes) (types.NamespaceId, error) {
		bytes := value.NextAbsolute(1)
		return types.NamespaceId(bytes), nil
	},
}
var TestNameSpaceScheme datamodeltypes.NamespaceScheme = datamodeltypes.NamespaceScheme{
	EncodingScheme: NameSpaceEncoding,
	IsEqual: func(a types.NamespaceId, b types.NamespaceId) bool {
		return utils.OrderBytes(a, b) == 0
	},
	DefaultNamespaceId: types.NamespaceId(""),
}

var SubspaceEncoding utils.EncodingScheme[types.SubspaceId] = utils.EncodingScheme[types.SubspaceId]{
	Encode: func(id types.SubspaceId) []byte {
		return []byte(id)
	},
	Decode: func(bytes []byte) (types.SubspaceId, error) {
		return types.SubspaceId(bytes), nil
	},
	EncodedLength: func(id types.SubspaceId) uint64 {
		return uint64(len(id))
	},
	DecodeStream: func(value *utils.GrowingBytes) (types.SubspaceId, error) {
		bytes := value.NextAbsolute(1)
		return types.SubspaceId(bytes), nil
	},
}

var TestSubspaceScheme datamodeltypes.SubspaceScheme = datamodeltypes.SubspaceScheme{
	EncodingScheme:      SubspaceEncoding,
	SuccessorSubspaceFn: utils.SuccessorSubspaceId,
	Order:               utils.OrderSubspace,
	MinimalSubspaceId:   types.SubspaceId(""),
}
var TestFingerprintScheme datamodeltypes.FingerprintScheme[uint64, uint64] = datamodeltypes.FingerprintScheme[uint64, uint64]{} //Dummy scheme

var TestAuthorisationScheme datamodeltypes.AuthorisationScheme[[]byte, string] = datamodeltypes.AuthorisationScheme[[]byte, string]{
	Authorise: func(entry types.Entry, opts []byte) (string, error) {
		if strings.Compare(string(entry.Subspace_id), string(opts)) == 0 {
			return string(entry.Subspace_id), nil
		}

		return string(""), fmt.Errorf("user not authorised")
	},
	IsAuthoriseWrite: func(entry types.Entry, token string) bool {
		return utils.OrderBytes(entry.Subspace_id, types.SubspaceId(token)) == 0
	},
	TokenEncoding: utils.EncodingScheme[string]{
		Encode: func(id string) []byte {
			return []byte(id)
		},
		Decode: func(bytes []byte) (string, error) {
			return string(bytes), nil
		},
		EncodedLength: func(id string) uint64 {
			return uint64(len(id))
		},
		DecodeStream: func(value *utils.GrowingBytes) (string, error) {
			bytes := value.NextAbsolute(1)
			return string(bytes), nil
		},
	},
}

var TestPathParams types.PathParams[uint8] = types.PathParams[uint8]{
	MaxComponentCount:  50,
	MaxComponentLength: 50,
	MaxPathLength:      50,
}

var TestPayloadScheme datamodeltypes.PayloadScheme = datamodeltypes.PayloadScheme{
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
var StoreSchemes datamodeltypes.StoreSchemes[uint64, uint64, uint8, []byte, string] = datamodeltypes.StoreSchemes[uint64, uint64, uint8, []byte, string]{
	PathParams:          TestPathParams,
	NamespaceScheme:     TestNameSpaceScheme,
	AuthorisationScheme: TestAuthorisationScheme,
	SubspaceScheme:      TestSubspaceScheme,
	PayloadScheme:       TestPayloadScheme,
}
