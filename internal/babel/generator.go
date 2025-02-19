package babel

import (
	"fmt"
	"os"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
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
	ret += "interface br0 type wired\n"
	ret += "interface br0 rxcost 96\n"
	ret += "interface br0 split-horizon true\n"
	ret += "out if br0 deny\n"

	if config.Supernode {
		ret += "import-table 21\n"
		ret += "redistribute ip 10.0.0.0/8 allow\n"
		ret += "redistribute ip 44.0.0.0/9 allow\n"
		ret += "redistribute ip 44.128.0.0/10 allow\n"
		ret += fmt.Sprintf("redistribute ip %s/32 eq 32 allow\n", config.NodeIP)
		ret += fmt.Sprintf("out ip %s/32 eq 32 allow\n", config.NodeIP)
		ret += "out ip 10.0.0.0/8 eq 8 allow\n"
		ret += "out ip 44.0.0.0/9 eq 9 allow\n"
		ret += "out ip 44.128.0.0/10 eq 10 allow\n"
		ret += "redistribute ip 172.30.0.0/16 deny\n"
		ret += "install ip 10.0.0.0/8 eq 8 deny\n"
		ret += "install ip 44.0.0.0/9 eq 9 deny\n"
		ret += "install ip 44.128.0.0/10 eq 10 deny\n"
	} else {
		ret += "redistribute ip 10.0.0.0/8 ge 24 allow\n"
		ret += "redistribute ip 44.0.0.0/9 allow\n"
		ret += "redistribute ip 44.128.0.0/10 allow\n"
		ret += "redistribute ip 172.31.0.0/16 deny\n"
		ret += "install ip 10.0.0.0/8 eq 8 allow table 21\n"
		ret += "install ip 44.0.0.0/9 eq 9 allow table 21\n"
		ret += "install ip 44.128.0.0/10 eq 10 allow table 21\n"
	}
	ret += "redistribute local if eth0 deny\n"
	ret += "install ip 0.0.0.0/0 eq 0 allow table 22\n"
	return ret
}
