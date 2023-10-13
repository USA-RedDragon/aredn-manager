package vtun

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"gorm.io/gorm"
)

type VTunClientWatcher struct {
	started bool
	db      *gorm.DB
	config  *config.Config
}

func NewVTunClientWatcher(db *gorm.DB, config *config.Config) *VTunClientWatcher {
	return &VTunClientWatcher{
		started: false,
		db:      db,
		config:  config,
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
			if !IsRunningClient(tunnel.Hostname, tunnel.IP) {
				err = v.runClient(tunnel)
				if err != nil {
					fmt.Printf("VTunClientWatcher: Error running vtun client %s %s: %v\n", tunnel.Hostname, tunnel.IP, err)
					continue
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (v *VTunClientWatcher) runClient(tunnel models.Tunnel) error {
	// All we need to do is run the vtund client, it will daemonize itself and exit
	// vtund \
	//   -f /etc/vtund-${tunnel.hostname}-${dashed-net}.conf \
	//   -r /usr/var/run/vtundclient-${tunnel.hostname}-${dashed-net}.pid \
	//   -P ${tunnel.port} \
	//   ${v.config.ServerName}-${dashed-net} \
	//   ${tunnel.hostname}

	split := strings.Split(tunnel.Hostname, ":")
	host := split[0]
	port := "5525"
	if len(split) > 1 {
		port = split[1]
	}

	cmd := exec.Command(
		"vtund",
		"-f", fmt.Sprintf("/etc/vtund-%s-%s.conf", strings.ReplaceAll(tunnel.Hostname, ":", "-"), strings.ReplaceAll(tunnel.IP, ".", "-")),
		"-r", fmt.Sprintf("/usr/var/run/vtundclient-%s-%s.pid", strings.ReplaceAll(tunnel.Hostname, ":", "-"), strings.ReplaceAll(tunnel.IP, ".", "-")),
		"-P", port,
		fmt.Sprintf("%s-%s", v.config.ServerName, strings.ReplaceAll(tunnel.IP, ".", "-")),
		host,
	)

	return cmd.Run()
}
