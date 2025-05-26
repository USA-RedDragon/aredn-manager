package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/events"
	"github.com/USA-RedDragon/aredn-manager/internal/ifacewatcher"
	"github.com/USA-RedDragon/aredn-manager/internal/metrics"
	"github.com/USA-RedDragon/aredn-manager/internal/server"
	"github.com/USA-RedDragon/aredn-manager/internal/services"
	"github.com/USA-RedDragon/aredn-manager/internal/services/arednlink"
	"github.com/USA-RedDragon/aredn-manager/internal/services/babel"
	"github.com/USA-RedDragon/aredn-manager/internal/services/dnsmasq"
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/USA-RedDragon/aredn-manager/internal/wireguard"
	"github.com/USA-RedDragon/configulator"
	"github.com/spf13/cobra"
	"github.com/ztrue/shutdown"
	"golang.org/x/sync/errgroup"
)

func newServerCommand(version, commit string) *cobra.Command {
	return &cobra.Command{
		Use:     "server",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Short:   "Start the daemon server",
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:              runServer,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
}

func runServer(cmd *cobra.Command, _ []string) error {
	err := runRoot(cmd, nil)
	if err != nil {
		slog.Error("Encountered an error.", "error", err.Error())
	}

	ctx := cmd.Context()

	c, err := configulator.FromContext[config.Config](ctx)
	if err != nil {
		return fmt.Errorf("failed to get config from context")
	}

	config, err := c.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Start the server
	slog.Info("Starting server")

	serviceRegistry := services.NewServiceRegistry()
	if config.OLSR {
		serviceRegistry.Register(services.OLSRServiceName, olsr.NewService(config))
	}
	if config.Babel.Enabled {
		serviceRegistry.Register(services.BabelServiceName, babel.NewService(config))
		serviceRegistry.Register(services.AREDNLinkServiceName, arednlink.NewService(config))
	}
	serviceRegistry.Register(services.DNSMasqServiceName, dnsmasq.NewService(config))

	go serviceRegistry.StartAll()

	// Start the metrics server
	go metrics.CreateMetricsServer(config, cmd.Root().Version)
	log.Printf("Metrics server started")

	db, err := db.MakeDB(config)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	log.Printf("DB connection established")

	// Clear active status from all tunnels in the db
	err = models.ClearActiveFromAllTunnels(db)
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

	// Start the server
	srv := server.NewServer(config, db, ifWatcher.Stats, eventBus.GetChannel(), wireguardManager)
	err = srv.Run(cmd.Root().Version, serviceRegistry)
	if err != nil {
		return err
	}
	log.Printf("Server started")

	stopChan := make(chan error)
	defer close(stopChan)
	stop := func(sig os.Signal) {
		// Extra newline to clear potential terminal control character
		fmt.Println()
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
		errGrp.SetLimit(1)

		errGrp.Go(func() error {
			return wireguardManager.Stop()
		})

		errGrp.Go(func() error {
			return srv.Stop()
		})

		errGrp.Go(func() error {
			return ifWatcher.Stop()
		})

		errGrp.Go(func() error {
			return models.ClearActiveFromAllTunnels(db)
		})

		errGrp.Go(func() error {
			eventBus.Close()
			return nil
		})

		errGrp.Go(func() error {
			return serviceRegistry.StopAll()
		})

		stopChan <- errGrp.Wait()
	}
	shutdown.AddWithParam(stop)
	log.Println("Signal hooks added")
	shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	log.Println("Signal listener returned")

	return <-stopChan
}
