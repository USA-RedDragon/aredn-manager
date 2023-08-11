package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/spf13/cobra"
)

//nolint:golint,gochecknoglobals
var (
	RootCmd = &cobra.Command{
		Use:               "aredn-manager",
		RunE:              runRoot,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

//nolint:golint,gochecknoinits
func init() {
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
}

func runRoot(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)

	if config.Debug {
		fmt.Println("debug logging enabled")
	}

	return nil
}
