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

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Notes management commands",
	Long:  "Commands for adding, listing, and managing notes for tasks in the coordination system",
}

// Add note command
var addNoteCmd = &cobra.Command{
	Use:   "add <task-id> <content>",
	Short: "Add a note to a task",
	Args:  cobra.ExactArgs(2),
	Long: `Add a note to a specific task.

Examples:
  amp-tasks notes add <task-id> "Found an issue with error handling"
  amp-tasks notes add <task-id> "Implementation progress: 80% complete"
  amp-tasks notes add <task-id> "Need to refactor the authentication logic"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		content := args[1]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Verify task exists
		_, err = tm.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}

		// Get agent ID from flag or use default
		agentID, _ := cmd.Flags().GetString("agent")
		if agentID == "" {
			agentID = "cli-user"
		}

		note, err := tm.CreateNote(taskID, agentID, content)
		if err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(note)
		}

		fmt.Printf("Added note: %s\n", note.ID)
		fmt.Printf("Task: %s\n", note.TaskID)
		fmt.Printf("Content: %s\n", note.Content)
		fmt.Printf("Agent: %s\n", note.AgentID)

		return nil
	},
}

// List notes command
var listNotesCmd = &cobra.Command{
	Use:   "list <task-id>",
	Short: "List notes for a specific task",
	Args:  cobra.ExactArgs(1),
	Long: `List all notes for a specific task with optional filtering by agent.

Examples:
  amp-tasks notes list <task-id>                   # List all notes for task
  amp-tasks notes list <task-id> --agent <agent-id> # List notes from specific agent
  amp-tasks notes list <task-id> --output json      # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Verify task exists and get task info
		task, err := tm.GetTask(taskID)
		if err != nil {
			return fmt.Errorf("task not found: %w", err)
		}

		var agentID *string
		if cmd.Flags().Changed("agent") {
			agent, _ := cmd.Flags().GetString("agent")
			agentID = &agent
		}

		notes, err := tm.ListNotes(&taskID, agentID)
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show task context in table mode
		if output == "table" {
			fmt.Printf("Task: %s - %s\n", task.ID[:8], task.Title)
			fmt.Printf("Status: %s\n\n", task.Status)
		}

		return outputNotes(notes, output)
	},
}

// Show specific note command
var showNoteCmd = &cobra.Command{
	Use:   "show <note-id>",
	Short: "Show detailed information about a specific note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Get note by searching through all notes (since there's no GetNote method)
		notes, err := tm.ListNotes(nil, nil)
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		var foundNote *Note
		for _, note := range notes {
			if note.ID == noteID {
				foundNote = &note
				break
			}
		}

		if foundNote == nil {
			return fmt.Errorf("note with ID %s not found", noteID)
		}

		output, _ := cmd.Flags().GetString("output")
		return outputNoteDetail(foundNote, output)
	},
}

func outputNotes(notes []Note, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(notes)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(notes)
	case "csv":
		return outputNotesCSV(notes)
	default: // table
		return outputNotesTable(notes)
	}
}

func outputNotesTable(notes []Note) error {
	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCONTENT\tAGENT\tCREATED")

	for _, note := range notes {
		id := truncateString(note.ID, 8)
		content := truncateString(note.Content, 50)
		agent := truncateString(note.AgentID, 8)
		created := note.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			id, content, agent, created)
	}

	return w.Flush()
}

func outputNotesCSV(notes []Note) error {
	fmt.Println("id,task_id,agent_id,content,created_at,updated_at")
	for _, note := range notes {
		fmt.Printf("%s,%s,%s,%q,%s,%s\n",
			note.ID,
			note.TaskID,
			note.AgentID,
			note.Content,
			note.CreatedAt.Format(time.RFC3339),
			note.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func outputNoteDetail(note *Note, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(note)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(note)
	default: // detailed text
		fmt.Printf("Note: %s\n", note.ID)
		fmt.Printf("Task: %s\n", note.TaskID)
		fmt.Printf("Content: %s\n", note.Content)
		fmt.Printf("Agent: %s\n", note.AgentID)
		fmt.Printf("Created: %s\n", note.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated: %s\n", note.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(notesCmd)

	// Add subcommands
	notesCmd.AddCommand(addNoteCmd)
	notesCmd.AddCommand(listNotesCmd)
	notesCmd.AddCommand(showNoteCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listNotesCmd, showNoteCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// Add specific flags
	addNoteCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	addNoteCmd.Flags().String("agent", "", "Agent ID for the note (defaults to 'cli-user')")

	// List specific flags
	listNotesCmd.Flags().String("agent", "", "Filter by agent ID")
}
