package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"log"
	"os"
	"path"
)

type CodeBlock struct {
	Type    string
	Content string
}

func extractCodeBlocksFromMarkdown(content []byte) ([]CodeBlock, error) {
	var codeBlocks []CodeBlock

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)

	reader := text.NewReader(content)
	document := md.Parser().Parse(reader)

	err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == ast.KindFencedCodeBlock {
			fc := node.(*ast.FencedCodeBlock)
			var buf bytes.Buffer
			lines := fc.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				buf.Write(line.Value(reader.Source()))
			}
			codeBlocks = append(codeBlocks, CodeBlock{
				Type:    string(fc.Language(reader.Source())),
				Content: buf.String(),
			})
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return codeBlocks, nil
}

func writeBlocksToFile(blocks []CodeBlock, baseFileName string) error {
	for i, block := range blocks {
		fileName := fmt.Sprintf("%s-%d.%s", baseFileName, i+1, block.Type)
		fmt.Printf("Writing block to file '%s'\n", fileName)
		if err := os.WriteFile(fileName, []byte(block.Content), 0644); err != nil {
			return err
		}
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "md-tool [flags] [files...]",
	Short: "Extract code blocks from markdown files",
	Run: func(cmd *cobra.Command, args []string) {
		join, _ := cmd.Flags().GetBool("join")

		var allBlocks []CodeBlock
		for _, filename := range args {
			content, err := os.ReadFile(filename)
			if err != nil {
				log.Fatalf("Failed to read file '%s': %v", filename, err)
			}

			blocks, err := extractCodeBlocksFromMarkdown(content)
			if err != nil {
				log.Fatalf("Failed to extract code blocks from '%s': %v", filename, err)
			}

			if join {
				allBlocks = append(allBlocks, blocks...)
			} else {
				// remove extension
				extension := path.Ext(filename)
				baseName := filename[0 : len(filename)-len(extension)]
				if err := writeBlocksToFile(blocks, baseName); err != nil {
					log.Fatalf("Failed to write blocks to file: %v", err)
				}
			}
		}

		if join {
			for _, block := range allBlocks {
				fmt.Println(block.Content)
			}
		}
	},
}

func main() {
	rootCmd.Flags().BoolP("join", "j", false, "Concatenate all code blocks and output them on stdout")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}
