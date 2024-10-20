package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
)

var CatterCmd = &cobra.Command{
	Use:   "catter",
	Short: "Print file contents with token counting for LLM context",
	Long:  `A CLI tool to print file contents, recursively process directories, and count tokens for LLM context preparation.`,
	Run:   runCatter,
}

func init() {
	CatterCmd.Flags().Int64("max-file-size", 1024*1024, "Maximum size of individual files in bytes")
	CatterCmd.Flags().Int64("max-total-size", 10*1024*1024, "Maximum total size of all files in bytes")
	CatterCmd.Flags().Bool("disable-gitignore", false, "Disable .gitignore filter")
	CatterCmd.Flags().Bool("disable-default-filters", false, "Disable default file and directory filters")

	CatterCmd.Flags().StringSliceP("include", "i", []string{}, "List of file extensions to include (e.g., .go,.js)")
	CatterCmd.Flags().StringSliceP("exclude", "e", []string{}, "List of file extensions to exclude (e.g., .exe,.dll)")
	CatterCmd.Flags().StringSliceP("stats", "s", []string{}, "Types of statistics to show: overview, dir, full")
	CatterCmd.Flags().StringSliceP("match-filename", "f", []string{}, "List of regular expressions to match filenames")
	CatterCmd.Flags().StringSliceP("match-path", "p", []string{}, "List of regular expressions to match full paths")
	CatterCmd.Flags().BoolP("list", "l", false, "List filenames only without printing content")
	CatterCmd.Flags().StringSliceP("exclude-dirs", "x", []string{}, "List of directories to exclude")
	CatterCmd.Flags().StringP("delimiter", "d", "default", "Type of delimiter to use between files: default, xml, markdown, simple, begin-end")
	CatterCmd.Flags().StringSliceP("exclude-match-filename", "F", []string{}, "List of regular expressions to exclude matching filenames")
	CatterCmd.Flags().StringSliceP("exclude-match-path", "P", []string{}, "List of regular expressions to exclude matching full paths")

	CatterCmd.Flags().Int("max-lines", 0, "Maximum number of lines to print per file (0 for no limit)")
	CatterCmd.Flags().Int("max-tokens", 0, "Maximum number of tokens to print per file (0 for no limit)")
	CatterCmd.Flags().Bool("print-filters", false, "Print configured filters")
	CatterCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging of filtered/unfiltered paths")
	CatterCmd.Flags().String("filter-yaml", "", "Path to YAML file containing filter configuration")
	CatterCmd.Flags().String("filter-profile", "", "Name of the filter profile to use from the YAML configuration")
	CatterCmd.Flags().Bool("print-filter-yaml", false, "Print the current filter configuration as YAML and exit")
}

func runCatter(cmd *cobra.Command, args []string) {
	maxTotalSize, _ := cmd.Flags().GetInt64("max-total-size")
	statsTypes, _ := cmd.Flags().GetStringSlice("stats")
	listOnly, _ := cmd.Flags().GetBool("list")
	delimiterType, _ := cmd.Flags().GetString("delimiter")
	maxLines, _ := cmd.Flags().GetInt("max-lines")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	filterYAMLPath, _ := cmd.Flags().GetString("filter-yaml")
	filterProfile, _ := cmd.Flags().GetString("filter-profile")
	printFilterYAML, _ := cmd.Flags().GetBool("print-filter-yaml")

	fileFilter, err := loadFileFilter(filterYAMLPath, filterProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading filter configuration: %v\n", err)
		os.Exit(1)
	}

	applyFlagOverrides(cmd, fileFilter)

	if printFilterYAML {
		yamlData, err := fileFilter.ToYAML()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating YAML: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(yamlData))
		return
	}

	fileProcessorOptions := []FileProcessorOption{
		WithMaxTotalSize(maxTotalSize),
		WithStatsTypes(statsTypes),
		WithListOnly(listOnly),
		WithDelimiterType(delimiterType),
		WithMaxLines(maxLines),
		WithMaxTokens(maxTokens),
		WithFileFilter(fileFilter),
	}

	fp := NewFileProcessor(fileProcessorOptions...)

	if len(args) < 1 {
		args = append(args, ".")
	}

	fp.ProcessPaths(args)
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

func applyFlagOverrides(cmd *cobra.Command, ff *FileFilter) {
	if maxFileSize, _ := cmd.Flags().GetInt64("max-file-size"); maxFileSize != 0 {
		ff.MaxFileSize = maxFileSize
	}
	if includeExts, _ := cmd.Flags().GetStringSlice("include"); len(includeExts) > 0 {
		ff.IncludeExts = includeExts
	}
	if excludeExts, _ := cmd.Flags().GetStringSlice("exclude"); len(excludeExts) > 0 {
		ff.ExcludeExts = excludeExts
	}
	if matchFilenameStrs, _ := cmd.Flags().GetStringSlice("match-filename"); len(matchFilenameStrs) > 0 {
		ff.MatchFilenames = compileRegexps(matchFilenameStrs)
	}
	if matchPathStrs, _ := cmd.Flags().GetStringSlice("match-path"); len(matchPathStrs) > 0 {
		ff.MatchPaths = compileRegexps(matchPathStrs)
	}
	if excludeDirs, _ := cmd.Flags().GetStringSlice("exclude-dirs"); len(excludeDirs) > 0 {
		ff.ExcludeDirs = excludeDirs
	}
	if excludeMatchFilenameStrs, _ := cmd.Flags().GetStringSlice("exclude-match-filename"); len(excludeMatchFilenameStrs) > 0 {
		ff.ExcludeMatchFilenames = compileRegexps(excludeMatchFilenameStrs)
	}
	if excludeMatchPathStrs, _ := cmd.Flags().GetStringSlice("exclude-match-path"); len(excludeMatchPathStrs) > 0 {
		ff.ExcludeMatchPaths = compileRegexps(excludeMatchPathStrs)
	}
	if disableGitIgnore, _ := cmd.Flags().GetBool("disable-gitignore"); disableGitIgnore {
		ff.DisableGitIgnore = disableGitIgnore
	}
	if disableDefaultFilters, _ := cmd.Flags().GetBool("disable-default-filters"); disableDefaultFilters {
		ff.DisableDefaultFilters = disableDefaultFilters
	}
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		ff.Verbose = verbose
	}
}

func loadFileFilter(filterYAMLPath, filterProfile string) (*FileFilter, error) {
	if filterYAMLPath == "" {
		if filterProfile != "" {
			return nil, fmt.Errorf("filter profile specified but no filter YAML file provided")
		}
		return NewFileFilter(), nil
	}

	config, err := LoadFromFile(filterYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("error loading filter configuration from YAML: %v", err)
	}

	if filterProfile == "" {
		return config, nil
	}

	profileFilter, ok := config.Profiles[filterProfile]
	if !ok {
		return nil, fmt.Errorf("specified filter profile '%s' not found in the configuration", filterProfile)
	}

	return profileFilter, nil
}
