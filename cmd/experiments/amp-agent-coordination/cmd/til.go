package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var tilCmd = &cobra.Command{
	Use:   "til",
	Short: "Today I Learned (TIL) management commands",
	Long:  "Commands for creating, listing, and managing TIL entries in the coordination system",
}

// Create TIL command
var createTILCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new TIL entry",
	Args:  cobra.ExactArgs(1),
	Long: `Create a new TIL (Today I Learned) entry with the specified title.

Examples:
  amp-tasks til create "How to use goroutines" --content "Learned about sync.WaitGroup"
  amp-tasks til create "Docker best practices" --content "Always use multi-stage builds" --task <task-id>
  amp-tasks til create "Testing patterns" --content "Table-driven tests are powerful"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		content, _ := cmd.Flags().GetString("content")

		if content == "" {
			return fmt.Errorf("content is required (use --content flag)")
		}

		var taskID *string
		if cmd.Flags().Changed("task") {
			task, _ := cmd.Flags().GetString("task")
			taskID = &task
		}

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Get default project
		project, err := tm.GetDefaultProject()
		if err != nil {
			return fmt.Errorf("failed to get default project: %w", err)
		}

		// Use a default agent ID for now (could be made configurable)
		agentID := "cli-user"

		til, err := tm.CreateTIL(project.ID, taskID, agentID, title, content)
		if err != nil {
			return fmt.Errorf("failed to create TIL: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(til)
		}

		fmt.Printf("Created TIL: %s\n", til.ID)
		fmt.Printf("Title: %s\n", til.Title)
		fmt.Printf("Content: %s\n", til.Content)
		if til.TaskID != nil {
			fmt.Printf("Task: %s\n", *til.TaskID)
		}
		fmt.Printf("Project: %s\n", til.ProjectID)

		return nil
	},
}

// List TILs command
var listTILsCmd = &cobra.Command{
	Use:   "list",
	Short: "List TIL entries with optional filtering",
	Long: `List TIL entries with optional filtering by project, task, or agent.

Examples:
  amp-tasks til list                        # List all TILs for default project
  amp-tasks til list --project <project-id> # List TILs for specific project
  amp-tasks til list --task <task-id>       # List TILs for specific task
  amp-tasks til list --agent <agent-id>     # List TILs for specific agent
  amp-tasks til list --output json          # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		var projectID *string
		var taskID *string
		var agentID *string

		if cmd.Flags().Changed("project") {
			project, _ := cmd.Flags().GetString("project")
			projectID = &project
		} else {
			// Use default project if not specified
			project, err := tm.GetDefaultProject()
			if err != nil {
				return fmt.Errorf("failed to get default project: %w", err)
			}
			projectID = &project.ID
		}

		if cmd.Flags().Changed("task") {
			task, _ := cmd.Flags().GetString("task")
			taskID = &task
		}

		if cmd.Flags().Changed("agent") {
			agent, _ := cmd.Flags().GetString("agent")
			agentID = &agent
		}

		tils, err := tm.ListTILs(projectID, taskID, agentID)
		if err != nil {
			return fmt.Errorf("failed to list TILs: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in table mode
		if output == "table" && projectID != nil {
			project, err := tm.GetProject(*projectID)
			if err == nil {
				fmt.Printf("Project: %s\n", project.Name)
				if project.Guidelines != "" {
					fmt.Printf("Guidelines: %s\n\n", project.Guidelines)
				} else {
					fmt.Println()
				}
			}
		}

		return outputTILs(tils, output)
	},
}

// Show specific TIL command
var showTILCmd = &cobra.Command{
	Use:   "show <til-id>",
	Short: "Show detailed information about a specific TIL entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tilID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Get TIL by searching through all TILs (since there's no GetTIL method)
		tils, err := tm.ListTILs(nil, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to list TILs: %w", err)
		}

		var foundTIL *TIL
		for _, til := range tils {
			if til.ID == tilID {
				foundTIL = &til
				break
			}
		}

		if foundTIL == nil {
			return fmt.Errorf("TIL with ID %s not found", tilID)
		}

		output, _ := cmd.Flags().GetString("output")
		return outputTILDetail(foundTIL, output)
	},
}

func outputTILs(tils []TIL, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(tils)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(tils)
	case "csv":
		return outputTILsCSV(tils)
	default: // table
		return outputTILsTable(tils)
	}
}

func outputTILsTable(tils []TIL) error {
	if len(tils) == 0 {
		fmt.Println("No TILs found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tTASK\tAGENT\tCREATED")

	for _, til := range tils {
		id := truncateString(til.ID, 8)
		title := truncateString(til.Title, 40)
		task := "none"
		if til.TaskID != nil {
			task = truncateString(*til.TaskID, 8)
		}
		agent := truncateString(til.AgentID, 8)
		created := til.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			id, title, task, agent, created)
	}

	return w.Flush()
}

func outputTILsCSV(tils []TIL) error {
	fmt.Println("id,project_id,task_id,agent_id,title,content,created_at,updated_at")
	for _, til := range tils {
		taskID := ""
		if til.TaskID != nil {
			taskID = *til.TaskID
		}

		fmt.Printf("%s,%s,%s,%s,%q,%q,%s,%s\n",
			til.ID,
			til.ProjectID,
			taskID,
			til.AgentID,
			til.Title,
			til.Content,
			til.CreatedAt.Format(time.RFC3339),
			til.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func outputTILDetail(til *TIL, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(til)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(til)
	default: // detailed text
		fmt.Printf("TIL: %s\n", til.ID)
		fmt.Printf("Title: %s\n", til.Title)
		fmt.Printf("Content: %s\n", til.Content)
		fmt.Printf("Project: %s\n", til.ProjectID)

		if til.TaskID != nil {
			fmt.Printf("Task: %s\n", *til.TaskID)
		} else {
			fmt.Printf("Task: none (project-level)\n")
		}

		fmt.Printf("Agent: %s\n", til.AgentID)
		fmt.Printf("Created: %s\n", til.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated: %s\n", til.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(tilCmd)

	// Add subcommands
	tilCmd.AddCommand(createTILCmd)
	tilCmd.AddCommand(listTILsCmd)
	tilCmd.AddCommand(showTILCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listTILsCmd, showTILCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// Create specific flags
	createTILCmd.Flags().StringP("content", "c", "", "TIL content (required)")
	createTILCmd.Flags().StringP("task", "t", "", "Task ID to associate with this TIL")
	createTILCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// List specific flags
	listTILsCmd.Flags().String("project", "", "Filter by project ID (uses default if not specified)")
	listTILsCmd.Flags().String("task", "", "Filter by task ID")
	listTILsCmd.Flags().String("agent", "", "Filter by agent ID")
}
