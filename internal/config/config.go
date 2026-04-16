package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/napilab/discoverd/internal/discovery"
	"github.com/spf13/pflag"
)

const (
	// ModeClient runs multicast discovery requests.
	ModeClient = "client"
	// ModeServer runs multicast listener and responds to requests.
	ModeServer = "server"

	defaultMode           = ModeClient
	defaultMulticastGroup = "239.255.255.250"
	defaultPort           = 9999
	defaultTimeout        = 5 * time.Second
	defaultMaxClockSkew   = 10 * time.Second
	defaultTTL            = 2
	defaultRetries        = 1
	defaultOutput         = "text"
	defaultSecret         = "napilab"
)

// Config stores runtime settings for discovery client and server modes.
type Config struct {
	Mode           string
	MulticastGroup string
	Port           int
	Secret         string
	Timeout        time.Duration
	MaxClockSkew   time.Duration
	TTL            int
	Retries        int
	Output         string
	Device         discovery.DeviceInfo
	Server         discovery.ServerInfo
}

// Parse builds Config from command-line flags and environment values.
func Parse(args []string, lookupEnv func(string) string) (Config, error) {
	if lookupEnv == nil {
		lookupEnv = func(string) string { return "" }
	}

	secretFromEnv := lookupEnv("DISCOVERD_SECRET")
	if secretFromEnv == "" {
		secretFromEnv = defaultSecret
	}

	cfg := Config{
		Mode:           defaultMode,
		MulticastGroup: defaultMulticastGroup,
		Port:           defaultPort,
		Secret:         secretFromEnv,
		Timeout:        defaultTimeout,
		MaxClockSkew:   defaultMaxClockSkew,
		TTL:            defaultTTL,
		Retries:        defaultRetries,
		Output:         defaultOutput,
		Device: discovery.DeviceInfo{
			ID: discovery.SystemDeviceID(),
		},
	}

	fs := pflag.NewFlagSet("discoverd", pflag.ContinueOnError)
	fs.StringVar(&cfg.Mode, "mode", cfg.Mode, "work mode: client or server")
	fs.StringVar(&cfg.MulticastGroup, "mcast", cfg.MulticastGroup, "multicast IPv4 group")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "multicast UDP port")
	fs.StringVar(&cfg.Secret, "secret", cfg.Secret, "shared HMAC secret")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "client wait timeout per retry")
	fs.DurationVar(&cfg.MaxClockSkew, "max-clock-skew", cfg.MaxClockSkew, "maximum allowed timestamp skew")
	fs.IntVar(&cfg.TTL, "ttl", cfg.TTL, "multicast TTL for client requests")
	fs.IntVar(&cfg.Retries, "retries", cfg.Retries, "client retries count")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "client output format: text or json")

	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	cfg.Server = discovery.SystemServerInfo(cfg.MulticastGroup)

	cfg.Mode = strings.ToLower(cfg.Mode)
	cfg.Output = strings.ToLower(cfg.Output)

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks whether Config values are valid.
func Validate(cfg Config) error {
	switch cfg.Mode {
	case ModeClient, ModeServer:
	default:
		return fmt.Errorf("invalid --mode %q: expected client or server", cfg.Mode)
	}

	groupIP := net.ParseIP(cfg.MulticastGroup)
	if groupIP == nil || groupIP.To4() == nil {
		return fmt.Errorf("invalid --mcast %q: expected IPv4 multicast address", cfg.MulticastGroup)
	}
	if !groupIP.IsMulticast() {
		return fmt.Errorf("invalid --mcast %q: address must be multicast", cfg.MulticastGroup)
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid --port %d: expected 1..65535", cfg.Port)
	}
	if cfg.TTL < 1 || cfg.TTL > 255 {
		return fmt.Errorf("invalid --ttl %d: expected 1..255", cfg.TTL)
	}
	if cfg.Retries < 1 {
		return fmt.Errorf("invalid --retries %d: expected >= 1", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("invalid --timeout %s: expected > 0", cfg.Timeout)
	}
	if cfg.MaxClockSkew <= 0 {
		return fmt.Errorf("invalid --max-clock-skew %s: expected > 0", cfg.MaxClockSkew)
	}
	if cfg.Secret == "" {
		return errors.New("invalid --secret: must not be empty")
	}
	if cfg.Device.ID == "" {
		return errors.New("invalid --device-id: must not be empty")
	}
	if cfg.Server.ID == "" {
		return errors.New("detected server id is empty")
	}

	if net.ParseIP(cfg.Server.IP) == nil {
		return fmt.Errorf("detected server ip is invalid %q", cfg.Server.IP)
	}

	switch cfg.Output {
	case "text", "json":
	default:
		return fmt.Errorf("invalid --output %q: expected text or json", cfg.Output)
	}

	return nil
}

// IsUsingDefaultSecret reports whether the development secret is still configured.
func IsUsingDefaultSecret(cfg Config) bool {
	return cfg.Secret == defaultSecret
}
