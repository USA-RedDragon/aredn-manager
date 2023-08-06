package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/server"
	"github.com/spf13/cobra"
	"github.com/ztrue/shutdown"
)

var (
	serverCmd = &cobra.Command{
		Use:               "server",
		Short:             "Start the daemon server",
		RunE:              runServer,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func init() {
	serverCmd.Flags().String("pid-file", "/var/run/aredn-manager.pid", "file to write the daemon PID to")
	serverCmd.Flags().IntP("port", "p", 3333, "port to listen on")
	serverCmd.Flags().Bool("no-daemon", false, "do not daemonize the process")
	RootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	config := config.GetConfig(cmd)

	// Check if the PID file exists
	if _, err := os.Stat(config.PIDFile); err == nil {
		// Read the PID file
		pidBytes, err := os.ReadFile(config.PIDFile)
		if err != nil {
			return err
		}
		pidStr := string(pidBytes)
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			return err
		}

		// Check if the PID is running
		if process, err := os.FindProcess(int(pid)); err == nil {
			// Workaround since FindProcess doesn't actually check if the process is running
			err := process.Signal(syscall.Signal(0))
			if err == nil {
				return fmt.Errorf("only one instance of the daemon can be running at a time")
			}
		}
	}

	if config.Daemonize {
		// Fork a child process that runs this same command, but with --no-daemon
		// The child process will write its PID to the PID file and start the server
		_, err := syscall.ForkExec(os.Args[0], append(os.Args, "--no-daemon"),
			&syscall.ProcAttr{
				Env: os.Environ(),
				Sys: &syscall.SysProcAttr{
					Setsid: true,
				},
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
			},
		)
		if err != nil {
			return err
		}
		fmt.Println("aredn-manager daemon started")
		return nil
	} else {
		// Write the current PID to the PID file
		pidStr := fmt.Sprintf("%d", os.Getpid())
		err := os.WriteFile(config.PIDFile, []byte(pidStr), 0644)
		if err != nil {
			return err
		}

		// Start the server
		fmt.Println("starting server")

		db := db.MakeDB(config)
		srv := server.NewServer(config, db)
		srv.Run()
		stop := func(sig os.Signal) {
			wg := new(sync.WaitGroup)

			wg.Add(1)
			go func() {
				defer wg.Done()
				srv.Stop()
			}()

			const timeout = 10 * time.Second
			c := make(chan struct{})
			go func() {
				defer close(c)
				wg.Wait()
			}()
			clearPID := func() {
				err := os.Remove(config.PIDFile)
				if err != nil {
					log.Fatal("failed to remove PID file")
				}
			}
			select {
			case <-c:
				clearPID()
				os.Exit(0)
			case <-time.After(timeout):
				clearPID()
				os.Exit(1)
			}
		}
		shutdown.AddWithParam(stop)
		shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	}

	return nil
}
