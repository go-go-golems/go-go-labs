---
slug: introduction
title: Getting Started with Datadog CLI
---

# Getting Started with Datadog CLI

The Datadog CLI is a powerful, YAML-driven tool for querying Datadog logs. It follows the go-go-golems architecture pattern, providing composable commands, streaming results, and flexible output formats.

## Quick Setup

1. **Install the CLI:**
   ```bash
   go install github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli@latest
   ```

2. **Set up authentication:**
   ```bash
   export DATADOG_CLI_API_KEY="your-api-key"
   export DATADOG_CLI_APP_KEY="your-app-key"
   export DATADOG_CLI_SITE="datadoghq.com"  # Optional
   ```

3. **Run your first query:**
   ```bash
   datadog-cli logs top_errors --limit 5
   ```

## Key Features

- **YAML-driven queries** - Define reusable query templates
- **Streaming results** - Handle large datasets efficiently
- **Multiple output formats** - table, CSV, JSON, YAML, etc.
- **Repository system** - Organize and share query collections
- **Profile management** - Multiple environments and configurations
- **Raw query support** - Direct Datadog search queries

## Getting Help

- `datadog-cli --help` - General help
- `datadog-cli logs --help` - Logs command help
- `datadog-cli help <topic>` - Detailed topic help
