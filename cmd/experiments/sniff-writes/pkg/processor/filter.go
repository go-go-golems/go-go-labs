package processor

import (
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
func matchesGlobFilters(path string, config *models.Config) bool {
	filename := filepath.Base(path)

	// If we have include patterns, file must match at least one
	if len(config.GlobPatterns) > 0 {
		matched := false
		for _, pattern := range config.GlobPatterns {
			if match, _ := filepath.Match(pattern, filename); match {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If we have exclude patterns, file must not match any
	for _, pattern := range config.GlobExclude {
		if match, _ := filepath.Match(pattern, filename); match {
			return false
		}
	}

	return true
}

// matchesProcessGlobFilters checks if the process name matches include patterns and doesn't match exclude patterns
func matchesProcessGlobFilters(processName string, config *models.Config) bool {
	// If we have include patterns, process must match at least one
	if len(config.ProcessGlob) > 0 {
		matched := false
		for _, pattern := range config.ProcessGlob {
			if match, _ := filepath.Match(pattern, processName); match {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If we have exclude patterns, process must not match any
	for _, pattern := range config.ProcessGlobExclude {
		if match, _ := filepath.Match(pattern, processName); match {
			return false
		}
	}

	return true
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