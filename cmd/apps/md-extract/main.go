package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ExtractMdCommand struct {
	*cmds.CommandDescription
}

type ExtractMdSettings struct {
	Output           string   `glazed.parameter:"output"`
	WithQuotes       bool     `glazed.parameter:"with-quotes"`
	File             string   `glazed.parameter:"file"`
	AllowedLanguages []string `glazed.parameter:"allowed-languages"`
}

func NewExtractMdCommand() (*ExtractMdCommand, error) {
	return &ExtractMdCommand{
		CommandDescription: cmds.NewCommandDescription(
			"extract-md",
			cmds.WithShort("Extract code blocks from markdown"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"output",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Output format"),
					parameters.WithDefault("concatenated"),
					parameters.WithChoices("concatenated", "list", "yaml"),
				),
				parameters.NewParameterDefinition(
					"with-quotes",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include code block quotes"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"allowed-languages",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of allowed languages to extract"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Input file (use - for stdin)"),
					parameters.WithDefault("-"),
				),
			),
		),
	}, nil
}

func (c *ExtractMdCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	// Wrap the writer with a bufio.Writer
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	bw = bufio.NewWriter(os.Stderr)

	s := &ExtractMdSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	var input strings.Builder
	var lastOutput string

	processAndWrite := func(data string) error {
		input.WriteString(data)
		blocks := markdown.ExtractAllBlocks(input.String())
		output, err := generateOutput(blocks, s)
		if err != nil {
			return err
		}

		if s.File == "-" {
			// For stdin, only write the new output
			newOutput := strings.TrimPrefix(output, lastOutput)
			_, err = fmt.Fprint(w, newOutput)
			lastOutput = output
		} else {
			// For file input, write the entire output
			_, err = fmt.Fprint(w, output)
		}
		return err
	}

	if s.File == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if err := processAndWrite(scanner.Text() + "\n"); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		data, err := os.ReadFile(s.File)
		if err != nil {
			return err
		}
		return processAndWrite(string(data))
	}

	return nil
}

func generateOutput(blocks []markdown.MarkdownBlock, s *ExtractMdSettings) (string, error) {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)

	isLanguageAllowed := func(lang string) bool {
		if len(s.AllowedLanguages) == 0 {
			return true
		}
		for _, allowed := range s.AllowedLanguages {
			if allowed == lang {
				return true
			}
		}
		return false
	}

	switch s.Output {
	case "concatenated":
		for _, block := range blocks {
			if block.Type == markdown.Code && isLanguageAllowed(block.Language) {
				if s.WithQuotes {
					_, _ = fmt.Fprintf(bw, "```%s\n%s\n```\n", block.Language, block.Content)
				} else {
					_, _ = fmt.Fprintln(bw, block.Content)
				}
				bw.Flush() // Flush after each write
			}
		}
	case "list":
		for _, block := range blocks {
			if block.Type == markdown.Code && isLanguageAllowed(block.Language) {
				_, _ = fmt.Fprintf(bw, "Language: %s\n", block.Language)
				if s.WithQuotes {
					_, _ = fmt.Fprintf(bw, "```%s\n%s\n```\n", block.Language, block.Content)
				} else {
					_, _ = fmt.Fprintln(bw, block.Content)
				}
				_, _ = fmt.Fprintln(bw, "---")
				bw.Flush() // Flush after each write
			}
		}
	case "yaml":
		filteredBlocks := make([]markdown.MarkdownBlock, 0)
		for _, block := range blocks {
			if block.Type == markdown.Code && isLanguageAllowed(block.Language) {
				filteredBlocks = append(filteredBlocks, block)
			}
		}
		yamlEncoder := yaml.NewEncoder(bw)
		err := yamlEncoder.Encode(filteredBlocks)
		if err != nil {
			return "", err
		}
		bw.Flush() // Flush after YAML encoding
	}

	bw.Flush()
	return buf.String(), nil
}

func main() {
	cmd, err := NewExtractMdCommand()
	cobra.CheckErr(err)

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	cobra.CheckErr(err)

	//rootCmd.AddCommand(cobraCmd)

	err = cobraCmd.Execute()
	cobra.CheckErr(err)
}
