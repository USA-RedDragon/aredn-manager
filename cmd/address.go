package cmd

import (
	"fmt"

	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/spf13/cobra"
)

//nolint:golint,gochecknoglobals
var (
	addressCmd = &cobra.Command{
		Use:               "address",
		Short:             "generates IPv6 addresses",
		RunE:              runAddress,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
)

func init() {
	addressCmd.PersistentFlags().BoolP("random", "r", false, "generate a random IPv6 address")
}

func runAddress(cmd *cobra.Command, _ []string) error {
	// check if random flag is set
	random, err := cmd.Flags().GetBool("random")
	if err != nil {
		return err
	}

	if random {
		// generate random IPv6 address
		ip, err := utils.GenerateIPv6RandomAddress()
		if err != nil {
			return err
		}
		fmt.Println(ip)
	} else {
		// generate link-local IPv6 address
		ip, err := utils.GenerateIPv6LinkLocalAddress()
		if err != nil {
			return err
		}
		fmt.Println(ip)
	}

	return nil
}
