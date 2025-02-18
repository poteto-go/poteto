package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang-jwt/jwt/v5"
	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
	"github.com/poteto-go/poteto/utils"
)

type jwtUserClaims struct {
	Name string `json:"name"`
	jwt.RegisteredClaims
}

type user struct {
	name string `json:"name"`
}

func generateUserClaims(user user, time_duration time.Duration) *jwtUserClaims {
	return &jwtUserClaims{
		user.name,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time_duration)),
		},
	}
}

func TestKeyFunc(t *testing.T) {
	claims := generateUserClaims(user{name: "hello"}, time.Hour*14*24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tests := []struct {
		name          string
		config        potetoJWSConfig
		expectedErr   bool
		expectSignKey []byte
	}{
		{
			"Test valid token",
			potetoJWSConfig{
				SignMethod: constant.AlgorithmHS256,
				ContextKey: "user",
				SignKey:    []byte("secret"),
			},
			false,
			[]byte("secret"),
		},
		{
			"Test not equal sign method",
			potetoJWSConfig{
				SignMethod: "SHA256",
				ContextKey: "user",
				SignKey:    []byte("secret"),
			},
			true,
			[]byte("secret"),
		},
		{
			"Test nil sign key throw error",
			potetoJWSConfig{
				SignMethod: constant.AlgorithmHS256,
				ContextKey: "user",
				SignKey:    nil,
			},
			true,
			[]byte("secret"),
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			val, err := it.config.KeyFunc(token)
			if it.expectedErr {
				if err == nil {
					t.Errorf("Unmatched")
				}
			} else {
				if err != nil {
					t.Errorf("Unmatched: %v", err)
				}

				switch asserted := any(val).(type) {
				case []byte:
					if !utils.IsSliceEqual[byte](asserted, it.expectSignKey) {
						t.Errorf(
							"Unmatched actual(%v) -> expected(%v)",
							val,
							it.expectSignKey,
						)
					}
				default:
					t.Errorf("Unmatched type")
				}
			}
		})
	}
}

func TestParseJWSToken(t *testing.T) {
	defer monkey.UnpatchAll()

	config := NewPotetoJWSConfig("user", []byte("secret"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := poteto.NewContext(w, req)

	claims := generateUserClaims(user{name: "hello"}, time.Hour*14*24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tk, _ := token.SignedString([]byte("secret"))

	tests := []struct {
		name        string
		mock        bool
		mockParse   func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error)
		expectedErr bool
	}{
		{
			"Test valid case",
			false,
			nil,
			false,
		},
		{
			"Test jwt Parse occur error",
			true,
			func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				return nil, errors.New("error")
			},
			true,
		},
		{
			"Test invalid token",
			true,
			func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
				token := &jwt.Token{}
				token.Valid = false
				return token, nil
			},
			true,
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			if it.mock {
				monkey.Patch(jwt.ParseWithClaims, it.mockParse)
			}

			_, err := config.ParseToken(ctx, tk)
			if it.expectedErr {
				if err == nil {
					t.Errorf("Not throw error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected err: %v", err)
				}
			}
		})
	}
}

func TestJWSMiddleware(t *testing.T) {
	jwsConfig := NewPotetoJWSConfig(
		"user", []byte("secret"),
	)

	p := poteto.New()
	p.Register(JWSWithConfig(jwsConfig))

	p.GET("/users", func(ctx poteto.Context) error {
		return ctx.JSON(200, "hello")
	})

	u := user{
		name: "hello",
	}

	// create jws token
	claims := generateUserClaims(u, time.Hour*14*24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tk, _ := token.SignedString([]byte("secret"))

	tests := []struct {
		name         string
		withToken    bool
		isReplace    bool
		expectedCode int
	}{
		{"Test valid token", true, false, http.StatusOK},
		{"Test not included token", false, false, http.StatusBadRequest},
		{"Test invalid token", true, true, http.StatusUnauthorized},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/users", nil)

			if it.isReplace {
				tk = "invalid token"
			}

			if it.withToken {
				bearer := "Bearer " + tk
				req.Header.Set("Authorization", bearer)
			}

			p.ServeHTTP(w, req)
			if w.Code != it.expectedCode {
				t.Errorf("Unmatched")
			}
		})
	}
}
