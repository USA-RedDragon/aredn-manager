package pollers

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/services/babel"
)

type Interface struct {
	babel.Interface
	Best net.IP
}

type NeighborhoodPoller struct {
	interfaces map[string]Interface
}

func (p *NeighborhoodPoller) Poll() error {
	slog.Info("Neighborhood poller is running")
	babelClient, err := babel.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create babel client: %w", err)
	}

	ifaces, err := babelClient.GetInterfaces()
	oldInterfaces := p.interfaces
	p.interfaces = make(map[string]Interface)
	for _, iface := range ifaces {
		name := iface.Name
		if oldIface, ok := oldInterfaces[name]; ok {
			oldIface.Best = iface.IPv6
			p.interfaces[name] = oldIface
			delete(oldInterfaces, name)
		} else {
			p.interfaces[name] = Interface{
				Interface: iface,
				Best:      iface.IPv6,
			}
		}
	}

	for _, iface := range p.interfaces {
		if iface.Best.Equal(iface.Interface.IPv6) {
			// new best
			// CLOSECON
		} else {
			// NEWADDRESS
		}
	}

	return nil
}

func (p *NeighborhoodPoller) PollRate() time.Duration {
	return 60 * time.Second
}

func (p *NeighborhoodPoller) Name() string {
	return "Neighborhood"
}
