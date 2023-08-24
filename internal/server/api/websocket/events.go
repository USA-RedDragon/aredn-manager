package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/events"
	"github.com/USA-RedDragon/aredn-manager/internal/server/websocket"
	"github.com/gin-contrib/sessions"
	gorillaWebsocket "github.com/gorilla/websocket"
)

type EventsWebsocket struct {
	websocket.Websocket
	cancel        context.CancelFunc
	eventsChannel chan events.Event
}

func CreateEventsWebsocket(eventsChannel chan events.Event) *EventsWebsocket {
	return &EventsWebsocket{
		eventsChannel: eventsChannel,
	}
}

func (c *EventsWebsocket) OnMessage(_ context.Context, _ *http.Request, _ websocket.Writer, _ sessions.Session, _ []byte, _ int) {
}

func (c *EventsWebsocket) OnConnect(ctx context.Context, _ *http.Request, w websocket.Writer, _ sessions.Session) {
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
			case event := <-c.eventsChannel:
				eventDataJSON, err := json.Marshal(event)
				if err != nil {
					fmt.Println("Error marshalling event data:", err)
					continue
				}
				w.WriteMessage(websocket.Message{
					Type: gorillaWebsocket.TextMessage,
					Data: eventDataJSON,
				})
			case <-channel:
				// We don't actually want to receive messages from the client
				continue
			}
		}
	}()
}

func (c *EventsWebsocket) OnDisconnect(_ context.Context, _ *http.Request, _ sessions.Session) {
	c.cancel()
}
