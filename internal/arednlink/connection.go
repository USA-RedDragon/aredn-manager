package arednlink

import (
	"io"
	"log/slog"
	"net"
	"sync"
)

type Connection struct {
	conn      net.Conn
	server    *Server
	writeLock sync.Mutex
}

func NewConnection(conn net.Conn, server *Server) *Connection {
	return &Connection{
		conn:      conn,
		server:    server,
		writeLock: sync.Mutex{},
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) SendMessage(msg Message) error {
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

func (c *Connection) Start() {
	buf := make([]byte, 2048)
	var currentMessage *Message

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
						currentMessage.Hops--
						c.server.SendAll(*currentMessage)
					}
					currentMessage = nil
					n = 0
					continue
				}

				// Check if the current buffer is larger than the message
				if n > int(currentMessage.Length) {
					msgLen := int(currentMessage.Length)
					currentMessage.Payload = buf[8:msgLen]
					forward := c.handleMessage(*currentMessage)
					if forward {
						currentMessage.Hops--
						c.server.SendAll(*currentMessage)
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
				// Current message is already being parsed, so we need to append the new data to the payload
				bytesStillWanted := int(currentMessage.Length) - len(currentMessage.Payload) + 8
				if n == bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:bytesStillWanted]...)
					forward := c.handleMessage(*currentMessage)
					if forward {
						currentMessage.Hops--
						c.server.SendAll(*currentMessage)
					}
					currentMessage = nil
					n = 0
					continue
				}

				if n > bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:bytesStillWanted]...)
					forward := c.handleMessage(*currentMessage)
					if forward {
						currentMessage.Hops--
						c.server.SendAll(*currentMessage)
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

func (c *Connection) handleMessage(msg Message) bool {
	slog.Info("arednlink: received message", "length", msg.Length-8, "command", msg.Command, "source", msg.Source, "hops", msg.Hops, "payload", string(msg.Payload))

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
		return false
	case CommandKeepAlive:
		return false
	case CommandUpdateHosts:
		return true
	case CommandUpdateServices:
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
