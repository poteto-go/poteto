package utils

import (
	"encoding/base64"
	"strings"
)

func JwtDecodeSegment(raw string) ([]byte, error) {
	paddingLength := ((4 - len(raw)%4) % 4)
	padding := strings.Repeat("=", paddingLength)
	padded := strings.Join([]string{raw, padding}, "")

	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		return []byte(""), err
	}

	return decoded, nil
}
