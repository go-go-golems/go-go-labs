package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
	"github.com/weaviate/tiktoken-go"
)

type CodePrinter struct {
	MaxFileSize           int64
	MaxTotalSize          int64
	IncludeExts           []string
	ExcludeExts           []string
	CurrentSize           int64
	TotalTokens           int
	FileCount             int
	TokenCounter          *tiktoken.Tiktoken
	TokenCounts           map[string]int
	StatsLevel            string
	MatchFilenames        []*regexp.Regexp
	MatchPaths            []*regexp.Regexp
	ListOnly              bool
	ExcludeDirs           []string
	GitIgnoreFilter       gitignore.GitIgnore
	DisableGitIgnore      bool
	DelimiterType         string
	DefaultExcludeExts    []string
	ExcludeMatchFilenames []*regexp.Regexp
	ExcludeMatchPaths     []*regexp.Regexp
}

type FlagResults struct {
	MaxFileSize              int64
	MaxTotalSize             int64
	IncludeExts              []string
	ExcludeExts              []string
	StatsLevel               string
	MatchFilenameStrs        []string
	MatchPathStrs            []string
	ListOnly                 bool
	ExcludeDirs              []string
	DisableGitIgnore         bool
	DelimiterType            string
	ExcludeMatchFilenameStrs []string
	ExcludeMatchPathStrs     []string
}

func NewCodePrinter() *CodePrinter {
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing tiktoken: %v\n", err)
		os.Exit(1)
	}

	return &CodePrinter{
		TokenCounter: tokenCounter,
		TokenCounts:  make(map[string]int),
		StatsLevel:   "none",
		ExcludeDirs:  []string{},
		DefaultExcludeExts: []string{
			".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff",
			".mp3", ".wav", ".ogg", ".flac",
			".mp4", ".avi", ".mov", ".wmv",
			".zip", ".tar", ".gz", ".rar",
			".exe", ".dll", ".so", ".dylib",
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
			".bin", ".dat", ".db", ".sqlite", ".ico", ".lock", ".tmp", ".woff", ".ttf", ".eot", ".svg",
		},
	}
}

func (cp *CodePrinter) Run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		args = append(args, ".")
	}

	for _, path := range args {
		cp.processPath(path)
		if cp.CurrentSize >= cp.MaxTotalSize {
			_, _ = fmt.Fprintf(os.Stderr, "Reached maximum total size limit of %d bytes\n", cp.MaxTotalSize)
			break
		}
	}

	if cp.StatsLevel != "none" {
		cp.printSummary()
	}
}

func (cp *CodePrinter) processPath(path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
		return
	}

	if fileInfo.IsDir() {
		cp.processDirectory(path)
	} else {
		if cp.shouldProcessFile(path, fileInfo) {
			cp.printFileContent(path, fileInfo)
		}
	}
}

func (cp *CodePrinter) processDirectory(dirPath string) {
	if strings.HasSuffix(dirPath, ".git") {
		return
	}
	for _, excludedDir := range cp.ExcludeDirs {
		if strings.Contains(dirPath, excludedDir) {
			return
		}
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dirPath, err)
		return
	}

	dirTokens := 0
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		cp.processPath(fullPath)
		dirTokens += cp.TokenCounts[fullPath]
		if cp.CurrentSize >= cp.MaxTotalSize {
			break
		}
	}
	cp.TokenCounts[dirPath] = dirTokens
}

func (cp *CodePrinter) shouldProcessFile(filePath string, fileInfo os.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check against default excluded extensions
	for _, excludedExt := range cp.DefaultExcludeExts {
		if ext == excludedExt {
			return false
		}
	}

	if fileInfo.Size() > cp.MaxFileSize {
		return false
	}

	if len(cp.IncludeExts) > 0 {
		included := false
		for _, includedExt := range cp.IncludeExts {
			if ext == strings.ToLower(includedExt) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, excludedExt := range cp.ExcludeExts {
		if ext == strings.ToLower(excludedExt) {
			return false
		}
	}

	if len(cp.MatchFilenames) > 0 || len(cp.MatchPaths) > 0 {
		filenameMatch := false
		pathMatch := false

		for _, re := range cp.MatchFilenames {
			if re.MatchString(filepath.Base(filePath)) {
				filenameMatch = true
				break
			}
		}

		for _, re := range cp.MatchPaths {
			if re.MatchString(filePath) {
				pathMatch = true
				break
			}
		}

		if !filenameMatch && !pathMatch {
			return false
		}
	}

	for _, re := range cp.ExcludeMatchFilenames {
		if re.MatchString(filepath.Base(filePath)) {
			return false
		}
	}

	for _, re := range cp.ExcludeMatchPaths {
		if re.MatchString(filePath) {
			return false
		}
	}

	if !cp.DisableGitIgnore && cp.GitIgnoreFilter.Ignore(filePath) {
		return false
	}

	return true
}

func (cp *CodePrinter) printFileContent(filePath string, _ os.FileInfo) {
	if cp.ListOnly {
		fmt.Println(filePath)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		return
	}

	if cp.CurrentSize+int64(len(content)) > cp.MaxTotalSize {
		remainingSize := cp.MaxTotalSize - cp.CurrentSize
		if remainingSize > 0 {
			content = content[:remainingSize]
		} else {
			return
		}
	}

	switch cp.DelimiterType {
	case "xml":
		fmt.Printf("<file name=\"%s\">\n<content>\n%s\n</content>\n</file>\n", filePath, string(content))
	case "markdown":
		fmt.Printf("## File: %s\n\n```\n%s\n```\n\n", filePath, string(content))
	case "simple":
		fmt.Printf("===\n\nFile: %s\n\n---\n\n%s\n\n===\n\n", filePath, string(content))
	default:
		fmt.Printf("=== BEGIN: %s ===\n%s\n=== END: %s ===\n\n", filePath, string(content), filePath)
	}

	cp.CurrentSize += int64(len(content))
	cp.FileCount++

	tokens := cp.TokenCounter.Encode(string(content), nil, nil)
	tokenCount := len(tokens)
	cp.TokenCounts[filePath] = tokenCount
	cp.TotalTokens += tokenCount
}

func (cp *CodePrinter) printSummary() {
	if cp.StatsLevel == "total" {
		_, _ = fmt.Fprintf(os.Stderr, "\nTotal tokens: %d\n", cp.TotalTokens)
	} else if cp.StatsLevel == "detailed" {
		_, _ = fmt.Fprintf(os.Stderr, "\nSummary:\n")
		_, _ = fmt.Fprintf(os.Stderr, "Total files processed: %d\n", cp.FileCount)
		_, _ = fmt.Fprintf(os.Stderr, "Total size processed: %d bytes\n", cp.CurrentSize)
		_, _ = fmt.Fprintf(os.Stderr, "Total tokens: %d\n", cp.TotalTokens)
		_, _ = fmt.Fprintf(os.Stderr, "\nToken counts per file/directory:\n")

		paths := make([]string, 0, len(cp.TokenCounts))
		for path := range cp.TokenCounts {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %d tokens\n", path, cp.TokenCounts[path])
		}
	}
}

func init() {
	CatterCmd.Flags().Int64("max-file-size", 1024*1024, "Maximum size of individual files in bytes")
	CatterCmd.Flags().Int64("max-total-size", 10*1024*1024, "Maximum total size of all files in bytes")
	CatterCmd.Flags().Bool("disable-gitignore", false, "Disable .gitignore filter")

	// Add shorthand flags
	CatterCmd.Flags().StringSliceP("include", "i", []string{}, "List of file extensions to include (e.g., .go,.js)")
	CatterCmd.Flags().StringSliceP("exclude", "e", []string{}, "List of file extensions to exclude (e.g., .exe,.dll)")
	CatterCmd.Flags().StringP("stats", "s", "none", "Level of statistics to show: none, total, or detailed")
	CatterCmd.Flags().StringSliceP("match-filename", "f", []string{}, "List of regular expressions to match filenames")
	CatterCmd.Flags().StringSliceP("match-path", "p", []string{}, "List of regular expressions to match full paths")
	CatterCmd.Flags().BoolP("list", "l", false, "List filenames only without printing content")
	CatterCmd.Flags().StringSliceP("exclude-dirs", "x", []string{}, "List of directories to exclude")
	CatterCmd.Flags().StringP("delimiter", "d", "default", "Type of delimiter to use between files: default, xml, markdown, simple, begin-end")
	CatterCmd.Flags().StringSliceP("exclude-match-filename", "F", []string{}, "List of regular expressions to exclude matching filenames")
	CatterCmd.Flags().StringSliceP("exclude-match-path", "P", []string{}, "List of regular expressions to exclude matching full paths")
}

func main() {
	if err := CatterCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var CatterCmd = &cobra.Command{
	Use:   "catter",
	Short: "Print file contents with token counting for LLM context",
	Long:  `A CLI tool to print file contents, recursively process directories, and count tokens for LLM context preparation.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagResults := &FlagResults{}

		// Populate the FlagResults struct
		flagResults.MaxFileSize, _ = cmd.Flags().GetInt64("max-file-size")
		flagResults.MaxTotalSize, _ = cmd.Flags().GetInt64("max-total-size")
		flagResults.IncludeExts, _ = cmd.Flags().GetStringSlice("include")
		flagResults.ExcludeExts, _ = cmd.Flags().GetStringSlice("exclude")
		flagResults.StatsLevel, _ = cmd.Flags().GetString("stats")
		flagResults.MatchFilenameStrs, _ = cmd.Flags().GetStringSlice("match-filename")
		flagResults.MatchPathStrs, _ = cmd.Flags().GetStringSlice("match-path")
		flagResults.ListOnly, _ = cmd.Flags().GetBool("list")
		flagResults.ExcludeDirs, _ = cmd.Flags().GetStringSlice("exclude-dirs")
		flagResults.DisableGitIgnore, _ = cmd.Flags().GetBool("disable-gitignore")
		flagResults.DelimiterType, _ = cmd.Flags().GetString("delimiter")
		flagResults.ExcludeMatchFilenameStrs, _ = cmd.Flags().GetStringSlice("exclude-match-filename")
		flagResults.ExcludeMatchPathStrs, _ = cmd.Flags().GetStringSlice("exclude-match-path")

		// Update CodePrinter with flag results
		cp := NewCodePrinter()
		cp.MaxFileSize = flagResults.MaxFileSize
		cp.MaxTotalSize = flagResults.MaxTotalSize
		cp.IncludeExts = flagResults.IncludeExts
		cp.ExcludeExts = flagResults.ExcludeExts
		cp.StatsLevel = flagResults.StatsLevel
		cp.ListOnly = flagResults.ListOnly
		cp.ExcludeDirs = flagResults.ExcludeDirs
		cp.DisableGitIgnore = flagResults.DisableGitIgnore
		cp.DelimiterType = flagResults.DelimiterType

		// Convert string flags to regular expressions
		for _, matchFilenameStr := range flagResults.MatchFilenameStrs {
			re, err := regexp.Compile(matchFilenameStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid match-filename regex: %v\n", err)
				os.Exit(1)
			}
			cp.MatchFilenames = append(cp.MatchFilenames, re)
		}

		for _, matchPathStr := range flagResults.MatchPathStrs {
			re, err := regexp.Compile(matchPathStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid match-path regex: %v\n", err)
				os.Exit(1)
			}
			cp.MatchPaths = append(cp.MatchPaths, re)
		}

		// Convert exclude match string flags to regular expressions
		for _, excludeMatchFilenameStr := range flagResults.ExcludeMatchFilenameStrs {
			re, err := regexp.Compile(excludeMatchFilenameStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid exclude-match-filename regex: %v\n", err)
				os.Exit(1)
			}
			cp.ExcludeMatchFilenames = append(cp.ExcludeMatchFilenames, re)
		}

		for _, excludeMatchPathStr := range flagResults.ExcludeMatchPathStrs {
			re, err := regexp.Compile(excludeMatchPathStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid exclude-match-path regex: %v\n", err)
				os.Exit(1)
			}
			cp.ExcludeMatchPaths = append(cp.ExcludeMatchPaths, re)
		}

		// Initialize gitignore filter if not disabled
		if !cp.DisableGitIgnore {
			if _, err := os.Stat(".gitignore"); err == nil {
				gitIgnoreFilter, err := gitignore.NewFromFile(".gitignore")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error initializing gitignore filter from file: %v\n", err)
					os.Exit(1)
				}
				cp.GitIgnoreFilter = gitIgnoreFilter
			} else {
				gitIgnoreFilter, err := gitignore.NewRepository(".")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error initializing gitignore filter: %v\n", err)
					os.Exit(1)
				}
				cp.GitIgnoreFilter = gitIgnoreFilter
			}
		}

		cp.Run(cmd, args)
	},
}
