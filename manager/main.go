package main

import (
	"log"

	"github.com/USA-RedDragon/aredn-manager/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
