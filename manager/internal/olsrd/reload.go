package olsrd

import (
	"os"
	"strconv"
	"syscall"
)

// This file will run olsrd
const (
	pidFile = "/tmp/olsrd.pid"
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
