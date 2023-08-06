package utils

import (
	"os"
	"strconv"
	"syscall"
)

func ProcessIsRunning(pid int) bool {
	// Check if the PID is running
	if process, err := os.FindProcess(int(pid)); err == nil {
		// Workaround since FindProcess doesn't actually check if the process is running
		err := process.Signal(syscall.Signal(0))
		if err == nil {
			return false
		}
		return true
	}
	return false
}

func PIDFileIsRunning(pidFile string) bool {
	// Check if the PID file exists
	if _, err := os.Stat(pidFile); err == nil {
		// Read the PID file
		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			return false
		}
		pidStr := string(pidBytes)
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			return false
		}

		if ProcessIsRunning(int(pid)) {
			return true
		}
	}
	return false
}
