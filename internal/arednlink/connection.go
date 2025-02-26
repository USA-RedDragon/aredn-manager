package arednlink

import (
	"io"
	"log/slog"
	"net"
	"regexp"
	"sync"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/puzpuzpuz/xsync/v3"
)

const HopDefault = 64

type Connection struct {
	conn          net.Conn
	writeLock     sync.Mutex
	config        *config.Config
	broadcastChan chan Message
	hosts         *xsync.MapOf[string, string]
	services      *xsync.MapOf[string, string]
	routes        ***xsync.MapOf[string, string]
	iface         string
}

func HandleConnection(
	config *config.Config,
	conn net.Conn,
	broadcastChan chan Message,
	hosts *xsync.MapOf[string, string],
	services *xsync.MapOf[string, string],
	routes ***xsync.MapOf[string, string],
) {
	// conn.RemoteAddr().String() should be in the format [fe80::ac1e:2c4%wgc32]:38428
	// where wgc32 is the interface name
	interfaceRegex := regexp.MustCompile(`^\[[a-fA-F0-9:]+%(.*)\]`)
	matches := interfaceRegex.FindStringSubmatch(conn.RemoteAddr().String())
	if len(matches) != 2 {
		slog.Error("arednlink: failed to parse remote address", "address", conn.RemoteAddr().String())
		return
	}

	iface := matches[1]
	if iface == "" {
		slog.Error("arednlink: failed to parse remote address", "address", conn.RemoteAddr().String())
		return
	}

	connection := &Connection{
		conn:          conn,
		writeLock:     sync.Mutex{},
		config:        config,
		broadcastChan: broadcastChan,
		hosts:         hosts,
		services:      services,
		routes:        routes,
		iface:         iface,
	}
	connection.start()
}

func (c *Connection) sendMessage(msg Message) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	msgBytes := msg.Bytes()
	n, err := c.conn.Write(msgBytes)
	if err != nil {
		return err
	}
	if n != len(msgBytes) {
		return io.ErrShortWrite
	}
	return nil
}

func (c *Connection) broadcastMessage(msg Message) {
	msg.ConnID = c.conn.RemoteAddr().String()
	msg.Hops--
	if msg.Hops > 0 {
		c.broadcastChan <- msg
	}
}

func (c *Connection) start() {
	buf := make([]byte, 2048)
	var currentMessage *Message

	stopChan := make(chan interface{})
	defer func() {
		close(stopChan)
		err := c.conn.Close()
		if err != nil {
			slog.Error("arednlink: failed to close connection", "error", err)
		}
	}()

	go func() {
		for {
			select {
			case msg := <-c.broadcastChan:
				if msg.ConnID == c.conn.RemoteAddr().String() {
					continue
				}
				if msg.DestIface != "" && msg.DestIface != c.iface {
					continue
				}
				err := c.sendMessage(msg)
				if err != nil {
					slog.Error("arednlink: failed to send message", "error", err)
					return
				}
			case <-stopChan:
				return
			}
		}
	}()

	for {
		n, err := c.conn.Read(buf)
		if err != nil && err != io.EOF {
			slog.Error("arednlink: failed to read from connection", "error", err)
			return
		}
		if n == 0 {
			return
		}

		slog.Debug("arednlink: received", "remote", c.conn.RemoteAddr(), "data", string(buf[:n]))

		for n > 0 {
			if currentMessage == nil {
				// We're not parsing a message yet, so we need to start reading the header
				if n < 8 {
					slog.Error("arednlink: received message with less than 8 bytes")
					return
				}

				currentMessage = readMessageHeader(buf[:8])

				// Check if the current message takes up the entire buffer
				if n == int(currentMessage.Length) {
					currentMessage.Payload = buf[8:n]
					forward := c.handleMessage(*currentMessage)
					if forward {
						c.broadcastMessage(*currentMessage)
					}
					currentMessage = nil
					n = 0
					continue
				}

				// Check if the current buffer is larger than the message
				if n > int(currentMessage.Length) {
					msgLen := int(currentMessage.Length)
					currentMessage.Payload = buf[8 : msgLen-8]
					forward := c.handleMessage(*currentMessage)
					if forward {
						c.broadcastMessage(*currentMessage)
					}
					currentMessage = nil

					buf = buf[msgLen:n]
					n -= int(msgLen)
					continue
				}

				// If we're here, the message is larger than the buffer
				// so we need to read the rest of the message in the next iteration
				currentMessage.Payload = buf[8:n]
				n = 0
			} else {
				// Current message is already being parsed
				bytesStillWanted := int(currentMessage.Length) - len(currentMessage.Payload) - 8
				if n == bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:bytesStillWanted]...)
					forward := c.handleMessage(*currentMessage)
					if forward {
						c.broadcastMessage(*currentMessage)
					}
					currentMessage = nil
					n = 0
					continue
				}

				if n > bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:bytesStillWanted]...)
					forward := c.handleMessage(*currentMessage)
					if forward {
						c.broadcastMessage(*currentMessage)
					}
					currentMessage = nil

					buf = buf[bytesStillWanted:n]
					n -= bytesStillWanted
					continue
				}

				currentMessage.Payload = append(currentMessage.Payload, buf[:n]...)
				n = 0
			}
		}
	}
}

func (c *Connection) validNextHop(cmd Command, srcIP net.IP) bool {
	route, hasRoute := (**c.routes).Load(srcIP.String())
	if srcIP != nil && !srcIP.Equal(net.ParseIP(c.config.NodeIP)) && (cmd == CommandSync || (hasRoute && route == c.iface)) {
		return true
	}
	return false
}

func (c *Connection) handleMessage(msg Message) bool {
	slog.Info("arednlink: received message", "command", msg.Command, "source", msg.Source, "hops", msg.Hops, "payload", msg.Payload)
	if !c.validNextHop(msg.Command, msg.Source) {
		slog.Warn("arednlink: invalid next hop", "command", msg.Command, "source", msg.Source, "hops", msg.Hops)
		return false
	}

	switch msg.Command {
	case CommandVersion:
		// Payload should be the version string
		if string(msg.Payload) != arednlinkVersion {
			slog.Warn("arednlink: version mismatch", "peer", msg.Source, "expected", arednlinkVersion, "received", string(msg.Payload))
			err := c.conn.Close()
			if err != nil {
				slog.Error("arednlink: failed to close connection", "peer", msg.Source, "error", err)
			}
			return false
		}
	case CommandSync:
		// Payload is a list of IPs
		// Check that the length of the payload is a multiple of 4
		if len(msg.Payload)%4 != 0 {
			slog.Warn("arednlink: received invalid sync message", "peer", msg.Source)
		}

		for i := 0; i < len(msg.Payload); i += 4 {
			ip := net.IP(msg.Payload[i : i+4])
			if ip == nil {
				slog.Warn("arednlink: received invalid IP in sync message", "peer", msg.Source)
				continue
			}
			slog.Info("arednlink: received sync message", "peer", msg.Source, "ip", ip)
			hosts, ok := c.hosts.Load(ip.String())
			if !ok {
				slog.Warn("arednlink: received sync request for unknown ip", "peer", msg.Source, "ip", ip)
				continue
			}
			services, ok := c.services.Load(ip.String())
			if !ok {
				slog.Warn("arednlink: received sync request for unknown ip", "peer", msg.Source, "ip", ip)
				continue
			}

			c.sendMessage(Message{
				Length:  8 + uint16(len(hosts)),
				Command: CommandUpdateHosts,
				Source:  ip,
				Hops:    HopDefault,
				Payload: []byte(hosts),
			})

			c.sendMessage(Message{
				Length:  8 + uint16(len(services)),
				Command: CommandUpdateServices,
				Source:  ip,
				Hops:    HopDefault,
				Payload: []byte(services),
			})
		}
		return false
	case CommandKeepAlive:
		return false
	case CommandUpdateHosts:
		if msg.Source.Equal(net.ParseIP(c.config.NodeIP)) {
			// Ignore messages from ourselves
			return false
		}
		existing, loaded := c.hosts.LoadAndStore(msg.Source.String(), string(msg.Payload))
		if loaded {
			if existing == string(msg.Payload) {
				return false
			}
		}
		return true
	case CommandUpdateServices:
		if msg.Source.Equal(net.ParseIP(c.config.NodeIP)) {
			// Ignore messages from ourselves
			return false
		}
		existing, loaded := c.services.LoadAndStore(msg.Source.String(), string(msg.Payload))
		if loaded {
			if existing == string(msg.Payload) {
				return false
			}
		}
		return true
	default:
		slog.Warn("arednlink: received unknown command", "command", msg.Command)
	}

	return false
}

func readMessageHeader(buf []byte) *Message {
	messageLength := int(buf[0])<<8 | int(buf[1])
	command := Command(buf[2])
	hopCount := buf[3]
	ipv4 := net.IP(buf[4:8])

	return &Message{
		Length:  uint16(messageLength),
		Command: command,
		Source:  ipv4,
		Hops:    uint8(hopCount),
	}
}
