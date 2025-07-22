package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run a demonstration of the task coordination system",
	Long: `Create sample data and demonstrate the task coordination system functionality.

This command will:
1. Create a sample task hierarchy
2. Add dependencies between tasks 
3. Create sample agents
4. Demonstrate task assignment and completion`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		logger.Info().Msg("Creating sample project and task hierarchy...")

		// Create project
		project, err := tm.CreateProject("Agent Coordination System", "A task management system for AI agents", "Work collaboratively, follow dependencies, communicate progress clearly", nil)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		// Set as default project
		err = tm.SetGlobalKV("default_project", project.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to set default project: %w", err)
		}

		// Create agent types
		codeReviewType, err := tm.CreateAgentType("Code Reviewer", "Reviews code for quality and standards", project.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent type: %w", err)
		}

		testRunnerType, err := tm.CreateAgentType("Test Runner", "Executes tests and validates functionality", project.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent type: %w", err)
		}

		docAgentType, err := tm.CreateAgentType("Documentation Writer", "Creates and maintains documentation", project.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent type: %w", err)
		}

		// Create root task
		rootTask, err := tm.CreateTask("Build Agent Coordination System", "Main project to build agent coordination system", nil, project.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to create root task: %w", err)
		}

		// Create subtasks with preferred agent types
		designTask, err := tm.CreateTask("Design Database Schema", "Design the task management database schema", &rootTask.ID, project.ID, &codeReviewType.ID)
		if err != nil {
			return fmt.Errorf("failed to create design task: %w", err)
		}

		implementTask, err := tm.CreateTask("Implement Task Manager", "Implement the task manager with CRUD operations", &rootTask.ID, project.ID, &codeReviewType.ID)
		if err != nil {
			return fmt.Errorf("failed to create implement task: %w", err)
		}

		testTask, err := tm.CreateTask("Write Tests", "Write unit tests for task manager", &rootTask.ID, project.ID, &testRunnerType.ID)
		if err != nil {
			return fmt.Errorf("failed to create test task: %w", err)
		}

		cliTask, err := tm.CreateTask("Build CLI Tools", "Create command-line interface tools", &rootTask.ID, project.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to create CLI task: %w", err)
		}

		// Create documentation task to demonstrate the feature
		_, err = tm.CreateTask("Write Documentation", "Create user and API documentation", &rootTask.ID, project.ID, &docAgentType.ID)
		if err != nil {
			return fmt.Errorf("failed to create doc task: %w", err)
		}

		// Add dependencies
		err = tm.AddDependency(implementTask.ID, designTask.ID)
		if err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}

		err = tm.AddDependency(testTask.ID, implementTask.ID)
		if err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}

		err = tm.AddDependency(cliTask.ID, implementTask.ID)
		if err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}

		// Create agents
		agent1, err := tm.CreateAgent("Code Review Agent", &codeReviewType.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		agent2, err := tm.CreateAgent("Test Runner Agent", &testRunnerType.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		_, err = tm.CreateAgent("Documentation Agent", &docAgentType.ID)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		logger.Info().Msg("Created sample data successfully")

		// Demonstrate workflow
		logger.Info().Msg("Simulating agent workflow...")

		// Get available tasks
		availableTasks, err := tm.GetAvailableTasks(nil)
		if err != nil {
			return fmt.Errorf("failed to get available tasks: %w", err)
		}

		fmt.Printf("\nInitial available tasks: %d\n", len(availableTasks))
		for _, task := range availableTasks {
			fmt.Printf("  - %s: %s\n", task.ID[:8], task.Title)
		}

		// Assign and complete first available task (design task)
		if len(availableTasks) > 0 {
			firstTask := availableTasks[0]
			err = tm.AssignTask(firstTask.ID, agent1.ID)
			if err != nil {
				return fmt.Errorf("failed to assign task: %w", err)
			}

			fmt.Printf("\nAssigned '%s' to %s\n", firstTask.Title, agent1.Name)

			err = tm.UpdateTaskStatus(firstTask.ID, TaskStatusCompleted)
			if err != nil {
				return fmt.Errorf("failed to update task status: %w", err)
			}

			fmt.Printf("Completed '%s'\n", firstTask.Title)

			// Check available tasks again
			availableTasks, err = tm.GetAvailableTasks(nil)
			if err != nil {
				return fmt.Errorf("failed to get available tasks: %w", err)
			}

			fmt.Printf("\nAvailable tasks after completion: %d\n", len(availableTasks))
			for _, task := range availableTasks {
				fmt.Printf("  - %s: %s\n", task.ID[:8], task.Title)
			}
		}

		// Assign implementation task
		if len(availableTasks) > 0 {
			for _, task := range availableTasks {
				if task.Title == "Implement Task Manager" {
					err = tm.AssignTask(task.ID, agent2.ID)
					if err != nil {
						return fmt.Errorf("failed to assign task: %w", err)
					}
					fmt.Printf("\nAssigned '%s' to %s\n", task.Title, agent2.Name)
					break
				}
			}
		}

		fmt.Printf("\nDemo completed successfully! Use the CLI tools to explore:\n")
		fmt.Printf("  amp-tasks tasks list\n")
		fmt.Printf("  amp-tasks agents list\n")
		fmt.Printf("  amp-tasks deps graph\n")
		fmt.Printf("  amp-tasks tasks available\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}
