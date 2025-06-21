package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// PersistenceManager handles saving and loading of application state
type PersistenceManager struct {
	dataDir string
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager() (*PersistenceManager, error) {
	dataDir, err := getDataDirectory()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data directory")
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to create data directory: %s", dataDir)
	}

	// Ensure history directory exists
	historyDir := filepath.Join(dataDir, "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to create history directory: %s", historyDir)
	}

	return &PersistenceManager{
		dataDir: dataDir,
	}, nil
}

// getDataDirectory returns the appropriate data directory for the application
func getDataDirectory() (string, error) {
	// Try XDG_DATA_HOME first
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "prompt-builder"), nil
	}

	// Fall back to ~/.local/share
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get home directory")
	}

	return filepath.Join(homeDir, ".local", "share", "prompt-builder"), nil
}

// SaveCurrentState saves the current selection state to last.yml
func (p *PersistenceManager) SaveCurrentState(selection *SelectionState) error {
	lastPath := filepath.Join(p.dataDir, "last.yml")
	return p.saveSelectionToFile(selection, lastPath)
}

// LoadCurrentState loads the current selection state from last.yml
func (p *PersistenceManager) LoadCurrentState() (*SelectionState, error) {
	lastPath := filepath.Join(p.dataDir, "last.yml")
	return p.loadSelectionFromFile(lastPath)
}

// SaveToHistory saves the selection state to a timestamped history file
func (p *PersistenceManager) SaveToHistory(selection *SelectionState) error {
	timestamp := selection.Timestamp.Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.yml", timestamp, selection.TemplateID)
	historyPath := filepath.Join(p.dataDir, "history", filename)
	
	return p.saveSelectionToFile(selection, historyPath)
}

// saveSelectionToFile saves a selection state to the specified file
func (p *PersistenceManager) saveSelectionToFile(selection *SelectionState, path string) error {
	selection.Timestamp = time.Now()
	
	data, err := yaml.Marshal(selection)
	if err != nil {
		return errors.Wrap(err, "failed to marshal selection state")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrapf(err, "failed to write selection to file: %s", path)
	}

	return nil
}

// loadSelectionFromFile loads a selection state from the specified file
func (p *PersistenceManager) loadSelectionFromFile(path string) (*SelectionState, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // File doesn't exist, return nil without error
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read selection file: %s", path)
	}

	var selection SelectionState
	if err := yaml.Unmarshal(data, &selection); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal selection from file: %s", path)
	}

	return &selection, nil
}

// ListHistory returns a list of saved history files
func (p *PersistenceManager) ListHistory() ([]string, error) {
	historyDir := filepath.Join(p.dataDir, "history")
	
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, errors.Wrapf(err, "failed to read history directory: %s", historyDir)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// LoadFromHistory loads a selection state from a history file
func (p *PersistenceManager) LoadFromHistory(filename string) (*SelectionState, error) {
	historyPath := filepath.Join(p.dataDir, "history", filename)
	return p.loadSelectionFromFile(historyPath)
}

// GetDataDirectory returns the data directory path
func (p *PersistenceManager) GetDataDirectory() string {
	return p.dataDir
}
