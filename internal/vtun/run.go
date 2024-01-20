package vtun

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
	vtunCmd *exec.Cmd
)

func Run(ctx context.Context) error {
	vtunCmd = exec.CommandContext(ctx, "vtund", "-s", "-f", "/etc/vtundsrv.conf", "-n")
	processResults, err := runner.Run(ctx, vtunCmd)
	defer close(processResults)
	if err != nil {
		return err
	}
	fmt.Println("VTun started")

	select {
	case err := <-processResults:
		if err != nil {
			fmt.Printf("VTun process exited with error: %v, restarting it\n", err)
		} else {
			fmt.Println("VTun process exited, restarting it")
		}
		return Run(ctx)
	case <-ctx.Done():
		fmt.Println("Context cancelled")
	}
	fmt.Println("Waiting for VTun process to exit")
	err = <-processResults
	if err != nil {
		return fmt.Errorf("vtun process exited with error: %v", err)
	}

	return nil
}

func IsRunning() bool {
	return vtunCmd != nil && vtunCmd.Process != nil && vtunCmd.ProcessState != nil && !vtunCmd.ProcessState.Exited()
}

func Reload() error {
	return vtunCmd.Process.Signal(os.Signal(syscall.SIGHUP))
}
