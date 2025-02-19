package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/spf13/cobra"
)

func NewCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "aredn-manager",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:              runRoot,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
	cmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
	cmd.AddCommand(generateCmd)
	cmd.AddCommand(notifyCmd)
	cmd.AddCommand(serverCmd)
	cmd.AddCommand(addressCmd)
	return cmd
}

func runRoot(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)

	if config.Debug {
		fmt.Println("debug logging enabled")
	}

	return nil
}
