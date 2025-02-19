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
	return ""
}
