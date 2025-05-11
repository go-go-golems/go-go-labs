# AGENT.md - Coding Assistant Guide

## Backend (Go) Commands
- Build: `go build ./...`
- Dev server: `go run main.go`
- Lint: `golangci-lint run`
- Format: `gofmt -w .`
- Test: `go test ./...`
- Single test: `go test -run "TestName" ./package`

## Frontend (React) Commands
- Build: `bun run build`
- Dev server: `bun run dev`
- Lint: `bun run lint`
- Format: `bun run format`
- Test: `bun test`
- Single test: `bun test -t "test name"`

## Backend (Go) Guidelines
- **Language**: Go with SQLite for data storage
- **Web Framework**: Echo for REST APIs
- **CLI Framework**: Cobra for command-line applications
- **UI Technology**: HTMX with Bootstrap and templ templating
- **Interfaces**: Use `var _ Interface = &Foo{}` pattern to enforce interface implementation
- **Context**: Always use context argument when appropriate
- **Error Handling**: Use github.com/pkg/errors for wrapping errors
- **Concurrency**: Use errgroup when starting goroutines
- **Package Naming**: Use "defaults" package name instead of "default" (reserved in Go)

## Frontend (React) Guidelines
- **Framework**: React with TypeScript
- **State Management**: Redux Toolkit (RTK) and RTK Query
- **Styling**: TailwindCSS for all styling
- **Icons**: Lucide React for icons
- **Formatting**: Use consistent indentation (2 spaces)
- **Naming**: camelCase for variables/functions, PascalCase for components
- **Types**: Strong TypeScript typing for all components
- **Error Handling**: Use try/catch for async operations

## Project Structure
- Full-stack application with Go backend and React frontend
- Go backend uses SQLite for data persistence
- React frontend follows "Severance" TV show aesthetic