package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/poteto-go/poteto/utils"
)

// Cache jwk
//
// To be default at v2
func CachedVerifyTokenSignature(idToken IdToken, pCache *cache.Cache, jwksUrl string) error {
	// decode header
	byteHeader, err := utils.JwtDecodeSegment(idToken.RawHeader)
	if err != nil {
		return err
	}

	header := Header{}
	if err := json.Unmarshal(byteHeader, &header); err != nil {
		return err
	}
	// Assign header back to idToken.Header. This was missing before but needed by getJwk.
	idToken.Header = header

	var key jwk
	cachedKey, ok := pCache.Get(jwksUrl)
	if ok {
		key = cachedKey.(jwk)
	} else {
		// verify signature
		keyFound, err := getJwk(idToken, jwksUrl)
		if err != nil {
			return err
		}

		pCache.Set(
			jwksUrl,
			keyFound,
			cache.DefaultExpiration,
		)
		key = keyFound
	}

	// base64.RawURLEncoding.DecodeString(key.N)
	byteN, err := utils.JwtUrlDecodeSegment(key.N)
	if err != nil {
		return err
	}

	exponent, err := getExponentialFromKey(key.E)
	if err != nil {
		return err
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(byteN),
		E: exponent,
	}

	headerAndPayload := fmt.Sprintf("%s.%s", idToken.RawHeader, idToken.RawPayload)
	sha := sha256.New()
	sha.Write([]byte(headerAndPayload))

	decSignature, err := utils.JwtUrlDecodeSegment(idToken.RawSignature)
	if err != nil {
		return err
	}

	// Assuming ALG is RS256 based on crypto/sha256 usage
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sha.Sum(nil), decSignature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// DefaultVerifyTokenSignature verifies the signature of an ID token using JWKS.
func DefaultVerifyTokenSignature(idToken IdToken, jwksUrl string) error {
	// decode header
	byteHeader, err := utils.JwtDecodeSegment(idToken.RawHeader)
	if err != nil {
		return err
	}

	header := Header{}
	if err := json.Unmarshal(byteHeader, &header); err != nil {
		return err
	}
	// Assign header back to idToken.Header. This was missing before but needed by getJwk.
	idToken.Header = header

	// verify signature
	key, err := getJwk(idToken, jwksUrl)
	if err != nil {
		return err
	}

	// base64.RawURLEncoding.DecodeString(key.N)
	byteN, err := utils.JwtUrlDecodeSegment(key.N)
	if err != nil {
		return err
	}

	exponent, err := getExponentialFromKey(key.E)
	if err != nil {
		return err
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(byteN),
		E: exponent,
	}

	headerAndPayload := fmt.Sprintf("%s.%s", idToken.RawHeader, idToken.RawPayload)
	sha := sha256.New()
	sha.Write([]byte(headerAndPayload))

	decSignature, err := utils.JwtUrlDecodeSegment(idToken.RawSignature)
	if err != nil {
		return err
	}

	// Assuming ALG is RS256 based on crypto/sha256 usage
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sha.Sum(nil), decSignature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func getJwk(token IdToken, jwksUrl string) (jwk, error) {
	parsedUrl, err := url.Parse(jwksUrl)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to parse jwks url: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reqWithCtx, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, parsedUrl.String(), nil)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to create request of GET JWKs endpoint: %w", err)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(reqWithCtx)
	if err != nil {
		return jwk{}, fmt.Errorf("failed to GET JWKs endpoint: %w", err)
	}

	defer resp.Body.Close()
	byteArray, _ := io.ReadAll(resp.Body)

	keys := &jwks{}
	if err := json.Unmarshal(byteArray, keys); err != nil {
		return jwk{}, fmt.Errorf("failed to unmarshal JWKs response: %w", err)
	}

	foundKey, err := keys.find(token.Header.Kid)
	if err != nil {
		return jwk{}, fmt.Errorf("jwks keys not found")
	}

	return foundKey, nil
}

// 'e' is typically "AQAB" which is 65537
func getExponentialFromKey(e string) (int, error) {
	if e == "AQAB" || e == "" { // Default exponent if missing or standard
		return 65537, nil
	}

	// Ensure E is base64url encoded before decoding
	byteE, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return 0, fmt.Errorf("failed to decode exponent E: %w", err)
	}

	return int(new(big.Int).SetBytes(byteE).Int64()), nil
}
