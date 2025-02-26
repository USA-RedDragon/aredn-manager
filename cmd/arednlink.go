package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/arednlink"
	"github.com/USA-RedDragon/aredn-manager/internal/arednlink/pollers"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/cobra"
	"github.com/ztrue/shutdown"
)

//nolint:golint,gochecknoglobals
var (
	arednlinkCommand = &cobra.Command{
		Use:               "arednlink",
		Short:             "Reimplementation of the arednlink service",
		RunE:              runArednlink,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func runArednlink(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)
	routes := xsync.NewMapOf[string, string]()
	services := xsync.NewMapOf[string, string]()
	hosts := xsync.NewMapOf[string, string]()
	broadcastChan := make(chan arednlink.Message, 1024)

	arednlinkServer, err := arednlink.NewServer(config, &routes, hosts, services, broadcastChan)
	if err != nil {
		return err
	}

	pollers := pollers.NewManager(cmd.Context(), config, &routes, hosts, services, broadcastChan)
	pollers.Start()

	stopChan := make(chan interface{})
	stop := func(sig os.Signal) {
		// Extra newline to clear potential terminal control character
		fmt.Println()
		switch sig {
		case syscall.SIGINT:
			log.Println("arednlink: received SIGINT, shutting down")
		case syscall.SIGKILL:
			log.Println("arednlink: received SIGKILL, shutting down")
		case syscall.SIGTERM:
			log.Println("arednlink: received SIGTERM, shutting down")
		case syscall.SIGQUIT:
			log.Println("arednlink: received SIGQUIT, shutting down")
		}

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			arednlinkServer.Stop()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			pollers.Stop()
		}()

		wg.Wait()

		close(stopChan)
	}
	shutdown.AddWithParam(stop)
	shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	<-stopChan
	return nil
}
