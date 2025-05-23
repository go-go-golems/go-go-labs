package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
