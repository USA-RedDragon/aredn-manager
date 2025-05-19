package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/USA-RedDragon/aredn-manager/cmd"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/configulator"
)

// https://goreleaser.com/cookbooks/using-main.version/
//
//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
)

func main() {
	rootCmd := cmd.NewCommand(version, commit)

	c := configulator.New[config.Config]().
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{
			Separator: "_",
		}).
		WithFile(&configulator.FileOptions{
			Paths: []string{"config.yaml"},
		}).
		WithPFlags(rootCmd.Flags(), nil)

	rootCmd.SetContext(c.WithContext(context.TODO()))

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Encountered an error.", "error", err.Error())
		os.Exit(1)
	}
}
