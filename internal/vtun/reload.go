package vtun

import (
	"os"
	"strconv"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

// This file will run vtund
const (
	pidFile = "/usr/var/run/vtund.pid"
)

func Reload() error {
	// Read the PID file
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pidStr := string(pidBytes)
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGHUP)
}

func ReloadAllClients(db *gorm.DB, watcher *VTunClientWatcher) error {
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
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pidStr := string(pidBytes)
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		return false
	}
	err = syscall.Kill(int(pid), 0)
	return err == nil
}
