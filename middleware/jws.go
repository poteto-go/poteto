package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
)

type PotetoJWSConfig struct {
	AuthScheme string
	SignMethod string
	SignKey    any
	ContextKey string
	ClaimsFunc func(c poteto.Context) jwt.Claims
}

type IPotetoJWSConfig interface {
	KeyFunc(token *jwt.Token) (any, error)
	ParseToken(ctx poteto.Context, auth string) (any, error)
}

var DefaultJWSConfig = &PotetoJWSConfig{
	AuthScheme: constant.AuthScheme,
	SignMethod: constant.AlgorithmHS256,
	ContextKey: "user",
	ClaimsFunc: func(c poteto.Context) jwt.Claims {
		return jwt.MapClaims{}
	},
}

func (cfg *PotetoJWSConfig) KeyFunc(token *jwt.Token) (any, error) {
	if token.Method.Alg() != cfg.SignMethod {
		return nil, errors.New("unexpected jwt signing method: " + cfg.SignMethod)
	}

	if cfg.SignKey == nil {
		return nil, errors.New("undefined sign key")
	}

	return cfg.SignKey, nil
}

func (cfg *PotetoJWSConfig) ParseToken(ctx poteto.Context, auth string) (any, error) {
	token, err := jwt.ParseWithClaims(auth, cfg.ClaimsFunc(ctx), cfg.KeyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

func JWSWithConfig(cfg IPotetoJWSConfig) poteto.MiddlewareFunc {
	config := cfg.(*PotetoJWSConfig)
	if config.SignKey == nil {
		panic(config.SignKey)
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			authValue, err := extractBearer(ctx)
			if err != nil {
				return poteto.NewHttpError(http.StatusBadRequest, err)
			}

			token, err := cfg.ParseToken(ctx, authValue)
			if err != nil {
				return poteto.NewHttpError(http.StatusUnauthorized, err)
			}

			ctx.Set(config.ContextKey, token)
			return next(ctx)
		}
	}
}

func extractBearer(ctx poteto.Context) (string, error) {
	authHeader := ctx.GetRequest().Header.Get(constant.HeaderAuthorization)
	target := constant.AuthScheme
	bearers := strings.Split(authHeader, target)
	if len(bearers) <= 1 {
		return "", errors.New("not included bearer token")
	}
	return strings.Trim(bearers[1], " "), nil
}
