package vtun

import (
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"gorm.io/gorm"
)

// This file will run vtund
const (
	pidFile = "/usr/var/run/vtund.pid"
)

func Reload() error {
	pid, err := utils.PIDFromPIDFile(pidFile)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGHUP)
}

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

func IsRunning() bool {
	return utils.PIDFileIsRunning(pidFile)
}
