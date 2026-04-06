package discovery

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// GenerateNonce creates a random base64 nonce.
func GenerateNonce(size int) (string, error) {
	if size <= 0 {
		return "", fmt.Errorf("nonce size must be positive: %d", size)
	}

	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

// SignMessage signs JSON-encoded data with HMAC-SHA256 and returns a base64 signature.
func SignMessage(secret []byte, data any) (string, error) {
	payload, err := canonicalJSON(data)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(payload); err != nil {
		return "", fmt.Errorf("write hmac payload: %w", err)
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// VerifyMessage checks whether the supplied signature matches data.
func VerifyMessage(secret []byte, data any, signature string) (bool, error) {
	expected, err := SignMessage(secret, data)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(expected), []byte(signature)), nil
}

func canonicalJSON(data any) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical json: %w", err)
	}

	return payload, nil
}
