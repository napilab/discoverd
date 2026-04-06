package discovery

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestSignMessageAndVerifyMessage(t *testing.T) {
	secret := []byte("test-secret")
	msg := DiscoverResponse{
		Type:      MessageTypeDiscoverResponse,
		Version:   discoverVersion,
		RequestID: "request-1",
		Timestamp: time.Now().Unix(),
		Nonce:     "nonce-1",
		Server: ServerInfo{
			ID: "server-1",
			IP: "127.0.0.1",
		},
	}

	sig, err := SignMessage(secret, msg)
	if err != nil {
		t.Fatalf("SignMessage returned error: %v", err)
	}

	ok, err := VerifyMessage(secret, msg, sig)
	if err != nil {
		t.Fatalf("VerifyMessage returned error: %v", err)
	}
	if !ok {
		t.Fatal("VerifyMessage returned false for valid signature")
	}

	ok, err = VerifyMessage(secret, msg, "invalid-signature")
	if err != nil {
		t.Fatalf("VerifyMessage returned error for invalid signature: %v", err)
	}
	if ok {
		t.Fatal("VerifyMessage returned true for invalid signature")
	}
}

func TestGenerateNonce(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "valid size", size: 16, wantErr: false},
		{name: "invalid size", size: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, err := GenerateNonce(tt.size)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GenerateNonce returned error: %v", err)
			}
			if nonce == "" {
				t.Fatal("GenerateNonce returned empty nonce")
			}
			if _, err := base64.StdEncoding.DecodeString(nonce); err != nil {
				t.Fatalf("nonce is not valid base64: %v", err)
			}
		})
	}
}
