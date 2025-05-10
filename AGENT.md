# AGENT.md

## Go Commands
- Build: `make build` (runs go generate ./... and go build ./...)
- Test: `make test` (runs all tests with `go test ./...`)
- Run specific test: `go test ./path/to/package -run TestName`
- Lint: `make lint` (uses golangci-lint)

## Python Commands
- Install dependencies: `pip install -r requirements.txt`
- Run Flask app: `python app.py`
- Run specific test: `python -m unittest path/to/test.py::TestClass::test_method`

## Code Style Guidelines
- Go: Uses gofmt, go 1.23+, github.com/pkg/errors for error wrapping
- Go: Uses zerolog for logging, cobra for CLI, viper for config
- Go: Follow standard naming (CamelCase for exported, camelCase for unexported)
- Python: PEP 8 formatting, uses logging module for structured logging
- Python: Try/except blocks with specific exceptions and error logging
- Use interfaces to define behavior, prefer structured concurrency
- Pre-commit hooks use lefthook (configured in lefthook.yml)