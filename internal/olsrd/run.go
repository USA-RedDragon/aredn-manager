package olsrd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/runner"
)

//nolint:golint,gochecknoglobals
var (
	olsrCmd *exec.Cmd
)

func Run(ctx context.Context) error {
	olsrCmd = exec.CommandContext(ctx, "olsrd", "-f", "/etc/olsrd/olsrd.conf", "-nofork")
	processResults, err := runner.Run(ctx, olsrCmd)
	defer close(processResults)
	if err != nil {
		return err
	}
	fmt.Println("OLSR started")

	select {
	case err := <-processResults:
		if err != nil {
			fmt.Printf("OLSR process exited with error: %v, restarting it\n", err)
		} else {
			fmt.Println("OLSR process exited, restarting it")
		}
		return Run(ctx)
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	}
	fmt.Println("Waiting for OLSR process to exit")
	err = <-processResults
	if err != nil {
		return fmt.Errorf("olsrd process exited with error: %v", err)
	}

	return nil
}

func IsRunning() bool {
	return olsrCmd != nil && olsrCmd.Process != nil && olsrCmd.ProcessState != nil && !olsrCmd.ProcessState.Exited()
}

func Reload() error {
	return olsrCmd.Process.Signal(os.Signal(syscall.SIGHUP))
}
