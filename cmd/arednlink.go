package cmd

import (
	"log"
	"os"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/arednlink"
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

func runArednlink(_ *cobra.Command, _ []string) error {
	arednlinkServer, err := arednlink.NewServer()
	if err != nil {
		return err
	}

	stopChan := make(chan interface{})
	stop := func(sig os.Signal) {
		switch sig {
		case syscall.SIGINT:
			log.Println("Received SIGINT, shutting down")
		case syscall.SIGKILL:
			log.Println("Received SIGKILL, shutting down")
		case syscall.SIGTERM:
			log.Println("Received SIGTERM, shutting down")
		case syscall.SIGQUIT:
			log.Println("Received SIGQUIT, shutting down")
		}
		arednlinkServer.Stop()
		close(stopChan)
	}
	shutdown.AddWithParam(stop)
	shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	<-stopChan
	return nil
}
