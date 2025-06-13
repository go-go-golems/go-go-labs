# Built-in Datadog Queries

This directory contains built-in query templates for common Datadog operations.

## Available Queries

### recent-logs
Get recent logs from Datadog with optional filtering.

**Parameters:**
- `service`: Service name to filter by
- `level`: Log level to filter by (default: all levels)
- `limit`: Maximum number of logs to return (default: 100)

**Example:**
```bash
datadog-cli recent-logs --service web --level ERROR --limit 50
```

### service-logs  
Get logs for a specific service with time range filtering.

**Parameters:**
- `service`: Service name (required)
- `from`: Start time (default: -1h)
- `to`: End time (default: now)
- `limit`: Maximum number of logs (default: 100)

**Example:**
```bash
datadog-cli service-logs --service api --from "-2h" --limit 200
```

### top-errors
Find the most frequent error messages in your logs.

**Parameters:**
- `service`: Service name to filter by (optional)
- `hours`: Number of hours to look back (default: 1)
- `limit`: Maximum number of error groups (default: 10)

**Example:**
```bash
datadog-cli top-errors --service web --hours 6 --limit 20
```

## Custom Queries

You can create your own queries by:

1. Creating a local repository directory:
   ```bash
   mkdir -p ~/.datadog-cli/queries
   ```

2. Adding your query YAML files to the directory

3. Using the `datadog-cli repositories` commands to manage your repositories

## Query Template Format

All queries use the same YAML format. See the main documentation for details on creating custom query templates.
