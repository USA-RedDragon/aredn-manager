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
	RootCmd.AddCommand(serverCmd)
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
			log.Fatal(err)
		}
	}()

	go func() {
		err := vtun.Run(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Start the metrics server
	go metrics.CreateMetricsServer(config)

	db := db.MakeDB(config)

	// Clear active status from all tunnels in the db
	err := models.ClearActiveFromAllTunnels(db)
	if err != nil {
		return err
	}

	// Run the OLSR metrics watcher
	go metrics.OLSRWatcher(db)

	// Initialize the websocket event bus
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Start the interface watcher
	ifWatcher := ifacewatcher.NewWatcher(db, eventBus.GetChannel())
	err = ifWatcher.Watch()
	if err != nil {
		return err
	}

	// Start the vtun client watcher
	vtunClientWatcher := vtun.NewVTunClientWatcher(db, config)
	vtunClientWatcher.Run()

	// Start the wireguard manager
	wireguardManager, err := wireguard.NewManager(db)
	if err != nil {
		return err
	}
	err = wireguardManager.Run()
	if err != nil {
		return err
	}

	// Start the server
	srv := server.NewServer(config, db, ifWatcher.Stats, eventBus.GetChannel(), vtunClientWatcher, wireguardManager)
	err = srv.Run()
	if err != nil {
		return err
	}

	stopChan := make(chan error)
	defer close(stopChan)
	stop := func(sig os.Signal) {
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
	shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

	return <-stopChan
}
