package pollers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/arednlink"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/puzpuzpuz/xsync/v3"
)

type Poller interface {
	Poll() error
	PollRate() time.Duration
	Name() string
}

type Manager struct {
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	routes        ***xsync.MapOf[string, string]
	hosts         *xsync.MapOf[string, string]
	services      *xsync.MapOf[string, string]
	config        *config.Config
	broadcastChan chan arednlink.Message
}

func NewManager(
	ctx context.Context,
	config *config.Config,
	routes ***xsync.MapOf[string, string],
	hosts *xsync.MapOf[string, string],
	services *xsync.MapOf[string, string],
	broadcastChan chan arednlink.Message,
) *Manager {
	subctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:           subctx,
		cancel:        cancel,
		wg:            sync.WaitGroup{},
		routes:        routes,
		hosts:         hosts,
		services:      services,
		config:        config,
		broadcastChan: broadcastChan,
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
		NewRoutePoller(m.config, m.routes, m.hosts, m.services, m.broadcastChan),
		&NeighborhoodPoller{},
	}

	err := pollers[0].Poll()
	if err != nil {
		slog.Error("failed to poll routes", "error", err)
	}

	for _, poller := range pollers {
		m.wg.Add(1)
		go func(poller Poller) {
			defer m.wg.Done()
			tick := time.NewTicker(poller.PollRate())
			for {
				select {
				case <-tick.C:
					slog.Info("poller is running", "poller", poller.Name())
					err := poller.Poll()
					if err != nil {
						slog.Error("poller failed", "poller", poller.Name(), "error", err)
						continue
					}
					slog.Info("poller finished", "poller", poller.Name())
				case <-m.ctx.Done():
					tick.Stop()
					return
				}
			}
		}(poller)
	}
}
