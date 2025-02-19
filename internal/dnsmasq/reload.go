package dnsmasq

import (
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
)

// This file will run dnsmasq
const (
	pidFile = "/var/run/dnsmasq.pid"
)

func Reload() error {
	pid, err := utils.PIDFromPIDFile(pidFile)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGHUP)
}

func IsRunning() bool {
	return utils.PIDFileIsRunning(pidFile)
}
