package vtun

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/runner"
	"gorm.io/gorm"
)

type vtunClient struct {
	cancel context.CancelFunc
	cmd    exec.Cmd
}

type ClientWatcher struct {
	started bool
	db      *gorm.DB
	config  *config.Config
	cancels map[uint]vtunClient
}

func NewVTunClientWatcher(db *gorm.DB, config *config.Config) *ClientWatcher {
	return &ClientWatcher{
		started: false,
		db:      db,
		config:  config,
		cancels: make(map[uint]vtunClient),
	}
}

func (v *ClientWatcher) Run() {
	if v.started {
		return
	}
	v.started = true
	go v.watch()
}

func (v *ClientWatcher) Stop() error {
	if !v.started {
		return fmt.Errorf("vtun client watcher not started")
	}
	v.started = false
	for _, cancel := range v.cancels {
		cancel.cancel()
	}
	return nil
}

func (v *ClientWatcher) ReloadTunnel(id uint) {
	if !v.started {
		return
	}
	cancel, ok := v.cancels[id]
	if !ok {
		return
	}
	if cancel.cancel != nil {
		cancel.cancel()
	}
}

func (v *ClientWatcher) Running(id uint) bool {
	if !v.started {
		return false
	}
	_, ok := v.cancels[id]
	return ok
}

func (v *ClientWatcher) watch() {
	for {
		if !v.started {
			return
		}
		tunnels, err := models.ListClientTunnels(v.db)
		if err != nil {
			fmt.Printf("VTunClientWatcher: Error listing tunnels: %v\n", err)
			continue
		}
		for _, tunnel := range tunnels {
			if tunnel.Wireguard {
				continue
			}
			if !v.Running(tunnel.ID) {
				withCancel, cancel := context.WithCancel(context.Background())
				v.cancels[tunnel.ID] = vtunClient{
					cancel: cancel,
				}
				err = v.runClient(withCancel, tunnel)
				if err != nil {
					fmt.Printf("VTunClientWatcher: Error running vtun client %s %s: %v\n", tunnel.Hostname, tunnel.IP, err)
					continue
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (v *ClientWatcher) wait(ctx context.Context, processResults chan error, tunnel models.Tunnel) {
	err := <-processResults
	if err != nil {
		if !v.started {
			return
		}
		fmt.Printf("VTunClientWatcher: Error running vtun client %s %s: %v\n", tunnel.Hostname, tunnel.IP, err)
		tunnel, err := models.FindTunnelByID(v.db, tunnel.ID)
		if err != nil {
			fmt.Printf("VTunClientWatcher: Error finding tunnel %d: %v\n", tunnel.ID, err)
			return
		}
		if tunnel.Client {
			err = v.runClient(ctx, tunnel)
			if err != nil {
				fmt.Printf("VTunClientWatcher: Error restarting vtun client %s %s: %v\n", tunnel.Hostname, tunnel.IP, err)
				return
			}
		}
		return
	}
}

func (v *ClientWatcher) runClient(ctx context.Context, tunnel models.Tunnel) error {
	// All we need to do is run the vtund client, it will daemonize itself and exit
	// vtund \
	//   -n
	//   -f /etc/vtund-${tunnel.hostname}-${dashed-net}.conf \
	//   -P ${port}
	//   ${v.config.ServerName}-${dashed-net} \
	//   ${tunnel.hostname}

	split := strings.Split(tunnel.Hostname, ":")
	host := split[0]
	port := "5525"
	if len(split) > 1 {
		port = split[1]
	}

	//nolint:golint,gosec
	cmd := exec.CommandContext(
		ctx,
		"vtund",
		"-P", port,
		"-n",
		"-f", fmt.Sprintf("/etc/vtund-%s-%s.conf", strings.ReplaceAll(tunnel.Hostname, ":", "-"), strings.ReplaceAll(tunnel.IP, ".", "-")),
		fmt.Sprintf("%s-%s", v.config.ServerName, strings.ReplaceAll(tunnel.IP, ".", "-")),
		host,
	)

	tunInfo, ok := v.cancels[tunnel.ID]
	if !ok {
		return fmt.Errorf("tunnel %d not found in cancels map", tunnel.ID)
	}
	tunInfo.cmd = *cmd
	v.cancels[tunnel.ID] = tunInfo

	processResults, err := runner.Run(cmd)
	defer close(processResults)
	if err != nil {
		return err
	}

	go v.wait(ctx, processResults, tunnel)

	return nil
}
