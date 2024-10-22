package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/go-go-golems/clay/pkg/filefilter"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"

	"github.com/denormal/go-gitignore"
)

type CatterSettings struct {
	MaxFileSize           int64    `glazed.parameter:"max-file-size"`
	MaxTotalSize          int64    `glazed.parameter:"max-total-size"`
	DisableGitIgnore      bool     `glazed.parameter:"disable-gitignore"`
	DisableDefaultFilters bool     `glazed.parameter:"disable-default-filters"`
	Include               []string `glazed.parameter:"include"`
	Exclude               []string `glazed.parameter:"exclude"`
	Stats                 []string `glazed.parameter:"stats"`
	MatchFilename         []string `glazed.parameter:"match-filename"`
	MatchPath             []string `glazed.parameter:"match-path"`
	List                  bool     `glazed.parameter:"list"`
	ExcludeDirs           []string `glazed.parameter:"exclude-dirs"`
	Delimiter             string   `glazed.parameter:"delimiter"`
	ExcludeMatchFilename  []string `glazed.parameter:"exclude-match-filename"`
	ExcludeMatchPath      []string `glazed.parameter:"exclude-match-path"`
	MaxLines              int      `glazed.parameter:"max-lines"`
	MaxTokens             int      `glazed.parameter:"max-tokens"`
	PrintFilters          bool     `glazed.parameter:"print-filters"`
	Verbose               bool     `glazed.parameter:"verbose"`
	FilterYAML            string   `glazed.parameter:"filter-yaml"`
	PrintFilterYAML       bool     `glazed.parameter:"print-filter-yaml"`
	FilterBinary          bool     `glazed.parameter:"filter-binary"`
	FilterProfile         string   `glazed.parameter:"filter-profile"`
	Paths                 []string `glazed.parameter:"paths"`
}

type CatterCommand struct {
	*cmds.CommandDescription
}

func NewCatterCommand() (*CatterCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &CatterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"catter",
			cmds.WithShort("Print file contents with token counting for LLM context"),
			cmds.WithLong("A CLI tool to print file contents, recursively process directories, and count tokens for LLM context preparation."),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"max-file-size",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum size of individual files in bytes"),
					parameters.WithDefault(int64(1024*1024)),
				),
				parameters.NewParameterDefinition(
					"max-total-size",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum total size of all files in bytes"),
					parameters.WithDefault(int64(10*1024*1024)),
				),
				parameters.NewParameterDefinition(
					"disable-gitignore",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Disable .gitignore filter"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"disable-default-filters",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Disable default file and directory filters"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"include",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of file extensions to include (e.g., .go,.js)"),
				),
				parameters.NewParameterDefinition(
					"exclude",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of file extensions to exclude (e.g., .exe,.dll)"),
				),
				parameters.NewParameterDefinition(
					"stats",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Types of statistics to show: overview, dir, full"),
				),
				parameters.NewParameterDefinition(
					"match-filename",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of regular expressions to match filenames"),
				),
				parameters.NewParameterDefinition(
					"match-path",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of regular expressions to match full paths"),
				),
				parameters.NewParameterDefinition(
					"list",
					parameters.ParameterTypeBool,
					parameters.WithHelp("List filenames only without printing content"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"exclude-dirs",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of directories to exclude"),
				),
				parameters.NewParameterDefinition(
					"delimiter",
					parameters.ParameterTypeString,
					parameters.WithHelp("Type of delimiter to use between files: default, xml, markdown, simple, begin-end"),
					parameters.WithDefault("default"),
				),
				parameters.NewParameterDefinition(
					"exclude-match-filename",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of regular expressions to exclude matching filenames"),
				),
				parameters.NewParameterDefinition(
					"exclude-match-path",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of regular expressions to exclude matching full paths"),
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
					"verbose",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable verbose logging of filtered/unfiltered paths"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"filter-yaml",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to YAML file containing filter configuration"),
				),
				parameters.NewParameterDefinition(
					"print-filter-yaml",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print the current filter configuration as YAML and exit"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"filter-binary",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Filter out binary files"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"filter-profile",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the filter profile to use"),
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
			),
		),
	}, nil
}

func (c *CatterCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &CatterSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return fmt.Errorf("error initializing settings: %w", err)
	}

	var fileFilter *filefilter.FileFilter

	if s.FilterYAML != "" {
		fileFilter, err = filefilter.LoadFromFile(s.FilterYAML, s.FilterProfile)
		if err != nil {
			return fmt.Errorf("error loading filter configuration from YAML: %w", err)
		}
	} else {
		// Check for default .catter-filter.yaml in the current directory
		if _, err := os.Stat(".catter-filter.yaml"); err == nil {
			fileFilter, err = filefilter.LoadFromFile(".catter-filter.yaml", s.FilterProfile)
			if err != nil {
				return fmt.Errorf("error loading default filter configuration: %w", err)
			}
		} else {
			fileFilter = filefilter.NewFileFilter()
		}
	}

	applyFlagOverrides(s, fileFilter)

	if s.PrintFilterYAML {
		yamlData, err := fileFilter.ToYAML()
		if err != nil {
			return fmt.Errorf("error generating YAML: %w", err)
		}
		fmt.Println(string(yamlData))
		return nil
	}

	var gitIgnoreFilter gitignore.GitIgnore
	if !s.DisableGitIgnore {
		gitIgnoreFilter = initGitIgnoreFilter()
	}

	fileProcessorOptions := []FileProcessorOption{
		WithMaxTotalSize(s.MaxTotalSize),
		WithStatsTypes(s.Stats),
		WithListOnly(s.List),
		WithDelimiterType(s.Delimiter),
		WithMaxLines(s.MaxLines),
		WithMaxTokens(s.MaxTokens),
		WithGitIgnoreFilter(gitIgnoreFilter),
		WithFileFilter(fileFilter),
	}

	fp := NewFileProcessor(fileProcessorOptions...)

	if len(s.Paths) < 1 {
		s.Paths = append(s.Paths, ".")
	}

	fp.ProcessPaths(s.Paths)
	return nil
}

func applyFlagOverrides(s *CatterSettings, ff *filefilter.FileFilter) {
	if s.MaxFileSize != 0 {
		ff.MaxFileSize = s.MaxFileSize
	}
	if len(s.Include) > 0 {
		ff.IncludeExts = s.Include
	}
	if len(s.Exclude) > 0 {
		ff.ExcludeExts = s.Exclude
	}
	if len(s.MatchFilename) > 0 {
		ff.MatchFilenames = compileRegexps(s.MatchFilename)
	}
	if len(s.MatchPath) > 0 {
		ff.MatchPaths = compileRegexps(s.MatchPath)
	}
	if len(s.ExcludeDirs) > 0 {
		ff.ExcludeDirs = s.ExcludeDirs
	}
	if len(s.ExcludeMatchFilename) > 0 {
		ff.ExcludeMatchFilenames = compileRegexps(s.ExcludeMatchFilename)
	}
	if len(s.ExcludeMatchPath) > 0 {
		ff.ExcludeMatchPaths = compileRegexps(s.ExcludeMatchPath)
	}
	ff.DisableGitIgnore = s.DisableGitIgnore
	ff.DisableDefaultFilters = s.DisableDefaultFilters
	ff.Verbose = s.Verbose
	ff.FilterBinaryFiles = s.FilterBinary
}

func initGitIgnoreFilter() gitignore.GitIgnore {
	if _, err := os.Stat(".gitignore"); err == nil {
		gitIgnoreFilter, err := gitignore.NewFromFile(".gitignore")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error initializing gitignore filter from file: %v\n", err)
			os.Exit(1)
		}
		return gitIgnoreFilter
	}

	gitIgnoreFilter, err := gitignore.NewRepository(".")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing gitignore filter: %v\n", err)
		os.Exit(1)
	}
	return gitIgnoreFilter
}

func compileRegexps(patterns []string) []*regexp.Regexp {
	var regexps []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid regex pattern: %v\n", err)
			os.Exit(1)
		}
		regexps = append(regexps, re)
	}
	return regexps
}
