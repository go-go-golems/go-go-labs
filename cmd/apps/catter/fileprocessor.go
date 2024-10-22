package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/go-go-golems/clay/pkg/filefilter"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/weaviate/tiktoken-go"
)

type FileProcessor struct {
	MaxTotalSize    int64
	CurrentSize     int64
	TotalTokens     int
	FileCount       int
	TokenCounter    *tiktoken.Tiktoken
	TokenCounts     map[string]int
	StatsTypes      []string
	ListOnly        bool
	DelimiterType   string
	MaxLines        int
	MaxTokens       int
	Filter          *filefilter.FileFilter
	GitIgnoreFilter gitignore.GitIgnore
	PrintFilters    bool
	Processor       *middlewares.TableProcessor
	Stats           *Stats
}

type FileProcessorOption func(*FileProcessor)

func NewFileProcessor(options ...FileProcessorOption) *FileProcessor {
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing tiktoken: %v\n", err)
		os.Exit(1)
	}

	fp := &FileProcessor{
		TokenCounter:    tokenCounter,
		TokenCounts:     make(map[string]int),
		StatsTypes:      []string{},
		MaxLines:        0,
		MaxTokens:       0,
		GitIgnoreFilter: nil,
		PrintFilters:    false,
	}

	for _, option := range options {
		option(fp)
	}

	return fp
}

func WithMaxTotalSize(size int64) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.MaxTotalSize = size
	}
}

func WithStatsTypes(types []string) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.StatsTypes = types
	}
}

func WithListOnly(listOnly bool) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.ListOnly = listOnly
	}
}

func WithDelimiterType(delimiterType string) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.DelimiterType = delimiterType
	}
}

func WithMaxLines(maxLines int) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.MaxLines = maxLines
	}
}

func WithMaxTokens(maxTokens int) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.MaxTokens = maxTokens
	}
}

func WithFileFilter(filter *filefilter.FileFilter) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.Filter = filter
		if fp.GitIgnoreFilter != nil {
			fp.Filter.GitIgnoreFilter = fp.GitIgnoreFilter
		}
	}
}

func WithGitIgnoreFilter(gitIgnoreFilter gitignore.GitIgnore) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.GitIgnoreFilter = gitIgnoreFilter
	}
}

func WithPrintFilters(printFilters bool) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.PrintFilters = printFilters
	}
}

func WithProcessor(processor *middlewares.TableProcessor) FileProcessorOption {
	return func(fp *FileProcessor) {
		fp.Processor = processor
	}
}

func (fp *FileProcessor) ProcessPaths(paths []string) {
	if fp.PrintFilters {
		fp.printConfiguredFilters()
		return
	}

	fp.Stats = NewStats()
	err := fp.Stats.ComputeStats(paths, fp.Filter)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error computing stats: %v\n", err)
		return
	}

	if len(fp.StatsTypes) > 0 {
		fp.processStats(paths)
	} else {
		for _, path := range paths {
			fp.processPath(path)
			if fp.CurrentSize >= fp.MaxTotalSize {
				_, _ = fmt.Fprintf(os.Stderr, "Reached maximum total size limit of %d bytes\n", fp.MaxTotalSize)
				break
			}
		}
	}
}

func (fp *FileProcessor) processStats(paths []string) {
	config := Config{}
	for _, statType := range fp.StatsTypes {
		switch strings.ToLower(statType) {
		case "overview":
			config.OutputFlags |= OutputOverview
		case "dir":
			config.OutputFlags |= OutputDirStructure
		case "full":
			config.OutputFlags |= OutputFullStructure
		default:
			_, _ = fmt.Fprintf(os.Stderr, "Unknown stat type: %s\n", statType)
		}
	}

	err := fp.Stats.PrintStats(config, fp.Processor)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error printing stats: %v\n", err)
	}
}

func (fp *FileProcessor) processPath(path string) {
	if fp.Filter == nil || fp.Filter.FilterPath(path) {
		if fileInfo, err := os.Stat(path); err == nil {
			if fileInfo.IsDir() {
				fp.processDirectory(path)
			} else {
				fp.printFileContent(path, fileInfo)
			}
		}
	}
}

func (fp *FileProcessor) processDirectory(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dirPath, err)
		return
	}

	dirTokens := 0
	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		if fp.Filter != nil && fp.Filter.FilterPath(fullPath) {
			fp.processPath(fullPath)
			dirTokens += fp.TokenCounts[fullPath]
			if fp.CurrentSize >= fp.MaxTotalSize {
				break
			}
		}
	}
	fp.TokenCounts[dirPath] = dirTokens
}

func (fp *FileProcessor) printFileContent(filePath string, fileInfo os.FileInfo) {
	if fp.ListOnly {
		fmt.Println(filePath)
		return
	}

	fileStats, ok := fp.Stats.GetStats(filePath)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Stats not found for file %s\n", filePath)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		return
	}

	if fp.CurrentSize+fileStats.Size > fp.MaxTotalSize {
		remainingSize := fp.MaxTotalSize - fp.CurrentSize
		if remainingSize > 0 {
			content = content[:remainingSize]
		} else {
			return
		}
	}

	fp.TokenCounts[filePath] = fileStats.TokenCount
	fp.TotalTokens += fileStats.TokenCount

	// Apply max lines and max tokens limits
	limitedContent := fp.applyLimits(content, fileStats)

	if fp.Processor == nil {
		switch fp.DelimiterType {
		case "xml":
			fmt.Printf("<file name=\"%s\">\n<content>\n%s\n</content>\n</file>\n", filePath, limitedContent)
		case "markdown":
			fmt.Printf("## File: %s\n\n```\n%s\n```\n\n", filePath, limitedContent)
		case "simple":
			fmt.Printf("===\n\nFile: %s\n\n---\n\n%s\n\n===\n\n", filePath, limitedContent)
		default:
			fmt.Printf("=== BEGIN: %s ===\n%s\n=== END: %s ===\n\n", filePath, limitedContent, filePath)
		}
	}

	fp.CurrentSize += fileStats.Size
	fp.FileCount++

	if fp.Processor != nil {
		ctx := context.Background()
		err := fp.Processor.AddRow(ctx, types.NewRow(
			types.MRP("Path", filePath),
			types.MRP("Size", fileStats.Size),
			types.MRP("TokenCount", fileStats.TokenCount),
			types.MRP("LineCount", fileStats.LineCount),
			types.MRP("Content", limitedContent),
		))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error adding row to processor: %v\n", err)
		}
	}
}

func (fp *FileProcessor) applyLimits(content []byte, fileStats FileStats) string {
	if fp.MaxLines == 0 && fp.MaxTokens == 0 {
		return string(content)
	}

	var limitedContent bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineCount := 0
	tokenCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineTokens := fp.TokenCounter.Encode(line, nil, nil)

		if fp.MaxLines > 0 && lineCount >= fp.MaxLines {
			break
		}

		if fp.MaxTokens > 0 && tokenCount+len(lineTokens) > fp.MaxTokens {
			remainingTokens := fp.MaxTokens - tokenCount
			if remainingTokens > 0 {
				decodedLine := fp.TokenCounter.Decode(lineTokens[:remainingTokens])
				limitedContent.WriteString(decodedLine)
			}
			break
		}

		limitedContent.WriteString(line + "\n")
		lineCount++
		tokenCount += len(lineTokens)
	}

	return limitedContent.String()
}

func (fp *FileProcessor) printConfiguredFilters() {
	fmt.Println("Configured Filters:")
	fmt.Println("-------------------")

	if fp.Filter == nil {
		fmt.Println("No filters configured.")
		return
	}

	fmt.Printf("Max File Size: %d bytes\n", fp.Filter.MaxFileSize)
	fmt.Printf("Disable Default Filters: %v\n", fp.Filter.DisableDefaultFilters)
	fmt.Printf("Disable GitIgnore: %v\n", fp.Filter.DisableGitIgnore)
	fmt.Printf("Filter Binary Files: %v\n", fp.Filter.FilterBinaryFiles)
	fmt.Printf("Verbose: %v\n", fp.Filter.Verbose)

	printStringList("Include Extensions", fp.Filter.IncludeExts)
	printStringList("Exclude Extensions", fp.Filter.ExcludeExts)
	printStringList("Exclude Directories", fp.Filter.ExcludeDirs)

	printRegexpList("Match Filenames", fp.Filter.MatchFilenames)
	printRegexpList("Match Paths", fp.Filter.MatchPaths)
	printRegexpList("Exclude Match Filenames", fp.Filter.ExcludeMatchFilenames)
	printRegexpList("Exclude Match Paths", fp.Filter.ExcludeMatchPaths)

	fmt.Println("\nFile Processor Settings:")
	fmt.Printf("Max Total Size: %d bytes\n", fp.MaxTotalSize)
	fmt.Printf("Max Lines: %d\n", fp.MaxLines)
	fmt.Printf("Max Tokens: %d\n", fp.MaxTokens)
	printStringList("Stats Types", fp.StatsTypes)
	fmt.Printf("List Only: %v\n", fp.ListOnly)
	fmt.Printf("Delimiter Type: %s\n", fp.DelimiterType)
}

func printStringList(name string, list []string) {
	if len(list) > 0 {
		fmt.Printf("%s: %s\n", name, strings.Join(list, ", "))
	}
}

func printRegexpList(name string, list []*regexp.Regexp) {
	if len(list) > 0 {
		patterns := make([]string, len(list))
		for i, re := range list {
			patterns[i] = re.String()
		}
		fmt.Printf("%s: %s\n", name, strings.Join(patterns, ", "))
	}
}
