package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultVerifyTokenSignature verifies the signature of an ID token using JWKS.
func DefaultVerifyTokenSignature(idToken IdToken, jwksUrl string) error {
	// decode header
	byteHeader, err := jwtDecodeSegment(idToken.RawHeader)
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

	byteN, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return err
	}

	// 'e' is typically "AQAB" which is 65537
	var exponent int
	if key.E == "AQAB" || key.E == "" { // Default exponent if missing or standard
		exponent = 65537
	} else {
		// Ensure E is base64url encoded before decoding
		byteE, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return fmt.Errorf("failed to decode exponent E: %w", err)
		}
		exponent = int(new(big.Int).SetBytes(byteE).Int64())
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(byteN),
		E: exponent,
	}

	headerAndPayload := fmt.Sprintf("%s.%s", idToken.RawHeader, idToken.RawPayload)
	sha := sha256.New()
	sha.Write([]byte(headerAndPayload))

	decSignature, err := base64.RawURLEncoding.DecodeString(idToken.RawSignature)
	if err != nil {
		return err
	}

	// Assuming ALG is RS256 based on crypto/sha256 usage
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sha.Sum(nil), decSignature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// jwtDecodeSegment decodes JWT specific base64url encoding with padding.
func jwtDecodeSegment(raw string) ([]byte, error) {
	paddingLength := ((4 - len(raw)%4) % 4)
	padding := strings.Repeat("=", paddingLength)
	padded := strings.Join([]string{raw, padding}, "")

	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		return []byte(""), err
	}

	return decoded, nil
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
		return jwk{}, err
	}

	return foundKey, nil
}

func (keys jwks) find(kid string) (jwk, error) {
	var foundKey jwk
	for _, key := range keys.Keys {
		if key.Kid == kid {
			foundKey = key
			break
		}
	}

	if foundKey != (jwk{}) {
		return foundKey, nil
	} else {
		return jwk{}, errors.New("jwks keys not found")
	}
}
