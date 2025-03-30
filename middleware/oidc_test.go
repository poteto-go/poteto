package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/middleware"
	"github.com/poteto-go/poteto/oidc"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_OidcWithConfig(t *testing.T) {
	oidcMiddleware := middleware.OidcWithConfig(
		middleware.OidcWithoutVerifyConfig,
	)

	t.Run("valid token", func(t *testing.T) {
		tokenBytes, _ := os.ReadFile("../_fixture/token/jwt.txt")
		token := string(tokenBytes)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := poteto.NewContext(w, req)

		var claims oidc.GoogleOidcClaims
		handler := func(ctx poteto.Context) error {
			token, _ := ctx.Get("googleToken")
			json.Unmarshal(token.([]byte), &claims)

			return ctx.JSON(http.StatusOK, claims)
		}
		oidc_handler := oidcMiddleware(handler)

		oidc_handler(ctx)

		assert.Equal(t, claims.Iss, "https://accounts.google.com")

		assert.Equal(t, claims.Email, "test@exmaple.com")
	})

	t.Run("invalid token", func(t *testing.T) {
		token := "invalid"

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := poteto.NewContext(w, req)

		var claims oidc.GoogleOidcClaims
		handler := func(ctx poteto.Context) error {
			token, _ := ctx.Get("googleToken")
			json.Unmarshal(token.([]byte), &claims)

			return ctx.JSON(http.StatusOK, claims)
		}
		oidc_handler := oidcMiddleware(handler)

		err := oidc_handler(ctx)

		assert.NotNil(t, err)
	})
}
