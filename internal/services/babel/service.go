package babel

import (
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
)

const (
	pidFile = "/tmp/babeld.pid"
)

type Service struct {
	config *config.Config
}

func NewService(config *config.Config) *Service {
	return &Service{
		config: config,
	}
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() error {
	return nil
}

func (s *Service) Reload() error {
	pid, err := utils.PIDFromPIDFile(pidFile)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGHUP)
}

func (s *Service) IsRunning() bool {
	return utils.PIDFileIsRunning(pidFile)
}

func (s *Service) IsEnabled() bool {
	return s.config.EnableBabel
}
