package utils

import (
	"fmt"
	"net"
)

// Generate a link-local address beginning with fe80 and ending with the 4 octets of the IPv4 address
func GenerateIPv6LinkLocalAddress(ipv4 net.IP) (string, error) {
	v4Bytes := ipv4.To4()
	ipv6 := net.ParseIP(fmt.Sprintf("fe80::%02x%02x:%02x%02x", v4Bytes[0], v4Bytes[1], v4Bytes[2], v4Bytes[3]))
	if ipv6 == nil {
		return "", fmt.Errorf("failed to parse IPv6 address")
	}
	return ipv6.String(), nil
}
