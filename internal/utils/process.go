package utils

import (
	"os"
	"strconv"
	"syscall"
)

func ProcessIsRunning(pid int) bool {
	// Check if the PID is running
	if process, err := os.FindProcess(pid); err == nil {
		// Workaround since FindProcess doesn't actually check if the process is running
		err := process.Signal(syscall.Signal(0))
		return err != nil
	}
	return false
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
	return -1, nil
}

func PIDFileIsRunning(pidFile string) bool {
	if pid, err := PIDFromPIDFile(pidFile); err == nil {
		if ProcessIsRunning(int(pid)) {
			return true
		}
	}
	return false
}
