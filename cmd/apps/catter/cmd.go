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
}

func runCatter(cmd *cobra.Command, args []string) {
	maxFileSize, _ := cmd.Flags().GetInt64("max-file-size")
	maxTotalSize, _ := cmd.Flags().GetInt64("max-total-size")
	includeExts, _ := cmd.Flags().GetStringSlice("include")
	excludeExts, _ := cmd.Flags().GetStringSlice("exclude")
	statsTypes, _ := cmd.Flags().GetStringSlice("stats")
	matchFilenameStrs, _ := cmd.Flags().GetStringSlice("match-filename")
	matchPathStrs, _ := cmd.Flags().GetStringSlice("match-path")
	listOnly, _ := cmd.Flags().GetBool("list")
	excludeDirs, _ := cmd.Flags().GetStringSlice("exclude-dirs")
	delimiterType, _ := cmd.Flags().GetString("delimiter")
	excludeMatchFilenameStrs, _ := cmd.Flags().GetStringSlice("exclude-match-filename")
	excludeMatchPathStrs, _ := cmd.Flags().GetStringSlice("exclude-match-path")
	maxLines, _ := cmd.Flags().GetInt("max-lines")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	disableGitIgnore, _ := cmd.Flags().GetBool("disable-gitignore")
	disableDefaultFilters, _ := cmd.Flags().GetBool("disable-default-filters")
	printFilters, _ := cmd.Flags().GetBool("print-filters")
	verbose, _ := cmd.Flags().GetBool("verbose")

	fileFilterOptions := []FileFilterOption{
		WithMaxFileSize(maxFileSize),
		WithIncludeExts(includeExts),
		WithExcludeExts(excludeExts),
		WithMatchFilenames(matchFilenameStrs),
		WithMatchPaths(matchPathStrs),
		WithExcludeDirs(excludeDirs),
		WithExcludeMatchFilenames(excludeMatchFilenameStrs),
		WithExcludeMatchPaths(excludeMatchPathStrs),
		WithDisableGitIgnore(disableGitIgnore),
		WithDisableDefaultFilters(disableDefaultFilters),
		WithVerbose(verbose),
	}

	if !disableGitIgnore {
		fileFilterOptions = append(fileFilterOptions, WithGitIgnoreFilter(initGitIgnoreFilter()))
	}

	fileFilter := NewFileFilter(fileFilterOptions...)

	if printFilters {
		fileFilter.PrintConfiguredFilters()
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
