package pollers

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/arednlink"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/vishvananda/netlink"
)

const (
	routingTable = 20
)

type RoutePoller struct {
	routes        **xsync.MapOf[string, string]
	hosts         *xsync.MapOf[string, string]
	services      *xsync.MapOf[string, string]
	config        *config.Config
	broadcastChan chan arednlink.Message
}

type Route struct {
	Destination   net.IPNet
	Gateway       net.IP
	OutboundIface string
}

func NewRoutePoller(
	config *config.Config,
	routes **xsync.MapOf[string, string],
	hosts *xsync.MapOf[string, string],
	services *xsync.MapOf[string, string],
	broadcastChan chan arednlink.Message,
) *RoutePoller {
	return &RoutePoller{
		routes:        routes,
		hosts:         hosts,
		services:      services,
		config:        config,
		broadcastChan: broadcastChan,
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
	newRoutes := xsync.NewMapOf[string, []net.IP]()
	hostRoutes := xsync.NewMapOf[string, string]()
	for _, route := range routes {
		hostRoutes.Store(route.Destination.IP.String(), route.OutboundIface)
		link, ok := oldRoutes.Load(route.Destination.IP.String())
		if ok {
			oldRoutes.Delete(route.Destination.IP.String())
			if link != route.OutboundIface {
				existingIPs, ok := newRoutes.Load(route.OutboundIface)
				if !ok {
					existingIPs = []net.IP{}
				}
				existingIPs = append(existingIPs, route.Destination.IP)
				newRoutes.Store(route.OutboundIface, existingIPs)
			}
		} else {
			existingIPs, ok := newRoutes.Load(route.OutboundIface)
			if !ok {
				existingIPs = []net.IP{}
			}
			existingIPs = append(existingIPs, route.Destination.IP)
			newRoutes.Store(route.OutboundIface, existingIPs)
		}

		// If we have a route and we don't have a host entry, add it
		// to newRoutes
		_, ok = p.hosts.Load(route.Destination.IP.String())
		if !ok {
			existingIPs, ok := newRoutes.Load(route.OutboundIface)
			if !ok {
				existingIPs = []net.IP{}
			}
			existingIPs = append(existingIPs, route.Destination.IP)
			newRoutes.Store(route.OutboundIface, existingIPs)
		}
	}
	p.routes = &hostRoutes

	oldRoutes.Range(func(ip string, _ string) bool {
		p.hosts.Delete(ip)
		p.services.Delete(ip)
		return true
	})

	newRoutes.Range(func(iface string, ips []net.IP) bool {
		slog.Info("Route poller: want to request sync for", "ips", ips, "iface", iface)
		payload := make([]byte, 0)
		for _, ip := range ips {
			payload = append(payload, ip.To4()...)
		}
		msg := arednlink.Message{
			Command:   arednlink.CommandSync,
			Source:    net.ParseIP(p.config.NodeIP),
			Hops:      0,
			Payload:   payload,
			Length:    8 + uint16(len(payload)),
			DestIface: iface,
		}
		p.broadcastChan <- msg
		return true
	})

	return nil
}

func (p *RoutePoller) PollRate() time.Duration {
	return 30 * time.Second
}
