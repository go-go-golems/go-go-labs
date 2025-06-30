package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-labs/cmd/apps/filter-binaries-from-git-history/cmd"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize default logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
