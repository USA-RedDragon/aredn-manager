package utils_test

import (
	"net"
	"testing"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
)

func TestGenerateIPv6LinkLocalAddress(t *testing.T) {
	// Test that the generated IPs are valid
	for i := 0; i < 1000000; i++ {
		ip, err := utils.GenerateIPv6LinkLocalAddress()
		if err != nil {
			t.Fatalf("Failed to generate link-local IPv6 address: %v", err)
		}
		ipObj := net.ParseIP(ip)
		if ipObj == nil {
			t.Fatalf("Invalid IP address: %s", ip)
		}
		if !ipObj.IsLinkLocalUnicast() {
			t.Fatalf("Non-link-local unicast IP address generated: %s", ip)
		}
	}
}
