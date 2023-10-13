package vtun

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

type vtunClient struct {
	cancel context.CancelFunc
	cmd    exec.Cmd
}

type VTunClientWatcher struct {
	started bool
	db      *gorm.DB
	config  *config.Config
	cancels map[uint]vtunClient
}

func NewVTunClientWatcher(db *gorm.DB, config *config.Config) *VTunClientWatcher {
	return &VTunClientWatcher{
		started: false,
		db:      db,
		config:  config,
		cancels: make(map[uint]vtunClient),
	}
}

func (v *VTunClientWatcher) Run() {
	if v.started {
		return
	}
	go v.watch()
}

func (v *VTunClientWatcher) Stop() {
	if !v.started {
		return
	}
	v.started = false
	for _, cancel := range v.cancels {
		cancel.cancel()
	}
}

func (v *VTunClientWatcher) ReloadTunnel(id uint) {
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

func (v *VTunClientWatcher) Running(id uint) bool {
	if !v.started {
		return false
	}
	_, ok := v.cancels[id]
	return ok
}

func (v *VTunClientWatcher) watch() {
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

func (v *VTunClientWatcher) run(ctx context.Context, cmd exec.Cmd, tunnel models.Tunnel) {
	err := cmd.Wait()
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
			v.runClient(ctx, tunnel)
		}
		return
	}
}

func (v *VTunClientWatcher) runClient(ctx context.Context, tunnel models.Tunnel) error {
	// All we need to do is run the vtund client, it will daemonize itself and exit
	// vtund \
	//   -f /etc/vtund-${tunnel.hostname}-${dashed-net}.conf \
	//   ${v.config.ServerName}-${dashed-net} \
	//   ${tunnel.hostname}
	cmd := exec.CommandContext(
		ctx,
		"vtund",
		"-n",
		"-f", fmt.Sprintf("/etc/vtund-%s-%s.conf", strings.ReplaceAll(tunnel.Hostname, ":", "-"), strings.ReplaceAll(tunnel.IP, ".", "-")),
		fmt.Sprintf("%s-%s", v.config.ServerName, strings.ReplaceAll(tunnel.IP, ".", "-")),
		tunnel.Hostname,
	)

	tunInfo, ok := v.cancels[tunnel.ID]
	if !ok {
		return fmt.Errorf("Tunnel %d not found in cancels map", tunnel.ID)
	}
	tunInfo.cmd = *cmd
	v.cancels[tunnel.ID] = tunInfo

	err := cmd.Start()
	if err != nil {
		return err
	}

	go v.run(ctx, *cmd, tunnel)

	return nil
}
