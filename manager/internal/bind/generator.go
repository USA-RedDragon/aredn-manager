package bind

import (
	"fmt"
	"os"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"gorm.io/gorm"
)

const (
	namedConf = `options {
    directory "/var/bind";
    listen-on { any; };
    listen-on-v6 { any; };
    pid-file "/var/run/named/named.pid";

	forwarders {
		127.0.0.11;
	};

	dnssec-validation no;
	auth-nxdomain no;    # conform to RFC1035
};

zone "mesh" {
    type master;
    file "/etc/bind/mesh.zone";
};

zone "local.mesh" {
    type master;
    file "/etc/bind/local.mesh.zone";
};
`

	supernodeInclude   = "include \"/etc/bind/named.supernode.conf\";"
	namedSupernodeConf = `acl "supernodes" {
${OTHER_SUPERNODE_IPS}
};

masters "supernodes" {
${OTHER_SUPERNODE_IPS}
};

zone "${SUPERNODE_ZONE}.mesh" {
    type master;
    also-notify { supernodes; };
    allow-transfer { supernodes; };
    file "/etc/bind/${SUPERNODE_ZONE}.mesh.zone";
};
`

	supernodeSlaveZone = `zone "${SUPERNODE_ZONE}.mesh" {
    type slave;
    masters { ${SUPERNODE_IPS} };
    allow-notify { ${SUPERNODE_IPS} };
    masterfile-format text;
    file "/etc/bind/${SUPERNODE_ZONE}.mesh.zone";
};`

	supernodeMasterZone = `$TTL 60
$ORIGIN ${SUPERNODE_ZONE}.mesh.
@  SOA  ns0.${SUPERNODE_ZONE}.mesh. root.${SUPERNODE_ZONE}.mesh. (
  ${SERIAL} ; Serial
  3600      ; Refresh
  300       ; Retry
  604800    ; Expire
  60        ; TTL
)
           NS ns0
ns0        A  ${NODE_IP}
`

	localMeshZone = `$TTL 60
$ORIGIN local.mesh.
@  SOA ns.local.mesh. master.local.mesh. (
  ${SERIAL}   ; Serial
  3600        ; Refresh
  300         ; Retry
  604800      ; Expire
  60          ; TTL
)
      NS ns0
ns0   A  ${NODE_IP}
`

	meshZone = `$TTL 60
$ORIGIN mesh.
@  SOA ns.mesh. master.mesh. (
  ${SERIAL}   ; Serial
  3600        ; Refresh
  300         ; Retry
  604800      ; Expire
  60          ; TTL
)
          NS ns0.local
ns0.local A  ${NODE_IP}
local     NS ns0.local
`

	meshZoneSupernode = `         NS ns${COUNT}.${SUPERNODE_MESH_NAME}
ns${COUNT}.${SUPERNODE_MESH_NAME} A  ${SUPERNODE_MESH_IP}
${SUPERNODE_MESH_NAME}      NS ns${COUNT}.${SUPERNODE_MESH_NAME}
`
)

// Always:
// local.mesh.zone will be generated from the template
// mesh.zone will be generated from the template
// named.conf will stay as is

// If supernode is enabled:
// We might add the include statement if supernode is enabled
// named.supernode.conf will be generated from the template
// {supernodename}.mesh.zone will be generated from the template

func GenerateAndSave(config *config.Config, db *gorm.DB) error {
	gen := Generate(config, db)
	for path, content := range gen {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate creates the config files. The map contains the file name as key and the file content as value.
func Generate(config *config.Config, db *gorm.DB) map[string]string {
	ret := make(map[string]string)
	ret["/etc/bind/named.conf"] = namedConf
	ret["/etc/bind/local.mesh.zone"] = generateLocalMeshZone(config, db)
	ret["/etc/bind/mesh.zone"] = generateMeshZone(config, db)

	if config.Supernode {
		ret["/etc/bind/named.supernode.conf"] = generateNamedSupernodeConf(config, db)
		ret["/etc/bind/"+config.SupernodeZone+".mesh.zone"] = generateSupernodeMasterZone(config, db)
		ret["/etc/bind/named.conf"] = ret["/etc/bind/named.conf"] + "\n" + supernodeInclude
	}

	return ret
}

func generateLocalMeshZone(config *config.Config, db *gorm.DB) string {
	ret := localMeshZone
	utils.ShellReplace(&ret, map[string]string{
		"NODE_IP": config.NodeIP,
		"SERIAL":  fmt.Sprintf("%d", time.Now().Unix()),
	})
	hostParser := olsrd.NewHostsParser()
	err := hostParser.Parse()
	if err != nil {
		fmt.Printf("could not parse hosts file: %s\n", err)
		// Don't panic because we expect the hosts file to be empty on a fresh install
		return ret
	}
	hosts := hostParser.GetHosts()
	for _, host := range hosts {
		ret += host.Hostname + " A " + host.IP.String() + "\n"
	}
	return ret
}

func generateMeshZone(config *config.Config, db *gorm.DB) string {
	ret := meshZone
	utils.ShellReplace(&ret, map[string]string{
		"NODE_IP": config.NodeIP,
		"SERIAL":  fmt.Sprintf("%d", time.Now().Unix()),
	})
	if config.Supernode {
		supernodes, err := models.ListSupernodes(db)
		if err != nil {
			panic(fmt.Errorf("could not list supernodes: %w", err))
		}
		for _, node := range supernodes {
			r2 := meshZoneSupernode
			for idx, ip := range node.IPs {
				utils.ShellReplace(&r2, map[string]string{
					"COUNT":               fmt.Sprintf("%d", idx),
					"SUPERNODE_MESH_NAME": node.MeshName,
					"SUPERNODE_MESH_IP":   ip,
				})
				ret += "\n" + r2
			}
		}
	}

	return ret
}

func generateNamedSupernodeConf(config *config.Config, db *gorm.DB) string {
	ret := namedSupernodeConf
	utils.ShellReplace(&ret, map[string]string{
		"SUPERNODE_ZONE": config.SupernodeZone,
	})
	supernodes, err := models.ListSupernodes(db)
	if err != nil {
		panic(fmt.Errorf("could not list supernodes: %w", err))
	}
	perLineIPs := ""
	for _, node := range supernodes {
		r2 := supernodeSlaveZone
		nodeIPs := ""
		for _, ip := range node.IPs {
			nodeIPs += ip + "; "
			perLineIPs += ip + ";\n"
		}
		utils.ShellReplace(&r2, map[string]string{
			"SUPERNODE_IPS":  nodeIPs,
			"SUPERNODE_ZONE": node.MeshName,
		})
		ret += "\n" + r2
	}
	utils.ShellReplace(&ret, map[string]string{
		"OTHER_SUPERNODE_IPS": perLineIPs,
	})

	return ret
}

func generateSupernodeMasterZone(config *config.Config, db *gorm.DB) string {
	ret := supernodeMasterZone
	utils.ShellReplace(&ret, map[string]string{
		"NODE_IP":        config.NodeIP,
		"SUPERNODE_ZONE": config.SupernodeZone,
		"SERIAL":         fmt.Sprintf("%d", time.Now().Unix()),
	})
	hostParser := olsrd.NewHostsParser()
	err := hostParser.Parse()
	if err != nil {
		fmt.Printf("could not parse hosts file: %s\n", err)
		// Don't panic because we expect the hosts file to be empty on a fresh install
		return ret
	}
	hosts := hostParser.GetHosts()
	for _, host := range hosts {
		ret += host.Hostname + " A " + host.IP.String() + "\n"
	}
	return ret
}
