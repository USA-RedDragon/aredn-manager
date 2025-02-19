package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"
)

func ProcessIsRunning(pid int) bool {
	// Check if the PID is running
	process, err := os.FindProcess(pid)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to find process %d: %s", pid, err))
		return false
	}
	// Workaround since FindProcess doesn't actually check if the process is running
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func PIDFromPIDFile(pidFile string) (int, error) {
	// Check if the PID file exists
	if _, err := os.Stat(pidFile); err == nil {
		// Read the PID file
		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			return -1, err
		}
		pidStr := string(pidBytes)
		if pidStr[len(pidStr)-1] == '\n' {
			pidStr = pidStr[:len(pidStr)-1]
		}
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			return -1, err
		}
		return int(pid), nil
	}
	return -1, os.ErrNotExist
}

func PIDFileIsRunning(pidFile string) bool {
	pid, err := PIDFromPIDFile(pidFile)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to get PID from %s: %s", pidFile, err))
		return false
	}
	return ProcessIsRunning(pid)
}
