//go:build windows

package discovery

import (
	"fmt"
	"net"
	"syscall"
)

func setMulticastTTL(conn *net.UDPConn, ttl int) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("get raw connection: %w", err)
	}

	var controlErr error
	err = rawConn.Control(func(fd uintptr) {
		controlErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_TTL, ttl)
	})
	if err != nil {
		return fmt.Errorf("run raw control: %w", err)
	}
	if controlErr != nil {
		return fmt.Errorf("set socket option ip_multicast_ttl: %w", controlErr)
	}

	return nil
}
