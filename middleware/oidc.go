package middleware

import (
	"encoding/base64"
	"errors"
	"strings"

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
	CustomVerifyTokenSignature: oidc.DefaultVerifyTokenSignature,
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
