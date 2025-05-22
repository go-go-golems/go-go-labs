# n8n CLI Tool

A command-line interface for managing n8n workflows using the n8n REST API. This tool allows you to programmatically manage workflows, nodes, and connections in your n8n instance.

## Features

- List workflows with filtering options
- Get workflow details as JSON
- Create new workflows (empty or from JSON file)
- Add nodes to existing workflows
- List nodes in a workflow
- Connect/disconnect nodes
- Update node settings

## Installation

```bash
go install github.com/go-go-golems/go-go-labs/cmd/n8n-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/go-go-golems/go-go-labs.git
cd go-go-labs
go build -o n8n-cli ./cmd/n8n-cli
```

## Authentication

All commands require an API key. You can create one in the n8n UI under Settings → n8n API → "Create an API key".

Provide the API key using the `--api-key` flag with each command or set the `N8N_API_KEY` environment variable.

## Common Flags

All commands support these common options:

- `--base-url`: Base URL of your n8n instance (default: http://localhost:5678)
- `--api-key`: Your n8n API key (required)
- `--output`: Output format (json, yaml, csv, etc. - provided by Glazed framework)

## Commands

### List Workflows

List all workflows in your n8n instance with optional filtering.

```bash
n8n-cli list-workflows --api-key=YOUR_API_KEY [--active] [--limit=50] [--offset=0]
```

Options:
- `--active`: Only show active workflows
- `--limit`: Maximum number of workflows to return
- `--offset`: Pagination offset

### Get Workflow

Get a specific workflow by ID.

```bash
n8n-cli get-workflow --api-key=YOUR_API_KEY --id=WORKFLOW_ID [--output-file=workflow.json]
```

Options:
- `--id`: Workflow ID (required)
- `--output-file`: Save the workflow to a JSON file

### Create Workflow

Create a new workflow.

```bash
n8n-cli create-workflow --api-key=YOUR_API_KEY --name="My New Workflow" [--file=workflow.json] [--active]
```

Options:
- `--name`: Workflow name (required)
- `--file`: JSON file containing workflow definition
- `--active`: Set workflow as active immediately

### Add Node

Add a node to an existing workflow.

```bash
n8n-cli add-node --api-key=YOUR_API_KEY --workflow-id=WORKFLOW_ID --node-file=examples/webhook-node.json
```

Options:
- `--workflow-id`: Workflow ID (required)
- `--node-file`: JSON file containing node definition (required)

### List Nodes

List all nodes in a workflow.

```bash
n8n-cli list-nodes --api-key=YOUR_API_KEY --workflow-id=WORKFLOW_ID
```

Options:
- `--workflow-id`: Workflow ID (required)

### Connect Nodes

Connect two nodes in a workflow.

```bash
n8n-cli connect-nodes --api-key=YOUR_API_KEY --workflow-id=WORKFLOW_ID \
  --source="Webhook Trigger" --target="Respond" \
  [--source-index=0] [--target-index=0]
```

Options:
- `--workflow-id`: Workflow ID (required)
- `--source`: Source node name (required)
- `--target`: Target node name (required)
- `--source-index`: Source output index (default: 0)
- `--target-index`: Target input index (default: 0)
- `--disconnect`: Disconnect instead of connect

### Set Node Settings

Update a node's settings.

```bash
n8n-cli set-node-settings --api-key=YOUR_API_KEY --workflow-id=WORKFLOW_ID \
  --node-name="Respond" --settings-file=examples/node-settings.json
```

Options:
- `--workflow-id`: Workflow ID (required)
- `--node-name`: Node name (required)
- `--settings-file`: JSON file containing node settings (required)

## Example JSON Files

The `examples/` directory contains sample JSON files for nodes and workflows:

- `webhook-node.json`: Example Webhook node
- `respond-node.json`: Example Respond node
- `node-settings.json`: Example node settings
- `workflow.json`: Example workflow with webhook and respond nodes

## Complete Example: Create API Endpoint Workflow

Create a simple API endpoint workflow with a webhook and response:

```bash
# 1. Create a new workflow
n8n-cli create-workflow --api-key=YOUR_API_KEY --name="API Endpoint"
# Note the workflow ID from the output, e.g. "id": "123"

# 2. Add a webhook node
n8n-cli add-node --api-key=YOUR_API_KEY --workflow-id=123 --node-file=examples/webhook-node.json

# 3. Add a respond node
n8n-cli add-node --api-key=YOUR_API_KEY --workflow-id=123 --node-file=examples/respond-node.json

# 4. Connect the nodes
n8n-cli connect-nodes --api-key=YOUR_API_KEY --workflow-id=123 \
  --source="Webhook Trigger" --target="Respond"

# 5. Activate the workflow
n8n-cli create-workflow --api-key=YOUR_API_KEY --id=123 --active=true
```

## Environment Variables

- `N8N_API_KEY`: Your n8n API key
- `N8N_BASE_URL`: Base URL of your n8n instance

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.