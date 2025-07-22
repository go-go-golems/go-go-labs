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

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Task management commands",
	Long:  "Commands for creating, listing, and managing tasks in the coordination system",
}

// List tasks command
var listTasksCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks with optional filtering",
	Long: `List tasks with optional filtering by parent, status, or agent.

Examples:
  amp-tasks tasks list                    # List all tasks
  amp-tasks tasks list --status pending   # List pending tasks
  amp-tasks tasks list --parent ""        # List root tasks only
  amp-tasks tasks list --agent <agent-id> # List tasks for specific agent
  amp-tasks tasks list --output json      # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		var parentID *string
		var status *TaskStatus
		var agentID *string

		if cmd.Flags().Changed("parent") {
			parent, _ := cmd.Flags().GetString("parent")
			parentID = &parent
		}

		if cmd.Flags().Changed("status") {
			statusStr, _ := cmd.Flags().GetString("status")
			taskStatus := TaskStatus(statusStr)
			status = &taskStatus
		}

		if cmd.Flags().Changed("agent") {
			agent, _ := cmd.Flags().GetString("agent")
			agentID = &agent
		}

		var preferredAgentTypeID *string
		if cmd.Flags().Changed("agent-type") {
			agentType, _ := cmd.Flags().GetString("agent-type")
			preferredAgentTypeID = &agentType
		}

		tasks, err := tm.ListTasksWithAgentInfo(parentID, status, agentID, nil, preferredAgentTypeID)
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputTasksWithAgentInfo(tasks, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Show specific task command
var showTaskCmd = &cobra.Command{
	Use:   "show <task-id>",
	Short: "Show detailed information about a specific task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		task, err := tm.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		deps, err := tm.GetTaskDependencies(taskID)
		if err != nil {
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputTaskDetail(task, deps, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Create task command
var createTaskCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	Long: `Create a new task with the specified title.

Examples:
  amp-tasks tasks create "Implement user authentication"
  amp-tasks tasks create "Fix login bug" --parent <parent-task-id> --description "Detailed description"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		var parentID *string
		if cmd.Flags().Changed("parent") {
			parent, _ := cmd.Flags().GetString("parent")
			parentID = &parent
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

		var preferredAgentTypeID *string
		if cmd.Flags().Changed("agent-type") {
			agentType, _ := cmd.Flags().GetString("agent-type")
			preferredAgentTypeID = &agentType
		}

		task, err := tm.CreateTask(title, description, parentID, project.ID, preferredAgentTypeID)
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(task)
		}

		fmt.Printf("Created task: %s\n", task.ID)
		fmt.Printf("Title: %s\n", task.Title)
		if task.Description != "" {
			fmt.Printf("Description: %s\n", task.Description)
		}
		if task.ParentID != nil {
			fmt.Printf("Parent: %s\n", *task.ParentID)
		}

		return nil
	},
}

// Available tasks command
var availableTasksCmd = &cobra.Command{
	Use:   "available",
	Short: "List tasks available for assignment (dependencies satisfied)",
	Long: `List tasks that are ready for assignment to agents.
	
These are tasks with status 'pending' that have all their dependencies completed.
Shows assignment status and agent type information where available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		var preferredAgentTypeID *string
		if cmd.Flags().Changed("agent-type") {
			agentType, _ := cmd.Flags().GetString("agent-type")
			preferredAgentTypeID = &agentType
		}

		tasks, err := tm.GetAvailableTasksWithAgentInfo(preferredAgentTypeID)
		if err != nil {
			return fmt.Errorf("failed to get available tasks: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputTasksWithAgentInfo(tasks, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Assign task command
var assignTaskCmd = &cobra.Command{
	Use:   "assign <task-id> <agent-id>",
	Short: "Assign a task to an agent",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		agentID := args[1]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		err = tm.AssignTask(taskID, agentID)
		if err != nil {
			return fmt.Errorf("failed to assign task: %w", err)
		}

		fmt.Printf("Assigned task %s to agent %s\n", taskID, agentID)
		return nil
	},
}

// Update task status command
var updateTaskStatusCmd = &cobra.Command{
	Use:   "status <task-id> <status>",
	Short: "Update task status",
	Args:  cobra.ExactArgs(2),
	Long: `Update the status of a task.

Valid statuses: pending, in_progress, completed, failed

Examples:
  amp-tasks tasks status <task-id> completed
  amp-tasks tasks status <task-id> failed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		statusStr := args[1]

		validStatuses := map[string]TaskStatus{
			"pending":     TaskStatusPending,
			"in_progress": TaskStatusInProgress,
			"completed":   TaskStatusCompleted,
			"failed":      TaskStatusFailed,
		}

		status, valid := validStatuses[statusStr]
		if !valid {
			return fmt.Errorf("invalid status %q. Valid options: %v", statusStr, getValidStatusStrings())
		}

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		err = tm.UpdateTaskStatus(taskID, status)
		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		fmt.Printf("Updated task %s status to %s\n", taskID, statusStr)

		// If task completed, show available tasks in dual mode
		if status == TaskStatusCompleted {
			availableTasks, err := tm.GetAvailableTasks(nil)
			if err == nil && len(availableTasks) > 0 {
				fmt.Printf("\nAvailable tasks after completion:\n")
				for _, task := range availableTasks {
					fmt.Printf("  - %s: %s\n", task.ID[:8], task.Title)
				}
			}
		}

		return nil
	},
}

func outputTasks(tasks []Task, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(tasks)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(tasks)
	case "csv":
		return outputTasksCSV(tasks)
	default: // table
		return outputTasksTable(tasks)
	}
}

func outputTasksTable(tasks []Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tAGENT\tPARENT\tCREATED")

	for _, task := range tasks {
		id := truncateString(task.ID, 8)
		title := truncateString(task.Title, 30)
		agent := "none"
		if task.AgentID != nil {
			agent = truncateString(*task.AgentID, 8)
		}
		parent := "none"
		if task.ParentID != nil {
			parent = truncateString(*task.ParentID, 8)
		}
		created := task.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			id, title, task.Status, agent, parent, created)
	}

	return w.Flush()
}

func outputTasksCSV(tasks []Task) error {
	fmt.Println("id,title,description,status,agent_id,parent_id,created_at,updated_at")
	for _, task := range tasks {
		agentID := ""
		if task.AgentID != nil {
			agentID = *task.AgentID
		}
		parentID := ""
		if task.ParentID != nil {
			parentID = *task.ParentID
		}

		fmt.Printf("%s,%q,%q,%s,%s,%s,%s,%s\n",
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			agentID,
			parentID,
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func outputTaskDetail(task *Task, deps []TaskDependency, format string) error {
	switch format {
	case "json":
		detail := map[string]interface{}{
			"task":         task,
			"dependencies": deps,
		}
		return json.NewEncoder(os.Stdout).Encode(detail)
	case "yaml":
		detail := map[string]interface{}{
			"task":         task,
			"dependencies": deps,
		}
		return yaml.NewEncoder(os.Stdout).Encode(detail)
	default: // detailed text
		fmt.Printf("Task: %s\n", task.ID)
		fmt.Printf("Title: %s\n", task.Title)
		if task.Description != "" {
			fmt.Printf("Description: %s\n", task.Description)
		}
		fmt.Printf("Status: %s\n", task.Status)

		if task.AgentID != nil {
			fmt.Printf("Agent: %s\n", *task.AgentID)
		} else {
			fmt.Printf("Agent: none\n")
		}

		if task.ParentID != nil {
			fmt.Printf("Parent: %s\n", *task.ParentID)
		} else {
			fmt.Printf("Parent: none (root task)\n")
		}

		fmt.Printf("Created: %s\n", task.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated: %s\n", task.UpdatedAt.Format(time.RFC3339))

		if len(deps) > 0 {
			fmt.Printf("\nDependencies:\n")
			for _, dep := range deps {
				fmt.Printf("  - %s\n", dep.DependsOnID)
			}
		} else {
			fmt.Printf("\nDependencies: none\n")
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getValidStatusStrings() []string {
	return []string{"pending", "in_progress", "completed", "failed"}
}

func init() {
	rootCmd.AddCommand(tasksCmd)

	// Add subcommands
	tasksCmd.AddCommand(listTasksCmd)
	tasksCmd.AddCommand(showTaskCmd)
	tasksCmd.AddCommand(createTaskCmd)
	tasksCmd.AddCommand(availableTasksCmd)
	tasksCmd.AddCommand(assignTaskCmd)
	tasksCmd.AddCommand(updateTaskStatusCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listTasksCmd, showTaskCmd, availableTasksCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// List specific flags
	listTasksCmd.Flags().String("parent", "", "Filter by parent task ID (empty string for root tasks)")
	listTasksCmd.Flags().String("status", "", "Filter by status (pending, in_progress, completed, failed)")
	listTasksCmd.Flags().String("agent", "", "Filter by agent ID")
	listTasksCmd.Flags().String("agent-type", "", "Filter by preferred agent type ID")

	// Available tasks specific flags
	availableTasksCmd.Flags().String("agent-type", "", "Filter by preferred agent type ID")

	// Create specific flags
	createTaskCmd.Flags().StringP("description", "d", "", "Task description")
	createTaskCmd.Flags().StringP("parent", "p", "", "Parent task ID")
	createTaskCmd.Flags().String("agent-type", "", "Preferred agent type ID for this task")
	createTaskCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}

func outputTasksWithAgentInfo(tasks []TaskWithAgentInfo, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(tasks)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(tasks)
	case "csv":
		return outputTasksWithAgentInfoCSV(tasks)
	default: // table
		return outputTasksWithAgentInfoTable(tasks)
	}
}

func outputTasksWithAgentInfoTable(tasks []TaskWithAgentInfo) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tASSIGNED\tAGENT_TYPE\tPREFERRED_TYPE\tPARENT\tCREATED")

	for _, task := range tasks {
		id := truncateString(task.ID, 8)
		title := truncateString(task.Title, 30)

		assigned := "none"
		agentType := "none"
		if task.AgentID != nil {
			assigned = "yes"
			if task.AgentTypeName != nil {
				agentType = truncateString(*task.AgentTypeName, 12)
			} else {
				agentType = "untyped"
			}
		}

		preferredType := "none"
		if task.PreferredAgentTypeName != nil {
			preferredType = truncateString(*task.PreferredAgentTypeName, 12)
		}

		parent := "none"
		if task.ParentID != nil {
			parent = truncateString(*task.ParentID, 8)
		}
		created := task.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			id, title, task.Status, assigned, agentType, preferredType, parent, created)
	}

	return w.Flush()
}

func outputTasksWithAgentInfoCSV(tasks []TaskWithAgentInfo) error {
	fmt.Println("id,title,description,status,assigned,agent_id,agent_name,agent_type_name,preferred_agent_type_id,preferred_agent_type_name,parent_id,created_at,updated_at")
	for _, task := range tasks {
		agentID := ""
		agentName := ""
		agentTypeName := ""
		assigned := "no"

		if task.AgentID != nil {
			agentID = *task.AgentID
			assigned = "yes"
		}
		if task.AgentName != nil {
			agentName = *task.AgentName
		}
		if task.AgentTypeName != nil {
			agentTypeName = *task.AgentTypeName
		}

		preferredAgentTypeID := ""
		if task.PreferredAgentTypeID != nil {
			preferredAgentTypeID = *task.PreferredAgentTypeID
		}

		preferredAgentTypeName := ""
		if task.PreferredAgentTypeName != nil {
			preferredAgentTypeName = *task.PreferredAgentTypeName
		}

		parentID := ""
		if task.ParentID != nil {
			parentID = *task.ParentID
		}

		fmt.Printf("%s,%q,%q,%s,%s,%s,%q,%q,%s,%q,%s,%s,%s\n",
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			assigned,
			agentID,
			agentName,
			agentTypeName,
			preferredAgentTypeID,
			preferredAgentTypeName,
			parentID,
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}
