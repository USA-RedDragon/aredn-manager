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
AllowNoInt yes
IpVersion 4
LinkQualityAlgorithm "etx_ffeth"
Willingness 7

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
}

Interface "eth0"
{
    Mode "ether"
}`

	snippetOlsrdConfNameservice = `LoadPlugin "olsrd_nameservice.so.0.4"
{
    PlParam "interval" "30"
    PlParam "timeout" "300"
    PlParam "name-change-script" "aredn-manager notify"
    PlParam "name" "${SERVER_NAME}"
    PlParam "service" "http://${SERVER_NAME}:80/map|tcp|ki5vmf-cloud-tunnel-map"
    PlParam "service" "http://${SERVER_NAME}:81/|tcp|ki5vmf-cloud-tunnel-console"
}`

	snippetOlsrdConfSupernode = `Hna4
{
    10.0.0.0   255.128.0.0
    10.128.0.0 255.128.0.0
}`

	snippetOlsrdConfStandardTunnel = `Interface ${IFACES}
{
    Ip4Broadcast 255.255.255.255
    Mode "ether"
}`

	snippetOlsrdConfSupernodeTunnel = `Interface ${IFACES}
{
    Ip4Broadcast 255.255.255.255
    Mode "isolated"
    HnaInterval 1.0
    HnaValidityTime 600.0
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
	ret += "\n\n"

	// We need to replace shell variables in the template with the actual values
	cpSnippetOlsrdConfNameservice := snippetOlsrdConfNameservice
	utils.ShellReplace(
		&cpSnippetOlsrdConfNameservice,
		map[string]string{
			"SERVER_NAME": config.ServerName,
		},
	)
	ret += cpSnippetOlsrdConfNameservice

	if config.Supernode {
		ret += "\n\n"
		ret += snippetOlsrdConfSupernode
	}

	tunnelCount, err := models.CountTunnels(db)
	if err != nil {
		panic(err)
	}

	if tunnelCount > 0 {
		tun := 50
		tunnelString := ""
		for tunnelNumber := 0; tunnelNumber < tunnelCount; tunnelNumber++ {
			tunnelString += "\"tun" + fmt.Sprintf("%d", tun) + "\""
			if tunnelNumber != tunnelCount-1 {
				tunnelString += " "
			}
			tun++
		}

		if config.Supernode {
			ret += "\n\n"
			cpSnippetOlsrdConfSupernodeTunnel := snippetOlsrdConfSupernodeTunnel
			utils.ShellReplace(
				&cpSnippetOlsrdConfSupernodeTunnel,
				map[string]string{
					"IFACES": tunnelString,
				},
			)
			ret += cpSnippetOlsrdConfSupernodeTunnel
		} else {
			ret += "\n\n"
			cpSnippetOlsrdConfStandardTunnel := snippetOlsrdConfStandardTunnel
			utils.ShellReplace(
				&cpSnippetOlsrdConfStandardTunnel,
				map[string]string{
					"IFACES": tunnelString,
				},
			)
			ret += cpSnippetOlsrdConfStandardTunnel
		}
	}

	return ret
}
