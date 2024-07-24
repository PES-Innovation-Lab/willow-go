package decoding

import (
	"github.com/PES-Innovation-Lab/willow-go/pkg/wgps/wgpstypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

func DecodeCommitmentReveal(bytes *utils.GrowingBytes, challengeLength int) wgpstypes.MsgCommitmentReveal {
	CommitmentBytes := bytes.NextAbsolute(1 + challengeLength)

	bytes.Prune(1 + challengeLength)

	return wgpstypes.MsgCommitmentReveal{
		Kind: wgpstypes.CommitmentReveal,
		Data: wgpstypes.MsgCommitmentRevealData{
			Nonce: CommitmentBytes[1 : 1+challengeLength],
		},
	}

}
