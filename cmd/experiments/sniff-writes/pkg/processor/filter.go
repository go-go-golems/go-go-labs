package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

func ParseEvent(data []byte, event *models.Event) error {
	expectedSize := int(unsafe.Sizeof(*event))
	if len(data) < expectedSize {
		return fmt.Errorf("data too short: got %d bytes, expected %d bytes", len(data), expectedSize)
	}

	*event = *(*models.Event)(unsafe.Pointer(&data[0]))
	return nil
}

func IsRealFile(path string) bool {
	if path == "" {
		return false
	}

	// Filter out non-file descriptors by default
	if strings.Contains(path, "pipe:") ||
		strings.Contains(path, "anon_inode:") ||
		strings.Contains(path, "socket:") ||
		strings.HasPrefix(path, "/dev/") ||
		strings.HasPrefix(path, "/proc/") ||
		strings.HasPrefix(path, "/sys/") {
		return false
	}

	return true
}

func ShouldProcessEvent(event *models.Event, resolvedPath string, config *models.Config) bool {
	comm := cString(event.Comm[:])

	// Check process filter first (more efficient)
	if config.ProcessFilter != "" && !strings.Contains(comm, config.ProcessFilter) {
		return false
	}

	// Apply process glob filtering
	if !matchesProcessGlobFilters(comm, config) {
		return false
	}

	// Skip if we still don't have a filename and it's not a close event
	if resolvedPath == "" && event.Type != 3 {
		return false
	}

	// Filter out non-real files by default (unless in debug mode or show-all-files is enabled)
	if !config.Debug && !config.ShowAllFiles && !IsRealFile(resolvedPath) {
		return false
	}

	// Check directory filter - for close events without filename, we can't filter
	if resolvedPath != "" {
		// Get current working directory for relative path comparison
		cwd, _ := os.Getwd()

		// Convert both paths to absolute for proper comparison
		absFilename := resolvedPath
		if !filepath.IsAbs(resolvedPath) {
			absFilename = filepath.Join(cwd, resolvedPath)
		}
		absFilename = filepath.Clean(absFilename)

		absTargetDir := config.Directory
		if !filepath.IsAbs(config.Directory) {
			absTargetDir = filepath.Join(cwd, config.Directory)
		}
		absTargetDir = filepath.Clean(absTargetDir)

		// Check if file is within the target directory
		relPath, err := filepath.Rel(absTargetDir, absFilename)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return false
		}

		// Apply glob filtering
		if !matchesGlobFilters(resolvedPath, config) {
			return false
		}
	}

	return true
}

// matchesGlobFilters checks if the file path matches include patterns and doesn't match exclude patterns
// Supports both separate --glob-exclude flags and !pattern syntax in --glob flags
func matchesGlobFilters(path string, config *models.Config) bool {
	filename := filepath.Base(path)

	// Parse patterns and separate positive/negative patterns
	var includePatterns []string
	var excludePatterns []string

	// Process --glob patterns (support !prefix for exclusion)
	for _, pattern := range config.GlobPatterns {
		if strings.HasPrefix(pattern, "!") {
			excludePatterns = append(excludePatterns, pattern[1:]) // Remove ! prefix
		} else {
			includePatterns = append(includePatterns, pattern)
		}
	}

	// Add patterns from --glob-exclude (for backwards compatibility)
	excludePatterns = append(excludePatterns, config.GlobExclude...)

	// Check exclude patterns first - if any match, exclude the file
	for _, pattern := range excludePatterns {
		if match, _ := filepath.Match(pattern, filename); match {
			return false
		}
	}

	// If we have include patterns, file must match at least one
	if len(includePatterns) > 0 {
		for _, pattern := range includePatterns {
			if match, _ := filepath.Match(pattern, filename); match {
				return true
			}
		}
		return false // No include patterns matched
	}

	return true // No include patterns specified, file passes
}

// matchesProcessGlobFilters checks if the process name matches include patterns and doesn't match exclude patterns
// Supports both separate --process-glob-exclude flags and !pattern syntax in --process-glob flags
func matchesProcessGlobFilters(processName string, config *models.Config) bool {
	// Parse patterns and separate positive/negative patterns
	var includePatterns []string
	var excludePatterns []string

	// Process --process-glob patterns (support !prefix for exclusion)
	for _, pattern := range config.ProcessGlob {
		if strings.HasPrefix(pattern, "!") {
			excludePatterns = append(excludePatterns, pattern[1:]) // Remove ! prefix
		} else {
			includePatterns = append(includePatterns, pattern)
		}
	}

	// Add patterns from --process-glob-exclude (for backwards compatibility)
	excludePatterns = append(excludePatterns, config.ProcessGlobExclude...)

	// Check exclude patterns first - if any match, exclude the process
	for _, pattern := range excludePatterns {
		if match, _ := filepath.Match(pattern, processName); match {
			return false
		}
	}

	// If we have include patterns, process must match at least one
	if len(includePatterns) > 0 {
		for _, pattern := range includePatterns {
			if match, _ := filepath.Match(pattern, processName); match {
				return true
			}
		}
		return false // No include patterns matched
	}

	return true // No include patterns specified, process passes
}

func cString(b []int8) string {
	n := -1
	for i, v := range b {
		if v == 0 {
			n = i
			break
		}
	}
	if n == -1 {
		n = len(b)
	}
	// Convert []int8 to []byte
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = byte(b[i])
	}
	return string(bytes)
}
