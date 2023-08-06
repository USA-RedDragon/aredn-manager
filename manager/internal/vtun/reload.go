package vtun

import (
	"os"
	"strconv"
	"syscall"
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
