package config

import (
	"testing"
	"time"

	"github.com/napilab/discoverd/internal/discovery"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Mode:           ModeClient,
				MulticastGroup: "239.255.255.250",
				Port:           9999,
				Secret:         "secret",
				Timeout:        5 * time.Second,
				MaxClockSkew:   10 * time.Second,
				TTL:            2,
				Retries:        1,
				Output:         "text",
				Device:         discovery.DeviceInfo{ID: "device-1"},
				Server:         discovery.ServerInfo{ID: "server-1", IP: "127.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			cfg: Config{
				Mode:           "invalid",
				MulticastGroup: "239.255.255.250",
				Port:           9999,
				Secret:         "secret",
				Timeout:        5 * time.Second,
				MaxClockSkew:   10 * time.Second,
				TTL:            2,
				Retries:        1,
				Output:         "text",
				Device:         discovery.DeviceInfo{ID: "device-1"},
				Server:         discovery.ServerInfo{ID: "server-1", IP: "127.0.0.1"},
			},
			wantErr: true,
		},
		{
			name: "invalid multicast address",
			cfg: Config{
				Mode:           ModeClient,
				MulticastGroup: "127.0.0.1",
				Port:           9999,
				Secret:         "secret",
				Timeout:        5 * time.Second,
				MaxClockSkew:   10 * time.Second,
				TTL:            2,
				Retries:        1,
				Output:         "text",
				Device:         discovery.DeviceInfo{ID: "device-1"},
				Server:         discovery.ServerInfo{ID: "server-1", IP: "127.0.0.1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		envSecret string
		wantErr   bool
		check     func(t *testing.T, cfg Config)
	}{
		{
			name:      "defaults from env",
			args:      nil,
			envSecret: "env-secret",
			wantErr:   false,
			check: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Secret != "env-secret" {
					t.Fatalf("unexpected secret: %q", cfg.Secret)
				}
				if cfg.Mode != ModeClient {
					t.Fatalf("unexpected mode: %q", cfg.Mode)
				}
				if cfg.Device.ID != discovery.SystemDeviceID() {
					t.Fatalf("unexpected default device id: %q", cfg.Device.ID)
				}
			},
		},
		{
			name:      "flag override",
			args:      []string{"--mode=server", "--output=json", "--secret=flag-secret"},
			envSecret: "env-secret",
			wantErr:   false,
			check: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Mode != ModeServer {
					t.Fatalf("unexpected mode: %q", cfg.Mode)
				}
				if cfg.Output != "json" {
					t.Fatalf("unexpected output: %q", cfg.Output)
				}
				if cfg.Secret != "flag-secret" {
					t.Fatalf("unexpected secret: %q", cfg.Secret)
				}
				if cfg.Device.ID != discovery.SystemDeviceID() {
					t.Fatalf("unexpected system device id: %q", cfg.Device.ID)
				}
				if cfg.Server.ID == "" {
					t.Fatal("expected auto server id")
				}
				if cfg.Server.IP == "" {
					t.Fatal("expected auto server ip")
				}
			},
		},
		{
			name:      "device-id flag is not supported",
			args:      []string{"--device-id=node-01"},
			envSecret: "env-secret",
			wantErr:   true,
		},
		{
			name:      "invalid mode",
			args:      []string{"--mode=bad"},
			envSecret: "env-secret",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.args, func(string) string { return tt.envSecret })
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}
