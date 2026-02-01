package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

const (
	apiKeyCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	apiKeyLength  = 32
)

func GenerateAPIKey() (string, error) {
	bytes := make([]byte, apiKeyLength)
	for i := range bytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(apiKeyCharset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		bytes[i] = apiKeyCharset[num.Int64()]
	}
	return string(bytes), nil
}

func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		length = 16
	}

	bytes := make([]byte, length*2) // Generate extra bytes for base64 encoding
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Use base64 URL encoding without padding
	randomString := base64.RawURLEncoding.EncodeToString(bytes)
	return randomString[:length], nil
}

func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
