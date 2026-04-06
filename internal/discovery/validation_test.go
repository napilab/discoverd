package discovery

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseAndValidateResponse(t *testing.T) {
	secret := []byte("test-secret")
	request := DiscoverRequest{
		Type:      MessageTypeDiscoverRequest,
		Version:   discoverVersion,
		RequestID: "request-1",
		Timestamp: time.Now().Unix(),
		Nonce:     "nonce-1",
		Client:    DeviceInfo{ID: "device-1"},
	}

	buildPayload := func(modify func(*DiscoverResponse)) []byte {
		t.Helper()
		resp := DiscoverResponse{
			Type:      MessageTypeDiscoverResponse,
			Version:   discoverVersion,
			RequestID: request.RequestID,
			Timestamp: time.Now().Unix(),
			Nonce:     request.Nonce,
			Server: ServerInfo{
				ID: "server-1",
				IP: "127.0.0.1",
			},
		}
		if modify != nil {
			modify(&resp)
		}

		sig, err := SignMessage(secret, resp)
		if err != nil {
			t.Fatalf("SignMessage returned error: %v", err)
		}
		resp.Signature = sig

		payload, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}

		return payload
	}

	tests := []struct {
		name    string
		payload []byte
		wantErr bool
	}{
		{name: "valid response", payload: buildPayload(nil), wantErr: false},
		{
			name: "nonce mismatch",
			payload: buildPayload(func(resp *DiscoverResponse) {
				resp.Nonce = "wrong-nonce"
			}),
			wantErr: true,
		},
		{
			name: "request id mismatch",
			payload: buildPayload(func(resp *DiscoverResponse) {
				resp.RequestID = "wrong-request"
			}),
			wantErr: true,
		},
		{
			name: "missing signature",
			payload: func() []byte {
				resp := DiscoverResponse{
					Type:      MessageTypeDiscoverResponse,
					Version:   discoverVersion,
					RequestID: request.RequestID,
					Timestamp: time.Now().Unix(),
					Nonce:     request.Nonce,
					Server:    ServerInfo{ID: "server-1", IP: "127.0.0.1"},
				}
				payload, err := json.Marshal(resp)
				if err != nil {
					t.Fatalf("json.Marshal returned error: %v", err)
				}
				return payload
			}(),
			wantErr: true,
		},
		{
			name: "stale response",
			payload: buildPayload(func(resp *DiscoverResponse) {
				resp.Timestamp = time.Now().Add(-time.Minute).Unix()
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAndValidateResponse(tt.payload, request, secret, 10*time.Second)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestParseAndValidateRequest(t *testing.T) {
	buildPayload := func(req DiscoverRequest) []byte {
		t.Helper()
		payload, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		return payload
	}

	validReq := DiscoverRequest{
		Type:      MessageTypeDiscoverRequest,
		Version:   discoverVersion,
		RequestID: "request-1",
		Timestamp: time.Now().Unix(),
		Nonce:     "nonce-1",
		Client:    DeviceInfo{ID: "device-1"},
	}

	tests := []struct {
		name    string
		payload []byte
		wantErr bool
	}{
		{name: "valid request", payload: buildPayload(validReq), wantErr: false},
		{
			name: "invalid type",
			payload: buildPayload(DiscoverRequest{
				Type:      "OTHER",
				Version:   discoverVersion,
				RequestID: "request-1",
				Timestamp: time.Now().Unix(),
				Nonce:     "nonce-1",
				Client:    DeviceInfo{ID: "device-1"},
			}),
			wantErr: true,
		},
		{
			name: "stale timestamp",
			payload: buildPayload(DiscoverRequest{
				Type:      MessageTypeDiscoverRequest,
				Version:   discoverVersion,
				RequestID: "request-1",
				Timestamp: time.Now().Add(-time.Minute).Unix(),
				Nonce:     "nonce-1",
				Client:    DeviceInfo{ID: "device-1"},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAndValidateRequest(tt.payload, 10*time.Second)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
