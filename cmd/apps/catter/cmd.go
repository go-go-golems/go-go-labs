package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/clay/pkg/filefilter"
	"github.com/go-go-golems/go-go-labs/cmd/apps/catter/pkg"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
)

type CatterSettings struct {
	MaxTotalSize  int64    `glazed.parameter:"max-total-size"`
	Stats         []string `glazed.parameter:"stats"`
	List          bool     `glazed.parameter:"list"`
	Delimiter     string   `glazed.parameter:"delimiter"`
	MaxLines      int      `glazed.parameter:"max-lines"`
	MaxTokens     int      `glazed.parameter:"max-tokens"`
	PrintFilters  bool     `glazed.parameter:"print-filters"`
	FilterYAML    string   `glazed.parameter:"filter-yaml"`
	FilterProfile string   `glazed.parameter:"filter-profile"`
	Glazed        bool     `glazed.parameter:"glazed"`
	Paths         []string `glazed.parameter:"paths"`
}

type CatterCommand struct {
	*cmds.CommandDescription
}

func NewCatterCommand() (*CatterCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	fileFilterLayer, err := NewFileFilterParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create file filter parameter layer: %w", err)
	}

	return &CatterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"catter",
			cmds.WithShort("Print file contents with token counting for LLM context"),
			cmds.WithLong("A CLI tool to print file contents, recursively process directories, and count tokens for LLM context preparation."),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"max-total-size",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum total size of all files in bytes"),
					parameters.WithDefault(int64(10*1024*1024)),
				),
				parameters.NewParameterDefinition(
					"stats",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Types of statistics to show: overview, dir, full"),
					parameters.WithShortFlag("s"),
				),
				parameters.NewParameterDefinition(
					"list",
					parameters.ParameterTypeBool,
					parameters.WithHelp("List filenames only without printing content"),
					parameters.WithDefault(false),
					parameters.WithShortFlag("l"),
				),
				parameters.NewParameterDefinition(
					"delimiter",
					parameters.ParameterTypeString,
					parameters.WithHelp("Type of delimiter to use between files: default, xml, markdown, simple, begin-end"),
					parameters.WithDefault("default"),
					parameters.WithShortFlag("d"),
				),
				parameters.NewParameterDefinition(
					"max-lines",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of lines to print per file (0 for no limit)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"max-tokens",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of tokens to print per file (0 for no limit)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"print-filters",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print configured filters"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"filter-yaml",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML file containing filter configuration"),
				),
				parameters.NewParameterDefinition(
					"filter-profile",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the filter profile to use"),
				),
				parameters.NewParameterDefinition(
					"glazed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable Glazed structured output"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"paths",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Paths to process"),
					parameters.WithDefault([]string{"."}),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
				fileFilterLayer,
			),
		),
	}, nil
}

func (c *CatterCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &CatterSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return fmt.Errorf("error initializing settings: %w", err)
	}

	ff, err := CreateFileFilterFromSettings(parsedLayers)
	if err != nil {
		return fmt.Errorf("error creating file filter: %w", err)
	}

	if s.FilterYAML != "" {
		ff, err = filefilter.LoadFromFile(s.FilterYAML, s.FilterProfile)
		if err != nil {
			return fmt.Errorf("error loading filter configuration from YAML: %w", err)
		}
	} else {
		// Check for default .catter-filter.yaml in the current directory
		if _, err := os.Stat(".catter-filter.yaml"); err == nil {
			ff, err = filefilter.LoadFromFile(".catter-filter.yaml", s.FilterProfile)
			if err != nil {
				return fmt.Errorf("error loading default filter configuration: %w", err)
			}
		}
	}

	fileProcessorOptions := []pkg.FileProcessorOption{
		pkg.WithMaxTotalSize(s.MaxTotalSize),
		pkg.WithStatsTypes(s.Stats),
		pkg.WithListOnly(s.List),
		pkg.WithDelimiterType(s.Delimiter),
		pkg.WithMaxLines(s.MaxLines),
		pkg.WithMaxTokens(s.MaxTokens),
		pkg.WithFileFilter(ff),
	}
	if s.Glazed {
		fileProcessorOptions = append(fileProcessorOptions, pkg.WithProcessor(gp))
	}

	fp := pkg.NewFileProcessor(fileProcessorOptions...)

	if len(s.Paths) < 1 {
		s.Paths = append(s.Paths, ".")
	}

	fp.ProcessPaths(s.Paths)
	return nil
}
