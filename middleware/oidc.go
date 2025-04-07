package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/oidc"
	"github.com/poteto-go/poteto/utils"
)

type OidcConfig struct {
	Idp               string        `yaml:"idp"`
	ContextKey        string        `yaml:"context_key"`
	JwksUrl           string        `yaml:"jwks_url"`
	DefaultExpiration time.Duration `yaml:"default_expiration"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`
	CacheMode         bool          `yaml:"cache_mode"`
	Cache             *cache.Cache  `yaml:"-"`
	// you can set custom verify signature callback
	CustomVerifyTokenSignature func(idToken oidc.IdToken, jwksUrl string) error                      `yaml:"-"`
	CachedVerifyTokenSignature func(idToken oidc.IdToken, pCache *cache.Cache, jwksUrl string) error `yaml:"-"`
}

var OidcWithoutVerifyConfig = OidcConfig{
	Idp:               "google",
	ContextKey:        "googleToken",
	JwksUrl:           "",
	DefaultExpiration: cache.DefaultExpiration,
	CleanupInterval:   (24 * time.Hour),
	// default [true] at v2
	CacheMode:                  false,
	Cache:                      nil,
	CustomVerifyTokenSignature: nil,
}

var DefaultOidcConfig = OidcConfig{
	Idp:               "google",
	ContextKey:        "googleToken",
	JwksUrl:           "",
	DefaultExpiration: cache.DefaultExpiration,
	CleanupInterval:   (24 * time.Hour),
	// default [true] at v2
	CacheMode:                  true,
	Cache:                      nil,
	CustomVerifyTokenSignature: oidc.DefaultVerifyTokenSignature,
	CachedVerifyTokenSignature: oidc.CachedVerifyTokenSignature,
}

// Oidc verify signature by jwks url & set token -> context
//
// You can decode with oidc.GoogleOidcClaims
//
//	func main() {
//	  p := poteto.New()
//	  oidcConfig := middleware.OidcConfig {
//	    Idp: "google",
//	    ContextKey: "googleToken",
//	    CacheMode: true,
//	    JwksUrl: "https://www.googleapis.com/oauth2/v3/certs",
//	    CachedVerifyTokenSignature: oidc.CachedVerifyTokenSignature,
//	  }
//	  p.Register(
//	    middleware.OidcWithConfig(
//	      oidcConfig,
//	    )
//	  )
//	  p.POST("/login", func(ctx poteto.Context) error {
//	      var claims oidc.GoogleOidcClaims
//	      token, _ := ctx.Get("googleToken")
//	      json.Unmarshal(token.([]byte), &claims)
//	      ...
//	      return ctx.JSON(200, map[string]string{"message": "success"})
//	  })
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

	if cfg.DefaultExpiration == 0 {
		cfg.DefaultExpiration = DefaultOidcConfig.DefaultExpiration
	}

	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = DefaultOidcConfig.CleanupInterval
	}

	if cfg.CacheMode && cfg.Cache == nil {
		cfg.Cache = cache.New(cfg.DefaultExpiration, cfg.CleanupInterval)
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			authValue, err := extractBearer(ctx)
			if err != nil {
				return err
			}

			token, err := verifyDecode(authValue, cfg)
			if err != nil {
				return err
			}

			ctx.Set(cfg.ContextKey, token)
			return next(ctx)
		}
	}
}

func verifyDecode(token string, cfg OidcConfig) ([]byte, error) {
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

	// verify
	if err := applyVerifyFunc(idToken, cfg); err != nil {
		return []byte(""), err
	}

	// decode payload
	decodedPayload, err := utils.JwtDecodeSegment(idToken.RawPayload)
	if err != nil {
		return []byte(""), err
	}

	return decodedPayload, nil
}

func applyVerifyFunc(idToken oidc.IdToken, cfg OidcConfig) error {
	if cfg.CacheMode {
		if cfg.CachedVerifyTokenSignature != nil {
			return cfg.CachedVerifyTokenSignature(idToken, cfg.Cache, cfg.JwksUrl)
		}
		return nil
	}

	if cfg.CustomVerifyTokenSignature != nil {
		return cfg.CustomVerifyTokenSignature(idToken, cfg.JwksUrl)
	}

	return nil
}
