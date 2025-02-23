package pollers

import (
	"log/slog"
	"time"
)

type IdlePoller struct {
}

func (p *IdlePoller) Poll() error {
	slog.Info("Idle poller is running")
	return nil
}

func (p *IdlePoller) PollRate() time.Duration {
	return 1 * time.Hour
}
