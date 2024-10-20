package main

import (
	"fmt"
	"github.com/denormal/go-gitignore"
	"github.com/go-go-golems/clay/pkg/filewalker"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileFilter struct {
	MaxFileSize                   int64
	IncludeExts                   []string
	ExcludeExts                   []string
	MatchFilenames                []*regexp.Regexp
	MatchPaths                    []*regexp.Regexp
	ExcludeDirs                   []string
	GitIgnoreFilter               gitignore.GitIgnore
	DisableGitIgnore              bool
	DefaultExcludedExts           []string
	DefaultExcludedDirs           []string
	DefaultExcludedMatchFilenames []*regexp.Regexp
	ExcludeMatchFilenames         []*regexp.Regexp
	ExcludeMatchPaths             []*regexp.Regexp
	DisableDefaultFilters         bool
	Verbose                       bool
}

type FileFilterOption func(*FileFilter)

func NewFileFilter(options ...FileFilterOption) *FileFilter {
	ff := &FileFilter{
		MaxFileSize:                   1024 * 1024, // 1MB default
		DefaultExcludedExts:           DefaultExcludedExts,
		DefaultExcludedDirs:           DefaultExcludedDirs,
		DefaultExcludedMatchFilenames: DefaultExcludedMatchFilenames,
	}
	for _, option := range options {
		option(ff)
	}
	return ff
}

func WithMaxFileSize(size int64) FileFilterOption {
	return func(ff *FileFilter) {
		ff.MaxFileSize = size
	}
}

func WithIncludeExts(exts []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.IncludeExts = exts
	}
}

func WithExcludeExts(exts []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.ExcludeExts = exts
	}
}

func WithMatchFilenames(patterns []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.MatchFilenames = compileRegexps(patterns)
	}
}

func WithMatchPaths(patterns []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.MatchPaths = compileRegexps(patterns)
	}
}

func WithExcludeDirs(dirs []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.ExcludeDirs = dirs
	}
}

func WithGitIgnoreFilter(filter gitignore.GitIgnore) FileFilterOption {
	return func(ff *FileFilter) {
		ff.GitIgnoreFilter = filter
	}
}

func WithDisableGitIgnore(disable bool) FileFilterOption {
	return func(ff *FileFilter) {
		ff.DisableGitIgnore = disable
	}
}

func WithDisableDefaultFilters(disable bool) FileFilterOption {
	return func(ff *FileFilter) {
		ff.DisableDefaultFilters = disable
	}
}

func WithExcludeMatchFilenames(patterns []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.ExcludeMatchFilenames = compileRegexps(patterns)
	}
}

func WithExcludeMatchPaths(patterns []string) FileFilterOption {
	return func(ff *FileFilter) {
		ff.ExcludeMatchPaths = compileRegexps(patterns)
	}
}

func WithVerbose(verbose bool) FileFilterOption {
	return func(ff *FileFilter) {
		ff.Verbose = verbose
	}
}

// Initialize default values
var (
	DefaultExcludedExts = []string{
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff",
		".mp3", ".wav", ".ogg", ".flac",
		".mp4", ".avi", ".mov", ".wmv",
		".zip", ".tar", ".gz", ".rar",
		".exe", ".dll", ".so", ".dylib",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".bin", ".dat", ".db", ".sqlite",
		".woff", ".ttf", ".eot", ".svg", ".webp", ".woff2",
	}

	DefaultExcludedDirs = []string{
		".git", ".svn", "node_modules", "vendor", ".history", ".idea", ".vscode",
	}

	DefaultExcludedMatchFilenames = []*regexp.Regexp{
		regexp.MustCompile(`.*-lock\.json$`),
		regexp.MustCompile(`go\.sum$`),
		regexp.MustCompile(`yarn\.lock$`),
		regexp.MustCompile(`package-lock\.json$`),
	}
)

func (ff *FileFilter) PrintConfiguredFilters() {
	fmt.Println("Configured Filters:")
	fmt.Printf("  Max File Size: %d bytes\n", ff.MaxFileSize)
	fmt.Printf("  Include Extensions: %v\n", ff.IncludeExts)
	fmt.Printf("  Exclude Extensions: %v\n", ff.ExcludeExts)
	fmt.Printf("  Match Filenames: %v\n", ff.MatchFilenames)
	fmt.Printf("  Match Paths: %v\n", ff.MatchPaths)
	fmt.Printf("  Exclude Directories: %v\n", ff.ExcludeDirs)
	fmt.Printf("  Exclude Match Filenames: %v\n", ff.ExcludeMatchFilenames)
	fmt.Printf("  Exclude Match Paths: %v\n", ff.ExcludeMatchPaths)
	fmt.Printf("  Disable GitIgnore: %v\n", ff.DisableGitIgnore)
	fmt.Printf("  Disable Default Filters: %v\n", ff.DisableDefaultFilters)
	fmt.Printf("  Default Excluded Extensions: %v\n", ff.DefaultExcludedExts)
	fmt.Printf("  Default Excluded Directories: %v\n", ff.DefaultExcludedDirs)
	fmt.Printf("  Default Excluded Match Filenames: %v\n", ff.DefaultExcludedMatchFilenames)
	fmt.Printf("  Verbose: %v\n", ff.Verbose)
}

func (ff *FileFilter) FilterNode(node *filewalker.Node) bool {
	result := false
	if node.GetType() == filewalker.DirectoryNode {
		result = !ff.isExcludedDir(node.GetPath())
	} else {
		result = ff.FilterPath(node.GetPath())
	}

	if ff.Verbose {
		if result {
			fmt.Printf("Including: %s\n", node.GetPath())
		} else {
			fmt.Printf("Excluding: %s\n", node.GetPath())
		}
	}

	return result
}

func (ff *FileFilter) FilterPath(filePath string) bool {
	result := ff.shouldProcessFile(filePath)

	if ff.Verbose {
		if result {
			fmt.Printf("Including: %s\n", filePath)
		} else {
			fmt.Printf("Excluding: %s\n", filePath)
		}
	}

	return result
}

func (ff *FileFilter) isExcludedDir(dirPath string) bool {
	if !ff.DisableDefaultFilters {
		for _, excludedDir := range ff.DefaultExcludedDirs {
			if strings.Contains(dirPath, excludedDir) {
				return true
			}
		}
	}
	for _, excludedDir := range ff.ExcludeDirs {
		if strings.Contains(dirPath, excludedDir) {
			return true
		}
	}
	return false
}

func (ff *FileFilter) shouldProcessFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// If we can't get file info, we'll exclude the file
		return false
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// Check against default excluded extensions
	if !ff.DisableDefaultFilters {
		for _, excludedExt := range ff.DefaultExcludedExts {
			if ext == excludedExt {
				return false
			}
		}
	}

	if fileInfo.Size() > ff.MaxFileSize {
		return false
	}

	if len(ff.IncludeExts) > 0 {
		included := false
		for _, includedExt := range ff.IncludeExts {
			if ext == strings.ToLower(includedExt) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, excludedExt := range ff.ExcludeExts {
		if ext == strings.ToLower(excludedExt) {
			return false
		}
	}

	if len(ff.MatchFilenames) > 0 || len(ff.MatchPaths) > 0 {
		filenameMatch := false
		pathMatch := false

		for _, re := range ff.MatchFilenames {
			if re.MatchString(filepath.Base(filePath)) {
				filenameMatch = true
				break
			}
		}

		for _, re := range ff.MatchPaths {
			if re.MatchString(filePath) {
				pathMatch = true
				break
			}
		}

		if !filenameMatch && !pathMatch {
			return false
		}
	}

	// Check against default excluded match filenames
	if !ff.DisableDefaultFilters {
		for _, re := range ff.DefaultExcludedMatchFilenames {
			if re.MatchString(filepath.Base(filePath)) {
				return false
			}
		}
	}

	for _, re := range ff.ExcludeMatchFilenames {
		if re.MatchString(filepath.Base(filePath)) {
			return false
		}
	}

	for _, re := range ff.ExcludeMatchPaths {
		if re.MatchString(filePath) {
			return false
		}
	}

	// TODO: fix upstream bug where "." / root panics
	if filePath != "." && !ff.DisableGitIgnore && ff.GitIgnoreFilter != nil && ff.GitIgnoreFilter.Ignore(filePath) {
		return false
	}

	return true
}
