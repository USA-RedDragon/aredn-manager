package babel

import (
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
)

const (
	pidFile = "/tmp/babeld.pid"
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
