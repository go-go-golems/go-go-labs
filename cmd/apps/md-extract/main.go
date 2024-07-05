package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"io"
	"os"

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
	Output     string `glazed.parameter:"output"`
	WithQuotes bool   `glazed.parameter:"with-quotes"`
	File       string `glazed.parameter:"file"`
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
	s := &ExtractMdSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	var input []byte
	if s.File == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(s.File)
	}
	if err != nil {
		return err
	}

	blocks := markdown.ExtractAllBlocks(string(input))

	switch s.Output {
	case "concatenated":
		for _, block := range blocks {
			if block.Type == markdown.Code {
				if s.WithQuotes {
					_, _ = fmt.Fprintf(w, "```%s\n%s\n```\n", block.Language, block.Content)
				} else {
					_, _ = fmt.Fprintln(w, block.Content)
				}
			}
		}
	case "list":
		for _, block := range blocks {
			if block.Type == markdown.Code {
				_, _ = fmt.Fprintf(w, "Language: %s\n", block.Language)
				if s.WithQuotes {
					_, _ = fmt.Fprintf(w, "```%s\n%s\n```\n", block.Language, block.Content)
				} else {
					_, _ = fmt.Fprintln(w, block.Content)
				}
				_, _ = fmt.Fprintln(w, "---")
			}
		}
	case "yaml":
		err := yaml.NewEncoder(w).Encode(blocks)
		if err != nil {
			return err
		}
	}

	return nil
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
