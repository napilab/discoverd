package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ParseAndValidateRequest decodes and validates a discovery request payload.
func ParseAndValidateRequest(payload []byte, maxClockSkew time.Duration) (DiscoverRequest, error) {
	var req DiscoverRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return DiscoverRequest{}, fmt.Errorf("decode request json: %w", err)
	}
	if req.Type != MessageTypeDiscoverRequest {
		return DiscoverRequest{}, fmt.Errorf("unexpected request type %q", req.Type)
	}
	if req.RequestID == "" {
		return DiscoverRequest{}, errors.New("missing requestId")
	}
	if req.Nonce == "" {
		return DiscoverRequest{}, errors.New("missing nonce")
	}
	if req.Client.ID == "" {
		return DiscoverRequest{}, errors.New("missing client.id")
	}
	if IsTimestampStale(req.Timestamp, maxClockSkew) {
		return DiscoverRequest{}, errors.New("stale request timestamp")
	}

	return req, nil
}

// ParseAndValidateResponse decodes and validates a discovery response payload.
func ParseAndValidateResponse(payload []byte, req DiscoverRequest, secret []byte, maxClockSkew time.Duration) (DiscoverResponse, error) {
	var resp DiscoverResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return DiscoverResponse{}, fmt.Errorf("decode response json: %w", err)
	}
	if resp.Type != MessageTypeDiscoverResponse {
		return DiscoverResponse{}, fmt.Errorf("unexpected response type %q", resp.Type)
	}
	if resp.RequestID != req.RequestID {
		return DiscoverResponse{}, errors.New("requestId mismatch")
	}
	if resp.Nonce != req.Nonce {
		return DiscoverResponse{}, errors.New("nonce mismatch")
	}
	if resp.Signature == "" {
		return DiscoverResponse{}, errors.New("missing signature")
	}
	if resp.Server.ID == "" {
		return DiscoverResponse{}, errors.New("missing server.id")
	}
	if resp.Server.IP == "" {
		return DiscoverResponse{}, errors.New("missing server.ip")
	}
	if IsTimestampStale(resp.Timestamp, maxClockSkew) {
		return DiscoverResponse{}, errors.New("stale response timestamp")
	}

	signature := resp.Signature
	resp.Signature = ""
	valid, err := VerifyMessage(secret, resp, signature)
	if err != nil {
		return DiscoverResponse{}, fmt.Errorf("verify signature: %w", err)
	}
	if !valid {
		return DiscoverResponse{}, errors.New("invalid signature")
	}

	resp.Signature = signature
	return resp, nil
}

// IsTimestampStale reports whether timestamp differs from current time by more than maxClockSkew.
func IsTimestampStale(unixTS int64, maxClockSkew time.Duration) bool {
	if unixTS == 0 {
		return true
	}
	delta := time.Since(time.Unix(unixTS, 0))
	if delta < 0 {
		delta = -delta
	}

	return delta > maxClockSkew
}
