package utils

import (
	"github.com/PES-Innovation-Lab/willow-go/types"
)

func EncryptPath[T EncryptionKeyType](Key T, EncryptFn types.EncryptFn[T])
