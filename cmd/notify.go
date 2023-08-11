package cmd

import (
	"fmt"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/spf13/cobra"
)

//nolint:golint,gochecknoglobals
var (
	notifyCmd = &cobra.Command{
		Use:               "notify",
		Short:             "notify the daemon of a change in the mesh",
		RunE:              runNotify,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

//nolint:golint,gochecknoinits
func init() {
	RootCmd.AddCommand(notifyCmd)
}

func runNotify(cmd *cobra.Command, _ []string) error {
	config := config.GetConfig(cmd)

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, fmt.Sprintf("http://0.0.0.0:%d/api/v1/notify", config.Port), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error notifying daemon: %s", resp.Status)
	}

	return nil
}
