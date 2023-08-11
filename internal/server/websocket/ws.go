package websocket

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const bufferSize = 1024

type Websocket interface {
	OnMessage(ctx context.Context, r *http.Request, w Writer, session sessions.Session, msg []byte, t int)
	OnConnect(ctx context.Context, r *http.Request, w Writer, session sessions.Session)
	OnDisconnect(ctx context.Context, r *http.Request, session sessions.Session)
}

type WSHandler struct {
	wsUpgrader websocket.Upgrader
	handler    Websocket
	conn       *websocket.Conn
}

func CreateHandler(ws Websocket, config *config.Config) func(*gin.Context) {
	handler := &WSHandler{
		wsUpgrader: websocket.Upgrader{
			HandshakeTimeout: 0,
			ReadBufferSize:   bufferSize,
			WriteBufferSize:  bufferSize,
			WriteBufferPool:  nil,
			Subprotocols:     []string{},
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			},
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				for _, host := range config.CORSHosts {
					if strings.HasSuffix(host, ":443") && strings.HasPrefix(origin, "https://") {
						host = strings.TrimSuffix(host, ":443")
					}
					if strings.HasSuffix(host, ":80") && strings.HasPrefix(origin, "http://") {
						host = strings.TrimSuffix(host, ":80")
					}
					if strings.Contains(origin, host) {
						return true
					}
				}
				return false
			},
			EnableCompression: true,
		},
		handler: ws,
	}

	return func(c *gin.Context) {
		session := sessions.Default(c)
		conn, err := handler.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("Failed to set websocket upgrade: %v\n", err)
			return
		}
		handler.conn = conn

		defer func() {
			handler.handler.OnDisconnect(c, c.Request, session)
			err := handler.conn.Close()
			if err != nil {
				fmt.Printf("Failed to close websocket: %v\n", err)
			}
		}()
		handler.handle(c.Request.Context(), session, c.Request)
	}
}

func (h *WSHandler) handle(c context.Context, s sessions.Session, r *http.Request) {
	writer := wsWriter{
		writer: make(chan Message, bufferSize),
		error:  make(chan string),
	}
	h.handler.OnConnect(c, r, writer, s)

	go func() {
		for {
			t, msg, err := h.conn.ReadMessage()
			if err != nil {
				writer.Error("read failed")
				break
			}
			switch {
			case t == websocket.PingMessage:
				writer.WriteMessage(Message{
					Type: websocket.PongMessage,
				})
			case strings.EqualFold(string(msg), "ping"):
				writer.WriteMessage(Message{
					Type: websocket.TextMessage,
					Data: []byte("PONG"),
				})
			default:
				h.handler.OnMessage(c, r, writer, s, msg, t)
			}
		}
	}()

	for {
		select {
		case <-c.Done():
			return
		case <-writer.error:
			return
		case msg := <-writer.writer:
			err := h.conn.WriteMessage(msg.Type, msg.Data)
			if err != nil {
				return
			}
		}
	}
}
