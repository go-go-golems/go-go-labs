# Create Pull Request Tool

This is a Go-based CLI tool for generating pull request descriptions using an LLM. It analyzes your code changes and commit history to produce comprehensive, well-structured PR descriptions.

## Features

- Generate PR descriptions using an LLM (requires `pinocchio` to be installed)
- Automatically gather git diffs and commit history
- Configure LLM prompts and styles
- Customize diff context and exclusion patterns
- Include issue details and additional context

## Usage

```bash
# Basic usage
gopr create "Implemented feature X with tests"

# Specify an issue and title
gopr create --issue 123 --title "feat: implement X" "Added feature X with comprehensive tests"

# Customize diff context
gopr create --diff-context-size 5 --exclude "*.md,go.sum" "Refactored authentication system"

# Get just the diff without creating a PR
gopr get-diff --exclude "*.md"

# Create a PR from an existing YAML file
gopr create-from-yaml /path/to/pr.yaml
```

## Installation

The tool is currently a prototype. You can build it from source:

```bash
cd cmd/apps/create-pull-request
go build
```

## Current Status

This is a prototype implementation with the following limitations:

- Git and GitHub CLI adapters are currently mocked
- LLM adapter uses the real pinocchio tool
- No TUI implementation yet
- Limited error handling

## Dependencies

- Requires the `pinocchio` CLI tool to be installed and available in your PATH
- Uses `github.com/spf13/cobra` for CLI commands
- Uses `gopkg.in/yaml.v3` for YAML parsing