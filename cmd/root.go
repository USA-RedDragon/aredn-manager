package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:               "aredn-manager",
		RunE:              runRoot,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func init() {
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
}

func runRoot(cmd *cobra.Command, args []string) error {
	config := config.GetConfig(cmd)

	if config.Debug {
		fmt.Println("debug logging enabled")
	}

	return nil
}
