package main

import (
	"os"

	"github.com/go-go-golems/go-go-labs/cmd/apps/create-pull-request/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}