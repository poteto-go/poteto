package middleware

import (
	"net/http"
	"regexp"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/tslice"
)

type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
	AllowMethods []string `yaml:"allow_methods"`
}

var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
}

func CORSWithConfig(config CORSConfig) poteto.MiddlewareFunc {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}

	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}

	allowOriginPatterns := make([]string, len(config.AllowOrigins))
	for i, origin := range config.AllowOrigins {
		pattern := wrapRegExp(origin)
		allowOriginPatterns[i] = pattern
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			req := ctx.GetRequest()
			res := ctx.GetResponse()
			origin := req.Header.Get(constant.HEADER_ORIGIN)

			res.AddHeader(constant.HEADER_VARY, constant.HEADER_ORIGIN)
			preflight := req.Method == http.MethodOptions

			// Not From Browser
			if origin == "" {
				if !preflight {
					return next(ctx)
				}
				return ctx.NoContent()
			}

			allowSubDomain := getAllowSubDomain(origin, config.AllowOrigins)
			// allowed origin path
			allowOrigin := getAllowOrigin(allowSubDomain, allowOriginPatterns)

			// Origin not allowed
			if allowOrigin == "" {
				if !preflight {
					return next(ctx)
				}
				return ctx.NoContent()
			}

			// allowed method
			if tslice.IndexOf(config.AllowMethods, req.Method) >= 0 {
				res.SetHeader(constant.HEADER_ACCESS_CONTROL_ORIGIN, allowOrigin)
				return next(ctx)
			}

			// just pass not allow origin header
			return next(ctx)
		}
	}
}

func getAllowSubDomain(origin string, allowOrigins []string) string {
	for _, o := range allowOrigins {
		if o == "*" || o == origin {
			return origin
		}
		if matchSubdomain(origin, o) {
			return origin
		}
	}

	return ""
}

func getAllowOrigin(origin string, allowOriginPatterns []string) string {
	for _, pattern := range allowOriginPatterns {
		if match, _ := regexp.MatchString(pattern, origin); match {
			return origin
		}
	}
	return ""
}
