package arednlink

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	arednlinkVersion = "0.0.1"
)

type Server struct {
	listener    net.Listener
	quit        chan interface{}
	wg          sync.WaitGroup
	connections []*Connection
	config      *config.Config
	Hosts       *xsync.MapOf[string, string]
	Services    *xsync.MapOf[string, string]
}

func NewServer(config *config.Config) (*Server, error) {
	// Set SO_REUSEADDR
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
		},
	}
	listener, err := lc.Listen(context.TODO(), "tcp6", "[::]:9623")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on [::]:9623: %w", err)
	}
	slog.Info("arednlink: listening on [::]:9623")

	s := &Server{
		listener:    listener,
		quit:        make(chan interface{}),
		wg:          sync.WaitGroup{},
		connections: make([]*Connection, 0),
		config:      config,
		Hosts:       xsync.NewMapOf[string, string](),
		Services:    xsync.NewMapOf[string, string](),
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
	slog.Debug("arednlink: waiting up to 5 seconds for all connections to close")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		s.wg.Wait()
		cancel()
	}()
	<-ctx.Done()
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
			newConn := NewConnection(s.config, conn, s)
			s.connections = append(s.connections, newConn)
			newConn.Start()
			s.wg.Done()
		}()
	}
}

func (s *Server) SendAll(m Message) {
	for _, conn := range s.connections {
		err := conn.SendMessage(m)
		if err != nil {
			slog.Error("failed to send message", "error", err)
		}
	}
}
