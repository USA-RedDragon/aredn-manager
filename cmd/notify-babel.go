package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/configulator"
	"github.com/spf13/cobra"
)

func newNotifyBabelCommand(version, commit string) *cobra.Command {
	return &cobra.Command{
		Use:     "notify-babel",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Short:   "Notify the daemon of a change in the babel mesh",
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:              runNotifyBabel,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
}

func runNotifyBabel(cmd *cobra.Command, _ []string) error {
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

	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, fmt.Sprintf("http://0.0.0.0:%d/api/v1/notify-babel", config.Port), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error notifying daemon: %s", resp.Status)
	}

	return nil
}
