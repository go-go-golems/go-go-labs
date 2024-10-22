package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
)

func main() {
	ctx := context.Background()

	catterCmd, err := NewCatterCommand()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating catter command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(catterCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building Cobra command: %v\n", err)
		os.Exit(1)
	}

	if err := cobraCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
