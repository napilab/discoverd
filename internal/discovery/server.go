package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// ServerConfig defines runtime settings for discovery server mode.
type ServerConfig struct {
	MulticastGroup string
	Port           int
	Secret         string
	MaxClockSkew   time.Duration
	Server         ServerInfo
	Logger         *log.Logger
}

// RunServer listens on multicast and responds to valid discovery requests.
func RunServer(ctx context.Context, cfg ServerConfig) error {
	groupIP := net.ParseIP(cfg.MulticastGroup)
	if groupIP == nil {
		return fmt.Errorf("parse multicast group %q", cfg.MulticastGroup)
	}

	listenAddr := &net.UDPAddr{IP: groupIP, Port: cfg.Port}
	conn, err := net.ListenMulticastUDP("udp4", nil, listenAddr)
	if err != nil {
		return fmt.Errorf("listen multicast udp: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logOrDiscard(cfg.Logger, "close server socket: %v", closeErr)
		}
	}()

	if err := conn.SetReadBuffer(64 * 1024); err != nil {
		return fmt.Errorf("set read buffer: %w", err)
	}

	logOrDiscard(cfg.Logger, "server listening on multicast %s:%d", cfg.MulticastGroup, cfg.Port)

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return fmt.Errorf("set server read deadline: %w", err)
		}

		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("server read request: %w", err)
		}

		req, err := ParseAndValidateRequest(buf[:n], cfg.MaxClockSkew)
		if err != nil {
			logOrDiscard(cfg.Logger, "skip request from %s: %v", src.String(), err)
			continue
		}

		logOrDiscard(
			cfg.Logger,
			"client beacon detected: id=%s src=%s ts=%s",
			req.Client.ID,
			src.String(),
			time.Unix(req.Timestamp, 0).UTC().Format(time.RFC3339),
		)

		resp := DiscoverResponse{
			Type:      MessageTypeDiscoverResponse,
			Version:   discoverVersion,
			RequestID: req.RequestID,
			Timestamp: time.Now().Unix(),
			Nonce:     req.Nonce,
			Server:    cfg.Server,
		}

		signature, err := SignMessage([]byte(cfg.Secret), resp)
		if err != nil {
			logOrDiscard(cfg.Logger, "sign response for %s: %v", src.String(), err)
			continue
		}
		resp.Signature = signature

		payload, err := json.Marshal(resp)
		if err != nil {
			logOrDiscard(cfg.Logger, "marshal response for %s: %v", src.String(), err)
			continue
		}

		if _, err := conn.WriteToUDP(payload, src); err != nil {
			logOrDiscard(cfg.Logger, "send response to %s: %v", src.String(), err)
		}
	}
}
