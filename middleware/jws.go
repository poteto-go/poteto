package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
)

type potetoJWSConfig struct {
	AuthScheme string
	SignMethod string
	SignKey    any
	ContextKey string
	ClaimsFunc func(c poteto.Context) jwt.Claims
}

type PotetoJWSConfig interface {
	KeyFunc(token *jwt.Token) (any, error)
	ParseToken(ctx poteto.Context, auth string) (any, error)
}

var DefaultJWSConfig = &potetoJWSConfig{
	AuthScheme: constant.AuthScheme,
	SignMethod: constant.AlgorithmHS256,
	ContextKey: "user",
	ClaimsFunc: func(c poteto.Context) jwt.Claims {
		return jwt.MapClaims{}
	},
}

func NewPotetoJWSConfig(contextKey string, signKey any) PotetoJWSConfig {
	cfg := DefaultJWSConfig
	cfg.ContextKey = contextKey
	cfg.SignKey = signKey
	return cfg
}

func (cfg *potetoJWSConfig) KeyFunc(token *jwt.Token) (any, error) {
	if token.Method.Alg() != cfg.SignMethod {
		return nil, errors.New("unexpected jwt signing method: " + cfg.SignMethod)
	}

	if cfg.SignKey == nil {
		return nil, errors.New("undefined sign key")
	}

	return cfg.SignKey, nil
}

func (cfg *potetoJWSConfig) ParseToken(ctx poteto.Context, auth string) (any, error) {
	token, err := jwt.ParseWithClaims(auth, cfg.ClaimsFunc(ctx), cfg.KeyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

func JWSWithConfig(cfg PotetoJWSConfig) poteto.MiddlewareFunc {
	config := cfg.(*potetoJWSConfig)
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
