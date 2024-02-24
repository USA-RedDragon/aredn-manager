package runner

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func cancelAndWaitForExit(signal syscall.Signal, processResults chan error) error {
	newCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	select {
	case <-newCtx.Done():
		return fmt.Errorf("failed to send %s to process", signal.String())
	case <-processResults:
		return nil
	}
}

func Run(cmd *exec.Cmd) (chan error, error) {
	processResults := make(chan error)

	cmd.Cancel = func() error {
		err := cancelAndWaitForExit(syscall.SIGTERM, processResults)
		if err != nil {
			// SIGTERM didn't work, log it and try SIGKILL
			log.Printf("failed to send SIGTERM to process: %v\n", err)
			err = cancelAndWaitForExit(syscall.SIGKILL, processResults)
			if err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
		}
		return nil
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		processResults <- cmd.Wait()
	}()

	return processResults, nil
}
