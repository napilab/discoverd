package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// PrintServers writes discovered servers in the requested format.
func PrintServers(out io.Writer, found []ServerInfo, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(found); err != nil {
			return fmt.Errorf("encode json output: %w", err)
		}
		return nil
	case "text":
		if _, err := fmt.Fprintf(out, "Discovery finished, found %d server(s)\n", len(found)); err != nil {
			return fmt.Errorf("write text output header: %w", err)
		}
		for i, srv := range found {
			if _, err := fmt.Fprintf(out, "%d) id=%s ip=%s\n", i+1, srv.ID, srv.IP); err != nil {
				return fmt.Errorf("write text output row: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

// PrintServerEvent writes one discovered server event.
func PrintServerEvent(out io.Writer, server ServerInfo, format string) error {
	switch format {
	case "json":
		payload := map[string]string{
			"event": "server_discovered",
			"id":    server.ID,
			"ip":    server.IP,
			"ts":    time.Now().UTC().Format(time.RFC3339),
		}
		encoder := json.NewEncoder(out)
		if err := encoder.Encode(payload); err != nil {
			return fmt.Errorf("encode json event output: %w", err)
		}
		return nil
	case "text":
		if _, err := fmt.Fprintf(out, "[%s] server discovered: id=%s ip=%s\n", time.Now().UTC().Format(time.RFC3339), server.ID, server.IP); err != nil {
			return fmt.Errorf("write text event output: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
