package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaviate/tiktoken-go"
)

type FileProcessor struct {
	MaxTotalSize  int64
	CurrentSize   int64
	TotalTokens   int
	FileCount     int
	TokenCounter  *tiktoken.Tiktoken
	TokenCounts   map[string]int
	StatsTypes    []string
	ListOnly      bool
	DelimiterType string
	MaxLines      int
	MaxTokens     int
	Filter        *FileFilter
}

func NewFileProcessor() *FileProcessor {
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing tiktoken: %v\n", err)
		os.Exit(1)
	}

	return &FileProcessor{
		TokenCounter: tokenCounter,
		TokenCounts:  make(map[string]int),
		StatsTypes:   []string{},
		MaxLines:     0,
		MaxTokens:    0,
		Filter: &FileFilter{
			ExcludeDirs: []string{".git", ".history", ".idea"},
			DefaultExcludeExts: []string{
				".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff",
				".mp3", ".wav", ".ogg", ".flac",
				".mp4", ".avi", ".mov", ".wmv",
				".zip", ".tar", ".gz", ".rar",
				".exe", ".dll", ".so", ".dylib",
				".pdf", ".doc", ".docx", ".xls", ".xlsx",
				".bin", ".dat", ".db", ".sqlite",
				".woff", ".ttf", ".eot", ".svg", ".webp", ".woff2",
			},
		},
	}
}

func (fp *FileProcessor) ProcessPaths(paths []string) {
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
	stats, err := ComputeStats(paths, fp.Filter)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error computing stats: %v\n", err)
		return
	}

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

	PrintStats(stats, config)
}

func (fp *FileProcessor) processPath(path string) {
	if fp.Filter.FilterPath(path) {
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
		if fp.Filter.FilterPath(fullPath) {
			fp.processPath(fullPath)
			dirTokens += fp.TokenCounts[fullPath]
			if fp.CurrentSize >= fp.MaxTotalSize {
				break
			}
		}
	}
	fp.TokenCounts[dirPath] = dirTokens
}

func (fp *FileProcessor) printFileContent(filePath string, _ os.FileInfo) {
	if fp.ListOnly {
		fmt.Println(filePath)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		return
	}

	if fp.CurrentSize+int64(len(content)) > fp.MaxTotalSize {
		remainingSize := fp.MaxTotalSize - fp.CurrentSize
		if remainingSize > 0 {
			content = content[:remainingSize]
		} else {
			return
		}
	}

	tokens := fp.TokenCounter.Encode(string(content), nil, nil)
	tokenCount := len(tokens)
	fp.TokenCounts[filePath] = tokenCount
	fp.TotalTokens += tokenCount

	// Apply max lines and max tokens limits
	uintTokens := make([]uint, len(tokens))
	for i, t := range tokens {
		uintTokens[i] = uint(t)
	}
	limitedContent := fp.applyLimits(content, uintTokens)

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

	fp.CurrentSize += int64(len(content))
	fp.FileCount++
}

func (fp *FileProcessor) applyLimits(content []byte, tokens []uint) string {
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
