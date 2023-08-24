package events

type EventType uint

const (
	EventTypeTunnelDisconnection EventType = iota
	EventTypeTunnelConnection
	EventTypeTunnelBandwidth
	EventTypeTunnelSessionTraffic
	EventTypeTunnelTotalTraffic
	EventTypeTotalBandwidth
	EventTypeTotalTraffic
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
