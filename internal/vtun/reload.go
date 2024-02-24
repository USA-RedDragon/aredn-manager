package vtun

import (
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

func ReloadAllClients(db *gorm.DB, watcher *ClientWatcher) error {
	tunnels, err := models.ListClientTunnels(db)
	if err != nil {
		return err
	}

	for _, tunnel := range tunnels {
		watcher.ReloadTunnel(tunnel.ID)
	}
	return nil
}
