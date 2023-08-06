package websocket

import (
	"context"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/server/websocket"
	"github.com/gin-contrib/sessions"
	gorillaWebsocket "github.com/gorilla/websocket"
)

type HostsWebsocket struct {
	websocket.Websocket
	cancel context.CancelFunc
}

func CreateHostsWebsocket() *HostsWebsocket {
	return &HostsWebsocket{}
}

func (c *HostsWebsocket) OnMessage(ctx context.Context, r *http.Request, w websocket.Writer, _ sessions.Session, msg []byte, t int) {
}

func (c *HostsWebsocket) OnConnect(ctx context.Context, r *http.Request, w websocket.Writer, session sessions.Session) {
	newCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		channel := make(chan websocket.Message)
		for {
			select {
			case <-ctx.Done():
				return
			case <-newCtx.Done():
				return
			case msg := <-channel:
				w.WriteMessage(websocket.Message{
					Type: gorillaWebsocket.TextMessage,
					Data: []byte(msg.Data),
				})
			}
		}
	}()
}

func (c *HostsWebsocket) OnDisconnect(ctx context.Context, r *http.Request, _ sessions.Session) {
	c.cancel()
}
