package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "embeddings",
	Short: "A mock embeddings and similarity computation server",
	Long: `A server that provides mock APIs for computing embeddings and similarities between texts.
It exposes two endpoints:
- /compute-embeddings - Computes embeddings for given text
- /compute-similarity - Computes similarity between two texts`,
}
