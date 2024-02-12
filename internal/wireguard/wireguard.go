package wireguard

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/phayes/freeport"
	"github.com/vishvananda/netlink"
	"golang.org/x/sync/errgroup"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gorm.io/gorm"
)

const defTimeout = 10 * time.Second

type Manager struct {
	db                    *gorm.DB
	peerAddChan           chan models.Tunnel
	peerAddConfirmChan    chan models.Tunnel
	peerRemoveChan        chan models.Tunnel
	peerRemoveConfirmChan chan models.Tunnel
	shutdownChan          chan struct{}
	shutdownConfirmChan   chan struct{}
	activePeers           sync.Map
	wgClient              *wgctrl.Client
}

func NewManager(db *gorm.DB) (*Manager, error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &Manager{
		db:                    db,
		peerAddChan:           make(chan models.Tunnel),
		peerAddConfirmChan:    make(chan models.Tunnel),
		peerRemoveChan:        make(chan models.Tunnel),
		peerRemoveConfirmChan: make(chan models.Tunnel),
		shutdownChan:          make(chan struct{}),
		shutdownConfirmChan:   make(chan struct{}),
		activePeers:           sync.Map{},
		wgClient:              wgClient,
	}, nil
}

func (m *Manager) Run() error {
	go m.run()
	return m.initializeTunnels()
}

func (m *Manager) removeAllPeers() error {
	errGroup := &errgroup.Group{}
	m.activePeers.Range(func(_, value interface{}) bool {
		peer, ok := value.(models.Tunnel)
		if !ok {
			return true
		}
		errGroup.Go(func() error {
			return m.RemovePeer(peer)
		})
		return true
	})

	return errGroup.Wait()
}

func (m *Manager) Stop() error {
	// Remove all peers, then stop the thread and close the channels
	err := m.removeAllPeers()
	if err != nil {
		return err
	}
	m.shutdownChan <- struct{}{}
	<-m.shutdownConfirmChan
	return nil
}

func (m *Manager) initializeTunnels() error {
	tunnels, err := models.ListWireguardTunnels(m.db)
	if err != nil {
		return err
	}
	errGroup := &errgroup.Group{}
	for _, tunnel := range tunnels {
		tunnel := tunnel
		errGroup.Go(func() error {
			return m.AddPeer(tunnel)
		})
	}

	return errGroup.Wait()
}

func (m *Manager) run() {
	for {
		select {
		case peer := <-m.peerAddChan:
			go m.addPeer(peer)
		case peer := <-m.peerRemoveChan:
			go m.removePeer(peer)
		case <-m.shutdownChan:
			close(m.peerAddChan)
			close(m.peerRemoveChan)
			close(m.peerAddConfirmChan)
			close(m.peerRemoveConfirmChan)
			m.shutdownConfirmChan <- struct{}{}
			return
		}
	}
}

func GenerateWireguardInterfaceName(peer models.Tunnel) string {
	if peer.WireguardServerKey != "" {
		return fmt.Sprintf("wgs%d", peer.ID)
	}
	return fmt.Sprintf("wgc%d", peer.ID)
}

type WG struct {
	netlink.LinkAttrs
}

func (wglink *WG) Attrs() *netlink.LinkAttrs {
	return &wglink.LinkAttrs
}

func (wglink *WG) Type() string {
	return "wireguard"
}

func (m *Manager) addPeer(peer models.Tunnel) {
	// Create a new wireguard interface listening on the port from the peer tunnel
	// If the peer is a client, then the password is the public key of the client
	// If the peer is a server, then the password is the private key of the server
	log.Println("adding peer", peer)

	iface := GenerateWireguardInterfaceName(peer)

	_, err := netlink.LinkByName(iface)
	if err == nil {
		log.Println("wireguard interface already exists", iface)
	} else {
		la := netlink.NewLinkAttrs()
		la.Name = iface
		la.MTU = 1420
		wgdev := &WG{LinkAttrs: la}
		err := netlink.LinkAdd(wgdev)
		if err != nil {
			log.Println("failed to add wireguard device", err)
			return
		}
		err = netlink.LinkSetUp(wgdev)
		if err != nil {
			log.Println("failed to bring up wireguard device", err)
			return
		}
		peerIP := net.ParseIP(peer.IP)
		if peer.WireguardServerKey == "" {
			// Add one to the peer IP for the client side
			peerIP = peerIP.To4()
			peerIP[3]++
			if peerIP[3] == 0 {
				peerIP[2]++
				peerIP[3] = 1
				if peerIP[2] == 0 {
					peerIP[1]++
					if peerIP[1] == 0 {
						peerIP[0]++
					}
				}
			}
		}
		netlink.AddrReplace(wgdev, &netlink.Addr{IPNet: &net.IPNet{IP: peerIP, Mask: net.CIDRMask(32, 32)}})
	}

	var privkey wgtypes.Key
	portInt := int(peer.WireguardPort)
	var peers []wgtypes.PeerConfig

	_, netip, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		log.Println("failed to parse 0.0.0.0/0", err)
		return
	}

	duration, err := time.ParseDuration("25s")
	if err != nil {
		log.Println("failed to parse duration", err)
		return
	}

	if peer.WireguardServerKey != "" {
		var err error
		privkey, err = wgtypes.ParseKey(peer.WireguardServerKey)
		if err != nil {
			log.Println("failed to parse server key", err)
			return
		}

		// tunnel.Password is our server pubkey + client privkey + client pubkey
		pubkeyPart := peer.Password[88:]
		clientPubkey, err := wgtypes.ParseKey(pubkeyPart)
		if err != nil {
			log.Println("failed to parse client pubkey", err)
			return
		}

		peers = []wgtypes.PeerConfig{
			{
				PublicKey:                   wgtypes.Key(clientPubkey),
				AllowedIPs:                  []net.IPNet{*netip},
				PersistentKeepaliveInterval: &duration,
			},
		}

	} else {
		var err error
		portInt = freeport.GetPort()

		// tunnel.Password is the server pubkey + client privkey + client pubkey
		serverPubkeyStr := peer.Password[:44]
		serverPubkey, err := wgtypes.ParseKey(serverPubkeyStr)
		if err != nil {
			log.Println("failed to parse server pubkey", err)
			return
		}
		clientPrivkeyStr := peer.Password[44:88]
		privkey, err = wgtypes.ParseKey(clientPrivkeyStr)
		if err != nil {
			log.Println("failed to parse client privkey", err)
			return
		}

		// Parse tunnel.Hostname as an address and port
		hostnameParts := strings.Split(peer.Hostname, ":")
		if len(hostnameParts) != 2 {
			log.Println("invalid hostname", peer.Hostname)
			return
		}

		port64, err := strconv.ParseInt(hostnameParts[1], 10, 32)
		if err != nil {
			log.Println("failed to parse port", err)
			return
		}
		port := int(port64)

		peers = []wgtypes.PeerConfig{
			{
				PublicKey:                   wgtypes.Key(serverPubkey),
				AllowedIPs:                  []net.IPNet{*netip},
				PersistentKeepaliveInterval: &duration,
				Endpoint:                    &net.UDPAddr{IP: net.ParseIP(hostnameParts[0]), Port: port},
			},
		}
	}

	err = m.wgClient.ConfigureDevice(iface, wgtypes.Config{
		PrivateKey:   &privkey,
		ListenPort:   &portInt,
		ReplacePeers: true,
		Peers:        peers,
	})

	if err != nil {
		log.Println("failed to configure wireguard device", iface, ":", err)
		return
	}

	m.activePeers.Store(iface, peer)
	m.peerAddConfirmChan <- peer
}

func (m *Manager) removePeer(peer models.Tunnel) {
	log.Println("removing peer", peer)

	_, ok := m.activePeers.LoadAndDelete(GenerateWireguardInterfaceName(peer))
	if !ok {
		m.peerRemoveConfirmChan <- peer
		return
	}

	// TODO: remove peer

	m.peerRemoveConfirmChan <- peer
}

func (m *Manager) AddPeer(peer models.Tunnel) error {
	m.peerAddChan <- peer

	ctx, cancel := context.WithTimeout(context.Background(), defTimeout)
	defer cancel()

	return m.waitForPeerAddition(ctx, peer)
}

func (m *Manager) waitForPeerAddition(ctx context.Context, peer models.Tunnel) error {
	select {
	case <-m.shutdownChan:
		return fmt.Errorf("wireguard manager is shutting down")
	case <-ctx.Done():
		return ctx.Err()
	case addedPeer := <-m.peerAddConfirmChan:
		if addedPeer.ID != peer.ID {
			return m.waitForPeerAddition(ctx, peer)
		}
		return nil
	}
}

func (m *Manager) RemovePeer(peer models.Tunnel) error {
	m.peerRemoveChan <- peer

	ctx, cancel := context.WithTimeout(context.Background(), defTimeout)
	defer cancel()

	return m.waitForPeerRemoval(ctx, peer)
}

func (m *Manager) waitForPeerRemoval(ctx context.Context, peer models.Tunnel) error {
	select {
	case <-m.shutdownChan:
		return fmt.Errorf("wireguard manager is shutting down")
	case <-ctx.Done():
		return ctx.Err()
	case addedPeer := <-m.peerRemoveConfirmChan:
		if addedPeer.ID != peer.ID {
			return m.waitForPeerRemoval(ctx, peer)
		}
		return nil
	}
}
