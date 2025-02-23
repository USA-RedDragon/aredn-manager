package utils

import (
	"fmt"
	"net"
)

// Generate a link-local address beginning with fe80 and ending with the 4 octets of the IPv4 address
func GenerateIPv6LinkLocalAddress(ipv4 net.IP) (string, error) {
	ipv6 := net.ParseIP("fe80::")
	if ipv6 == nil {
		return "", fmt.Errorf("failed to parse IPv6 address")
	}
	ipv6 = append(ipv6, ipv4.To16()[12:]...)
	return ipv6.String(), nil
}
