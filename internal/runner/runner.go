package runner

import (
	"context"
	"os/exec"
)

func Run(ctx context.Context, cmd *exec.Cmd) (chan error, error) {
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	processResults := make(chan error)

	go func() {
		err := cmd.Wait()
		processResults <- err
	}()

	return processResults, nil
}
