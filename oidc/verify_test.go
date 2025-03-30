package oidc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/poteto-go/poteto/utils"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a minimal valid JWT structure for testing
func createTestJWT(header Header, payload map[string]interface{}, key *rsa.PrivateKey) (IdToken, error) {
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return IdToken{}, fmt.Errorf("failed to marshal header: %w", err)
	}
	rawHeader := base64.RawURLEncoding.EncodeToString(headerBytes)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return IdToken{}, fmt.Errorf("failed to marshal payload: %w", err)
	}
	rawPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	headerAndPayload := fmt.Sprintf("%s.%s", rawHeader, rawPayload)
	sha := sha256.New()
	sha.Write([]byte(headerAndPayload))
	hashed := sha.Sum(nil)

	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed)
	if err != nil {
		return IdToken{}, fmt.Errorf("failed to sign token: %w", err)
	}
	rawSignature := strings.TrimRight(base64.RawURLEncoding.EncodeToString(signatureBytes), "=")

	// Create IdToken struct (Payload field is not strictly needed for signature verification test)
	token := IdToken{
		RawHeader:    rawHeader,
		RawPayload:   rawPayload,
		RawSignature: rawSignature,
		Header:       header, // Assign parsed header
		// Payload:      payload, // Not strictly needed for this test
	}
	return token, nil
}

// Helper function to create a JWK from an RSA public key
func createJWK(key *rsa.PublicKey, kid string) jwk {
	encodedN := base64.URLEncoding.EncodeToString(key.N.Bytes())
	return jwk{
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		Kid: kid,
		N:   strings.TrimRight(encodedN, "="),                                       // Remove padding
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()), // Standard exponent AQAB is 65537
	}
}

func TestDefaultVerifyTokenSignature(t *testing.T) {
	// Generate RSA key pair for signing and verification
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// --- Test Setup ---
	kid := "test-key-id-1"
	testHeader := Header{
		Alg: "RS256",
		Typ: "JWT",
		Kid: kid,
	}
	testPayload := map[string]interface{}{
		"iss": "https://test.issuer.com",
		"sub": "test-subject",
		"aud": "test-audience",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	// Create a valid token
	validToken, err := createTestJWT(testHeader, testPayload, privateKey)
	assert.NoError(t, err)

	// Create JWK corresponding to the public key
	correctJWK := createJWK(publicKey, kid)
	jwksResponse := jwks{Keys: []jwk{correctJWK}}
	jwksBytes, err := json.Marshal(jwksResponse)
	assert.NoError(t, err)

	// --- Mock JWKS Server ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jwksBytes)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	jwksUrl := server.URL + "/.well-known/jwks.json"

	// --- Test Cases ---
	tests := []struct {
		name        string
		token       IdToken
		jwksUrl     string
		setupServer func() *httptest.Server // Optional server override for specific tests
		expectError bool
		errorMsg    string // Optional: check for specific error message part
	}{
		{
			name:        "Valid Token and JWKS",
			token:       validToken,
			jwksUrl:     jwksUrl,
			expectError: false,
		},
		{
			name: "Invalid Signature",
			token: IdToken{
				RawHeader:    validToken.RawHeader,
				RawPayload:   validToken.RawPayload,
				RawSignature: strings.TrimRight(base64.RawURLEncoding.EncodeToString([]byte("invalid")), "="),
				Header:       validToken.Header,
			},
			jwksUrl:     jwksUrl,
			expectError: true,
			errorMsg:    "signature verification failed",
		},
		{
			name: "Malformed Header Segment",
			token: IdToken{
				RawHeader:    "not-base64",
				RawPayload:   validToken.RawPayload,
				RawSignature: "not-base64",
				// Header will be derived from RawHeader, causing unmarshal error later if decode succeeds
			},
			jwksUrl:     jwksUrl,
			expectError: true,
			errorMsg:    "illegal base64 data", // Error from utils.JwtDecodeSegment
		},
		{
			name: "Malformed Signature Segment",
			token: IdToken{
				RawHeader:    "not-base64",
				RawPayload:   validToken.RawPayload,
				RawSignature: "not-base64",
				Header:       validToken.Header,
			},
			jwksUrl:     jwksUrl,
			expectError: true,
			errorMsg:    "illegal base64 data", // Error from utils.JwtDecodeSegment
		},
		{
			name:  "JWKS Server Down",
			token: validToken,
			setupServer: func() *httptest.Server {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate server being down by closing immediately
				}))
				s.Close() // Close the server immediately
				return s
			},
			expectError: true,
			errorMsg:    "failed to GET JWKs endpoint",
		},
		{
			name:  "JWKS Server Returns 500",
			token: validToken,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectError: true,
			// Note: The error might be about unmarshalling empty body or similar,
			// depending on how the http client handles the 500 before reading body.
			// Let's check for the unmarshal error which is likely.
			errorMsg: "failed to unmarshal JWKs response",
		},
		{
			name:  "JWKS Server Returns Malformed JSON",
			token: validToken,
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("{not json}"))
				}))
			},
			expectError: true,
			errorMsg:    "failed to unmarshal JWKs response",
		},
		{
			name: "Key ID (kid) Not Found in JWKS",
			token: func() IdToken {
				// Create a token with a kid that won't be in the mock JWKS
				headerWithWrongKid := Header{Alg: "RS256", Typ: "JWT", Kid: "wrong-kid"}
				token, _ := createTestJWT(headerWithWrongKid, testPayload, privateKey)
				return token
			}(),
			jwksUrl:     jwksUrl,
			expectError: true,
			errorMsg:    "jwks keys not found",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			currentJwksUrl := it.jwksUrl
			var currentServer *httptest.Server
			if it.setupServer != nil {
				currentServer = it.setupServer()
				currentJwksUrl = currentServer.URL + "/.well-known/jwks.json" // Assume same path
				defer currentServer.Close()
			}

			// Need to decode header first for DefaultVerifyTokenSignature
			// This simulates what would happen before calling the function usually.
			// If RawHeader is invalid, JwtDecodeSegment inside will fail first.
			if it.token.RawHeader != "" && it.token.RawHeader != "not-base64" {
				byteHeader, err := utils.JwtDecodeSegment(it.token.RawHeader)
				if err == nil { // Only proceed if header decodes
					header := Header{}
					_ = json.Unmarshal(byteHeader, &header) // Ignore unmarshal error here, let the function handle it
					it.token.Header = header                // Ensure header struct is populated if possible
				}
			}

			err := DefaultVerifyTokenSignature(it.token, currentJwksUrl)

			if it.expectError {
				assert.Error(t, err)
				if it.errorMsg != "" {
					assert.Contains(t, err.Error(), it.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetExponentialFromKey(t *testing.T) {
	tests := []struct {
		name          string
		e             string
		expectedValue int
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "Standard Exponent AQAB",
			e:             "AQAB",
			expectedValue: 65537,
			expectError:   false,
		},
		{
			name:          "Empty Exponent",
			e:             "",
			expectedValue: 65537,
			expectError:   false,
		},
		{
			name:          "Custom Exponent (3)",
			e:             "Aw", // base64url encoding of [3]
			expectedValue: 3,
			expectError:   false,
		},
		{
			name:          "Invalid Base64 Exponent",
			e:             "!",
			expectedValue: 0,
			expectError:   true,
			errorMsg:      "failed to decode exponent E: illegal base64 data",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			val, err := getExponentialFromKey(it.e)

			if it.expectError {
				assert.Error(t, err)
				assert.Equal(t, it.expectedValue, val) // Should return 0 on error
				if it.errorMsg != "" {
					assert.ErrorContains(t, err, it.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, it.expectedValue, val)
			}
		})
	}
}
