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

var agentTypesCmd = &cobra.Command{
	Use:   "agent-types",
	Short: "Agent type management commands",
	Long:  "Commands for creating, listing, and managing agent types",
}

// List agent types command
var listAgentTypesCmd = &cobra.Command{
	Use:   "list",
	Short: "List agent types",
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		var projectID *string
		if cmd.Flags().Changed("project") {
			proj, _ := cmd.Flags().GetString("project")
			projectID = &proj
		}

		agentTypes, err := tm.ListAgentTypes(projectID)
		if err != nil {
			return fmt.Errorf("failed to list agent types: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		return outputAgentTypes(agentTypes, output)
	},
}

// Create agent type command
var createAgentTypeCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new agent type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Get project ID (default or specified)
		var projectID string
		if cmd.Flags().Changed("project") {
			projectID, _ = cmd.Flags().GetString("project")
		} else {
			project, err := tm.GetDefaultProject()
			if err != nil {
				return fmt.Errorf("failed to get default project: %w", err)
			}
			projectID = project.ID
		}

		agentType, err := tm.CreateAgentType(name, description, projectID)
		if err != nil {
			return fmt.Errorf("failed to create agent type: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(agentType)
		}

		fmt.Printf("Created agent type: %s\n", agentType.ID)
		fmt.Printf("Name: %s\n", agentType.Name)
		if agentType.Description != "" {
			fmt.Printf("Description: %s\n", agentType.Description)
		}
		fmt.Printf("Project: %s\n", agentType.ProjectID)

		return nil
	},
}

// Assign task to agent type command
var assignToTypeCmd = &cobra.Command{
	Use:   "assign <task-id> <agent-type-id>",
	Short: "Assign a task to an agent type (finds available agent)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		agentTypeID := args[1]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		err = tm.AssignTaskToAgentType(taskID, agentTypeID)
		if err != nil {
			return fmt.Errorf("failed to assign task to agent type: %w", err)
		}

		fmt.Printf("Assigned task %s to agent type %s\n", taskID, agentTypeID)
		return nil
	},
}

func outputAgentTypes(agentTypes []AgentType, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(agentTypes)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(agentTypes)
	case "csv":
		return outputAgentTypesCSV(agentTypes)
	default: // table
		return outputAgentTypesTable(agentTypes)
	}
}

func outputAgentTypesTable(agentTypes []AgentType) error {
	if len(agentTypes) == 0 {
		fmt.Println("No agent types found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tPROJECT\tCREATED")

	for _, agentType := range agentTypes {
		id := truncateString(agentType.ID, 8)
		name := truncateString(agentType.Name, 20)
		description := truncateString(agentType.Description, 30)
		project := truncateString(agentType.ProjectID, 8)
		created := agentType.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, description, project, created)
	}

	return w.Flush()
}

func outputAgentTypesCSV(agentTypes []AgentType) error {
	fmt.Println("id,name,description,project_id,created_at,updated_at")
	for _, agentType := range agentTypes {
		fmt.Printf("%s,%q,%q,%s,%s,%s\n",
			agentType.ID,
			agentType.Name,
			agentType.Description,
			agentType.ProjectID,
			agentType.CreatedAt.Format(time.RFC3339),
			agentType.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(agentTypesCmd)

	// Add subcommands
	agentTypesCmd.AddCommand(listAgentTypesCmd)
	agentTypesCmd.AddCommand(createAgentTypeCmd)
	agentTypesCmd.AddCommand(assignToTypeCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listAgentTypesCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// List specific flags
	listAgentTypesCmd.Flags().String("project", "", "Filter by project ID")

	// Create specific flags
	createAgentTypeCmd.Flags().StringP("description", "d", "", "Agent type description")
	createAgentTypeCmd.Flags().StringP("project", "p", "", "Project ID (uses default if not specified)")
	createAgentTypeCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
