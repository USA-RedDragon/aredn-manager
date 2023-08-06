package dnsmasq

import (
	"os"
	"strconv"
	"syscall"
)

// This file will run dnsmasq
const (
	pidFile = "/var/run/dnsmasq.pid"
)

func Reload() error {
	// Read the PID file
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pidStr := string(pidBytes)
	// Trim newline
	pidStr = pidStr[:len(pidStr)-1]
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGHUP)
}
