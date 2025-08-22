package main

import (
	"context"
)

func main() {
    // Decide whether we're invoked as a toolexec wrapper (first arg is a tool path)
    if isWrapperInvocation() {
        runWrapper()
        return
    }

    // Otherwise, run the CLI for querying and managing runs
    runCLI(context.Background())
}


