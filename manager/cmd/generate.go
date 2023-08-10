package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/bind"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/spf13/cobra"
)

var (
	generateCmd = &cobra.Command{
		Use:               "generate",
		Short:             "Generate olsrd and vtund configs",
		RunE:              runGenerate,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func init() {
	RootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
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

	fmt.Println("Generating BIND config")
	return bind.GenerateAndSave(config, db)
}
