package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/go-go-golems/clay/pkg/filewalker"
	"github.com/weaviate/tiktoken-go"
)

type FileStats struct {
	TokenCount int
	LineCount  int
	Size       int64
	FileCount  int // Add FileCount field
}

type Stats struct {
	Files     map[string]FileStats
	FileTypes map[string]FileStats
	Dirs      map[string]FileStats
	DirFiles  map[string][]string // New field to store files per directory
	Total     FileStats
	mu        sync.Mutex
}

type OutputFlag int

const (
	OutputOverview OutputFlag = 1 << iota
	OutputDirStructure
	OutputFullStructure
)

type Config struct {
	OutputFlags OutputFlag
}

func NewStats() *Stats {
	return &Stats{
		Files:     make(map[string]FileStats),
		FileTypes: make(map[string]FileStats),
		Dirs:      make(map[string]FileStats),
		DirFiles:  make(map[string][]string), // Initialize the new map
	}
}

func (s *Stats) AddFile(path string, stats FileStats) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Files[path] = stats

	// Update filetype stats
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		ext = "no_extension"
	}
	fileTypeStats := s.FileTypes[ext]
	fileTypeStats.TokenCount += stats.TokenCount
	fileTypeStats.LineCount += stats.LineCount
	fileTypeStats.Size += stats.Size
	fileTypeStats.FileCount++ // Increment FileCount for file type
	s.FileTypes[ext] = fileTypeStats

	// Update directory stats
	dir := filepath.Dir(path)
	dirStats := s.Dirs[dir]
	dirStats.TokenCount += stats.TokenCount
	dirStats.LineCount += stats.LineCount
	dirStats.Size += stats.Size
	dirStats.FileCount++ // Increment FileCount for directory
	s.Dirs[dir] = dirStats

	// Update DirFiles map
	s.DirFiles[dir] = append(s.DirFiles[dir], path)

	// Update total stats
	s.Total.TokenCount += stats.TokenCount
	s.Total.LineCount += stats.LineCount
	s.Total.Size += stats.Size
	s.Total.FileCount++ // Increment FileCount for total
}

func ComputeStats(paths []string, filter *FileFilter) (*Stats, error) {
	stats := NewStats()
	tokenCounter, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("error initializing tiktoken: %v", err)
	}

	walker, err := filewalker.NewWalker(
		filewalker.WithPaths(paths),
		filewalker.WithFilter(filter.FilterNode),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating filewalker: %v", err)
	}

	preVisit := func(w *filewalker.Walker, node *filewalker.Node) error {
		if node.Type == filewalker.FileNode {
			content, err := os.ReadFile(node.Path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", node.Path, err)
			}

			tokens := tokenCounter.Encode(string(content), nil, nil)
			tokenCount := len(tokens)
			lineCount := strings.Count(string(content), "\n") + 1
			size := int64(len(content))

			fileStats := FileStats{
				TokenCount: tokenCount,
				LineCount:  lineCount,
				Size:       size,
				FileCount:  1, // Initialize FileCount for file stats
			}

			stats.AddFile(node.Path, fileStats)
		}
		return nil
	}

	if err := walker.Walk(paths, preVisit, nil); err != nil {
		return nil, fmt.Errorf("error walking files: %v", err)
	}

	return stats, nil
}

func PrintStats(stats *Stats, config Config) {
	if config.OutputFlags&OutputOverview != 0 {
		printOverview(stats)
	}

	if config.OutputFlags&OutputFullStructure != 0 {
		printFullStructure(stats)
	} else if config.OutputFlags&OutputDirStructure != 0 {
		printDirStructure(stats)
	}
}

func printOverview(stats *Stats) {
	fmt.Println("Overview:")
	fmt.Printf("Total Files: %d\n", stats.Total.FileCount) // Use Total.FileCount
	fmt.Printf("Total Directories: %d\n", len(stats.Dirs))
	fmt.Printf("Total Tokens: %d\n", stats.Total.TokenCount)
	fmt.Printf("Total Lines: %d\n", stats.Total.LineCount)
	fmt.Printf("Total Size: %d bytes\n", stats.Total.Size)

	fmt.Println("\nFile Type Statistics:")
	for ext, typeStats := range stats.FileTypes {
		fmt.Printf("  %s:\n    Files: %d, Tokens: %d, Lines: %d, Size: %d bytes\n",
			ext, typeStats.FileCount, typeStats.TokenCount, typeStats.LineCount, typeStats.Size)
	}
	fmt.Println() // Add a newline for better separation
}

func printDirStructure(stats *Stats) {
	fmt.Println("Directory Structure:")
	dirs := getSortedDirs(stats)
	for _, dir := range dirs {
		dirStats := stats.Dirs[dir]
		fmt.Printf("  %s:\n    Files: %d, Tokens: %d, Lines: %d, Size: %d bytes\n",
			dir, dirStats.FileCount, dirStats.TokenCount, dirStats.LineCount, dirStats.Size)
	}
	fmt.Println() // Add a newline for better separation
}

func printFullStructure(stats *Stats) {
	fmt.Println("Full Structure:")
	dirs := getSortedDirs(stats)
	for _, dir := range dirs {
		dirStats := stats.Dirs[dir]
		fmt.Printf("%s:\n  Tokens: %d, Lines: %d, Size: %d bytes\n",
			dir, dirStats.TokenCount, dirStats.LineCount, dirStats.Size)

		files := stats.DirFiles[dir]
		sort.Strings(files)

		for _, file := range files {
			fileStats := stats.Files[file]
			fmt.Printf("  %s:\n    Tokens: %d, Lines: %d, Size: %d bytes\n",
				filepath.Base(file), fileStats.TokenCount, fileStats.LineCount, fileStats.Size)
		}
		fmt.Println()
	}
}

func getSortedDirs(stats *Stats) []string {
	dirs := make([]string, 0, len(stats.Dirs))
	for dir := range stats.Dirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs
}

func countFilesByType(stats *Stats, ext string) int {
	count := 0
	for file := range stats.Files {
		if strings.ToLower(filepath.Ext(file)) == ext {
			count++
		}
	}
	return count
}

func countFilesInDir(stats *Stats, dir string) int {
	count := 0
	for file := range stats.Files {
		if filepath.Dir(file) == dir {
			count++
		}
	}
	return count
}
