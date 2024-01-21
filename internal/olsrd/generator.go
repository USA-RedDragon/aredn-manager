package olsrd

import (
	"fmt"
	"os"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"gorm.io/gorm"
)

const (
	snippetOlsrdConf = `# This file is generated by the AREDN Manager
# Do not edit this file directly
DebugLevel 0
Pollrate 0.05
AllowNoInt yes
IpVersion 4
LinkQualityAlgorithm "etx_ffeth"
Willingness 7
MainIp ${MAIN_IP}

LoadPlugin "olsrd_arprefresh.so.0.1"
{
}

LoadPlugin "olsrd_txtinfo.so.1.1"
{
    PlParam "accept" "0.0.0.0"
}

LoadPlugin "olsrd_jsoninfo.so.1.1"
{
    PlParam "accept" "0.0.0.0"
}

LoadPlugin "olsrd_dot_draw.so.0.3"
{
    PlParam "accept" "0.0.0.0"
    PlParam "port" "2004"
}

LoadPlugin "olsrd_watchdog.so.0.1"
{
    PlParam "file" "/tmp/olsrd.watchdog"
    PlParam "interval" "5"
}`

	snippetOlsrdSupernodeConf = `# This file is generated by the AREDN Manager
# Do not edit this file directly
DebugLevel 0
Pollrate 0.01
AllowNoInt yes
IpVersion 4
LinkQualityAlgorithm "etx_ffeth"
Willingness 7
MainIp ${MAIN_IP}

LoadPlugin "olsrd_arprefresh.so.0.1"
{
}

LoadPlugin "olsrd_txtinfo.so.1.1"
{
    PlParam "accept" "0.0.0.0"
}

LoadPlugin "olsrd_jsoninfo.so.1.1"
{
    PlParam "accept" "0.0.0.0"
}

LoadPlugin "olsrd_dot_draw.so.0.3"
{
    PlParam "accept" "0.0.0.0"
    PlParam "port" "2004"
}

LoadPlugin "olsrd_watchdog.so.0.1"
{
    PlParam "file" "/tmp/olsrd.watchdog"
    PlParam "interval" "5"
}`

	snippetOlsrdConfEth0Supernode = `Interface "br0"
{
    Mode "isolated"
    Ip4Broadcast 255.255.255.255
    HnaInterval 1.0
    HnaValidityTime 600.0
}`

	snippetOlsrdConfEth0Standard = `Interface "br0"
{
    Mode "ether"
}`

	snippetOlsrdConfNameservice = `LoadPlugin "olsrd_nameservice.so.0.4"
{
    PlParam "interval" "30"
    PlParam "timeout" "300"
    PlParam "name-change-script" "aredn-manager notify"
    PlParam "name" "${SERVER_NAME}"
    ${SERVICES}
}`

	snippetOlsrdConfSupernode = `Hna4
{
    10.0.0.0   255.0.0.0
}`

	snippetOlsrdConfTunnel = `Interface ${IFACES}
{
    Ip4Broadcast 255.255.255.255
    Mode "ether"
}`
)

// This file will generate the olsrd.conf file

func GenerateAndSave(config *config.Config, db *gorm.DB) error {
	conf := Generate(config, db)
	if conf == "" {
		return fmt.Errorf("failed to generate olsrd.conf")
	}

	//nolint:golint,gosec
	return os.WriteFile("/etc/olsrd/olsrd.conf", []byte(conf), 0644)
}

func Generate(config *config.Config, db *gorm.DB) string {
	ret := snippetOlsrdConf
	if config.Supernode {
		ret = snippetOlsrdSupernodeConf
	}
	utils.ShellReplace(
		&ret,
		map[string]string{
			"MAIN_IP": config.NodeIP,
		},
	)
	ret += "\n\n"
	if config.Supernode {
		ret += snippetOlsrdConfEth0Supernode
	} else {
		ret += snippetOlsrdConfEth0Standard
	}
	ret += "\n\n"

	// We need to replace shell variables in the template with the actual values
	cpSnippetOlsrdConfNameservice := snippetOlsrdConfNameservice
	servicesText := "PlParam \"service\" \"http://${SERVER_NAME}/|tcp|${SERVER_NAME}-console\""

	utils.ShellReplace(
		&servicesText,
		map[string]string{
			"SERVER_NAME": config.ServerName,
		},
	)

	utils.ShellReplace(
		&cpSnippetOlsrdConfNameservice,
		map[string]string{
			"SERVER_NAME": config.ServerName,
			"SERVICES":    servicesText,
		},
	)
	ret += cpSnippetOlsrdConfNameservice

	if config.Supernode {
		ret += "\n\n"
		ret += snippetOlsrdConfSupernode
	}

	tunnels, err := models.ListVtunTunnels(db)
	if err != nil {
		panic(err)
	}

	if len(tunnels) > 0 {
		server_tun := 50
		client_tun := 100
		tunnelString := ""
		for tunnelNumber := 0; tunnelNumber < len(tunnels); tunnelNumber++ {
			tunnel := tunnels[tunnelNumber]
			if tunnel.Client {
				tunnelString += "\"tun" + fmt.Sprintf("%d", client_tun) + "\""
				client_tun++
			} else {
				tunnelString += "\"tun" + fmt.Sprintf("%d", server_tun) + "\""
				server_tun++
			}
			if tunnelNumber != len(tunnels)-1 {
				tunnelString += " "
			}
		}

		ret += "\n\n"
		cpSnippetOlsrdConfTunnel := snippetOlsrdConfTunnel
		utils.ShellReplace(
			&cpSnippetOlsrdConfTunnel,
			map[string]string{
				"IFACES": tunnelString,
			},
		)
		ret += cpSnippetOlsrdConfTunnel
	}

	return ret
}
