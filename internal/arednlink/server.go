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
	listener, err := net.Listen("tcp6", ":9623")
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
	slog.Info("arednlink: stopping incoming connections to the server")
	close(s.quit)
	s.listener.Close()
	slog.Info("arednlink: waiting for all connections to close")
	s.wg.Wait()
	slog.Info("arednlink: all connections closed")
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

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			slog.Error("arednlink: failed to read from connection", "error", err)
			return
		}
		if n == 0 {
			return
		}
		slog.Info("arednlink: received", "remote", conn.RemoteAddr(), "data", string(buf[:n]))
	}
}
