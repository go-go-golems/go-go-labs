package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-go-labs/pkg/js"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Configure zerolog
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var rootCmd = &cobra.Command{
		Use:   "js-docstring-extractor-2 [file...]",
		Short: "Extract JavaScript docstrings and function signatures",
		Long: `A command-line tool that extracts docstrings and function signatures from JavaScript files
and outputs them in Markdown format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
				log.Debug().Msg("Debug logging enabled")
			}

			if len(args) == 0 {
				return extractFromStdin(cmd)
			}

			return extractFromFiles(cmd, args)
		},
	}

	// Output file flag
	rootCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().BoolP("recursive", "r", false, "Process directories recursively")
	rootCmd.Flags().BoolP("jsdoc-only", "j", false, "Only extract JSDoc comments (starting with /**)")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug logging")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func extractFromStdin(cmd *cobra.Command) error {
	outputFile, _ := cmd.Flags().GetString("output")
	jsdocOnly, _ := cmd.Flags().GetBool("jsdoc-only")

	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "failed to read from stdin")
	}

	// Extract docstrings
	result, err := js.ExtractDocstrings(data, "stdin.js", jsdocOnly)
	if err != nil {
		return errors.Wrap(err, "failed to extract docstrings")
	}

	// Generate markdown
	markdown := js.GenerateMarkdown(result)

	// Output result
	return writeOutput(markdown, outputFile)
}

func extractFromFiles(cmd *cobra.Command, paths []string) error {
	outputFile, _ := cmd.Flags().GetString("output")
	recursive, _ := cmd.Flags().GetBool("recursive")
	jsdocOnly, _ := cmd.Flags().GetBool("jsdoc-only")

	allResults := make(map[string][]js.FunctionWithDocs)

	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return errors.Wrapf(err, "failed to stat %s", path)
		}

		if fileInfo.IsDir() {
			if recursive {
				err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".js") {
						return nil
					}
					return processFile(filePath, jsdocOnly, allResults)
				})
				if err != nil {
					return errors.Wrapf(err, "error walking directory %s", path)
				}
			} else {
				// Process only JS files in the top level of the directory
				files, err := os.ReadDir(path)
				if err != nil {
					return errors.Wrapf(err, "failed to read directory %s", path)
				}
				for _, file := range files {
					if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".js") {
						continue
					}
					filePath := filepath.Join(path, file.Name())
					if err := processFile(filePath, jsdocOnly, allResults); err != nil {
						return err
					}
				}
			}
		} else {
			if !strings.HasSuffix(strings.ToLower(path), ".js") {
				fmt.Fprintf(os.Stderr, "Warning: %s does not have a .js extension, skipping\n", path)
				continue
			}
			if err := processFile(path, jsdocOnly, allResults); err != nil {
				return err
			}
		}
	}

	// Generate markdown
	markdown := js.GenerateMarkdownForMultipleFiles(allResults)

	// Output result
	return writeOutput(markdown, outputFile)
}

func processFile(filePath string, jsdocOnly bool, results map[string][]js.FunctionWithDocs) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %s", filePath)
	}

	// Extract docstrings
	result, err := js.ExtractDocstrings(data, filePath, jsdocOnly)
	if err != nil {
		return errors.Wrapf(err, "failed to extract docstrings from %s", filePath)
	}

	fmt.Printf("Found %d functions in %s\n", len(result), filePath)
	for i, fn := range result {
		fmt.Printf("  %d: %s (%d params)\n", i, fn.Name, len(fn.Parameters))
	}

	if len(result) > 0 {
		results[filePath] = result
	}

	return nil
}

func writeOutput(content string, outputFile string) error {
	if outputFile == "" {
		// Write to stdout
		_, err := fmt.Print(content)
		return err
	}

	// Write to file
	return os.WriteFile(outputFile, []byte(content), 0644)
}
