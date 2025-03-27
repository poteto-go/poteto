package middleware

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/oidc"
)

type OidcConfig struct {
	// google
	Idp        string `yaml:"idp"`
	ContextKey string `yaml:"context_key"`
}

var DefaultOidcConfig = &OidcConfig{
	Idp:        "google",
	ContextKey: "googleToken",
}

// Oidc support google.com
//
// case google: => oidc.GoogleOidcClaims
//
//	func main() {
//	  p := poteto.New()
//	  p.Register(
//	    middleware.OidcWithConfig(
//	      middleware.DefaultOidcConfig,
//	    )
//	  )
//	  p.POST("/login", func(ctx poteto.Context) error {
//	      token, _ := ctx.Get("googleToken")
//	      claims := token.(oidc.GoogleOidcClaims)
//	   })
//	}
//
// case other: => return []byte
func OidcWithConfig(cfg OidcConfig) poteto.MiddlewareFunc {
	if cfg.ContextKey == "" {
		cfg.ContextKey = DefaultOidcConfig.ContextKey
	}

	if cfg.Idp == "" {
		cfg.Idp = DefaultOidcConfig.Idp
	}

	var claims any
	switch cfg.Idp {
	case "google":
		claims = oidc.GoogleOidcClaims{}
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			authValue, err := extractBearer(ctx)
			if err != nil {
				return err
			}

			token, err := decode(authValue)
			if err != nil {
				return err
			}

			// unmarshal
			switch cfg.Idp {
			case "google":
				json.Unmarshal(token, &claims)
				ctx.Set(cfg.ContextKey, claims)
			default:
				ctx.Set(cfg.ContextKey, token)
				return next(ctx)
			}

			return next(ctx)
		}
	}
}

func decode(token string) ([]byte, error) {
	splitToken := strings.Split(token, ".")
	if len(splitToken) != 3 {
		return []byte(""), errors.New("invalid token")
	}
	payload := splitToken[1]

	// base64 needs 4* length
	paddingLength := ((4 - len(payload)%4) % 4)
	padding := strings.Repeat("=", paddingLength)
	paddedPayload := strings.Join([]string{payload, padding}, "")

	decodedPayload, err := base64.StdEncoding.DecodeString(paddedPayload)
	if err != nil {
		return []byte(""), err
	}

	return decodedPayload, nil
}
