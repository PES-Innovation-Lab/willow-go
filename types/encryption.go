package types

type EncryptionKeyType interface{} //TBD based on encryption algorithm???

type EncryptFn[T EncryptionKeyType] func(Key T, bytes []byte) []byte

type DeriveKey[T EncryptionKeyType] func(Key T, component []byte) T
