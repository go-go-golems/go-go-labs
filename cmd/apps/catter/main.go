package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaviate/tiktoken-go"
)

type CodePrinter struct {
	MaxFileSize  int64
	MaxTotalSize int64
	IncludeExts  []string
	ExcludeExts  []string
	CurrentSize  int64
	TotalTokens  int
	FileCount    int
	TokenCounter *tiktoken.Tiktoken
	TokenCounts  map[string]int
	StatsLevel   string
	GlobPattern  string
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
	}
}

func (cp *CodePrinter) Run(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Please provide at least one file or directory path")
		return
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
	files, err := os.ReadDir(dirPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dirPath, err)
		return
	}

	dirTokens := 0
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		if cp.GlobPattern != "" {
			matched, err := doublestar.PathMatch(cp.GlobPattern, fullPath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error matching glob pattern: %v\n", err)
				return
			}
			if !matched {
				continue
			}
		}
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

	return true
}

func (cp *CodePrinter) printFileContent(filePath string, _ os.FileInfo) {
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

	fmt.Printf("---\n\nFile: %s\n\n--\n\n%s\n\n---\n\n", filePath, string(content))
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

func main() {
	cp := NewCodePrinter()

	rootCmd := &cobra.Command{
		Use:   "codeprinter",
		Short: "Print file contents with token counting for LLM context",
		Long:  `A CLI tool to print file contents, recursively process directories, and count tokens for LLM context preparation.`,
		Run:   cp.Run,
	}

	rootCmd.Flags().Int64Var(&cp.MaxFileSize, "max-file-size", 1024*1024, "Maximum size of individual files in bytes")
	rootCmd.Flags().Int64Var(&cp.MaxTotalSize, "max-total-size", 10*1024*1024, "Maximum total size of all files in bytes")
	rootCmd.Flags().StringSliceVar(&cp.IncludeExts, "include", []string{}, "List of file extensions to include (e.g., .go,.js)")
	rootCmd.Flags().StringSliceVar(&cp.ExcludeExts, "exclude", []string{}, "List of file extensions to exclude (e.g., .exe,.dll)")
	rootCmd.Flags().StringVar(&cp.StatsLevel, "stats", "none", "Level of statistics to show: none, total, or detailed")
	rootCmd.Flags().StringVar(&cp.GlobPattern, "glob", "", "Glob pattern to match files and directories (e.g., **/*.go)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
