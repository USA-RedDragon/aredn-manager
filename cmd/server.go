package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/events"
	"github.com/USA-RedDragon/aredn-manager/internal/ifacewatcher"
	"github.com/USA-RedDragon/aredn-manager/internal/metrics"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/server"
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/USA-RedDragon/aredn-manager/internal/wireguard"
	"github.com/spf13/cobra"
	"github.com/ztrue/shutdown"
	"golang.org/x/sync/errgroup"
)

//nolint:golint,gochecknoglobals
var (
	serverCmd = &cobra.Command{
		Use:               "server",
		Short:             "Start the daemon server",
		RunE:              runServer,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

//nolint:golint,gochecknoinits
func init() {
	serverCmd.Flags().String("pid-file", "/var/run/aredn-manager.pid", "file to write the daemon PID to")
	serverCmd.Flags().IntP("port", "p", 3333, "port to listen on")
	serverCmd.Flags().Bool("no-daemon", false, "do not daemonize the process")
}

func runServer(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)

	// Start the server
	fmt.Println("starting server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run olsrd and vtun
	go func() {
		err := olsrd.Run(ctx)
		if err != nil {
			log.Println("OLSRD returned error", err)
		}
	}()

	go func() {
		err := vtun.Run(ctx)
		if err != nil {
			log.Println("VTUND returned error", err)
		}
	}()

	// Start the metrics server
	go metrics.CreateMetricsServer(config, cmd.Root().Version)
	log.Printf("Metrics server started")

	db := db.MakeDB(config)
	log.Printf("DB connection established")

	// Clear active status from all tunnels in the db
	err := models.ClearActiveFromAllTunnels(db)
	if err != nil {
		return err
	}
	log.Printf("Cleared active status from all tunnels")

	// Start the wireguard manager
	wireguardManager, err := wireguard.NewManager(db)
	if err != nil {
		return err
	}
	log.Printf("Wireguard manager started")

	err = wireguardManager.Run()
	if err != nil {
		return err
	}
	log.Printf("Wireguard manager running")

	// Run the OLSR metrics watcher
	go metrics.OLSRWatcher(db)
	log.Printf("OLSR watcher started")

	// Initialize the websocket event bus
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	log.Printf("Event bus started")

	// Start the interface watcher
	ifWatcher, err := ifacewatcher.NewWatcher(db, eventBus.GetChannel())
	if err != nil {
		return err
	}
	err = ifWatcher.Watch()
	if err != nil {
		return err
	}
	log.Printf("Interface watcher started")

	// Start the vtun client watcher
	vtunClientWatcher := vtun.NewVTunClientWatcher(db, config)
	vtunClientWatcher.Run()
	log.Printf("VTun client watcher started")

	// Start the server
	srv := server.NewServer(config, db, ifWatcher.Stats, eventBus.GetChannel(), vtunClientWatcher, wireguardManager)
	err = srv.Run(cmd.Root().Version)
	if err != nil {
		return err
	}
	log.Printf("Server started")

	stopChan := make(chan error)
	defer close(stopChan)
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
		errGrp := errgroup.Group{}

		errGrp.Go(func() error {
			return wireguardManager.Stop()
		})

		errGrp.Go(func() error {
			return srv.Stop()
		})

		errGrp.Go(func() error {
			return vtunClientWatcher.Stop()
		})

		errGrp.Go(func() error {
			eventBus.Close()
			return nil
		})

		errGrp.Go(func() error {
			ifWatcher.Stop()
			return models.ClearActiveFromAllTunnels(db)
		})

		stopChan <- errGrp.Wait()
	}
	shutdown.AddWithParam(stop)
	log.Println("Signal hooks added")
	shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	log.Println("Signal listener returned")

	return <-stopChan
}
