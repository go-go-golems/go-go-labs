package main

import (
	"github.com/denormal/go-gitignore"
	"github.com/go-go-golems/clay/pkg/filewalker"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileFilter struct {
	MaxFileSize           int64
	IncludeExts           []string
	ExcludeExts           []string
	MatchFilenames        []*regexp.Regexp
	MatchPaths            []*regexp.Regexp
	ExcludeDirs           []string
	GitIgnoreFilter       gitignore.GitIgnore
	DisableGitIgnore      bool
	DefaultExcludeExts    []string
	ExcludeMatchFilenames []*regexp.Regexp
	ExcludeMatchPaths     []*regexp.Regexp
}

func (ff *FileFilter) FilterNode(node *filewalker.Node) bool {
	if node.GetType() == filewalker.DirectoryNode {
		return !ff.isExcludedDir(node.GetPath())
	}
	return ff.FilterPath(node.GetPath())
}

func (ff *FileFilter) FilterPath(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if fileInfo.IsDir() {
		return !ff.isExcludedDir(filePath)
	}

	return ff.shouldProcessFile(filePath, fileInfo)
}

func (ff *FileFilter) isExcludedDir(dirPath string) bool {
	if strings.HasSuffix(dirPath, ".git") {
		return true
	}
	for _, excludedDir := range ff.ExcludeDirs {
		if strings.Contains(dirPath, excludedDir) {
			return true
		}
	}
	return false
}

func (ff *FileFilter) shouldProcessFile(filePath string, fileInfo os.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check against default excluded extensions
	for _, excludedExt := range ff.DefaultExcludeExts {
		if ext == excludedExt {
			return false
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

	if !ff.DisableGitIgnore && ff.GitIgnoreFilter != nil && ff.GitIgnoreFilter.Ignore(filePath) {
		return false
	}

	return true
}
