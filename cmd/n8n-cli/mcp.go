package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/go-go-labs/cmd/n8n-cli/pkg/n8n"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
)

// configureMCPLogging configures logging to output to stderr when in MCP mode
// This is critical because MCP uses stdout for protocol communication,
// so all application logging must go to stderr to avoid protocol interference.
var mcpLoggingConfigured = false

func configureMCPLogging() {
	if mcpLoggingConfigured {
		return
	}
	// Configure zerolog to output to stderr to avoid interfering with MCP protocol on stdout
	// log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	mcpLoggingConfigured = true
}

// registerMCPTools registers all n8n-cli commands as MCP tools
func registerMCPTools() []embeddable.ServerOption {
	return []embeddable.ServerOption{
		// List workflows
		embeddable.WithEnhancedTool("list_workflows", listWorkflowsHandler,
			embeddable.WithEnhancedDescription(`Retrieve a comprehensive list of all workflows from your n8n instance.

This tool fetches workflow summaries with essential metadata including ID, name, active status, creation/update timestamps, and tags. Unlike the full workflow details, this provides a clean overview without node configurations or connection data, making it perfect for browsing and discovery.

Use this tool to:
- Get an overview of all available workflows
- Check workflow statuses (active/inactive)
- Find workflow IDs for use with other tools
- Monitor your automation portfolio

The tool returns up to 50 workflows by default and only fetches inactive workflows to provide a comprehensive view while maintaining reasonable response times.`),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),

		// Get workflow
		embeddable.WithEnhancedTool("get_workflow", getWorkflowHandler,
			embeddable.WithEnhancedDescription(`Retrieve detailed information about a specific workflow including its complete structure, nodes, connections, and configuration.

This tool fetches the complete workflow definition including:
- Workflow metadata (name, description, active status, tags)
- All nodes with their types, names, parameters, and positions
- Node connections and data flow mapping
- Trigger configurations and execution settings
- Version information and update history

The workflow can be returned in two formats:
1. **JSON format** (default): Complete workflow definition with all technical details
2. **Mermaid diagram**: Visual flowchart representation showing node relationships

Use this tool to:
- Inspect workflow structure and logic
- Debug workflow configurations
- Document workflow architecture
- Analyze node connections and data flow
- Prepare workflows for cloning or modification`),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription(`Unique identifier of the workflow to retrieve. 

This is typically a numeric string (e.g., "123", "456") that uniquely identifies the workflow in your n8n instance. You can find workflow IDs using the list_workflows tool.

Example values: "1", "42", "1001"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithBooleanProperty("as_mermaid",
				embeddable.PropertyDescription(`Controls the output format of the workflow data.

When set to false (default): Returns the complete workflow as JSON with all technical details including node configurations, parameters, connections, and metadata.

When set to true: Returns a Mermaid flowchart diagram showing the visual structure of the workflow. This is ideal for:
- Creating documentation
- Understanding workflow flow at a glance  
- Sharing workflow architecture with stakeholders
- Visual debugging of complex workflows

The Mermaid output can be rendered in any Mermaid-compatible viewer or documentation system.`),
				embeddable.DefaultBool(false),
			),
		),

		// Create workflow
		embeddable.WithEnhancedTool("create_workflow", createWorkflowHandler,
			embeddable.WithEnhancedDescription(`Create a new workflow in your n8n instance with either a basic structure or from a complete workflow definition.

This tool provides two ways to create workflows:

**1. Basic Workflow Creation:**
When only providing a name, creates a minimal workflow with:
- Empty nodes array (ready for adding nodes)
- Empty connections object (ready for connecting nodes)
- Specified name and active status
- Default n8n workflow structure

**2. Advanced Workflow Creation:**
When providing a JSON file path, imports a complete workflow definition including:
- Pre-configured nodes with parameters
- Established node connections
- Complex workflow logic and triggers
- Custom settings and configurations

The created workflow will be immediately available in your n8n instance and can be activated for execution if the active parameter is set to true.

Use this tool to:
- Bootstrap new automation workflows
- Import workflows from backups or exports
- Clone workflows by exporting and re-importing
- Set up template workflows for common patterns
- Migrate workflows between n8n instances

**Important:** Newly created workflows start in inactive state by default for safety. Activate them manually or set active=true if you want immediate execution.`),
			embeddable.WithDestructiveHint(false),
			embeddable.WithStringProperty("name",
				embeddable.PropertyDescription(`Human-readable name for the new workflow.

This name will be displayed in the n8n UI and should be descriptive enough to identify the workflow's purpose. Names don't need to be unique but should be meaningful.

Best practices:
- Use clear, descriptive names (e.g., "Customer Onboarding Email", "Daily Sales Report")
- Include the main function or trigger (e.g., "Slack Alert on Error", "Weekly Backup Task")  
- Avoid special characters that might cause issues in URLs or file systems
- Keep names concise but informative (typically 20-50 characters)

Examples: "Process New Orders", "Backup Database Daily", "Notify Team on Deploy"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithBooleanProperty("active",
				embeddable.PropertyDescription(`Determines whether the workflow should be activated immediately after creation.

When set to false (default): Creates the workflow in inactive state. This is the safest option as it allows you to:
- Review the workflow configuration
- Add or modify nodes before execution
- Test the workflow manually
- Ensure all required credentials and configurations are set

When set to true: Activates the workflow immediately after creation. Use this when:
- The workflow is complete and ready to run
- All required integrations and credentials are configured
- You want the workflow to start processing triggers immediately
- Importing a tested workflow from another instance

**Recommendation:** Keep as false for new workflows, set to true only for fully configured imports.`),
				embeddable.DefaultBool(false),
			),
			embeddable.WithStringProperty("file",
				embeddable.PropertyDescription(`Optional path to a JSON file containing a complete workflow definition.

When provided, the tool will read the workflow structure from this file instead of creating a basic empty workflow. The file should contain a valid n8n workflow export in JSON format.

File format expectations:
- Valid JSON structure matching n8n workflow schema
- Contains nodes array with node definitions
- Contains connections object defining node relationships
- May include triggers, credentials references, and settings

Common use cases:
- Importing workflows exported from n8n UI
- Restoring workflows from backups
- Deploying pre-built workflow templates
- Migrating workflows between environments

**Note:** If not provided, a basic empty workflow will be created with the specified name and active status.

Example: "/path/to/exported-workflow.json", "./workflows/customer-onboarding.json"`),
			),
		),

		// Get nodes
		embeddable.WithEnhancedTool("get_nodes", getNodesHandler,
			embeddable.WithEnhancedDescription(`Retrieve a comprehensive catalog of all available node types in your n8n instance.

This tool fetches the complete node registry, providing detailed information about every node type that can be used in workflows. The response includes both core n8n nodes and any custom or community nodes installed in your instance.

For each node type, you'll receive:
- **Node metadata**: Name, display name, description, and version
- **Category classification**: Organization by function (trigger, action, transform, etc.)
- **Input/output specifications**: Expected data structure and connection types
- **Parameter schema**: Available configuration options and their types
- **Credential requirements**: Authentication and API key needs
- **Documentation links**: Help resources and usage examples

Use this tool to:
- Discover available integrations and capabilities
- Plan workflow architecture and node selection
- Understand parameter requirements before adding nodes
- Verify node availability before importing workflows
- Explore new nodes and integrations
- Troubleshoot missing node types in workflows

**Pro tip:** The response can be quite large due to the comprehensive nature of the node catalog. Consider filtering or searching the results for specific integration names or categories you're interested in.`),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),

		// Add node to workflow
		embeddable.WithEnhancedTool("add_node", addNodeHandler,
			embeddable.WithEnhancedDescription(`Add a new node to an existing workflow with specified configuration and positioning.

This tool extends a workflow by inserting a new node with the specified type, parameters, and visual positioning. The node will be added to the workflow but won't be connected to other nodes automatically - use the connect_nodes tool separately to establish data flow.

The process involves:
1. **Fetching the current workflow** to preserve existing structure
2. **Creating the new node** with specified type and parameters  
3. **Positioning the node** in the visual editor at given coordinates
4. **Updating the workflow** with the new node included

**Node Configuration:**
- Node type must match available types (use get_nodes to explore options)
- Parameters should match the node's schema requirements
- Position coordinates determine visual placement in the editor
- Node names should be unique within the workflow

Use this tool to:
- Extend existing workflows with new functionality
- Add data sources, transformations, or destinations
- Build workflows incrementally node by node
- Insert nodes at specific positions for organized layouts
- Add conditional logic or error handling nodes

**Important:** After adding nodes, you typically need to connect them using the connect_nodes tool to establish proper data flow. The newly added node will appear disconnected until explicitly connected.`),
			embeddable.WithStringProperty("workflow_id",
				embeddable.PropertyDescription(`Unique identifier of the workflow to modify.

This should be the numeric string ID of an existing workflow in your n8n instance. You can find workflow IDs using the list_workflows tool.

The workflow must exist and be accessible for modification. If the workflow is currently running executions, adding nodes may affect ongoing processes.

Example values: "1", "42", "1001"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("type",
				embeddable.PropertyDescription(`The specific node type to add to the workflow.

This must be an exact match for a node type available in your n8n instance. Node types are case-sensitive and follow n8n's naming conventions.

Common node types include:
- **Triggers**: "Webhook", "Cron", "Manual Trigger", "HTTP Request"
- **Data sources**: "HTTP Request", "Google Sheets", "Airtable", "MySQL"
- **Transformations**: "Code", "Set", "If", "Switch", "Merge"
- **Actions**: "Email Send", "Slack", "Discord", "File Write"
- **Utilities**: "Wait", "Stop and Error", "No Operation"

Use the get_nodes tool to see all available node types with their exact names.

Examples: "HTTP Request", "Google Sheets", "Set", "If", "Webhook"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("name",
				embeddable.PropertyDescription(`Human-readable name for the new node instance.

This name will be displayed in the workflow editor and should be descriptive of the node's specific purpose within this workflow. Names should be unique within the workflow to avoid confusion.

Naming best practices:
- Be specific about the node's function (e.g., "Fetch Customer Data", "Send Welcome Email")
- Include context relevant to your workflow (e.g., "Validate Email Format", "Check Inventory Level")
- Avoid generic names like "Node1" or "HTTP" - be descriptive
- Use consistent naming patterns across your workflows

Examples: "Get User Profile", "Calculate Shipping Cost", "Send Slack Notification", "Transform Data Format"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("parameters",
				embeddable.PropertyDescription(`Node-specific configuration parameters as a JSON string.

This field contains the actual configuration that determines how the node behaves. The parameter structure varies significantly between node types and should match the node's schema requirements.

**Format:** Valid JSON object as a string
**Content:** Node-specific parameters based on the node type

Common parameter patterns:
- **HTTP Request**: {"method": "GET", "url": "https://api.example.com/users"}
- **Set node**: {"values": {"key1": "value1", "key2": "value2"}}
- **If node**: {"conditions": {"string": [{"value1": "{{$json.status}}", "operation": "equal", "value2": "active"}]}}
- **Email**: {"toEmail": "user@example.com", "subject": "Welcome!", "text": "Hello!"}

**Default:** Empty object "{}" if not specified - creates node with default parameters.

**Tip:** Use the n8n UI to configure a node first, then export the workflow to see the exact parameter structure.`),
			),
			embeddable.WithIntProperty("position_x",
				embeddable.PropertyDescription(`Horizontal position (X coordinate) of the node in the visual workflow editor.

This determines where the node appears horizontally in the workflow canvas. Coordinates are in pixels from the left edge of the canvas.

Positioning guidelines:
- **0**: Left edge of the canvas
- **200-400**: Good spacing between nodes horizontally
- **Negative values**: Allowed, places nodes to the left of origin
- **Large values**: Nodes will be positioned far to the right

**Layout tips:**
- Space nodes 200-300 pixels apart horizontally for readability
- Align related nodes vertically with similar X coordinates
- Use consistent spacing for professional-looking workflows
- Consider the workflow's logical flow when positioning

**Default:** 0 (left edge of canvas)`),
				embeddable.DefaultNumber(0),
			),
			embeddable.WithIntProperty("position_y",
				embeddable.PropertyDescription(`Vertical position (Y coordinate) of the node in the visual workflow editor.

This determines where the node appears vertically in the workflow canvas. Coordinates are in pixels from the top edge of the canvas.

Positioning guidelines:
- **0**: Top edge of the canvas
- **100-200**: Good vertical spacing between nodes
- **Negative values**: Allowed, places nodes above the origin
- **Large values**: Nodes will be positioned further down

**Layout tips:**
- Space nodes 100-150 pixels apart vertically for clarity
- Use horizontal rows for parallel processing branches
- Follow the logical flow: triggers at top, outputs at bottom
- Group related nodes at similar Y coordinates for visual organization

**Default:** 0 (top edge of canvas)`),
				embeddable.DefaultNumber(0),
			),
		),

		// Connect nodes
		embeddable.WithEnhancedTool("connect_nodes", connectNodesHandler,
			embeddable.WithEnhancedDescription(`Establish data flow connections between two nodes in a workflow.

This tool creates the essential links that allow data to flow from one node to another, defining the execution order and data transfer paths in your workflow. Connections determine how information processed by one node is passed to subsequent nodes.

**Connection Fundamentals:**
- **Source node**: The node that produces/outputs data
- **Target node**: The node that receives/processes the incoming data  
- **Output types**: Different data streams a node can produce (main, error, etc.)
- **Input types**: Different data streams a node can accept (main, etc.)

**How Data Flows:**
1. Source node completes execution and produces output data
2. n8n transfers this data through the established connection
3. Target node receives the data and begins its execution
4. Process repeats for subsequent connected nodes

**Connection Types:**
- **Main connections**: Primary data flow for normal operations
- **Error connections**: Handle errors and exceptions
- **Conditional connections**: Based on IF/Switch node outcomes

Use this tool to:
- Link newly added nodes into workflow logic
- Establish proper execution sequences  
- Create data transformation pipelines
- Set up error handling paths
- Build complex branching workflows

**Important:** Nodes without proper connections may not execute or receive expected data. Always verify connection logic matches your intended workflow behavior.`),
			embeddable.WithStringProperty("workflow_id",
				embeddable.PropertyDescription(`Unique identifier of the workflow containing the nodes to connect.

This should be the numeric string ID of an existing workflow where both the source and target nodes already exist. You can find workflow IDs using the list_workflows tool.

The workflow will be updated with the new connection, potentially affecting execution flow if the workflow is active.

Example values: "1", "42", "1001"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("source_node",
				embeddable.PropertyDescription(`Name of the node that will send data (the data producer).

This must exactly match the name of an existing node in the specified workflow. The source node will execute first and its output data will be passed to the target node.

**Source Node Characteristics:**
- Must exist in the workflow before creating the connection
- Should produce output data compatible with the target node's expectations
- Can have multiple outgoing connections to different target nodes
- Name matching is case-sensitive and must be exact

**Common Source Node Types:**
- Trigger nodes (start the workflow)
- Data retrieval nodes (HTTP Request, Database queries)
- Transformation nodes (Set, Code, Function)
- Conditional nodes (If, Switch) - for branching logic

Examples: "Webhook Trigger", "Get Customer Data", "Process Orders", "Check Payment Status"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("target_node",
				embeddable.PropertyDescription(`Name of the node that will receive data (the data consumer).

This must exactly match the name of an existing node in the specified workflow. The target node will execute after the source node completes and will receive the source node's output data as input.

**Target Node Characteristics:**
- Must exist in the workflow before creating the connection
- Should accept input data in the format produced by the source node
- Can receive connections from multiple source nodes (data merging)
- Name matching is case-sensitive and must be exact

**Common Target Node Types:**
- Action nodes (Send Email, Create Database Record)
- Transformation nodes (Set, Code, Function)
- Conditional nodes (If, Switch) - for decision making
- Output nodes (Response, File Write)

Examples: "Send Welcome Email", "Update Customer Record", "Generate Report", "Slack Notification"`),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("source_output",
				embeddable.PropertyDescription(`The specific output stream from the source node to connect.

Most nodes have a "main" output for normal data flow, but some nodes provide multiple output types for different scenarios.

**Output Stream Types:**
- **"main"** (default): Primary data output for normal execution
- **"error"**: Error data output for exception handling  
- **"true"/"false"**: Conditional outputs from IF nodes
- **"0", "1", "2", etc.**: Numbered outputs from Switch nodes
- **Custom names**: Some nodes define specific output names

**When to Use Different Outputs:**
- Use "main" for standard data flow (most common)
- Use "error" when setting up error handling paths
- Use conditional outputs for branching logic
- Check node documentation for available output streams

**Default:** "main" (covers 90% of use cases)`),
				embeddable.DefaultString("main"),
			),
			embeddable.WithStringProperty("target_input",
				embeddable.PropertyDescription(`The specific input stream on the target node to connect to.

Most nodes accept data through their "main" input, but some specialized nodes have multiple input types for different data streams.

**Input Stream Types:**
- **"main"** (default): Primary data input for normal processing
- **"secondary"**: Additional data input for merge operations
- **"metadata"**: Configuration or metadata input
- **Custom names**: Some nodes define specific input names

**When to Use Different Inputs:**
- Use "main" for standard data flow (most common)
- Use "secondary" for merge nodes that combine data streams
- Use custom inputs when required by specialized nodes
- Check target node documentation for available input streams

**Special Cases:**
- Merge nodes often have "input1" and "input2" 
- Compare nodes may have "input1" and "input2"
- Some transformation nodes accept metadata through separate inputs

**Default:** "main" (covers 95% of use cases)`),
				embeddable.DefaultString("main"),
			),
		),

		// List executions
		embeddable.WithEnhancedTool("list_executions", listExecutionsHandler,
			embeddable.WithEnhancedDescription(`Retrieve a list of workflow execution records from your n8n instance.

This tool provides access to the execution history, showing when workflows have run, their completion status, duration, and basic outcome information. Execution records are essential for monitoring, debugging, and analyzing workflow performance.

**Execution Information Included:**
- **Execution ID**: Unique identifier for each run
- **Workflow details**: Which workflow was executed
- **Execution status**: Success, error, failed, waiting, running
- **Timing information**: Start time, end time, duration
- **Trigger information**: What initiated the execution
- **Data summary**: Input/output data size and basic structure
- **Error details**: Failure reasons and error messages (if applicable)

**Filtering and Scope:**
- Filter by specific workflow to see runs for just one automation
- Limit results to manage response size and performance
- Results are ordered by execution time (most recent first)
- Includes both manual and automatic executions

Use this tool to:
- Monitor workflow performance and reliability
- Debug failed executions and identify patterns
- Track execution frequency and timing
- Analyze workflow usage across your instance
- Identify high-volume or problematic workflows
- Generate usage reports and statistics

**Performance Note:** Large n8n instances may have thousands of executions. Use the workflow_id filter and reasonable limits for better performance.`),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
			embeddable.WithStringProperty("workflow_id",
				embeddable.PropertyDescription(`Optional filter to show executions only for a specific workflow.

When provided, only execution records for the specified workflow will be returned. This is useful for focused analysis of a particular automation's performance and behavior.

**When to use workflow filtering:**
- Debugging a specific workflow's issues
- Analyzing performance of a particular automation
- Tracking execution frequency for one workflow
- Investigating errors in a specific process

**When to omit (show all executions):**
- Getting an overview of all automation activity
- Identifying the most active workflows
- Finding recent errors across all workflows
- General monitoring and health checks

The workflow ID should be the numeric string identifier (e.g., "1", "42", "1001"). You can find workflow IDs using the list_workflows tool.

**Example:** "123" to see only executions of workflow #123`),
			),
			embeddable.WithIntProperty("limit",
				embeddable.PropertyDescription(`Maximum number of execution records to return in the response.

This controls the response size and helps manage performance when dealing with large execution histories. Results are returned in reverse chronological order (most recent first).

**Recommended limits by use case:**
- **Quick status check**: 5-10 executions
- **Recent activity review**: 20-50 executions  
- **Detailed analysis**: 50-100 executions
- **Historical research**: Up to 100 executions

**Performance considerations:**
- Larger limits take longer to process and return
- Very high limits may timeout on busy instances
- Consider multiple smaller requests for extensive analysis
- Balance between completeness and response time

**Range:** 1 to 100 executions
**Default:** 20 executions (good balance for most use cases)`),
				embeddable.DefaultNumber(20),
				embeddable.Minimum(1),
				embeddable.Maximum(100),
			),
		),

		// Get execution details
		embeddable.WithEnhancedTool("get_execution", getExecutionHandler,
			embeddable.WithEnhancedDescription(`Retrieve comprehensive details about a specific workflow execution.

This tool provides deep visibility into a single execution, including complete data flow, node-by-node results, timing information, and detailed error reports. This is essential for debugging, auditing, and understanding exactly what happened during a workflow run.

**Detailed Execution Information:**
- **Execution metadata**: ID, workflow, status, start/end times, duration
- **Node execution data**: Input data, output data, execution time for each node
- **Data flow tracking**: How data transformed as it moved between nodes
- **Error diagnostics**: Detailed error messages, stack traces, failed node information
- **Trigger details**: What initiated the execution and initial input data
- **Performance metrics**: Execution timing, memory usage, processing duration
- **Status progression**: Step-by-step execution flow and decision points

**Node-Level Details:**
Each node's execution includes:
- Input data received from previous nodes
- Parameters and configuration used during execution
- Output data produced by the node
- Execution time and status
- Error messages (if the node failed)
- Retry attempts and results

Use this tool to:
- **Debug failed executions**: Understand exactly where and why failures occurred
- **Analyze data transformations**: See how data changed through the workflow
- **Performance optimization**: Identify slow nodes and bottlenecks
- **Audit trails**: Create detailed records of workflow processing
- **Data validation**: Verify correct data handling and transformations
- **Error pattern analysis**: Study failure modes and error conditions

**Best for troubleshooting:** When you know a specific execution failed or behaved unexpectedly, this tool provides the forensic details needed to understand and fix the issue.`),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription(`Unique identifier of the specific execution to retrieve detailed information about.

This should be the execution ID from an execution record, typically obtained from the list_executions tool. Execution IDs are unique across your entire n8n instance and permanently identify a specific workflow run.

**Execution ID Characteristics:**
- Usually numeric strings (e.g., "12345", "98765")
- Unique across all workflows and time periods
- Permanent - IDs don't change or get reused
- Case-sensitive exact matching required

**How to find execution IDs:**
1. Use the list_executions tool to see recent executions
2. Look for executions with "failed" or "error" status for debugging
3. Check the n8n UI execution history for the ID
4. Use workflow-specific filtering to find relevant executions

**Common use cases by execution status:**
- **Failed executions**: Debug what went wrong and where
- **Successful executions**: Validate data processing and transformations  
- **Long-running executions**: Analyze performance and identify bottlenecks
- **Recent executions**: Verify current workflow behavior

**Example values:** "12345", "987654321", "1001"`),
				embeddable.PropertyRequired(),
			),
		),
	}
}

// Helper function to create N8N client from MCP context
func createN8NClientFromContext(ctx context.Context) (*n8n.N8NClient, error) {
	// Configure logging to stderr when in MCP mode (only once)
	configureMCPLogging()

	// For MCP tools, we'll use environment variables for configuration
	// This could be enhanced to use session-based configuration
	settings := &n8n.N8NAPISettings{
		BaseURL: getEnvOrDefault("N8N_BASE_URL", "http://localhost:5678"),
		APIKey:  getEnvOrDefault("N8N_API_KEY", ""),
	}

	if settings.APIKey == "" {
		return nil, fmt.Errorf("N8N_API_KEY environment variable is required")
	}

	return n8n.NewN8NClient(settings.BaseURL, settings.APIKey), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := getEnv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnv(key string) string {
	return os.Getenv(key)
}

// MCP Tool Handlers

func listWorkflowsHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	workflows, _, err := client.ListWorkflows(false, 50, "")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to list workflows: %v", err))), nil
	}

	// Extract only workflow summaries (id, name, active status) instead of full workflow data
	workflowSummaries := make([]map[string]interface{}, 0, len(workflows))
	for _, workflow := range workflows {
		summary := map[string]interface{}{
			"id":     workflow["id"],
			"name":   workflow["name"],
			"active": workflow["active"],
		}
		// Include createdAt and updatedAt if available
		if createdAt, ok := workflow["createdAt"]; ok {
			summary["createdAt"] = createdAt
		}
		if updatedAt, ok := workflow["updatedAt"]; ok {
			summary["updatedAt"] = updatedAt
		}
		// Include tags if available
		if tags, ok := workflow["tags"]; ok {
			summary["tags"] = tags
		}
		workflowSummaries = append(workflowSummaries, summary)
	}

	data, err := json.MarshalIndent(workflowSummaries, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal workflow summaries: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func getWorkflowHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	id, err := args.RequireString("id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	asMermaid := args.GetBool("as_mermaid", false)

	workflow, err := client.GetWorkflow(id)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to get workflow: %v", err))), nil
	}

	if asMermaid {
		mermaidResult := n8n.WorkflowToMermaid(workflow)
		return protocol.NewToolResult(
			protocol.WithText(mermaidResult.MermaidCode),
		), nil
	}

	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal workflow: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func createWorkflowHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	name, err := args.RequireString("name")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	active := args.GetBool("active", false)
	file := args.GetString("file", "")

	var workflow map[string]interface{}

	if file != "" {
		err := n8n.ReadJSONFile(file, &workflow)
		if err != nil {
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to read workflow file: %v", err))), nil
		}
	} else {
		// Create minimal workflow
		workflow = map[string]interface{}{
			"name":        name,
			"active":      active,
			"nodes":       []interface{}{},
			"connections": map[string]interface{}{},
		}
	}

	createdWorkflow, err := client.CreateWorkflow(workflow)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to create workflow: %v", err))), nil
	}

	data, err := json.MarshalIndent(createdWorkflow, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal created workflow: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func getNodesHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	nodes, err := client.GetNodes()
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to get nodes: %v", err))), nil
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal nodes: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func addNodeHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	workflowID, err := args.RequireString("workflow_id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	nodeType, err := args.RequireString("type")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	nodeName, err := args.RequireString("name")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	parametersStr := args.GetString("parameters", "{}")
	positionX := args.GetInt("position_x", 0)
	positionY := args.GetInt("position_y", 0)

	var parameters map[string]interface{}
	if err := json.Unmarshal([]byte(parametersStr), &parameters); err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Invalid parameters JSON: %v", err))), nil
	}

	// Get current workflow
	workflow, err := client.GetWorkflow(workflowID)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to get workflow: %v", err))), nil
	}

	// Create new node
	newNode := map[string]interface{}{
		"name":        nodeName,
		"type":        nodeType,
		"typeVersion": 1,
		"position":    []int{positionX, positionY},
		"parameters":  parameters,
	}

	// Add node to workflow
	workflow["nodes"] = append(workflow["nodes"].([]interface{}), newNode)

	// Update workflow
	updatedWorkflow, err := client.UpdateWorkflow(workflowID, workflow)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to update workflow: %v", err))), nil
	}

	data, err := json.MarshalIndent(updatedWorkflow, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal result: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func connectNodesHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	workflowID, err := args.RequireString("workflow_id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	sourceNode, err := args.RequireString("source_node")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	targetNode, err := args.RequireString("target_node")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	sourceOutput := args.GetString("source_output", "main")
	targetInput := args.GetString("target_input", "main")

	// Get current workflow
	workflow, err := client.GetWorkflow(workflowID)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to get workflow: %v", err))), nil
	}

	// Get existing connections
	connections, ok := workflow["connections"].(map[string]interface{})
	if !ok {
		connections = map[string]interface{}{}
	}

	// Add connection
	if connections[sourceNode] == nil {
		connections[sourceNode] = map[string]interface{}{}
	}
	sourceConns := connections[sourceNode].(map[string]interface{})
	if sourceConns[sourceOutput] == nil {
		sourceConns[sourceOutput] = []interface{}{}
	}

	// Add the connection
	sourceConns[sourceOutput] = append(sourceConns[sourceOutput].([]interface{}), map[string]interface{}{
		"node":  targetNode,
		"type":  targetInput,
		"index": 0,
	})

	workflow["connections"] = connections

	// Update workflow
	updatedWorkflow, err := client.UpdateWorkflow(workflowID, workflow)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to update workflow: %v", err))), nil
	}

	data, err := json.MarshalIndent(updatedWorkflow, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal result: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func listExecutionsHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	workflowID := args.GetString("workflow_id", "")
	limit := args.GetInt("limit", 20)

	params := map[string]string{}
	if workflowID != "" {
		params["workflowId"] = workflowID
	}
	params["limit"] = strconv.Itoa(limit)

	executions, err := client.ListExecutions(params)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to list executions: %v", err))), nil
	}

	data, err := json.MarshalIndent(executions, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal executions: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}

func getExecutionHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	client, err := createN8NClientFromContext(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	id, err := args.RequireString("id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	execution, err := client.GetExecution(id)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to get execution: %v", err))), nil
	}

	data, err := json.MarshalIndent(execution, "", "  ")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Failed to marshal execution: %v", err))), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(string(data)),
	), nil
}
