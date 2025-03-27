package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/middleware"
	"github.com/poteto-go/poteto/oidc"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_OidcWithConfig(t *testing.T) {
	oidcMiddleware := middleware.OidcWithConfig(
		middleware.DefaultOidcConfig,
	)

	t.Run("valid token", func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleUlkIiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiJleG1hcGxlLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwiYXVkIjoiZXhtYXBsZS5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsInN1YiI6IjEwMDAwMDAwMDAwMDAwMDAiLCJlbWFpbCI6InRlc3RAZXhtYXBsZS5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6Imhhc2giLCJuYW1lIjoiVGVzdCBVc2VyIiwicGljdHVyZSI6Imh0dHBzOi8vcGljdHVyZSIsImdpdmVuX25hbWUiOiJUZXN0IiwiZmFtaWx5X25hbWUiOiJVc2VyIiwiaWF0IjoxNzQzMDc2MTk4LCJleHAiOjE3NDMwNzk3OTh9.pPriv3JvTgtQectH3mfOcMSO6T2RmWOMjyCXl4qd5_2tZNLfUh1M4f2JAqyectfrS1c4515k92_kKRN5985GHS78etadEH-0lFW7T3ehfYD5fs0HVo0EwYlDTg_8tZZ4x8kWd_RPdg21BQiubnHWcIpFv-HyMI9mWrkCmw31bkMrS-5a5-CMfeZwh2kJaD8D1_84LLuQuzbzIEBXoDlWCwHVkMTx6MSXNJpSqakCcDpE_sZ1WMg4r2jNvsG3nT-jSKwTMZ4g960YfYW0BAQafUKKDOZM9JnQ0HrFHEnKAun83AHjjmGVu3bq7Jc4IqQ5sYVuZD7BeciaEvQKvsk3IA"

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
}
