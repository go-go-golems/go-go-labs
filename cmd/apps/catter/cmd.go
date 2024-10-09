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
}

func runCatter(cmd *cobra.Command, args []string) {
	fp := NewFileProcessor()

	fp.MaxFileSize, _ = cmd.Flags().GetInt64("max-file-size")
	fp.MaxTotalSize, _ = cmd.Flags().GetInt64("max-total-size")
	fp.IncludeExts, _ = cmd.Flags().GetStringSlice("include")
	fp.ExcludeExts, _ = cmd.Flags().GetStringSlice("exclude")
	fp.StatsTypes, _ = cmd.Flags().GetStringSlice("stats")
	fp.ListOnly, _ = cmd.Flags().GetBool("list")
	fp.ExcludeDirs, _ = cmd.Flags().GetStringSlice("exclude-dirs")
	fp.DisableGitIgnore, _ = cmd.Flags().GetBool("disable-gitignore")
	fp.DelimiterType, _ = cmd.Flags().GetString("delimiter")

	matchFilenameStrs, _ := cmd.Flags().GetStringSlice("match-filename")
	matchPathStrs, _ := cmd.Flags().GetStringSlice("match-path")
	excludeMatchFilenameStrs, _ := cmd.Flags().GetStringSlice("exclude-match-filename")
	excludeMatchPathStrs, _ := cmd.Flags().GetStringSlice("exclude-match-path")

	fp.MaxLines, _ = cmd.Flags().GetInt("max-lines")
	fp.MaxTokens, _ = cmd.Flags().GetInt("max-tokens")

	fp.MatchFilenames = compileRegexps(matchFilenameStrs)
	fp.MatchPaths = compileRegexps(matchPathStrs)
	fp.ExcludeMatchFilenames = compileRegexps(excludeMatchFilenameStrs)
	fp.ExcludeMatchPaths = compileRegexps(excludeMatchPathStrs)

	if !fp.DisableGitIgnore {
		fp.GitIgnoreFilter = initGitIgnoreFilter()
	}

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
			fmt.Fprintf(os.Stderr, "Invalid regex pattern: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "Error initializing gitignore filter from file: %v\n", err)
			os.Exit(1)
		}
		return gitIgnoreFilter
	}

	gitIgnoreFilter, err := gitignore.NewRepository(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing gitignore filter: %v\n", err)
		os.Exit(1)
	}
	return gitIgnoreFilter
}
