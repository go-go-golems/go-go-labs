# GitHub GraphQL CLI

A command-line tool for interacting with GitHub's GraphQL API, specifically designed for Projects v2 (Beta). Built with Go and the [Glazed](https://github.com/go-go-golems/glazed) framework.

## Features

- **Project Information**: Get details about GitHub Projects v2
- **Field Management**: List project fields, including custom fields and their options
- **Item Management**: List project items with their field values
- **Issue Creation**: Create issues and optionally add them to projects
- **Rich Output**: Support for multiple output formats (JSON, YAML, CSV, Markdown, etc.)
- **Structured Data**: Leverages Glazed framework for powerful data filtering and formatting

## Installation

### Prerequisites

- Go 1.22 or higher
- GitHub Personal Access Token with `project` scope

### Build from Source

```bash
git clone https://github.com/go-go-golems/go-go-labs
cd go-go-labs/cmd/apps/github-projects
go build .
```

### Configuration

Set your GitHub token as an environment variable:

```bash
export GITHUB_TOKEN="your_github_token_here"
```

Or create a `.envrc` file:

```bash
export GITHUB_TOKEN="your_github_token_here"
```

## Usage

### Get Project Information

```bash
./github-projects project --owner=myorg --number=5
```

Output formats:
```bash
# JSON output
./github-projects project --owner=myorg --number=5 --output=json

# Select specific fields
./github-projects project --owner=myorg --number=5 --fields=title,total_items
```

### List Project Fields

```bash
./github-projects fields --owner=myorg --number=5
```

This shows all fields including:
- Field names and IDs
- Field types (text, number, date, single-select, iteration)
- Options for single-select fields
- Iterations for iteration fields

### List Project Items

```bash
./github-projects items --owner=myorg --number=5 --limit=10
```

Shows:
- Item details (title, number, URL, assignees)
- Field values for each item
- Content type (Issue, Pull Request, Draft Issue)

### Create Issues

```bash
# Create a simple issue
./github-projects create-issue \
  --repo-owner=myorg \
  --repo-name=myrepo \
  --title="Bug fix needed" \
  --body="Description of the bug"

# Create issue and add to project
./github-projects create-issue \
  --repo-owner=myorg \
  --repo-name=myrepo \
  --title="New feature request" \
  --body="Feature description" \
  --project-owner=myorg \
  --project-number=5
```

## Command Reference

### `project`
Get information about a GitHub Project v2.

**Flags:**
- `--owner` (required): Organization or user name that owns the project
- `--number` (required): Project number
- `--log-level`: Log level (trace, debug, info, warn, error)

### `fields`
List all fields for a GitHub Project v2.

**Flags:**
- `--owner` (required): Organization or user name that owns the project
- `--number` (required): Project number
- `--log-level`: Log level

### `items`
List all items in a GitHub Project v2.

**Flags:**
- `--owner` (required): Organization or user name that owns the project
- `--number` (required): Project number
- `--limit`: Maximum number of items to return (default: 20)
- `--log-level`: Log level

### `create-issue`
Create a GitHub issue and optionally add it to a project.

**Flags:**
- `--repo-owner` (required): Repository owner
- `--repo-name` (required): Repository name
- `--title` (required): Issue title
- `--body`: Issue body
- `--project-owner`: Project owner (to add issue to project)
- `--project-number`: Project number (to add issue to project)
- `--log-level`: Log level

## Output Formats

Thanks to the Glazed framework, all commands support multiple output formats:

- `table` (default): ASCII table
- `json`: JSON format
- `yaml`: YAML format
- `csv`: Comma-separated values
- `markdown`: Markdown table
- `excel`: Excel spreadsheet

Use the `--output` flag to specify the format:

```bash
./github-projects project --owner=myorg --number=5 --output=json
```

## Field Filtering

Use Glazed's powerful field filtering capabilities:

```bash
# Select specific fields
./github-projects items --owner=myorg --number=5 --fields=title,type,assignees

# Remove specific fields
./github-projects items --owner=myorg --number=5 --filter=item_id,content_type

# Sort by field
./github-projects items --owner=myorg --number=5 --sort-by=title
```

## Advanced Usage

### Templates

Use Go templates for custom output:

```bash
./github-projects project --owner=myorg --number=5 \
  --template="Project: {{.title}} ({{.total_items}} items)"
```

### Pagination

```bash
# Skip first 10 items, limit to 5
./github-projects items --owner=myorg --number=5 \
  --glazed-skip=10 --glazed-limit=5
```

### JQ Queries

Apply jq queries to the output:

```bash
./github-projects items --owner=myorg --number=5 \
  --output=json --jq='.[] | select(.type == "ISSUE")'
```

## Architecture

This CLI is built using:

- **Go**: Core language
- **Glazed**: Framework for structured CLI applications
- **Cobra**: Command-line interface framework
- **machinebox/graphql**: GraphQL client library
- **Zerolog**: Structured logging

The architecture follows Glazed's command patterns:
- Commands implement the `GlazeCommand` interface
- Structured data is output as rows using Glazed's type system
- Parameter layers organize command options logically

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Related Projects

- [Glazed](https://github.com/go-go-golems/glazed) - Framework for building CLI applications
- [GitHub CLI](https://github.com/cli/cli) - Official GitHub CLI
- [GraphQL](https://graphql.org/) - Query language for APIs
