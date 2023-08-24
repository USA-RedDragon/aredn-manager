package events

type EventType string

const (
	EventTypeTunnelDisconnection EventType = "tunnel_disconnection"
	EventTypeTunnelConnection    EventType = "tunnel_connection"
	EventTypeTunnelStats         EventType = "tunnel_stats"
	EventTypeTotalBandwidth      EventType = "total_bandwidth"
	EventTypeTotalTraffic        EventType = "total_traffic"
)

type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type EventBus struct {
	eventQueue chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		eventQueue: make(chan Event, 100),
	}
}

func (eb *EventBus) GetChannel() chan Event {
	return eb.eventQueue
}

func (eb *EventBus) Close() {
	close(eb.eventQueue)
}
