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

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Project management commands",
	Long:  "Commands for creating, listing, and managing projects",
}

// List projects command
var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		projects, err := tm.ListProjects()
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		return outputProjects(projects, output)
	},
}

// Create project command
var createProjectCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		guidelines, _ := cmd.Flags().GetString("guidelines")

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		project, err := tm.CreateProject(name, description, guidelines, nil)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		setDefault, _ := cmd.Flags().GetBool("set-default")
		if setDefault {
			err = tm.SetGlobalKV("default_project", project.ID, nil)
			if err != nil {
				return fmt.Errorf("failed to set as default: %w", err)
			}
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(project)
		}

		fmt.Printf("Created project: %s\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		if project.Description != "" {
			fmt.Printf("Description: %s\n", project.Description)
		}
		if project.Guidelines != "" {
			fmt.Printf("Guidelines: %s\n", project.Guidelines)
		}
		if setDefault {
			fmt.Printf("Set as default project.\n")
		}

		return nil
	},
}

// Set default project command
var setDefaultProjectCmd = &cobra.Command{
	Use:   "set-default <project-id>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		// Verify project exists
		_, err = tm.GetProject(projectID)
		if err != nil {
			return fmt.Errorf("failed to verify project: %w", err)
		}

		err = tm.SetGlobalKV("default_project", projectID, nil)
		if err != nil {
			return fmt.Errorf("failed to set default project: %w", err)
		}

		fmt.Printf("Set %s as default project\n", projectID)
		return nil
	},
}

// Show current default project command
var defaultProjectCmd = &cobra.Command{
	Use:   "default",
	Short: "Show the current default project",
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		project, err := tm.GetDefaultProject()
		if err != nil {
			return fmt.Errorf("failed to get default project: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(project)
		}

		fmt.Printf("Default Project: %s\n", project.Name)
		fmt.Printf("ID: %s\n", project.ID)
		if project.Description != "" {
			fmt.Printf("Description: %s\n", project.Description)
		}
		if project.Guidelines != "" {
			fmt.Printf("\nGuidelines:\n%s\n", project.Guidelines)
		}

		return nil
	},
}

func outputProjects(projects []Project, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(projects)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(projects)
	case "csv":
		return outputProjectsCSV(projects)
	default: // table
		return outputProjectsTable(projects)
	}
}

func outputProjectsTable(projects []Project) error {
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tCREATED")

	for _, project := range projects {
		id := truncateString(project.ID, 8)
		name := truncateString(project.Name, 30)
		description := truncateString(project.Description, 40)
		created := project.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, description, created)
	}

	return w.Flush()
}

func outputProjectsCSV(projects []Project) error {
	fmt.Println("id,name,description,guidelines,author_id,created_at,updated_at")
	for _, project := range projects {
		authorID := ""
		if project.AuthorID != nil {
			authorID = *project.AuthorID
		}

		fmt.Printf("%s,%q,%q,%q,%s,%s,%s\n",
			project.ID,
			project.Name,
			project.Description,
			project.Guidelines,
			authorID,
			project.CreatedAt.Format(time.RFC3339),
			project.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Add subcommands
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(createProjectCmd)
	projectsCmd.AddCommand(setDefaultProjectCmd)
	projectsCmd.AddCommand(defaultProjectCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listProjectsCmd, defaultProjectCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// Create specific flags
	createProjectCmd.Flags().StringP("description", "d", "", "Project description")
	createProjectCmd.Flags().StringP("guidelines", "g", "", "Project guidelines")
	createProjectCmd.Flags().Bool("set-default", false, "Set as default project")
	createProjectCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
