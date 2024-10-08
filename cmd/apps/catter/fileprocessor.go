package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/weaviate/tiktoken-go"
)

type FileProcessor struct {
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

func NewFileProcessor() *FileProcessor {
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing tiktoken: %v\n", err)
		os.Exit(1)
	}

	return &FileProcessor{
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
			".bin", ".dat", ".db", ".sqlite",
		},
	}
}

func (fp *FileProcessor) ProcessPaths(paths []string) {
	for _, path := range paths {
		fp.processPath(path)
		if fp.CurrentSize >= fp.MaxTotalSize {
			_, _ = fmt.Fprintf(os.Stderr, "Reached maximum total size limit of %d bytes\n", fp.MaxTotalSize)
			break
		}
	}

	if fp.StatsLevel != "none" {
		fp.printSummary()
	}
}

func (fp *FileProcessor) processPath(path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
		return
	}

	if fileInfo.IsDir() {
		fp.processDirectory(path)
	} else {
		if fp.shouldProcessFile(path, fileInfo) {
			fp.printFileContent(path, fileInfo)
		}
	}
}

func (fp *FileProcessor) processDirectory(dirPath string) {
	if strings.HasSuffix(dirPath, ".git") {
		return
	}
	for _, excludedDir := range fp.ExcludeDirs {
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
		fp.processPath(fullPath)
		dirTokens += fp.TokenCounts[fullPath]
		if fp.CurrentSize >= fp.MaxTotalSize {
			break
		}
	}
	fp.TokenCounts[dirPath] = dirTokens
}

func (fp *FileProcessor) shouldProcessFile(filePath string, fileInfo os.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check against default excluded extensions
	for _, excludedExt := range fp.DefaultExcludeExts {
		if ext == excludedExt {
			return false
		}
	}

	if fileInfo.Size() > fp.MaxFileSize {
		return false
	}

	if len(fp.IncludeExts) > 0 {
		included := false
		for _, includedExt := range fp.IncludeExts {
			if ext == strings.ToLower(includedExt) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, excludedExt := range fp.ExcludeExts {
		if ext == strings.ToLower(excludedExt) {
			return false
		}
	}

	if len(fp.MatchFilenames) > 0 || len(fp.MatchPaths) > 0 {
		filenameMatch := false
		pathMatch := false

		for _, re := range fp.MatchFilenames {
			if re.MatchString(filepath.Base(filePath)) {
				filenameMatch = true
				break
			}
		}

		for _, re := range fp.MatchPaths {
			if re.MatchString(filePath) {
				pathMatch = true
				break
			}
		}

		if !filenameMatch && !pathMatch {
			return false
		}
	}

	for _, re := range fp.ExcludeMatchFilenames {
		if re.MatchString(filepath.Base(filePath)) {
			return false
		}
	}

	for _, re := range fp.ExcludeMatchPaths {
		if re.MatchString(filePath) {
			return false
		}
	}

	if !fp.DisableGitIgnore && fp.GitIgnoreFilter != nil && fp.GitIgnoreFilter.Ignore(filePath) {
		return false
	}

	return true
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

	switch fp.DelimiterType {
	case "xml":
		fmt.Printf("<file name=\"%s\">\n<content>\n%s\n</content>\n</file>\n", filePath, string(content))
	case "markdown":
		fmt.Printf("## File: %s\n\n```\n%s\n```\n\n", filePath, string(content))
	case "simple":
		fmt.Printf("===\n\nFile: %s\n\n---\n\n%s\n\n===\n\n", filePath, string(content))
	default:
		fmt.Printf("=== BEGIN: %s ===\n%s\n=== END: %s ===\n\n", filePath, string(content), filePath)
	}

	fp.CurrentSize += int64(len(content))
	fp.FileCount++

	tokens := fp.TokenCounter.Encode(string(content), nil, nil)
	tokenCount := len(tokens)
	fp.TokenCounts[filePath] = tokenCount
	fp.TotalTokens += tokenCount
}

func (fp *FileProcessor) printSummary() {
	if fp.StatsLevel == "total" {
		_, _ = fmt.Fprintf(os.Stderr, "\nTotal tokens: %d\n", fp.TotalTokens)
	} else if fp.StatsLevel == "detailed" {
		_, _ = fmt.Fprintf(os.Stderr, "\nSummary:\n")
		_, _ = fmt.Fprintf(os.Stderr, "Total files processed: %d\n", fp.FileCount)
		_, _ = fmt.Fprintf(os.Stderr, "Total size processed: %d bytes\n", fp.CurrentSize)
		_, _ = fmt.Fprintf(os.Stderr, "Total tokens: %d\n", fp.TotalTokens)
		_, _ = fmt.Fprintf(os.Stderr, "\nToken counts per file/directory:\n")

		paths := make([]string, 0, len(fp.TokenCounts))
		for path := range fp.TokenCounts {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %d tokens\n", path, fp.TokenCounts[path])
		}
	}
}