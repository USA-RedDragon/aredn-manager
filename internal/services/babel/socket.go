package babel

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
)

const (
	socketPath = "/var/run/babel.sock"
)

type Client struct {
}

// NewSocketClient creates a new SocketClient
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// Close closes the connection to the UNIX socket
func (c *Client) Close() error {
	return nil
}

type Interface struct {
	Name string
	IPv4 net.IP
	IPv6 net.IP
}

func (c *Client) GetInterfaces() ([]Interface, error) {
	dumpInterfaces := []byte("dump-interfaces\nquit\n")
	var ifaces []Interface

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	n, err := conn.Write(dumpInterfaces)
	if err != nil {
		return nil, err
	}
	if n != len(dumpInterfaces) {
		return nil, fmt.Errorf("failed to write all bytes to socket")
	}

	for scanner.Scan() {
		line := scanner.Text()
		// Process the line
		if !strings.HasPrefix(line, "add interface") {
			continue
		}
		ifacePuller := regexp.MustCompile(`interface\s([a-zA-Z0-9-]*)\s.*ipv6\s(.*)\sipv4\s(.*)`)
		matches := ifacePuller.FindStringSubmatch(line)
		if len(matches) != 4 {
			slog.Warn("failed to parse interface line", "line", line)
			continue
		}

		ifaces = append(ifaces, Interface{
			Name: matches[1],
			IPv6: net.ParseIP(matches[2]),
			IPv4: net.ParseIP(matches[3]),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ifaces, nil
}
