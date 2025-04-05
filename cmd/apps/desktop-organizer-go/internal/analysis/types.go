package analysis

import (
	"io/fs"
	"time"

	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/config"
)

// FileEntry holds information about a scanned file system entry.
type FileEntry struct {
	Path     string      `json:"path"` // Relative to TargetDir
	FullPath string      `json:"-"`    // Absolute path
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size"`
	ModTime  time.Time   `json:"mod_time"`
	Mode     fs.FileMode `json:"-"`

	TypeInfo map[string]any    `json:"type_info,omitempty"` // Source ("magika", "file"), Label, Group, MIME
	Metadata map[string]any    `json:"metadata,omitempty"`  // Source ("exiftool"), Key-Value pairs
	Hashes   map[string]string `json:"hashes,omitempty"`    // Algorithm ("md5", "sha256") -> Hash string
	Tags     []string          `json:"tags,omitempty"`      // "LargeFile", "RecentFile", "DuplicateSetID:xyz"
	Error    string            `json:"error,omitempty"`     // Record file-specific processing errors
}

// AnalysisResult contains the overall results of a directory analysis run.
type AnalysisResult struct {
	RootDir        string                 `json:"root_dir"`
	ScanStartTime  time.Time              `json:"scan_start_time"`
	ScanEndTime    time.Time              `json:"scan_end_time"`
	TotalFiles     int                    `json:"total_files"`
	TotalDirs      int                    `json:"total_dirs"`
	TotalSize      int64                  `json:"total_size"`
	FileEntries    []*FileEntry           `json:"file_entries,omitempty"` // May be omitted in summary-only reports
	TypeSummary    map[string]*TypeStats  `json:"type_summary"`           // Keyed by type label
	DuplicateSets  []*DuplicateSet        `json:"duplicate_sets,omitempty"`
	MonthlySummary map[string]*MonthStats `json:"monthly_summary"` // Keyed by "YYYY-MM"
	ToolStatus     map[string]*ToolInfo   `json:"tool_status"`
	OverallErrors  []string               `json:"overall_errors,omitempty"`
	Config         *config.Config         `json:"config_used"` // Include the config used for this run
}

// TypeStats represents aggregated statistics for a specific file type.
type TypeStats struct {
	Label string   `json:"label"`
	Count int      `json:"count"`
	Size  int64    `json:"size"`
	Paths []string `json:"-"` // Temporary storage during aggregation
}

// DuplicateSet represents a set of duplicate files.
type DuplicateSet struct {
	ID          string   `json:"id"`           // Hash or external tool ID
	FilePaths   []string `json:"file_paths"`   // Relative paths
	Size        int64    `json:"size"`         // Size of one file
	Count       int      `json:"count"`        // Number of files in the set
	WastedSpace int64    `json:"wasted_space"` // (Count-1) * Size
}

// MonthStats represents aggregated statistics for files modified in a specific month.
type MonthStats struct {
	YearMonth string `json:"year_month"` // YYYY-MM
	Count     int    `json:"count"`
	Size      int64  `json:"size"`
}

// ToolInfo holds information about an external tool's availability and status.
type ToolInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// NewFileEntry creates a new FileEntry with initialized maps.
func NewFileEntry(path, fullPath string, isDir bool, size int64, modTime time.Time, mode fs.FileMode) *FileEntry {
	return &FileEntry{
		Path:     path,
		FullPath: fullPath,
		IsDir:    isDir,
		Size:     size,
		ModTime:  modTime,
		Mode:     mode,
		TypeInfo: make(map[string]any),
		Metadata: make(map[string]any),
		Hashes:   make(map[string]string),
		Tags:     make([]string, 0),
	}
}

// NewAnalysisResult creates a new AnalysisResult with initialized maps and slices.
func NewAnalysisResult(rootDir string, cfg *config.Config) *AnalysisResult {
	return &AnalysisResult{
		RootDir:        rootDir,
		ScanStartTime:  time.Now(),
		FileEntries:    make([]*FileEntry, 0),
		TypeSummary:    make(map[string]*TypeStats),
		DuplicateSets:  make([]*DuplicateSet, 0),
		MonthlySummary: make(map[string]*MonthStats),
		ToolStatus:     make(map[string]*ToolInfo),
		OverallErrors:  make([]string, 0),
		Config:         cfg,
	}
}

// AddError adds an error to the list of overall errors.
func (r *AnalysisResult) AddError(err error) {
	if err != nil {
		r.OverallErrors = append(r.OverallErrors, errors.Cause(err).Error())
	}
}

// AddToolStatus records the status of an external tool.
func (r *AnalysisResult) AddToolStatus(name, path string, available bool, err error) {
	toolInfo := &ToolInfo{
		Name:      name,
		Path:      path,
		Available: available,
	}
	if err != nil {
		toolInfo.Error = err.Error()
	}
	r.ToolStatus[name] = toolInfo
}

// AddFileEntry adds a FileEntry to the result and updates totals.
func (r *AnalysisResult) AddFileEntry(entry *FileEntry) {
	r.FileEntries = append(r.FileEntries, entry)
	if !entry.IsDir {
		r.TotalFiles++
		r.TotalSize += entry.Size
	} else {
		r.TotalDirs++
	}
}

// AddTag adds a tag to a FileEntry if it doesn't already exist.
func (e *FileEntry) AddTag(tag string) {
	for _, t := range e.Tags {
		if t == tag {
			return
		}
	}
	e.Tags = append(e.Tags, tag)
}

// HasTag checks if a FileEntry has a specific tag.
func (e *FileEntry) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddError records an error message on a FileEntry.
func (e *FileEntry) AddError(err error) {
	if err != nil {
		if e.Error == "" {
			e.Error = err.Error()
		} else {
			e.Error += "; " + err.Error()
		}
	}
}
