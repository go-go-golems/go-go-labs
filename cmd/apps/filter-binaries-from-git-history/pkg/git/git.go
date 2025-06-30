package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-go-golems/go-go-labs/cmd/apps/filter-binaries-from-git-history/pkg/analyzer"
	"github.com/rs/zerolog/log"
)

// Repository wraps git repository operations
type Repository struct {
	repo *git.Repository
	path string
}

// OpenRepository opens a git repository at the given path
func OpenRepository(path string) (*Repository, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at %s: %w", path, err)
	}

	return &Repository{
		repo: repo,
		path: path,
	}, nil
}

// AnalyzeDiff analyzes the difference between two references
func (r *Repository) AnalyzeDiff(baseRef, compareRef string, sizeThreshold int64) (*analyzer.Stats, error) {
	log.Info().
		Str("base", baseRef).
		Str("compare", compareRef).
		Int64("threshold", sizeThreshold).
		Msg("Starting diff analysis")

	// Get commit hashes for both references
	baseHash, err := r.repo.ResolveRevision(plumbing.Revision(baseRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base reference %s: %w", baseRef, err)
	}

	compareHash, err := r.repo.ResolveRevision(plumbing.Revision(compareRef))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve compare reference %s: %w", compareRef, err)
	}

	log.Debug().
		Str("baseHash", baseHash.String()).
		Str("compareHash", compareHash.String()).
		Msg("Resolved commit hashes")

	// Use git command line for better diff analysis
	return r.analyzeDiffWithCommand(baseRef, compareRef, sizeThreshold)
}

// analyzeDiffWithCommand uses git command line to get detailed diff information
func (r *Repository) analyzeDiffWithCommand(baseRef, compareRef string, sizeThreshold int64) (*analyzer.Stats, error) {
	// Get list of changed files with sizes
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s..%s", baseRef, compareRef))
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return &analyzer.Stats{}, nil
	}

	stats := &analyzer.Stats{
		Files: make([]analyzer.FileInfo, 0, len(files)),
	}

	for _, file := range files {
		if file == "" {
			continue
		}

		fileInfo, err := r.getFileInfo(file, compareRef)
		if err != nil {
			log.Warn().Err(err).Str("file", file).Msg("Failed to get file info")
			continue
		}

		fileInfo.IsLargeBinary = analyzer.IsLikelyBinary(file) && fileInfo.Size >= sizeThreshold

		stats.Files = append(stats.Files, fileInfo)
		stats.TotalFiles++
		stats.TotalSize += fileInfo.Size

		if fileInfo.Size >= sizeThreshold {
			stats.LargeFiles++
			stats.LargeFileSize += fileInfo.Size
		}
	}

	stats.SortFilesBySize()
	return stats, nil
}

// getFileInfo gets detailed information about a file at a specific reference
func (r *Repository) getFileInfo(path, ref string) (analyzer.FileInfo, error) {
	// Get file size using git cat-file
	cmd := exec.Command("git", "cat-file", "-s", fmt.Sprintf("%s:%s", ref, path))
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		// File might not exist in this ref, try to get size from filesystem
		if strings.Contains(err.Error(), "does not exist") {
			fullPath := filepath.Join(r.path, path)
			if stat, statErr := os.Stat(fullPath); statErr == nil {
				return analyzer.FileInfo{
					Path: path,
					Size: stat.Size(),
				}, nil
			}
		}
		return analyzer.FileInfo{}, fmt.Errorf("failed to get file size for %s: %w", path, err)
	}

	size, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return analyzer.FileInfo{}, fmt.Errorf("failed to parse file size: %w", err)
	}

	// Get file hash
	cmd = exec.Command("git", "hash-object", fmt.Sprintf("%s:%s", ref, path))
	cmd.Dir = r.path
	hashOutput, err := cmd.Output()
	if err != nil {
		log.Warn().Err(err).Str("file", path).Msg("Failed to get file hash")
	}

	return analyzer.FileInfo{
		Path: path,
		Size: size,
		Hash: strings.TrimSpace(string(hashOutput)),
	}, nil
}

// RemoveFilesFromHistory removes selected files from git history using git filter-branch
func (r *Repository) RemoveFilesFromHistory(files []string, baseRef string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified for removal")
	}

	log.Info().
		Strs("files", files).
		Str("baseRef", baseRef).
		Msg("Starting git history rewrite")

	// Create index filter command to remove files
	var filterParts []string
	for _, file := range files {
		filterParts = append(filterParts, fmt.Sprintf("git rm --cached --ignore-unmatch '%s'", file))
	}
	indexFilter := strings.Join(filterParts, " && ")

	// Run git filter-branch
	cmd := exec.Command("git", "filter-branch", "-f", "--index-filter", indexFilter, fmt.Sprintf("%s..HEAD", baseRef))
	cmd.Dir = r.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to filter git history: %w", err)
	}

	log.Info().Msg("Git history rewrite completed successfully")
	return nil
}

// GetCommitInfo gets information about commits that modified a file
func (r *Repository) GetCommitInfo(filePath, ref string) ([]analyzer.FileInfo, error) {
	// Use git log to get commits that modified the file
	cmd := exec.Command("git", "log", "--oneline", "--follow", ref, "--", filePath)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history for %s: %w", filePath, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commits []analyzer.FileInfo

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		commitHash := parts[0]
		commitMsg := parts[1]

		commits = append(commits, analyzer.FileInfo{
			Path:       filePath,
			CommitHash: commitHash,
			CommitMsg:  commitMsg,
		})
	}

	return commits, nil
}
