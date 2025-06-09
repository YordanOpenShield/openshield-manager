package utils

import (
	"fmt"
	"net"
	"strings"
	"unicode/utf8"
)

// GetAllLocalAddresses returns all non-loopback, non-APIPA IPv4 addresses.
func GetAllLocalAddresses() ([]string, error) {
	var addresses []string

	// If osquery fails or returns no addresses, fallback to net.InterfaceAddrs
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip := ipnet.IP.To4()
			// Skip APIPA addresses (169.254.x.x)
			if ip[0] == 169 && ip[1] == 254 {
				continue
			}
			addresses = append(addresses, ip.String())
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no non-loopback addresses found")
	}
	return addresses, nil
}

func SanitizeString(input string) string {
	if utf8.ValidString(input) {
		return input
	}
	return strings.ToValidUTF8(input, "�")
}
