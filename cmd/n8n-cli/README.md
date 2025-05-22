# n8n-cli

A command line tool for managing n8n workflows via the REST API.

## Usage

```
n8n-cli [command] [flags]
```

## Global Flags

```
  --help                   Help for any command
  --log-file string        Log file (default: stderr)
  --log-format string      Log format (json, text) (default "text")
  --log-level string       Log level (trace, debug, info, warn, error, fatal) (default "info")
  --with-caller            Log caller information
```

## Environment Variables

You can also set logging options using environment variables:

```
N8N_CLI_LOG_LEVEL=debug     # Set log level to debug
N8N_CLI_LOG_FORMAT=json     # Set log format to JSON
N8N_CLI_LOG_FILE=n8n.log    # Log to a file instead of stderr
N8N_CLI_WITH_CALLER=true    # Include caller information in logs
```

## Available Commands

- `add-node` - Add a node to a workflow
- `connect-nodes` - Connect nodes in a workflow
- `create-workflow` - Create a new workflow
- `get-execution` - Get execution details
- `get-nodes` - Get available node types
- `get-workflow` - Get a workflow by ID
- `list-executions` - List workflow executions
- `list-workflows` - List all workflows in the n8n instance

## Examples

### List workflows with debug logging

```
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --log-level debug
```

### List workflows with pagination (using cursor)

```
# First page
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --limit 50

# Next page (use the nextCursor value from previous results)
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --limit 50 --cursor YOUR_CURSOR_VALUE
```

### Get a workflow with trace logging (shows full request/response)

```
n8n-cli get-workflow --id 123 --base-url http://localhost:5678 --api-key YOUR_API_KEY --log-level trace
```

### Create a workflow from a JSON file

```
n8n-cli create-workflow --name "My New Workflow" --file workflow.json --base-url http://localhost:5678 --api-key YOUR_API_KEY 
```

### Get all node types available in n8n

```
n8n-cli get-nodes --base-url http://localhost:5678 --api-key YOUR_API_KEY
```

### List executions for a specific workflow

```
n8n-cli list-executions --workflow-id 123 --base-url http://localhost:5678 --api-key YOUR_API_KEY
```

## Debugging API issues

If you encounter HTTP errors (like 400 Bad Request), use the `--log-level debug` or `--log-level trace` flags to see more information:

```
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --log-level debug
```

The `--log-level trace` flag will show the full request and response bodies, which can be helpful for diagnosing issues with the API:

```
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --log-level trace
```

You can also combine flags for more detailed logging:

```
n8n-cli list-workflows --base-url http://localhost:5678 --api-key YOUR_API_KEY --log-level trace --with-caller --log-format json
```