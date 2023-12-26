package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/files"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var name, prefix string

	var rootCmd = &cobra.Command{
		Use:   "tempfile",
		Short: "TempFile writes stdin to a temporary file and prints out its name",
		Run: func(cmd *cobra.Command, args []string) {
			// Garbage Collect Existing Files
			deletedFiles, err := files.GarbageCollectTemporaryFiles(os.TempDir(), "*.tmp", 10) // Adjust path and mask as needed
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error in garbage collection: %v\n", err)
				os.Exit(1)
			}
			if len(deletedFiles) > 0 {
				_, _ = fmt.Fprintln(os.Stderr, "Deleted files:", deletedFiles)
			}

			// Read from stdin and write to a temp file
			tempFilePath, err := files.CopyReaderToTemporaryFile(prefix, name, os.Stdin)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error writing to temporary file: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(tempFilePath)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&name, "name", "n", "default", "Name of the temporary file")
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "temporizer", "Prefix for the temporary file")

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
