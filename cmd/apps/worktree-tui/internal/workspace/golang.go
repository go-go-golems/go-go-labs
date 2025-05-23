package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
)

// GoWorkspaceOperations handles Go workspace-related operations
type GoWorkspaceOperations struct{}

// NewGoWorkspaceOperations creates a new GoWorkspaceOperations instance
func NewGoWorkspaceOperations() *GoWorkspaceOperations {
	return &GoWorkspaceOperations{}
}

// InitializeGoWork creates a go.work file in the workspace directory
func (g *GoWorkspaceOperations) InitializeGoWork(workspacePath string, repositories []config.Repository) error {
	// Find all Go modules in the workspace
	modules, err := g.findGoModules(workspacePath, repositories)
	if err != nil {
		return fmt.Errorf("failed to find Go modules: %w", err)
	}

	if len(modules) == 0 {
		// No Go modules found, don't create go.work
		return nil
	}

	// Create go.work file
	goWorkPath := filepath.Join(workspacePath, "go.work")
	return g.createGoWorkFile(goWorkPath, modules)
}

// findGoModules searches for go.mod files in the workspace
func (g *GoWorkspaceOperations) findGoModules(workspacePath string, repositories []config.Repository) ([]string, error) {
	var modules []string

	for _, repo := range repositories {
		repoPath := filepath.Join(workspacePath, repo.Name)
		
		// Check if repository path exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue // Skip if repository wasn't set up
		}

		// Find go.mod files in this repository
		repoModules, err := g.findGoModulesInPath(repoPath, workspacePath)
		if err != nil {
			return nil, fmt.Errorf("failed to find modules in %s: %w", repo.Name, err)
		}

		modules = append(modules, repoModules...)
	}

	return modules, nil
}

// findGoModulesInPath recursively searches for go.mod files in a directory
func (g *GoWorkspaceOperations) findGoModulesInPath(searchPath, workspacePath string) ([]string, error) {
	var modules []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directories and other hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Check for go.mod files
		if info.Name() == "go.mod" {
			// Get relative path from workspace root
			relPath, err := filepath.Rel(workspacePath, filepath.Dir(path))
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			// Convert to Unix-style path for go.work
			relPath = filepath.ToSlash(relPath)
			modules = append(modules, relPath)
			
			// Don't recurse into subdirectories of a Go module
			return filepath.SkipDir
		}

		return nil
	})

	return modules, err
}

// createGoWorkFile creates the go.work file with the specified modules
func (g *GoWorkspaceOperations) createGoWorkFile(goWorkPath string, modules []string) error {
	content := g.generateGoWorkContent(modules)
	
	file, err := os.Create(goWorkPath)
	if err != nil {
		return fmt.Errorf("failed to create go.work file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write go.work content: %w", err)
	}

	return nil
}

// generateGoWorkContent generates the content for the go.work file
func (g *GoWorkspaceOperations) generateGoWorkContent(modules []string) string {
	var content strings.Builder
	
	// Write go.work header
	content.WriteString("go 1.21\n\n")
	
	if len(modules) > 0 {
		content.WriteString("use (\n")
		for _, module := range modules {
			content.WriteString(fmt.Sprintf("\t./%s\n", module))
		}
		content.WriteString(")\n")
	}

	return content.String()
}

// ValidateGoWorkspace validates that the Go workspace is properly set up
func (g *GoWorkspaceOperations) ValidateGoWorkspace(workspacePath string) error {
	goWorkPath := filepath.Join(workspacePath, "go.work")
	
	// Check if go.work exists
	if _, err := os.Stat(goWorkPath); os.IsNotExist(err) {
		return fmt.Errorf("go.work file not found")
	}

	// Try to read the workspace
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return fmt.Errorf("failed to read go.work: %w", err)
	}

	// Basic validation - check if it contains 'go' directive
	if !strings.Contains(string(content), "go ") {
		return fmt.Errorf("invalid go.work file: missing go directive")
	}

	return nil
}

// GetGoVersion returns the Go version from go.mod or a default
func (g *GoWorkspaceOperations) GetGoVersion(modulePath string) string {
	goModPath := filepath.Join(modulePath, "go.mod")
	
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "1.21" // Default version
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return "1.21" // Default version
}