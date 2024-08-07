package types

import "golang.org/x/exp/constraints"

type PathParams[ValueType constraints.Unsigned] struct {
	/*
	  Setting up Path parameters, we are setting each type to be unisgned as these parameters cannot be negative
	  we are also not setting it to signed and not a fixed uint32 so that if the user does not want the params to be that long
	  we can save some space by using uint8 if required.
	*/
	MaxComponentCount  ValueType
	MaxComponentLength ValueType
	MaxPathLength      ValueType
}

type SignatureScheme[PublicKey, SecretKey, Signature any] struct {
	Sign   func(publicKey PublicKey, secretKey SecretKey, bytestring []byte) Signature
	Verify func(publicKey PublicKey, signature Signature, bytestring []byte) bool
}
