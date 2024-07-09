package payloadDriver

type PayloadDriver struct {
	path          string
	PayloadScheme PayloadScheme[PayloadDigest]
}
