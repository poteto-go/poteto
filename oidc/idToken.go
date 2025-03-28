package oidc

type Header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type IdToken struct {
	RawToken     string
	RawHeader    string
	RawPayload   string
	RawSignature string
	Header       Header
}
