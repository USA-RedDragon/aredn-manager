package pollers

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Poller interface {
	Poll() error
	PollRate() time.Duration
}

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewManager(ctx context.Context) *Manager {
	subctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:    subctx,
		cancel: cancel,
	}
}

func (m *Manager) Start() {
	m.wg.Add(1)
	go func() {
		m.run()
		m.wg.Done()
	}()
}

func (m *Manager) Stop() {
	slog.Debug("stopping pollers")
	m.cancel()
	slog.Debug("waiting for pollers to stop")
	m.wg.Wait()
	slog.Debug("pollers stopped")
}

func (m *Manager) run() {
	pollers := []Poller{
		&RoutePoller{},
		&IdlePoller{},
		&NeighborhoodPoller{},
	}

	for _, poller := range pollers {
		m.wg.Add(1)
		go func(poller Poller) {
			defer m.wg.Done()
			tick := time.NewTicker(poller.PollRate())
			select {
			case <-tick.C:
				poller.Poll()
			case <-m.ctx.Done():
				tick.Stop()
				return
			}
		}(poller)
	}
}
