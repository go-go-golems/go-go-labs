package main

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/pkg/errors"
)

// ClipboardManager handles clipboard operations
type ClipboardManager struct{}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{}
}

// CopyToClipboard copies text to the system clipboard
func (c *ClipboardManager) CopyToClipboard(text string) error {
	err := clipboard.WriteAll(text)
	if err != nil {
		// Fallback to stdout if clipboard fails
		fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
		fmt.Fprintf(os.Stderr, "Prompt content:\n%s\n", text)
		return errors.Wrap(err, "failed to copy to clipboard")
	}
	return nil
}

// IsClipboardAvailable checks if clipboard operations are supported
func (c *ClipboardManager) IsClipboardAvailable() bool {
	err := clipboard.WriteAll("test")
	return err == nil
}
