package discovery

const (
	// MessageTypeDiscoverRequest is a client discovery request message type.
	MessageTypeDiscoverRequest = "DISCOVER_REQUEST"
	// MessageTypeDiscoverResponse is a server discovery response message type.
	MessageTypeDiscoverResponse = "DISCOVER_RESPONSE"
)

// DeviceInfo describes a client device that initiates discovery.
type DeviceInfo struct {
	ID string `json:"id"`
}

// ServerInfo describes a server endpoint advertised over discovery.
type ServerInfo struct {
	ID string `json:"id"`
	IP string `json:"ip"`
}

// DiscoverRequest is sent by clients over multicast.
type DiscoverRequest struct {
	Type      string     `json:"type"`
	Version   int        `json:"version"`
	RequestID string     `json:"requestId"`
	Timestamp int64      `json:"ts"`
	Nonce     string     `json:"nonce"`
	Client    DeviceInfo `json:"client"`
}

// DiscoverResponse is sent by servers to a client unicast address.
type DiscoverResponse struct {
	Type      string     `json:"type"`
	Version   int        `json:"version"`
	RequestID string     `json:"requestId"`
	Timestamp int64      `json:"ts"`
	Nonce     string     `json:"nonce"`
	Server    ServerInfo `json:"server"`
	Signature string     `json:"signature,omitempty"`
}
