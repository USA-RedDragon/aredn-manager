package utils

import (
	"crypto/rand"
	"fmt"
)

// Generate a link-local address beginning with fe80
func GenerateIPv6LinkLocalAddress() (string, error) {
	address := make([]byte, 14)
	read, err := rand.Read(address)
	if err != nil {
		return "", err
	}
	if read != 14 {
		return "", fmt.Errorf("failed to read 6 random bytes")
	}

	return fmt.Sprintf(
		"fe80:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		address[0],
		address[1],
		address[2],
		address[3],
		address[4],
		address[5],
		address[6],
		address[7],
		address[8],
		address[9],
		address[10],
		address[11],
		address[12],
		address[13],
	), nil
}
