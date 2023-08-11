package cmd

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/spf13/cobra"
)

var (
	notifyCmd = &cobra.Command{
		Use:               "notify",
		Short:             "notify the daemon of a change in the mesh",
		RunE:              runNotify,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func init() {
	RootCmd.AddCommand(notifyCmd)
}

func runNotify(cmd *cobra.Command, args []string) error {
	config := config.GetConfig(cmd)

	resp, err := http.Post(fmt.Sprintf("http://0.0.0.0:%d/api/v1/notify", config.Port), "application/json", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error notifying daemon: %s", resp.Status)
	}

	return nil
}
