package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Task dependency management commands",
	Long:  "Commands for adding, listing, and managing task dependencies",
}

// Add dependency command
var addDepCmd = &cobra.Command{
	Use:   "add <task-id> <depends-on-id>",
	Short: "Add a dependency between tasks",
	Args:  cobra.ExactArgs(2),
	Long: `Add a dependency relationship between two tasks.

The first task will depend on the second task being completed.
Before adding the dependency, this command will show detailed information
about the parent task (the task being depended on) including:
- Task title, description, and current status
- All notes from agents who have worked on the task
- Agent information and timestamps

This helps you make informed decisions about task dependencies.

Examples:
  amp-tasks deps add <task-id> <depends-on-id>  # task-id depends on depends-on-id`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		dependsOnID := args[1]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Show parent task information before adding dependency
		err = displayParentTaskInfo(tm, dependsOnID)
		if err != nil {
			return fmt.Errorf("failed to display parent task info: %w", err)
		}

		// Add the dependency
		err = tm.AddDependency(taskID, dependsOnID)
		if err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}

		fmt.Printf("\n✓ Added dependency: %s depends on %s\n", taskID, dependsOnID)
		return nil
	},
}

// List dependencies command
var listDepsCmd = &cobra.Command{
	Use:   "list <task-id>",
	Short: "List dependencies for a specific task",
	Args:  cobra.ExactArgs(1),
	Long: `List all tasks that the specified task depends on.

Examples:
  amp-tasks deps list <task-id>
  amp-tasks deps list <task-id> --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		deps, err := tm.GetTaskDependencies(taskID)
		if err != nil {
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputDependencies(deps, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Graph command - show task graph
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Show task dependency graph",
	Long: `Display the task dependency graph in various formats.

Examples:
  amp-tasks deps graph              # ASCII art representation
  amp-tasks deps graph --format dot # Graphviz DOT format
  amp-tasks deps graph --output json # JSON representation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Get all tasks and dependencies
		tasks, err := tm.ListTasks(nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		// Build dependency map
		depMap := make(map[string][]string)
		for _, task := range tasks {
			deps, err := tm.GetTaskDependencies(task.ID)
			if err != nil {
				return fmt.Errorf("failed to get dependencies for task %s: %w", task.ID, err)
			}
			for _, dep := range deps {
				depMap[task.ID] = append(depMap[task.ID], dep.DependsOnID)
			}
		}

		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputGraph(tasks, depMap, format, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

func outputDependencies(deps []TaskDependency, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(deps)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(deps)
	default: // table
		return outputDependenciesTable(deps)
	}
}

func outputDependenciesTable(deps []TaskDependency) error {
	if len(deps) == 0 {
		fmt.Println("No dependencies found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TASK\tDEPENDS ON\tCREATED")

	for _, dep := range deps {
		taskID := truncateString(dep.TaskID, 8)
		dependsOnID := truncateString(dep.DependsOnID, 8)
		created := dep.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\n", taskID, dependsOnID, created)
	}

	return w.Flush()
}

func outputGraph(tasks []Task, depMap map[string][]string, format, output string) error {
	switch output {
	case "json":
		graph := map[string]interface{}{
			"tasks":        tasks,
			"dependencies": depMap,
		}
		return json.NewEncoder(os.Stdout).Encode(graph)
	case "yaml":
		graph := map[string]interface{}{
			"tasks":        tasks,
			"dependencies": depMap,
		}
		return yaml.NewEncoder(os.Stdout).Encode(graph)
	default:
		switch format {
		case "dot":
			return outputGraphDOT(tasks, depMap)
		default:
			return outputGraphASCII(tasks, depMap)
		}
	}
}

func outputGraphDOT(tasks []Task, depMap map[string][]string) error {
	fmt.Println("digraph TaskGraph {")
	fmt.Println("  rankdir=TB;")
	fmt.Println("  node [shape=box];")
	fmt.Println()

	// Output nodes
	for _, task := range tasks {
		label := truncateString(task.Title, 20)
		color := "lightblue"
		switch task.Status {
		case TaskStatusCompleted:
			color = "lightgreen"
		case TaskStatusInProgress:
			color = "yellow"
		case TaskStatusFailed:
			color = "lightcoral"
		}

		fmt.Printf("  \"%s\" [label=\"%s\\n%s\", fillcolor=%s, style=filled];\n",
			task.ID, label, task.Status, color)
	}

	fmt.Println()

	// Output edges
	for taskID, deps := range depMap {
		for _, depID := range deps {
			fmt.Printf("  \"%s\" -> \"%s\";\n", depID, taskID)
		}
	}

	fmt.Println("}")
	return nil
}

func outputGraphASCII(tasks []Task, depMap map[string][]string) error {
	fmt.Println("Task Dependency Graph:")
	fmt.Println("======================")
	fmt.Println()

	// Create task ID to title map
	taskMap := make(map[string]Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Find root tasks (no dependencies)
	rootTasks := make([]Task, 0)
	for _, task := range tasks {
		if len(depMap[task.ID]) == 0 {
			rootTasks = append(rootTasks, task)
		}
	}

	if len(rootTasks) == 0 {
		fmt.Println("No root tasks found.")
		return nil
	}

	// Print hierarchy starting from roots
	visited := make(map[string]bool)
	for _, root := range rootTasks {
		printTaskHierarchy(root, taskMap, depMap, visited, "")
	}

	return nil
}

func printTaskHierarchy(task Task, taskMap map[string]Task, depMap map[string][]string, visited map[string]bool, prefix string) {
	if visited[task.ID] {
		return
	}
	visited[task.ID] = true

	statusSymbol := "○"
	switch task.Status {
	case TaskStatusCompleted:
		statusSymbol = "●"
	case TaskStatusInProgress:
		statusSymbol = "◐"
	case TaskStatusFailed:
		statusSymbol = "✗"
	}

	title := truncateString(task.Title, 50)
	fmt.Printf("%s%s %s [%s]\n", prefix, statusSymbol, title, truncateString(task.ID, 8))

	// Find tasks that depend on this one
	dependents := make([]Task, 0)
	for _, otherTask := range taskMap {
		for _, depID := range depMap[otherTask.ID] {
			if depID == task.ID {
				dependents = append(dependents, otherTask)
				break
			}
		}
	}

	// Print dependents with increased indentation
	for i, dependent := range dependents {
		newPrefix := prefix
		if i == len(dependents)-1 {
			newPrefix += "  └─ "
		} else {
			newPrefix += "  ├─ "
		}
		printTaskHierarchy(dependent, taskMap, depMap, visited, newPrefix)
	}
}

// displayParentTaskInfo retrieves and displays detailed information about a parent task
// including all notes from agents who have worked on it
func displayParentTaskInfo(tm *TaskManager, parentTaskID string) error {
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("📋 PARENT TASK INFORMATION\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")

	// Get parent task details
	parentTask, err := tm.GetTask(parentTaskID)
	if err != nil {
		return fmt.Errorf("parent task not found: %w", err)
	}

	// Display task information
	statusSymbol := getTaskStatusSymbol(parentTask.Status)
	fmt.Printf("\n🔹 Task ID: %s\n", parentTask.ID)
	fmt.Printf("🔹 Title: %s\n", parentTask.Title)
	if parentTask.Description != "" {
		fmt.Printf("🔹 Description: %s\n", parentTask.Description)
	}
	fmt.Printf("🔹 Status: %s %s\n", statusSymbol, parentTask.Status)
	if parentTask.AgentID != nil {
		fmt.Printf("🔹 Assigned Agent: %s\n", *parentTask.AgentID)
	}
	fmt.Printf("🔹 Created: %s\n", parentTask.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("🔹 Updated: %s\n", parentTask.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Get and display notes
	notes, err := tm.ListNotes(&parentTaskID, nil)
	if err != nil {
		return fmt.Errorf("failed to get notes for parent task: %w", err)
	}

	fmt.Printf("\n📝 AGENT NOTES (%d total)\n", len(notes))
	fmt.Printf("───────────────────────────────────────────────────────────────────────────────\n")

	if len(notes) == 0 {
		fmt.Printf("💭 No notes found for this task.\n")
	} else {
		// Group notes by agent
		agentNotes := make(map[string][]Note)
		for _, note := range notes {
			agentNotes[note.AgentID] = append(agentNotes[note.AgentID], note)
		}

		// Display notes grouped by agent
		for agentID, agentNoteList := range agentNotes {
			fmt.Printf("\n👤 Agent: %s (%d notes)\n", agentID, len(agentNoteList))

			for i, note := range agentNoteList {
				timeStr := note.CreatedAt.Format("2006-01-02 15:04")
				fmt.Printf("   %d. [%s] %s\n", i+1, timeStr, note.Content)
				if i < len(agentNoteList)-1 {
					fmt.Printf("      ┆\n")
				}
			}
		}
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════════════════════════\n")
	return nil
}

// getTaskStatusSymbol returns a visual symbol for the task status
func getTaskStatusSymbol(status TaskStatus) string {
	switch status {
	case TaskStatusCompleted:
		return "✅"
	case TaskStatusInProgress:
		return "🔄"
	case TaskStatusFailed:
		return "❌"
	case TaskStatusPending:
		return "⏳"
	default:
		return "❓"
	}
}

func init() {
	rootCmd.AddCommand(depsCmd)

	// Add subcommands
	depsCmd.AddCommand(addDepCmd)
	depsCmd.AddCommand(listDepsCmd)
	depsCmd.AddCommand(graphCmd)

	// Flags
	listDepsCmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml)")
	graphCmd.Flags().String("format", "ascii", "Graph format (ascii, dot)")
	graphCmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml)")
}
