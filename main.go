package main

import (
	"log"

	"github.com/USA-RedDragon/aredn-manager/cmd"
)

// https://goreleaser.com/cookbooks/using-main.version/
//
//nolint:golint,gochecknoglobals
var (
	version = "dev"
	commit  = "none"
)

func main() {
	rootCmd := cmd.NewCommand(version, commit)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
