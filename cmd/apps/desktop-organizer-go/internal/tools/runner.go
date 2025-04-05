package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Runner defines an interface for executing external command-line tools.
type Runner interface {
	// Check verifies the tool exists and is executable.
	Check(ctx context.Context) error

	// Run executes the tool with given arguments.
	Run(ctx context.Context, args ...string) (stdout []byte, stderr []byte, err error)

	// RunWithInput executes the tool with stdin input.
	RunWithInput(ctx context.Context, input []byte, args ...string) (stdout []byte, stderr []byte, err error)

	// ToolName returns the name of the tool (e.g., "magika", "exiftool").
	ToolName() string

	// ToolPath returns the configured path to the tool.
	ToolPath() string
}

// CommandRunner implements the Runner interface using os/exec.
type CommandRunner struct {
	name string
	path string
}

// NewRunner creates a new CommandRunner for the specified tool.
// If path is empty, it attempts to locate the tool on the PATH.
func NewRunner(name, path string) (Runner, error) {
	// If path is empty, try to locate the tool
	if path == "" {
		var err error
		path, err = findCommand(name)
		if err != nil {
			return nil, err
		}
	}

	// Verify the path exists and is executable
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tool %s not found at %s", name, path)
		}
		return nil, fmt.Errorf("error accessing tool %s at %s: %w", name, path, err)
	}

	// Check if it's a file and executable
	// On Windows, we don't check for executable bit
	if info.IsDir() {
		return nil, fmt.Errorf("path %s for tool %s is a directory, not an executable", path, name)
	}

	return &CommandRunner{
		name: name,
		path: path,
	}, nil
}

// Check verifies the tool exists and is executable.
func (r *CommandRunner) Check(ctx context.Context) error {
	// Most verification is already done in NewRunner, but we can do a simple
	// command like "--version" or "--help" to further verify the tool works.
	// For now, we just check that the path exists.
	_, err := os.Stat(r.path)
	return err
}

// Run executes the tool with given arguments.
func (r *CommandRunner) Run(ctx context.Context, args ...string) ([]byte, []byte, error) {
	return r.RunWithInput(ctx, nil, args...)
}

// RunWithInput executes the tool with stdin input.
func (r *CommandRunner) RunWithInput(ctx context.Context, input []byte, args ...string) ([]byte, []byte, error) {
	// Create command with context
	cmd := exec.CommandContext(ctx, r.path, args...)

	// Set up stdout and stderr capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up stdin if input is provided
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	// Log command execution
	logger := log.Ctx(ctx)
	if logger.GetLevel() <= zerolog.DebugLevel {
		logger.Debug().
			Str("tool", r.name).
			Str("path", r.path).
			Str("args", strings.Join(args, " ")).
			Msg("Executing external tool")
	}

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Only log the error if it's not due to context cancellation
		select {
		case <-ctx.Done():
			return stdout.Bytes(), stderr.Bytes(), ctx.Err()
		default:
			logger.Error().
				Err(err).
				Str("tool", r.name).
				Str("stderr", stderr.String()).
				Msg("Tool execution failed")
		}
	}

	return stdout.Bytes(), stderr.Bytes(), err
}

// ToolName returns the name of the tool.
func (r *CommandRunner) ToolName() string {
	return r.name
}

// ToolPath returns the path to the tool.
func (r *CommandRunner) ToolPath() string {
	return r.path
}

// findCommand attempts to locate a command on the PATH.
func findCommand(name string) (string, error) {
	// On Windows, try some common extensions
	if isWindows() {
		// Try with exe extension first
		exeName := name + ".exe"
		path, err := exec.LookPath(exeName)
		if err == nil {
			return path, nil
		}

		// Try with cmd extension
		cmdName := name + ".cmd"
		path, err = exec.LookPath(cmdName)
		if err == nil {
			return path, nil
		}

		// Try with bat extension
		batName := name + ".bat"
		path, err = exec.LookPath(batName)
		if err == nil {
			return path, nil
		}
	}

	// Try without extension
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("tool %s not found on PATH: %w", name, err)
	}

	return filepath.Clean(path), nil
}

// isWindows returns true if the current OS is Windows.
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// NewMagikaRunner creates a new Runner for the Magika tool.
func NewMagikaRunner(customPath string) (Runner, error) {
	return NewRunner("magika", customPath)
}

// NewExiftoolRunner creates a new Runner for the ExifTool tool.
func NewExiftoolRunner(customPath string) (Runner, error) {
	return NewRunner("exiftool", customPath)
}

// NewJdupesRunner creates a new Runner for the jdupes tool.
func NewJdupesRunner(customPath string) (Runner, error) {
	return NewRunner("jdupes", customPath)
}

// NewFileRunner creates a new Runner for the file command.
func NewFileRunner(customPath string) (Runner, error) {
	return NewRunner("file", customPath)
}
