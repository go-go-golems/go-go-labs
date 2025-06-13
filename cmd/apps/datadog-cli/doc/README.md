# Datadog CLI Documentation

## Overview

The Datadog CLI is a YAML-driven tool for querying Datadog logs using the Logs Search API. It follows the go-go-golems architecture pattern with composable commands, streaming results, and flexible output formats.

## Quick Start

### Authentication

Set your Datadog credentials as environment variables:

```bash
export DATADOG_CLI_API_KEY="your-api-key"
export DATADOG_CLI_APP_KEY="your-app-key"  
export DATADOG_CLI_SITE="datadoghq.com"  # Optional, defaults to datadoghq.com
```

### Basic Usage

```bash
# List available commands
datadog-cli logs --help

# Get top errors from the last hour
datadog-cli logs top_errors

# Get logs for a specific service
datadog-cli logs service_logs --service web-api --level error

# Run a custom YAML query
datadog-cli logs run my-query.yaml

# Execute a raw Datadog query
datadog-cli logs query "service:web-api AND status:error"
```

## Commands

### Built-in Queries

- **top_errors** - Top error messages in a time range
- **service_logs** - Service-specific log filtering
- **recent_logs** - Recent logs with multi-facet filtering

### Utility Commands

- **run** - Execute a YAML query file
- **query** - Execute a raw Datadog search query

## Configuration

The CLI supports configuration via:
- Environment variables (`DATADOG_CLI_*`)
- Configuration files (`~/.datadog-cli/config.yaml`)
- Command-line flags
- Profiles (`~/.datadog-cli/profiles.yaml`)

## Output Formats

All commands support multiple output formats:
- **table** (default) - Pretty ASCII tables
- **csv** - Comma-separated values
- **json** - JSON output
- **yaml** - YAML output
- **markdown** - Markdown tables

Example:
```bash
datadog-cli logs top_errors --output json
datadog-cli logs service_logs --service api --output csv
```
