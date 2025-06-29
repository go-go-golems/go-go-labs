package browser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func BrowseForFiles() ([]string, error) {
	var selectedFiles []string
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current directory")
	}

	for {
		// List files and directories in current directory
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read directory")
		}

		// Create options for selection
		var options []huh.Option[string]
		
		// Add parent directory option (unless we're at root)
		if currentDir != "/" {
			options = append(options, huh.NewOption(".. (parent directory)", ".."))
		}

		// Add directories first
		for _, entry := range entries {
			if entry.IsDir() {
				displayName := fmt.Sprintf("📁 %s/", entry.Name())
				options = append(options, huh.NewOption(displayName, entry.Name()))
			}
		}

		// Add files
		for _, entry := range entries {
			if !entry.IsDir() {
				displayName := fmt.Sprintf("📄 %s", entry.Name())
				options = append(options, huh.NewOption(displayName, entry.Name()))
			}
		}

		// Add special options
		options = append(options, huh.NewOption("✅ Done selecting", "DONE"))
		options = append(options, huh.NewOption("❌ Cancel", "CANCEL"))

		var selection string
		err = huh.NewSelect[string]().
			Title(fmt.Sprintf("Browse files in: %s", currentDir)).
			Description(fmt.Sprintf("Selected files: %d", len(selectedFiles))).
			Options(options...).
			Value(&selection).
			Run()
		if err != nil {
			return nil, errors.Wrap(err, "failed to select item")
		}

		switch selection {
		case "DONE":
			return selectedFiles, nil
		case "CANCEL":
			return []string{}, nil
		case "..":
			currentDir = filepath.Dir(currentDir)
		default:
			targetPath := filepath.Join(currentDir, selection)
			stat, err := os.Stat(targetPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to stat selected item")
			}

			if stat.IsDir() {
				currentDir = targetPath
			} else {
				// It's a file, add it to selected files
				selectedFiles = append(selectedFiles, targetPath)
				log.Debug().Str("file", targetPath).Msg("Added file to selection")
			}
		}
	}
}
