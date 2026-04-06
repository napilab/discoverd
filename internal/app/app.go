package app

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/napilab/discoverd/internal/config"
	"github.com/napilab/discoverd/internal/discovery"
)

// Run executes the application based on the selected mode.
func Run(ctx context.Context, cfg config.Config, logger *log.Logger, out io.Writer) error {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	if out == nil {
		return fmt.Errorf("output writer is nil")
	}

	if config.IsUsingDefaultSecret(cfg) {
		logger.Println("warning: using development default secret; set --secret or DISCOVERD_SECRET in production")
	}

	switch cfg.Mode {
	case config.ModeClient:
		_, err := discovery.RunClient(ctx, discovery.ClientConfig{
			MulticastGroup: cfg.MulticastGroup,
			Port:           cfg.Port,
			Secret:         cfg.Secret,
			Timeout:        cfg.Timeout,
			MaxClockSkew:   cfg.MaxClockSkew,
			TTL:            cfg.TTL,
			Retries:        cfg.Retries,
			Device:         cfg.Device,
			Logger:         logger,
			OnDiscovered: func(server discovery.ServerInfo) error {
				return discovery.PrintServerEvent(out, server, cfg.Output)
			},
		})
		return err
	case config.ModeServer:
		return discovery.RunServer(ctx, discovery.ServerConfig{
			MulticastGroup: cfg.MulticastGroup,
			Port:           cfg.Port,
			Secret:         cfg.Secret,
			MaxClockSkew:   cfg.MaxClockSkew,
			Server:         cfg.Server,
			Logger:         logger,
		})
	default:
		return fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}
