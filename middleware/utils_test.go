package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poteto0/poteto"
	"github.com/poteto0/poteto/utils"
)

func TestWrapRegExp(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{"test * url", "https://example.com:*", `^https://example\.com:.*$`},
		{"test ? url", "https://example.com:300?", `^https://example\.com:300.$`},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := wrapRegExp(it.target)
			if result != it.expected {
				t.Errorf("Not matched")
			}
		})
	}
}

func TestMatchSubDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		pattern  string
		expected bool
	}{
		{"test same url", "http://hello.world.com.test", "http://hello.world.com.test", false},
		{"test http & https return false", "http://hello.world.com.test", "https://hello.world.com.*", false},
		{"test not :// type return false", "hello.world.com.test", "hello.world.com.test", false},
		{"test wild card pattern return true", "http://hello.world.com.test", "http://hello.world.com.*", true},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := matchSubdomain(it.domain, it.pattern)
			if result != it.expected {
				t.Errorf("Not matched")
				t.Errorf(fmt.Sprintf("expected: %t", it.expected))
				t.Errorf(fmt.Sprintf("actual: %t", result))
			}
		})
	}
}

func TestMatchScheme(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		pattern  string
		expected bool
	}{
		{"If : is not existed return false(both)", "example.com", "example.com", false},
		{"If : is not existed return false(pattern)", "example.com", "http://example.com", false},
		{"If : is not existed return false(domain)", "http://example.com", "example.com", false},
		{"matched", "http://example1.com", "http://example2.com", true},
		{"not matched", "http://example1.com", "https://example2.com", false},
	}

	for _, it := range tests {
		t.Run((it.name), func(t *testing.T) {
			result := matchScheme(it.domain, it.pattern)

			if result != it.expected {
				t.Errorf("Not matched")
				t.Errorf(fmt.Sprintf("expected: %t", it.expected))
				t.Errorf(fmt.Sprintf("actual: %t", result))
			}
		})
	}
}

func TestReverseStringArray(t *testing.T) {
	targets := []string{"!!", "world", "hello"}
	expected := []string{"hello", "world", "!!"}

	result := reverseStringArray(targets)
	if !utils.SliceEqual(result, expected) {
		t.Errorf("Not matched")
	}
}

func TestMatchMethod(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		allowMethods []string
		expected     bool
	}{
		{"test including method return true", http.MethodGet, []string{http.MethodGet}, true},
		{"test not including method return false", http.MethodPost, []string{http.MethodGet}, false},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := matchMethod(it.target, it.allowMethods)
			if result != it.expected {
				t.Errorf("Not matched")
				t.Errorf(fmt.Sprintf("expected: %t", it.expected))
				t.Errorf(fmt.Sprintf("actual: %t", result))
			}
		})
	}
}

func TestExtractFromHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("hello", "world")
	ctx := poteto.NewContext(w, req)

	result := ExtractFromHeader(ctx, "hello")
	if result != "world" {
		t.Errorf("unmatched")
	}
}
