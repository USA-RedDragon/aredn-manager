package wireguard

import (
	"context"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

const defTimeout = 10 * time.Second

type Manager struct {
	db             *gorm.DB
	ctx            context.Context
	peerAddChan    chan models.Tunnel
	peerRemoveChan chan models.Tunnel
	activePeerIDs  []uint
}

func NewManager(ctx context.Context, db *gorm.DB) *Manager {
	return &Manager{
		db:             db,
		ctx:            ctx,
		peerAddChan:    make(chan models.Tunnel),
		peerRemoveChan: make(chan models.Tunnel),
		activePeerIDs:  []uint{},
	}
}

func (m *Manager) Run() {
	go m.run()
}

func (m *Manager) run() {
	for {
		select {
		case peer := <-m.peerAddChan:
			go m.addPeer(peer)
		case peer := <-m.peerRemoveChan:
			go m.removePeer(peer)
		case <-m.ctx.Done():
			close(m.peerAddChan)
			close(m.peerRemoveChan)
			return
		}
	}
}

func (m *Manager) addPeer(peer models.Tunnel) {
	// Create a new wireguard interface listening on the port from the peer tunnel
	// If the peer is a client, then the password is the public key of the client
	// If the peer is a server, then the password is the private key of the server
	// TODO: remove peer
}

func (m *Manager) removePeer(peer models.Tunnel) {
	// TODO: remove peer
}

func (m *Manager) peerExists(peer models.Tunnel) bool {
	for _, id := range m.activePeerIDs {
		if id == peer.ID {
			return true
		}
	}
	return false
}

func (m *Manager) WaitForPeerAddition(ctx context.Context, peer models.Tunnel) chan error {
	resp := make(chan error)
	go func() {
		retChan := make(chan error)
		go func() {
			for {
				if m.peerExists(peer) {
					retChan <- nil
					return
				}
				if m.ctx.Err() != nil {
					retChan <- m.ctx.Err()
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-ctx.Done():
			resp <- ctx.Err()
			return
		case <-m.ctx.Done():
			resp <- m.ctx.Err()
			return
		case err := <-retChan:
			resp <- err
			return
		}
	}()
	return resp
}

func (m *Manager) WaitForPeerRemoval(ctx context.Context, peer models.Tunnel) chan error {
	resp := make(chan error)
	go func() {
		retChan := make(chan error)
		go func() {
			for {
				if !m.peerExists(peer) {
					retChan <- nil
					return
				}
				if m.ctx.Err() != nil {
					retChan <- m.ctx.Err()
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
		select {
		case <-ctx.Done():
			resp <- ctx.Err()
			return
		case <-m.ctx.Done():
			resp <- m.ctx.Err()
			return
		case err := <-retChan:
			resp <- err
			return
		}
	}()
	return resp
}

func (m *Manager) AddPeer(peer models.Tunnel) error {
	m.peerAddChan <- peer
	ctx, cancel := context.WithTimeout(m.ctx, defTimeout)
	defer cancel()

	resp := m.WaitForPeerAddition(ctx, peer)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ctx.Done():
		return m.ctx.Err()
	case err := <-resp:
		return err
	}
}

func (m *Manager) RemovePeer(peer models.Tunnel) error {
	m.peerRemoveChan <- peer
	ctx, cancel := context.WithTimeout(m.ctx, defTimeout)
	defer cancel()

	resp := m.WaitForPeerRemoval(ctx, peer)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.ctx.Done():
		return m.ctx.Err()
	case err := <-resp:
		return err
	}
}
