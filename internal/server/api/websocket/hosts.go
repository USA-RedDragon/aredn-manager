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

func (c *HostsWebsocket) OnMessage(_ context.Context, _ *http.Request, _ websocket.Writer, _ sessions.Session, _ []byte, _ int) {
}

func (c *HostsWebsocket) OnConnect(ctx context.Context, _ *http.Request, w websocket.Writer, _ sessions.Session) {
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
					Data: msg.Data,
				})
			}
		}
	}()
}

func (c *HostsWebsocket) OnDisconnect(_ context.Context, _ *http.Request, _ sessions.Session) {
	c.cancel()
}
