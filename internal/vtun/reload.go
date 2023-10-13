package vtun

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

// This file will run vtund
const (
	pidFile = "/var/run/vtund.pid"
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

func ReloadAllClients(db *gorm.DB) error {
	tunnels, err := models.ListClientTunnels(db)
	if err != nil {
		return err
	}

	for _, tunnel := range tunnels {
		err = ReloadClient(tunnel.Hostname, tunnel.IP)
		if err != nil {
			fmt.Printf("Error reloading vtun client %s: %v\n", tunnel.Hostname, err)
		}
	}
	return nil
}

func ReloadClient(hostname string, ip string) error {
	// Read the PID file
	pidBytes, err := os.ReadFile(fmt.Sprintf("/usr/var/run/vtundclient-%s-%s.pid", strings.ReplaceAll(hostname, ":", "-"), strings.ReplaceAll(ip, ".", "-")))
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

func IsRunningClient(hostname string, ip string) bool {
	pidBytes, err := os.ReadFile(fmt.Sprintf("/usr/var/run/vtundclient-%s-%s.pid", strings.ReplaceAll(hostname, ":", "-"), strings.ReplaceAll(ip, ".", "-")))
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
