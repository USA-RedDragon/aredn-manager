package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/spf13/cobra"
)

//nolint:golint,gochecknoglobals
var (
	generateCmd = &cobra.Command{
		Use:               "generate",
		Short:             "Generate olsrd and vtund configs",
		RunE:              runGenerate,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

//nolint:golint,gochecknoinits
func init() {
	RootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)
	db := db.MakeDB(config)

	fmt.Println("Generating olsrd config")
	err := olsrd.GenerateAndSave(config, db)
	if err != nil {
		return err
	}

	fmt.Println("Generating vtund config")
	err = vtun.GenerateAndSave(config, db)
	if err != nil {
		return err
	}

	fmt.Println("Generating vtund client config")
	return vtun.GenerateAndSaveClient(config, db)
}
