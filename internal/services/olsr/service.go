package olsr

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/runner"
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
	var err error
	s.processResults, err = runner.Run(s.olsrCmd)
	if err != nil {
		return fmt.Errorf("olsrd failed to start: %w", err)
	}
	fmt.Println("OLSR started")

	select {
	case err := <-s.processResults:
		var ret error
		if err != nil {
			ret = fmt.Errorf("OLSR process exited with error: %w, restarting it", err)
		}
		err = s.olsrCmd.Process.Signal(os.Signal(syscall.SIGKILL))
		if err != nil {
			log.Printf("failed to kill process: %v\n", err)
		}
		return ret
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
