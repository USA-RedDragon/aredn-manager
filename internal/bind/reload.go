package bind

import (
	"os"
	"strconv"
	"syscall"
)

// This file will run bind
const (
	pidFile = "/var/run/named/named.pid"
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

func Restart() error {
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
	return syscall.Kill(int(pid), syscall.SIGTERM)
	// SNEAKY: s6 will restart the process, so we don't need to do anything else
}

func IsRunning() bool {
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pidStr := string(pidBytes)
	// Trim newline
	pidStr = pidStr[:len(pidStr)-1]
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		return false
	}
	err = syscall.Kill(int(pid), 0)
	return err == nil
}
