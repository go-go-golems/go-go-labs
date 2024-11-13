package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/weaviate/tiktoken-go"
)

type Config struct {
	showStats bool
	preview   bool
}

func main() {
	cfg := &Config{}
	flag.BoolVar(&cfg.showStats, "s", false, "Show statistics about the output")
	flag.BoolVar(&cfg.preview, "p", false, "Preview the content in $PAGER")
	flag.Parse()

	var outputStr string
	if len(flag.Args()) == 0 {
		// Read from stdin when no arguments provided
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		outputStr = string(input)
	} else {
		// Create command from arguments
		cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)

		// Capture both stdout and stderr
		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output

		// Run the command
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
			os.Exit(1)
		}

		outputStr = output.String()
	}

	// Copy to clipboard
	err := clipboard.WriteAll(outputStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
		os.Exit(1)
	}

	// Show stats if requested
	if cfg.showStats {
		printStats(outputStr)
	}

	// Preview if requested
	if cfg.preview {
		previewInPager(outputStr)
	}
}

func printStats(content string) {
	// Initialize token counter
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing token counter: %v\n", err)
		return
	}

	// Count tokens
	tokens := tokenCounter.Encode(content, nil, nil)
	tokenCount := len(tokens)

	// Count lines
	lineCount := strings.Count(content, "\n") + 1

	// Get size in bytes
	size := len(content)

	fmt.Printf("Statistics:\n")
	fmt.Printf("  Tokens: %d\n", tokenCount)
	fmt.Printf("  Lines:  %d\n", lineCount)
	fmt.Printf("  Size:   %d bytes\n", size)
}

func previewInPager(content string) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less" // Default to less if $PAGER is not set
	}

	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running pager: %v\n", err)
	}
}
