package arednlink

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

type Server struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
}

func NewServer() (*Server, error) {
	listener, err := net.Listen("tcp6", "[::]:9623")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on [::]:9623: %w", err)
	}
	slog.Info("arednlink: listening on [::]:9623")

	s := &Server{
		listener: listener,
		quit:     make(chan interface{}),
		wg:       sync.WaitGroup{},
	}

	s.wg.Add(1)
	go func() {
		s.run()
		s.wg.Done()
	}()
	return s, nil
}

func (s *Server) Stop() {
	slog.Debug("arednlink: stopping incoming connections to the server")
	close(s.quit)
	s.listener.Close()
	slog.Debug("arednlink: waiting for all connections to close")
	s.wg.Wait()
	slog.Debug("arednlink: all connections closed")
}

func (s *Server) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				slog.Error("failed to accept connection", "error", err)
				continue
			}
		}
		s.wg.Add(1)
		go func() {
			s.handleConnection(conn)
			s.wg.Done()
		}()
	}
}

type Command byte

func (c Command) String() string {
	return string(c)
}

const (
	CommandVersion        Command = 'V'
	CommandUpdateNames    Command = 'N'
	CommandUpdateServices Command = 'U'
	CommandSync           Command = 'S'
)

type Message struct {
	Length  uint16
	Command Command
	Payload []byte
	Source  net.IP
	Hops    uint8
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 2048)
	var currentMessage *Message

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			slog.Error("arednlink: failed to read from connection", "error", err)
			return
		}
		if n == 0 {
			return
		}

		slog.Debug("arednlink: received", "remote", conn.RemoteAddr(), "data", string(buf[:n]))

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
					currentMessage.Payload = buf[8:]
					s.handleMessage(*currentMessage)
					currentMessage = nil
					n = 0
					continue
				}

				// Check if the current buffer is larger than the message
				// if so, we need to start a loop of reading the message in chunks
				if n > int(currentMessage.Length) {
					msgLen := int(currentMessage.Length)
					currentMessage.Payload = buf[8:currentMessage.Length]
					s.handleMessage(*currentMessage)
					currentMessage = nil

					buf = buf[msgLen:]
					n -= int(msgLen)
					continue
				}

				// If we're here, the message is larger than the buffer
				// so we need to read the rest of the message in the next iteration
				currentMessage.Payload = buf[8:]
				n = 0
			} else {
				// Current message is already being parsed, so we need to append the new data to the payload
				bytesStillWanted := int(currentMessage.Length) - len(currentMessage.Payload) + 8
				if n == bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:n]...)
					s.handleMessage(*currentMessage)
					currentMessage = nil
					n = 0
					continue
				}

				if n > bytesStillWanted {
					currentMessage.Payload = append(currentMessage.Payload, buf[:bytesStillWanted]...)
					s.handleMessage(*currentMessage)
					currentMessage = nil

					buf = buf[bytesStillWanted:]
					n -= bytesStillWanted
					continue
				}

				currentMessage.Payload = append(currentMessage.Payload, buf[:n]...)
				n = 0
			}
		}
	}
}

func (s *Server) handleMessage(m Message) {
	slog.Info("arednlink: received message", "command", m.Command, "source", m.Source, "hops", m.Hops, "payload", string(m.Payload))
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
