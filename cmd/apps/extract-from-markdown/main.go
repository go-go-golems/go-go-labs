package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type CodeBlock struct {
	Type    string
	Content string
}

var (
	join   bool
	output string

	rootCmd = &cobra.Command{
		Use:   "extractor",
		Short: "Extract code blocks from markdown files",
		Run:   execute,
		Args:  cobra.MinimumNArgs(1),
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&join, "join", false, "Join all the code blocks together")
	rootCmd.PersistentFlags().StringVar(&output, "output", "", "Output file for extracted code blocks")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func execute(cmd *cobra.Command, args []string) {
	var reader io.Reader
	for _, inputFile := range args {
		if inputFile == "-" {
			reader = os.Stdin
		} else {
			file, err := os.Open(inputFile)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to open input file: %s\n", err)
				os.Exit(1)
			}
			defer func(file *os.File) {
				_ = file.Close()
			}(file)
			reader = file
		}

		codeBlocks := extractCodeBlocks(reader)

		if output == "" {
			printCodeBlocks(os.Stdout, codeBlocks)
		} else {
			writeCodeBlocksToFile(output, codeBlocks)
		}
	}
}

func extractCodeBlocks(r io.Reader) []CodeBlock {
	re := regexp.MustCompile("(?s)```(.*?)\n(.*?)```")
	scanner := bufio.NewScanner(r)
	content := ""
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	matches := re.FindAllStringSubmatch(content, -1)
	var blocks []CodeBlock
	for _, match := range matches {
		blockType := strings.TrimSpace(match[1])
		if blockType == "" {
			blockType = "txt"
		}
		blocks = append(blocks, CodeBlock{Type: blockType, Content: strings.TrimSpace(match[2])})
	}
	return blocks
}

func printCodeBlocks(w io.Writer, blocks []CodeBlock) {
	for _, block := range blocks {
		_, _ = fmt.Fprintln(w, block.Content)
		if !join {
			_, _ = fmt.Fprintln(w, "--------------------------------")
		}
	}
}

func writeCodeBlocksToFile(outputFile string, blocks []CodeBlock) {
	for i, block := range blocks {
		filename := outputFile
		if !join && len(blocks) > 1 {
			filename = fmt.Sprintf("%s-%d.%s", strings.TrimSuffix(outputFile, ".txt"), i+1, block.Type)
		}
		file, err := os.Create(filename)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to create output file: %s\n", err)
			return
		}
		_, err = file.WriteString(block.Content)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to write to output file: %s\n", err)
			return
		}
		_ = file.Close()

		if join {
			break
		}
	}
}
