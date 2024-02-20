package olsrd

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
	olsrCmd *exec.Cmd
)

func Run(ctx context.Context) chan struct{} {
	stopChan := make(chan struct{})
	defer close(stopChan)
	go run(ctx, stopChan)
	return stopChan
}

func run(ctx context.Context, stopChan chan struct{}) {
	olsrCmd = exec.CommandContext(ctx, "olsrd", "-f", "/etc/olsrd/olsrd.conf", "-nofork")
	processResults, err := runner.Run(ctx, olsrCmd)
	defer close(processResults)
	if err != nil {
		fmt.Println("olsrd failed to start:", err)
		return
	}
	fmt.Println("OLSR started")

	select {
	case err := <-processResults:
		if err != nil {
			fmt.Printf("OLSR process exited with error: %v, restarting it\n", err)
		} else {
			fmt.Println("OLSR process exited, restarting it")
		}
		err = olsrCmd.Process.Signal(os.Signal(syscall.SIGKILL))
		if err != nil {
			log.Printf("failed to kill process: %v\n", err)
		}
		go run(ctx, stopChan)
	case <-ctx.Done():
		err = <-processResults
		if err != nil {
			log.Printf("olsrd process exited with error: %v", err)
		}
	}
}

func IsRunning() bool {
	return olsrCmd != nil && olsrCmd.Process != nil
}

func Reload() error {
	return olsrCmd.Process.Signal(os.Signal(syscall.SIGHUP))
}
