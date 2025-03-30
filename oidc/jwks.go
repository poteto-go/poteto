package oidc

type jwk struct {
	E   string `json:"e"`
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	Alg string `json:"alg"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

var JWKsUrls = map[string]string{
	"google": "https://www.googleapis.com/oauth2/v1/certs",
}
