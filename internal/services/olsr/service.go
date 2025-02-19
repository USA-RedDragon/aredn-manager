package olsr

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
)

type Service struct {
	config         *config.Config
	olsrCmd        *exec.Cmd
	processResults chan error
}

func NewService(config *config.Config) *Service {
	return &Service{
		config:  config,
		olsrCmd: exec.Command("olsrd", "-f", "/etc/olsrd/olsrd.conf", "-nofork"),
	}
}

func (s *Service) Start() error {
	for {
		if s.olsrCmd == nil {
			log.Println("olsr command is nil")
			return nil
		}
		err := s.olsrCmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start process: %w", err)
		}
	}
}

func (s *Service) Stop() error {
	if s.olsrCmd != nil && s.olsrCmd.Process != nil {
		err := s.olsrCmd.Process.Signal(os.Signal(syscall.SIGTERM))
		if err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	defer close(s.processResults)
	return s.olsrCmd.Wait()
}

func (s *Service) Reload() error {
	return s.olsrCmd.Process.Signal(os.Signal(syscall.SIGHUP))
}

func (s *Service) IsRunning() bool {
	return s.olsrCmd != nil && s.olsrCmd.Process != nil
}

func (s *Service) IsEnabled() bool {
	return true
}
