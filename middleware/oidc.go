package middleware

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/oidc"
)

type OidcConfig struct {
	Idp        string `yaml:"idp"`
	ContextKey string `yaml:"context_key"`
	JwksUrl    string `yaml:"jwks_url"`
	// you can set custom verify signature callback
	CustomVerifyTokenSignature func(idToken oidc.IdToken, jwksUrl string) error `yaml:"-"`
}

var OidcWithoutVerifyConfig = OidcConfig{
	Idp:                        "google",
	ContextKey:                 "googleToken",
	JwksUrl:                    "",
	CustomVerifyTokenSignature: nil,
}

var DefaultOidcConfig = OidcConfig{
	Idp:                        "google",
	ContextKey:                 "googleToken",
	JwksUrl:                    "",
	CustomVerifyTokenSignature: DefaultVerifyTokenSignature,
}

// Oidc set token -> context
//
// You can decode with oidc.GoogleOidcClaims
//
//	func main() {
//	  p := poteto.New()
//	  p.Register(
//	    middleware.OidcWithConfig(
//	      middleware.DefaultOidcConfig,
//	    )
//	  )
//	  p.POST("/login", func(ctx poteto.Context) error {
//	      var claims oidc.GoogleOidcClaims
//	      token, _ := ctx.Get("googleToken")
//	      json.Unmarshal(token.([]byte), &claims)
//	   })
//	}
func OidcWithConfig(cfg OidcConfig) poteto.MiddlewareFunc {
	if cfg.ContextKey == "" {
		cfg.ContextKey = DefaultOidcConfig.ContextKey
	}

	if cfg.Idp == "" {
		cfg.Idp = DefaultOidcConfig.Idp
	}

	if cfg.JwksUrl == "" {
		cfg.JwksUrl = oidc.JWKsUrls[cfg.Idp]
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			authValue, err := extractBearer(ctx)
			if err != nil {
				return err
			}

			token, err := verifyDecode(authValue, cfg.JwksUrl, cfg.CustomVerifyTokenSignature)
			if err != nil {
				return err
			}

			ctx.Set(cfg.ContextKey, token)
			return next(ctx)
		}
	}
}

func verifyDecode(token, jwksUrl string, customVerifyTokenSignature func(oidc.IdToken, string) error) ([]byte, error) {
	splitToken := strings.Split(token, ".")
	if len(splitToken) != 3 {
		return []byte(""), errors.New("invalid token")
	}

	idToken := oidc.IdToken{
		RawToken:     token,
		RawHeader:    splitToken[0],
		RawPayload:   splitToken[1],
		RawSignature: splitToken[2],
	}

	if customVerifyTokenSignature != nil {
		err := customVerifyTokenSignature(idToken, jwksUrl)
		if err != nil {
			return []byte(""), err
		}
	}

	// decode payload
	decodedPayload, err := jwtDecodeSegment(idToken.RawPayload)
	if err != nil {
		return []byte(""), err
	}

	return decodedPayload, nil
}

func jwtDecodeSegment(raw string) ([]byte, error) {
	paddingLength := ((4 - len(raw)%4) % 4)
	padding := strings.Repeat("=", paddingLength)
	padded := strings.Join([]string{raw, padding}, "")

	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		return []byte(""), err
	}

	return decoded, nil
}

func DefaultVerifyTokenSignature(idToken oidc.IdToken, jwksUrl string) error {
	// decode header
	byteHeader, err := jwtDecodeSegment(idToken.RawHeader)
	if err != nil {
		return err
	}

	header := oidc.Header{}
	if err := json.Unmarshal(byteHeader, &header); err != nil {
		return err
	}
	idToken.Header = header

	// verify signature
	key, err := getJwk(idToken, jwksUrl)
	if err != nil {
		return err
	}

	byteN, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return err
	}

	const standardExponent = 65537
	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(byteN),
		E: standardExponent,
	}

	headerAndPayload := fmt.Sprintf("%s.%s", idToken.RawHeader, idToken.RawPayload)
	sha := sha256.New()
	sha.Write([]byte(headerAndPayload))

	decSignature, err := base64.RawURLEncoding.DecodeString(idToken.RawSignature)
	if err != nil {
		return err
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sha.Sum(nil), decSignature); err != nil {
		return err
	}

	return nil
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

func (keys jwks) find(kid string) (jwk, error) {
	var foundKey jwk
	for _, key := range keys.Keys {
		if key.Kid == kid {
			foundKey = key

			break
		}
	}

	if foundKey != (jwk{}) {
		return foundKey, nil
	} else {
		return jwk{}, errors.New("jwks keys not found")
	}
}

type jwk struct {
	E   string `json:"e"`
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	Alg string `json:"alg"`
}

func getJwk(token oidc.IdToken, jwksUrl string) (jwk, error) {
	parsedUrl, err := url.Parse(jwksUrl)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to parse jwks url: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reqWithCtx, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, parsedUrl.String(), nil)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to create request of GET JWKs endpoint: %w", err)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(reqWithCtx)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to GET JWKs endpoint: %w", err)
	}

	defer resp.Body.Close()
	byteArray, _ := io.ReadAll(resp.Body)

	keys := &jwks{}
	if err := json.Unmarshal(byteArray, keys); err != nil {
		return jwk{}, fmt.Errorf("failed to unmarshal JWKs response: %w", err)
	}

	foundKey, err := keys.find(token.Header.Kid)
	if err != nil {
		return jwk{}, err
	}

	return foundKey, nil
}
