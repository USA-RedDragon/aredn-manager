package ifacewatcher

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/bandwidth"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/events"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"golang.zx2c4.com/wireguard/wgctrl"
	"gorm.io/gorm"
)

type _iface struct {
	net.Interface
	AssociatedTunnel *models.Tunnel
}

type Watcher struct {
	stopped                  bool
	db                       *gorm.DB
	interfaces               []_iface
	interfacesToMarkInactive []_iface
	Stats                    *bandwidth.StatCounterManager
	eventChannel             chan events.Event
	wgClient                 *wgctrl.Client
}

func NewWatcher(db *gorm.DB, events chan events.Event) (*Watcher, error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		stopped:      true,
		db:           db,
		Stats:        bandwidth.NewStatCounterManager(db, events),
		eventChannel: events,
		wgClient:     wgClient,
	}
	w.Stats.Start()
	return w, nil
}

func (w *Watcher) Watch() error {
	if w.stopped {
		w.stopped = false
		go func() {
			for !w.stopped {
				w.watch()
			}
		}()
	} else {
		return fmt.Errorf("watcher already running")
	}
	return nil
}

func netInterfaceContainsIface(s []net.Interface, e _iface) bool {
	for _, a := range s {
		if a.Name == e.Name && a.Index == e.Index && a.HardwareAddr.String() == e.HardwareAddr.String() {
			return true
		}
	}
	return false
}

func ifaceContainsNetInterface(s []_iface, e net.Interface) bool {
	for _, a := range s {
		if a.Name == e.Name && a.Index == e.Index && a.HardwareAddr.String() == e.HardwareAddr.String() {
			return true
		}
	}
	return false
}

func remove(s []_iface, e _iface) []_iface {
	for i, a := range s {
		if a.Name == e.Name && a.Index == e.Index && a.HardwareAddr.String() == e.HardwareAddr.String() {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (w *Watcher) wgInterfaceActive(iface _iface) bool {
	if iface.Name == "wg0" {
		return false
	}
	if !strings.HasPrefix(iface.Name, "wg") {
		return false
	}
	dev, err := w.wgClient.Device(iface.Name)
	if err != nil {
		return false
	}
	if len(dev.Peers) > 0 {
		for _, peer := range dev.Peers {
			// If the last handshake time is more than 3 minutes ago, consider the interface inactive
			if peer.LastHandshakeTime.IsZero() || time.Since(peer.LastHandshakeTime) > 180*time.Second {
				return false
			} else {
				return true
			}
		}
	}
	return false
}

func (w *Watcher) watch() {
	w.interfacesToMarkInactive = []_iface{}
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
	} else {
		// Loop through w.interfaces and check if any are present but missing from net.Interfaces()
		for _, iface := range w.interfaces {
			if strings.HasPrefix(iface.Name, "wg") && iface.Name != "wg0" && !w.wgInterfaceActive(iface) {
				fmt.Printf("Interface %s is no longer present\n", iface.Name)
				w.eventChannel <- events.Event{
					Type: events.EventTypeTunnelDisconnection,
					Data: apimodels.WebsocketTunnelDisconnect{
						ID: iface.AssociatedTunnel.ID,
					},
				}
				err = w.Stats.Remove(iface.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				w.interfaces = remove(w.interfaces, iface)
				w.interfacesToMarkInactive = append(w.interfacesToMarkInactive, iface)
			} else if strings.HasPrefix(iface.Name, "tun") && !netInterfaceContainsIface(interfaces, iface) {
				fmt.Printf("Interface %s is no longer present\n", iface.Name)
				w.eventChannel <- events.Event{
					Type: events.EventTypeTunnelDisconnection,
					Data: apimodels.WebsocketTunnelDisconnect{
						ID: iface.AssociatedTunnel.ID,
					},
				}
				err = w.Stats.Remove(iface.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				w.interfaces = remove(w.interfaces, iface)
				w.interfacesToMarkInactive = append(w.interfacesToMarkInactive, iface)
			}
		}

		// Loop through net.Interfaces() and check if any are missing from w.interfaces
		for _, iface := range interfaces {
			if strings.HasPrefix(iface.Name, "wg") && iface.Name != "wg0" && w.wgInterfaceActive(_iface{iface, nil}) {
				fmt.Printf("Interface %s is now present\n", iface.Name)
				tunnel := w.findTunnel(iface)
				if tunnel == nil {
					fmt.Printf("No tunnel found for interface %s\n", iface.Name)
					continue
				}
				err = w.Stats.Add(iface.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				w.interfaces = append(w.interfaces, _iface{
					iface,
					tunnel,
				})
			} else if strings.HasPrefix(iface.Name, "tun") && !ifaceContainsNetInterface(w.interfaces, iface) {
				fmt.Printf("Interface %s is now present\n", iface.Name)
				tunnel := w.findTunnel(iface)
				if tunnel == nil {
					fmt.Printf("No tunnel found for interface %s\n", iface.Name)
					continue
				}
				err = w.Stats.Add(iface.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				w.interfaces = append(w.interfaces, _iface{
					iface,
					tunnel,
				})
			}
		}
	}
	w.reconcileDB()
	time.Sleep(1 * time.Second)
}

func (w *Watcher) findTunnel(iface net.Interface) *models.Tunnel {
	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			fmt.Println(err)
			continue
		}
		ip = ip.To4()
		var tun models.Tunnel
		if strings.HasPrefix(iface.Name, "wg") && iface.Name != "wg0" {
			var err error
			if strings.HasPrefix(iface.Name, "wgs") {
				tun, err = models.FindTunnelByIP(w.db, ip)
				if err != nil {
					fmt.Println(err)
					continue
				}
			} else if strings.HasPrefix(iface.Name, "wgc") {
				ip[3] -= 1
				tun, err = models.FindTunnelByIP(w.db, ip)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		} else if strings.HasPrefix(iface.Name, "tun") {
			var err error
			ip[3] -= 2 // AREDN tunnel IPs are always the interface IP - 2 if a client
			tun, err = models.FindTunnelByIP(w.db, ip)
			if err != nil {
				ip[3] += 1 // AREDN tunnel IPs are always the interface IP - 1 if a server
				tun, err = models.FindTunnelByIP(w.db, ip)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		}
		return &tun
	}
	return nil
}

// reconcileDB will loop through w.interfaces and change the database to reflect the current state
func (w *Watcher) reconcileDB() {
	for _, iface := range w.interfacesToMarkInactive {
		if iface.AssociatedTunnel != nil {
			fmt.Printf("Marking tunnel %s as inactive\n", iface.AssociatedTunnel.Hostname)
			iface.AssociatedTunnel.Active = false
			iface.AssociatedTunnel.TunnelInterface = ""
			iface.AssociatedTunnel.RXBytesPerSec = 0
			iface.AssociatedTunnel.TXBytesPerSec = 0
			iface.AssociatedTunnel.TotalRXMB += float64(iface.AssociatedTunnel.RXBytes) / 1024 / 1024
			iface.AssociatedTunnel.TotalTXMB += float64(iface.AssociatedTunnel.TXBytes) / 1024 / 1024
			iface.AssociatedTunnel.RXBytes = 0
			iface.AssociatedTunnel.TXBytes = 0
			w.eventChannel <- events.Event{
				Type: events.EventTypeTunnelDisconnection,
				Data: apimodels.WebsocketTunnelDisconnect{
					ID: iface.AssociatedTunnel.ID,
				},
			}

			wsTunnel := apimodels.WebsocketTunnelStats{
				ID:               iface.AssociatedTunnel.ID,
				RXBytesPerSecond: iface.AssociatedTunnel.RXBytesPerSec,
				TXBytesPerSecond: iface.AssociatedTunnel.TXBytesPerSec,
				RXBytes:          iface.AssociatedTunnel.RXBytes,
				TXBytes:          iface.AssociatedTunnel.TXBytes,
				TotalRXMB:        iface.AssociatedTunnel.TotalRXMB,
				TotalTXMB:        iface.AssociatedTunnel.TotalTXMB,
			}
			w.eventChannel <- events.Event{
				Type: events.EventTypeTunnelStats,
				Data: wsTunnel,
			}
			w.db.Save(iface.AssociatedTunnel)
		}
	}

	for _, iface := range w.interfaces {
		if iface.AssociatedTunnel != nil {
			if !iface.AssociatedTunnel.Active {
				fmt.Printf("Marking tunnel %s as active\n", iface.AssociatedTunnel.Hostname)
				iface.AssociatedTunnel.Active = true
				iface.AssociatedTunnel.TunnelInterface = iface.Name
				iface.AssociatedTunnel.ConnectionTime = time.Now()

				w.eventChannel <- events.Event{
					Type: events.EventTypeTunnelConnection,
					Data: apimodels.WebsocketTunnelConnect{
						ID: iface.AssociatedTunnel.ID,
					},
				}
				w.db.Save(iface.AssociatedTunnel)
			}
		}
	}
}

func (w *Watcher) Stop() {
	w.stopped = true
	w.Stats.Stop()
}
