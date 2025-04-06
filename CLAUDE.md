# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test/Lint Commands
- Build: `make build` (runs go generate ./... and go build ./...)
- Test: `make test` (runs all tests with `go test ./...`)
- Run specific test: `go test ./path/to/package -run TestName`
- Lint: `make lint` (uses golangci-lint)

## Code Style Guidelines
- Go version: 1.23+
- Use gofmt formatting (enforced by linter)
- Error handling: Use `github.com/pkg/errors` for wrapping errors with context
- Logging: Use `github.com/rs/zerolog` for structured logging
- CLI apps: Use `github.com/spf13/cobra` for command-line interfaces
- Config: Use `github.com/spf13/viper` for configuration management
- Use proper code organization with clear separation of concerns
- Follow standard Go naming conventions (CamelCase for exported, camelCase for unexported)
- Use interfaces to define behavior and allow for testing
- Prefer structured concurrency with errgroup for worker pools

## Git Workflow
- Pre-commit hooks: Lint and test run automatically before commits
- Pre-push hooks: Lint and test run before pushing