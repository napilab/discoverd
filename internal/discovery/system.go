package discovery

import (
	"net"
	"os"
)

const fallbackServerIP = "127.0.0.1"

// SystemServerInfo resolves server identity and IPv4 address from the current host.
func SystemServerInfo(multicastGroup string) ServerInfo {
	return ServerInfo{
		ID: systemHostname(),
		IP: systemIPv4(multicastGroup),
	}
}

// SystemDeviceID resolves default device identifier from the current host.
func SystemDeviceID() string {
	return systemHostname()
}

func systemHostname() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "unknown-host"
	}

	return hostname
}

func systemIPv4(multicastGroup string) string {
	if ip := outboundIPv4(multicastGroup); ip != "" {
		return ip
	}
	if ip := firstNonLoopbackIPv4(); ip != "" {
		return ip
	}

	return fallbackServerIP
}

func outboundIPv4(multicastGroup string) string {
	if net.ParseIP(multicastGroup) == nil {
		return ""
	}

	conn, err := net.Dial("udp4", net.JoinHostPort(multicastGroup, "9"))
	if err != nil {
		return ""
	}

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if closeErr := conn.Close(); closeErr != nil {
		return ""
	}
	if !ok || localAddr.IP == nil {
		return ""
	}
	ip4 := localAddr.IP.To4()
	if ip4 == nil {
		return ""
	}

	return ip4.String()
}

func firstNonLoopbackIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}

			return ip4.String()
		}
	}

	return ""
}
