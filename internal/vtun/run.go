package vtun

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/runner"
)

//nolint:golint,gochecknoglobals
var (
	vtunCmd *exec.Cmd
)

func Run(ctx context.Context) chan struct{} {
	stopChan := make(chan struct{})
	defer close(stopChan)
	go run(ctx, stopChan)
	return stopChan
}

func run(ctx context.Context, stopChan chan struct{}) {
	vtunCmd = exec.CommandContext(ctx, "vtund", "-s", "-f", "/etc/vtundsrv.conf", "-n")
	processResults, err := runner.Run(ctx, vtunCmd)
	defer close(processResults)
	if err != nil {
		fmt.Println("vtund failed to start:", err)
		return
	}
	fmt.Println("VTun started")

	select {
	case err := <-processResults:
		if err != nil {
			fmt.Printf("VTun process exited with error: %v, restarting it\n", err)
		} else {
			fmt.Println("VTun process exited, restarting it")
		}
		err = vtunCmd.Process.Signal(os.Signal(syscall.SIGKILL))
		if err != nil {
			log.Printf("failed to kill process: %v\n", err)
		}
		go run(ctx, stopChan)
	case <-ctx.Done():
		err = <-processResults
		if err != nil {
			log.Printf("vtund process exited with error: %v", err)
		}
	}
}

func IsRunning() bool {
	return vtunCmd != nil && vtunCmd.Process != nil
}

func Reload() error {
	return vtunCmd.Process.Signal(os.Signal(syscall.SIGHUP))
}
