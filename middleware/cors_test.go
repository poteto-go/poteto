package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
)

type TestVal struct {
	Name string `json:"name"`
	Val  string `json:"val"`
}

func TestCORSWithConfigByDefault(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		url      string
		origin   string
		aOrigins []string
		aMethods []string
		expected string
	}{
		{
			"Test Default allow all",
			"GET",
			"https://example.com/test",
			"https://example.com",
			[]string{},
			[]string{},
			"https://example.com",
		},
		{
			"Test Allowed origin",
			"GET",
			"https://example.com/test",
			"https://example.com",
			[]string{"https://example.com"},
			[]string{},
			"https://example.com",
		},
		{
			"Test Not Allowed origin",
			"GET",
			"https://example.com/test",
			"https://example.com",
			[]string{"https://unexpected.com"},
			[]string{},
			"",
		},
		{
			"Test Allowed origin",
			"POST",
			"https://example.com/test",
			"https://example.com",
			[]string{"https://example.com"},
			[]string{"GET"},
			"",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			config := CORSConfig{
				AllowOrigins: it.aOrigins,
				AllowMethods: it.aMethods,
			}
			cors := CORSWithConfig(config)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(it.method, it.url, nil)
			req.Header.Set(constant.HeaderOrigin, it.origin)
			context := poteto.NewContext(w, req)

			handler := func(ctx poteto.Context) error {
				return ctx.JSON(http.StatusOK, TestVal{Name: "test", Val: "val"})
			}

			cors_handler := cors(handler)
			cors_handler(context)
			result := w.Header().Get(constant.HeaderAccessControlOrigin)
			if result != it.expected {
				t.Errorf("Unmatched")
				t.Errorf(fmt.Sprintf("result: %s", result))
				t.Errorf(fmt.Sprintf("expected: %s", it.expected))
			}
		})
	}
}

func TestGetAllowSubDomain(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		allowOrigins []string
		expected     string
	}{
		{"test wildcard return true", "https://example.com", []string{"*"}, "https://example.com"},
		{"test match same domain", "https://example.com", []string{"https://example.com"}, "https://example.com"},
		{"test matched subdomain", "https://exmaple.com.test", []string{"https://example.com.*"}, "https://exmaple.com.test"},
		{"test not matched", "https://hello.world.com", []string{"https://exmaple.com"}, ""},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := getAllowSubDomain(it.origin, it.allowOrigins)
			if result != it.expected {
				t.Errorf("Not matched")
				t.Errorf(fmt.Sprintf("expected: %s", it.expected))
				t.Errorf(fmt.Sprintf("actual: %s", result))
			}
		})
	}
}

func TestGetAllowOrigin(t *testing.T) {
	tests := []struct {
		name                string
		origin              string
		allowOriginPatterns []string
		expected            string
	}{
		{"test match case", "https://example.com", []string{wrapRegExp("https://example.*")}, "https://example.com"},
		{"test not match case", "https://example.com", []string{wrapRegExp("https://hello.world.com")}, ""},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := getAllowOrigin(it.origin, it.allowOriginPatterns)
			if result != it.expected {
				t.Errorf("Not matched")
				t.Errorf(fmt.Sprintf("expected: %s", it.expected))
				t.Errorf(fmt.Sprintf("actual: %s", result))
			}
		})
	}
}
