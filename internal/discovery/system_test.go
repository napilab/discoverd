package discovery

import (
	"net"
	"testing"
)

func TestSystemServerInfo(t *testing.T) {
	info := SystemServerInfo("239.255.255.250")

	if info.ID == "" {
		t.Fatal("SystemServerInfo returned empty ID")
	}
	if info.IP == "" {
		t.Fatal("SystemServerInfo returned empty IP")
	}
	if ip := net.ParseIP(info.IP); ip == nil || ip.To4() == nil {
		t.Fatalf("SystemServerInfo returned invalid IPv4 %q", info.IP)
	}
}
