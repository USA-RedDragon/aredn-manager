package babel

import (
	"fmt"
	"os"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/wireguard"
	"gorm.io/gorm"
)

// This file will generate the babel.conf file

func GenerateAndSave(config *config.Config, db *gorm.DB) error {
	conf := Generate(config, db)
	if conf == "" {
		return fmt.Errorf("failed to generate babel.conf")
	}

	//nolint:golint,gosec
	return os.WriteFile("/tmp/babel-generated.conf", []byte(conf), 0644)
}

func Generate(config *config.Config, db *gorm.DB) string {
	// Yay this config format is much easier to generate.
	var ret string
	ret += "router-id " + config.BabelRouterID + "\n"
	ret += "interface br-dtdlink type wired\n"
	ret += "interface br-dtdlink rxcost 96\n"
	ret += "interface br-dtdlink split-horizon true\n"

	if config.Supernode {
		ret += "import-table 21\n"
		ret += "redistribute ip 10.0.0.0/8 allow\n"
		ret += fmt.Sprintf("redistribute ip %s/32 eq 32 allow\n", config.NodeIP)
		ret += fmt.Sprintf("out ip %s/32 eq 32 allow\n", config.NodeIP)
		ret += "out ip 10.0.0.0/8 eq 8 allow\n"
		ret += "redistribute ip 172.30.0.0/16 deny\n"
		ret += "install ip 10.0.0.0/8 eq 8 deny\n"
		ret += "out if br-dtdlink deny\n"
	} else {
		ret += "redistribute ip 10.0.0.0/8 ge 24 allow\n"
		ret += "redistribute ip 172.31.0.0/16 deny\n"
		ret += "install ip 10.0.0.0/8 eq 8 allow table 21\n"
	}
	ret += "redistribute local if br0 deny\n"
	ret += "install ip 0.0.0.0/0 eq 0 allow table 22\n"

	tunnelInterfaces := make([]string, 0)
	tunnels, err := models.ListWireguardTunnels(db)
	if err != nil {
		panic(err)
	}

	if len(tunnels) > 0 {
		for _, tunnel := range tunnels {
			if !tunnel.Enabled {
				continue
			}
			tunnelInterfaces = append(tunnelInterfaces, wireguard.GenerateWireguardInterfaceName(tunnel))
		}
	}

	for _, iface := range tunnelInterfaces {
		ret += fmt.Sprintf("interface %s type tunnel\n", iface)
		ret += fmt.Sprintf("interface %s enable-timestamps true\n", iface)
		ret += fmt.Sprintf("interface %s rxcost 206\n", iface)
		ret += fmt.Sprintf("interface %s max-rtt-penalty 150\n", iface)
		ret += fmt.Sprintf("redistribute local if %s deny\n", iface)
	}

	return ret
}
