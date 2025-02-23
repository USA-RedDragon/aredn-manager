package pollers

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

const (
	routingTable = 20
)

type RoutePoller struct {
}

type Route struct {
	Destination   *net.IPNet
	Gateway       net.IP
	OutboundIface *netlink.Link
}

func (p *RoutePoller) Poll() error {
	slog.Info("Route poller is running")
	netRoutes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	routes := make([]Route, 0)
	for _, route := range netRoutes {
		if route.Table == routingTable && strings.HasSuffix(route.Dst.String(), "/32") {
			link, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil {
				slog.Error("failed to get link by index", "error", err)
				continue
			}
			slog.Info("found route", "dst", route.Dst.String(), "gw", route.Gw.String(), "route", link.Attrs().Name)
			routes = append(routes, Route{
				Destination:   route.Dst,
				Gateway:       route.Gw,
				OutboundIface: &link,
			})
		}
	}

	return nil
}

func (p *RoutePoller) PollRate() time.Duration {
	return 30 * time.Second
}
