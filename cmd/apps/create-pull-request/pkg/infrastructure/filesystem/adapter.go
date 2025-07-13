package filesystem

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

// Adapter defines the interface for filesystem operations
type Adapter interface {
	// ReadFile reads the content of a file
	ReadFile(path string) ([]byte, error)
	// WriteFile writes content to a file
	WriteFile(path string, content []byte) error
	// EditFile opens a file in the user's editor and returns the edited content
	EditFile(path string) ([]byte, error)
	// EditTempFile creates a temporary file with content, opens it in editor, and returns edited content
	EditTempFile(content string, extension string) (string, error)
	// ViewWithPager displays content using a pager
	ViewWithPager(content string) error
}

// RealAdapter provides a real implementation of the filesystem adapter
type RealAdapter struct{}

// ReadFile implements the Adapter interface
func (a *RealAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile implements the Adapter interface
func (a *RealAdapter) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// EditFile implements the Adapter interface
func (a *RealAdapter) EditFile(path string) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default to vi if EDITOR is not set
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run editor")
	}

	return os.ReadFile(path)
}

// EditTempFile implements the Adapter interface
func (a *RealAdapter) EditTempFile(content string, extension string) (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gopr-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file
	tmpPath := filepath.Join(tmpDir, "content"+extension)
	err = os.WriteFile(tmpPath, []byte(content), 0644)
	if err != nil {
		return "", errors.Wrap(err, "failed to write to temp file")
	}

	// Edit the file
	editedContent, err := a.EditFile(tmpPath)
	if err != nil {
		return "", err
	}

	return string(editedContent), nil
}

// ViewWithPager implements the Adapter interface
func (a *RealAdapter) ViewWithPager(content string) error {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less" // Default to less if PAGER is not set
	}

	// Create a temporary file with the content
	tmpFile, err := os.CreateTemp("", "gopr-view-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return errors.Wrap(err, "failed to write to temp file")
	}
	tmpFile.Close() // Close now so the pager can read it

	// Run the pager
	cmd := exec.Command(pager, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// NewRealAdapter creates a new real filesystem adapter
func NewRealAdapter() *RealAdapter {
	return &RealAdapter{}
}
