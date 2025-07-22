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

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Agent management commands",
	Long:  "Commands for creating, listing, and managing agents in the coordination system",
}

// List agents command
var listAgentsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
	Long: `List all registered agents in the system.

Examples:
  amp-tasks agents list
  amp-tasks agents list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		agents, err := tm.ListAgents()
		if err != nil {
			return fmt.Errorf("failed to list agents: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputAgents(agents, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Create agent command
var createAgentCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new agent",
	Args:  cobra.ExactArgs(1),
	Long: `Create a new agent with the specified name.

Examples:
  amp-tasks agents create "Code Review Agent"
  amp-tasks agents create "Test Runner" --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		var agentTypeID *string
		if cmd.Flags().Changed("type") {
			typeID, _ := cmd.Flags().GetString("type")
			agentTypeID = &typeID
		}

		agent, err := tm.CreateAgent(name, agentTypeID)
		if err != nil {
			return fmt.Errorf("failed to create agent: %w", err)
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(agent)
		}

		fmt.Printf("Created agent: %s\n", agent.ID)
		fmt.Printf("Name: %s\n", agent.Name)
		fmt.Printf("Status: %s\n", agent.Status)

		return nil
	},
}

// Agent workload command
var workloadCmd = &cobra.Command{
	Use:   "workload",
	Short: "Show agent workload distribution",
	Long: `Display the current workload distribution across all agents.

Shows how many tasks are assigned to each agent by status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		agents, err := tm.ListAgents()
		if err != nil {
			return fmt.Errorf("failed to list agents: %w", err)
		}

		// Get task counts per agent
		workload := make(map[string]map[TaskStatus]int)
		for _, agent := range agents {
			workload[agent.ID] = make(map[TaskStatus]int)

			// Count tasks by status for this agent
			for _, status := range []TaskStatus{TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusFailed} {
				tasks, err := tm.ListTasks(nil, &status, &agent.ID, nil, nil)
				if err != nil {
					return fmt.Errorf("failed to get tasks for agent %s: %w", agent.ID, err)
				}
				workload[agent.ID][status] = len(tasks)
			}
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputWorkload(agents, workload, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

// Agent stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show agent statistics",
	Long: `Display overall statistics about agents and their task completion rates.

Shows metrics like total tasks completed, success rate, etc.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		agents, err := tm.ListAgents()
		if err != nil {
			return fmt.Errorf("failed to list agents: %w", err)
		}

		var stats []AgentStats
		for _, agent := range agents {
			stat := AgentStats{Agent: agent}

			// Count tasks by status
			for _, status := range []TaskStatus{TaskStatusPending, TaskStatusInProgress, TaskStatusCompleted, TaskStatusFailed} {
				tasks, err := tm.ListTasks(nil, &status, &agent.ID, nil, nil)
				if err != nil {
					return fmt.Errorf("failed to get tasks for agent %s: %w", agent.ID, err)
				}

				count := len(tasks)
				switch status {
				case TaskStatusPending:
					stat.Pending = count
				case TaskStatusInProgress:
					stat.InProgress = count
				case TaskStatusCompleted:
					stat.Completed = count
				case TaskStatusFailed:
					stat.Failed = count
				}
			}

			stat.Total = stat.Pending + stat.InProgress + stat.Completed + stat.Failed
			if stat.Completed+stat.Failed > 0 {
				stat.SuccessRate = float64(stat.Completed) / float64(stat.Completed+stat.Failed) * 100
			}

			stats = append(stats, stat)
		}

		output, _ := cmd.Flags().GetString("output")

		// Show project context in dual mode (default table output)
		if output == "table" {
			showProjectTitle(tm)
		}

		err = outputAgentStats(stats, output)

		// Show TIL/notes reminders in table mode
		if output == "table" {
			showTILNotesReminders()
		}

		return err
	},
}

func outputAgents(agents []Agent, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(agents)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(agents)
	case "csv":
		return outputAgentsCSV(agents)
	default: // table
		return outputAgentsTable(agents)
	}
}

func outputAgentsTable(agents []Agent) error {
	if len(agents) == 0 {
		fmt.Println("No agents found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED")

	for _, agent := range agents {
		id := truncateString(agent.ID, 8)
		name := truncateString(agent.Name, 30)
		created := agent.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, agent.Status, created)
	}

	return w.Flush()
}

func outputAgentsCSV(agents []Agent) error {
	fmt.Println("id,name,status,created_at,updated_at")
	for _, agent := range agents {
		fmt.Printf("%s,%q,%s,%s,%s\n",
			agent.ID,
			agent.Name,
			agent.Status,
			agent.CreatedAt.Format(time.RFC3339),
			agent.UpdatedAt.Format(time.RFC3339),
		)
	}
	return nil
}

func outputWorkload(agents []Agent, workload map[string]map[TaskStatus]int, format string) error {
	switch format {
	case "json":
		result := make(map[string]interface{})
		for _, agent := range agents {
			result[agent.ID] = map[string]interface{}{
				"agent":    agent,
				"workload": workload[agent.ID],
			}
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "yaml":
		result := make(map[string]interface{})
		for _, agent := range agents {
			result[agent.ID] = map[string]interface{}{
				"agent":    agent,
				"workload": workload[agent.ID],
			}
		}
		return yaml.NewEncoder(os.Stdout).Encode(result)
	default: // table
		return outputWorkloadTable(agents, workload)
	}
}

func outputWorkloadTable(agents []Agent, workload map[string]map[TaskStatus]int) error {
	if len(agents) == 0 {
		fmt.Println("No agents found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "AGENT\tNAME\tPENDING\tIN_PROGRESS\tCOMPLETED\tFAILED\tTOTAL")

	for _, agent := range agents {
		id := truncateString(agent.ID, 8)
		name := truncateString(agent.Name, 20)

		pending := workload[agent.ID][TaskStatusPending]
		inProgress := workload[agent.ID][TaskStatusInProgress]
		completed := workload[agent.ID][TaskStatusCompleted]
		failed := workload[agent.ID][TaskStatusFailed]
		total := pending + inProgress + completed + failed

		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
			id, name, pending, inProgress, completed, failed, total)
	}

	return w.Flush()
}

func outputAgentStats(stats []AgentStats, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(stats)
	case "yaml":
		return yaml.NewEncoder(os.Stdout).Encode(stats)
	default: // table
		return outputAgentStatsTable(stats)
	}
}

func outputAgentStatsTable(stats []AgentStats) error {
	if len(stats) == 0 {
		fmt.Println("No agent statistics found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "AGENT\tNAME\tTOTAL\tCOMPLETED\tFAILED\tSUCCESS_RATE")

	for _, stat := range stats {
		id := truncateString(stat.Agent.ID, 8)
		name := truncateString(stat.Agent.Name, 20)

		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%.1f%%\n",
			id, name, stat.Total, stat.Completed, stat.Failed, stat.SuccessRate)
	}

	return w.Flush()
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	// Add subcommands
	agentsCmd.AddCommand(listAgentsCmd)
	agentsCmd.AddCommand(createAgentCmd)
	agentsCmd.AddCommand(workloadCmd)
	agentsCmd.AddCommand(statsCmd)

	// Common flags
	for _, cmd := range []*cobra.Command{listAgentsCmd, workloadCmd, statsCmd} {
		cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml, csv)")
	}

	// Create specific flags
	createAgentCmd.Flags().StringP("type", "t", "", "Agent type ID")
	createAgentCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
