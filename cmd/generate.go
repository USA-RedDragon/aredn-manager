package cmd

import (
	"fmt"
	"log/slog"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/services/babel"
	"github.com/USA-RedDragon/aredn-manager/internal/services/olsr"
	"github.com/USA-RedDragon/configulator"
	"github.com/spf13/cobra"
)

//nolint:golint,gochecknoglobals
var (
	generateCmd = &cobra.Command{
		Use:               "generate",
		Short:             "Generate olsrd, babeld configs",
		RunE:              runGenerate,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func runGenerate(cmd *cobra.Command, _ []string) error {
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

	db := db.MakeDB(config)

	fmt.Println("Generating olsrd config")
	err = olsr.GenerateAndSave(config, db)
	if err != nil {
		return err
	}

	if config.Babel.Enabled {
		fmt.Println("Generating babeld config")
		err = babel.GenerateAndSave(config, db)
		if err != nil {
			return err
		}
	}

	return nil
}
