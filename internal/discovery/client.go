package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/google/uuid"
)

const discoverVersion = 1

// ClientConfig defines runtime settings for discovery client mode.
type ClientConfig struct {
	MulticastGroup string
	Port           int
	Secret         string
	Timeout        time.Duration
	MaxClockSkew   time.Duration
	TTL            int
	Retries        int
	Device         DeviceInfo
	Logger         *log.Logger
	OnDiscovered   func(ServerInfo) error
}

// RunClient broadcasts discovery requests and collects valid server responses.
func RunClient(ctx context.Context, cfg ClientConfig) ([]ServerInfo, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("listen udp client socket: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logOrDiscard(cfg.Logger, "close client socket: %v", closeErr)
		}
	}()

	if err := setMulticastTTL(conn, cfg.TTL); err != nil {
		return nil, fmt.Errorf("set multicast ttl: %w", err)
	}

	targetIP := net.ParseIP(cfg.MulticastGroup)
	if targetIP == nil {
		return nil, fmt.Errorf("parse multicast group %q", cfg.MulticastGroup)
	}
	target := &net.UDPAddr{IP: targetIP, Port: cfg.Port}

	serversByID := make(map[string]ServerInfo)

	for {
		for attempt := 1; attempt <= cfg.Retries; attempt++ {
			select {
			case <-ctx.Done():
				return sortedServers(serversByID), nil
			default:
			}

			req, err := newDiscoverRequest(cfg.Device)
			if err != nil {
				return nil, fmt.Errorf("build request: %w", err)
			}

			payload, err := json.Marshal(req)
			if err != nil {
				return nil, fmt.Errorf("marshal request: %w", err)
			}

			if _, err := conn.WriteToUDP(payload, target); err != nil {
				return nil, fmt.Errorf("send discover request: %w", err)
			}

			if err := receiveResponses(ctx, conn, req, cfg, serversByID); err != nil {
				return nil, err
			}
		}
	}
}

func receiveResponses(ctx context.Context, conn *net.UDPConn, req DiscoverRequest, cfg ClientConfig, found map[string]ServerInfo) error {
	deadline := time.Now().Add(cfg.Timeout)
	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := conn.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("set read deadline: %w", err)
		}

		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil
			}
			return fmt.Errorf("read response: %w", err)
		}

		resp, err := ParseAndValidateResponse(buf[:n], req, []byte(cfg.Secret), cfg.MaxClockSkew)
		if err != nil {
			logOrDiscard(cfg.Logger, "skip response from %s: %v", addr.String(), err)
			continue
		}

		if changed := upsertServer(found, resp.Server); changed && cfg.OnDiscovered != nil {
			if err := cfg.OnDiscovered(resp.Server); err != nil {
				return fmt.Errorf("report discovered server: %w", err)
			}
		}
	}
}

func upsertServer(found map[string]ServerInfo, server ServerInfo) bool {
	current, exists := found[server.ID]
	if exists && current.IP == server.IP {
		return false
	}

	found[server.ID] = server
	return true
}

func sortedServers(serversByID map[string]ServerInfo) []ServerInfo {
	found := make([]ServerInfo, 0, len(serversByID))
	for _, srv := range serversByID {
		found = append(found, srv)
	}
	sort.Slice(found, func(i, j int) bool {
		return found[i].ID < found[j].ID
	})

	return found
}

func newDiscoverRequest(client DeviceInfo) (DiscoverRequest, error) {
	nonce, err := GenerateNonce(16)
	if err != nil {
		return DiscoverRequest{}, fmt.Errorf("generate nonce: %w", err)
	}

	return DiscoverRequest{
		Type:      MessageTypeDiscoverRequest,
		Version:   discoverVersion,
		RequestID: uuid.NewString(),
		Timestamp: time.Now().Unix(),
		Nonce:     nonce,
		Client:    client,
	}, nil
}

func logOrDiscard(logger *log.Logger, format string, args ...any) {
	if logger == nil {
		return
	}
	logger.Printf(format, args...)
}
