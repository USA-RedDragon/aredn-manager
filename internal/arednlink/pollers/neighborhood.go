package pollers

import (
	"log/slog"
	"time"
)

type NeighborhoodPoller struct {
}

func (p *NeighborhoodPoller) Poll() error {
	slog.Info("Neighborhood poller is running")
	return nil
}

func (p *NeighborhoodPoller) PollRate() time.Duration {
	return 60 * time.Second
}
