package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

// FileInfo represents information about a file in git history
type FileInfo struct {
	Path         string
	Size         int64
	Hash         string
	IsLargeBinary bool
	CommitHash   string
	CommitMsg    string
}

// Stats represents analysis statistics
type Stats struct {
	TotalFiles    int
	LargeFiles    int
	TotalSize     int64
	LargeFileSize int64
	Files         []FileInfo
}

// FormatSize formats bytes into human readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsLikelyBinary determines if a file is likely binary based on extension
func IsLikelyBinary(path string) bool {
	binaryExts := []string{
		".exe", ".bin", ".dll", ".so", ".dylib", ".a", ".o",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp",
		".mp4", ".avi", ".mov", ".mkv", ".mp3", ".wav", ".flac",
		".zip", ".tar", ".gz", ".rar", ".7z", ".pdf", ".doc", ".docx",
	}
	
	path = strings.ToLower(path)
	for _, ext := range binaryExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// SortFilesBySize sorts files by size in descending order
func (s *Stats) SortFilesBySize() {
	sort.Slice(s.Files, func(i, j int) bool {
		return s.Files[i].Size > s.Files[j].Size
	})
}

// GetLargeFiles returns only files above the size threshold
func (s *Stats) GetLargeFiles(threshold int64) []FileInfo {
	var large []FileInfo
	for _, file := range s.Files {
		if file.Size >= threshold {
			large = append(large, file)
		}
	}
	return large
}

// Summary returns a formatted summary of the analysis
func (s *Stats) Summary(threshold int64) string {
	large := s.GetLargeFiles(threshold)
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Analysis Summary:\n"))
	sb.WriteString(fmt.Sprintf("  Total files: %d\n", s.TotalFiles))
	sb.WriteString(fmt.Sprintf("  Large files (>%s): %d\n", FormatSize(threshold), len(large)))
	sb.WriteString(fmt.Sprintf("  Total size: %s\n", FormatSize(s.TotalSize)))
	
	var largeSize int64
	for _, f := range large {
		largeSize += f.Size
	}
	sb.WriteString(fmt.Sprintf("  Large files size: %s\n", FormatSize(largeSize)))
	
	return sb.String()
}
