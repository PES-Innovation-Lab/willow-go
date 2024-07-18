package encoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"golang.org/x/exp/constraints"
)

func EncodeDataSendEntry[NamespaceId, SubspaceId, PayloadDigest constraints.Ordered, DynamicToken interface{}](
	msg wgpstypes.MsgDataSendEntry[NamespaceId, SubspaceId, PayloadDigest, DynamicToken],
opts struct{
	EncodeDynamicToken func(token DynamicToken) []byte
	CurrentlySentEntry types.Entry[]
}
)
