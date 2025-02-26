package pollers

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/vishvananda/netlink"
)

const (
	routingTable = 20
)

type RoutePoller struct {
	routes   **xsync.MapOf[string, string]
	hosts    *xsync.MapOf[string, string]
	services *xsync.MapOf[string, string]
}

type Route struct {
	Destination   net.IPNet
	Gateway       net.IP
	OutboundIface string
}

func NewRoutePoller(
	routes **xsync.MapOf[string, string],
	hosts *xsync.MapOf[string, string],
	services *xsync.MapOf[string, string],
) *RoutePoller {
	return &RoutePoller{
		routes:   routes,
		hosts:    hosts,
		services: services,
	}
}

func (p *RoutePoller) Poll() error {
	slog.Info("Route poller is running")
	netRoutes, err := netlink.RouteListFiltered(
		netlink.FAMILY_V4,
		&netlink.Route{
			Table: routingTable,
		},
		netlink.RT_FILTER_TABLE,
	)
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
			linkAttrs := link.Attrs()
			if linkAttrs == nil {
				slog.Error("failed to get link attributes", "error", err)
				continue
			}
			routes = append(routes, Route{
				Destination:   *route.Dst,
				Gateway:       route.Gw,
				OutboundIface: linkAttrs.Name,
			})
		}
	}

	oldRoutes := *p.routes
	newRoutes := xsync.NewMapOf[string, net.IP]()
	hostRoutes := xsync.NewMapOf[string, string]()
	for _, route := range routes {
		hostRoutes.Store(route.Destination.IP.String(), route.OutboundIface)
		link, ok := (*oldRoutes).Load(route.Destination.IP.String())
		if ok {
			oldRoutes.Delete(route.Destination.IP.String())
			if link != route.OutboundIface {
				newRoutes.Store(route.OutboundIface, route.Destination.IP)
			}
		}
	}
	p.routes = &hostRoutes

	oldRoutes.Range(func(ip string, _ string) bool {
		p.hosts.Delete(ip)
		p.services.Delete(ip)
		return true
	})

	newRoutes.Range(func(iface string, ip net.IP) bool {
		// TODO: send sync message to neighbors based on interface
		slog.Info("Route poller: want to request sync for", "ip", ip, "iface", iface)
		return true
	})

	return nil
}

func (p *RoutePoller) PollRate() time.Duration {
	return 30 * time.Second
}
