package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show project dashboard overview",
	Long:  "Display comprehensive overview of current project including tasks, agents, and recent activity",
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		return showDashboard(verbose)
	},
}

type DashboardData struct {
	Project     *Project
	Tasks       []Task
	Agents      []Agent
	AgentTypes  []AgentType
	RecentNotes []Note
	RecentTILs  []TIL
}

func showDashboard(verbose bool) error {
	dbPath, _ := rootCmd.PersistentFlags().GetString("db")
	logger := zerolog.Nop()

	tm, err := NewTaskManager(dbPath, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize task manager: %w", err)
	}

	// Get current project
	project, err := tm.GetDefaultProject()
	if err != nil {
		return fmt.Errorf("failed to get current project: %w", err)
	}

	// Gather all dashboard data
	data, err := gatherDashboardData(tm, project)
	if err != nil {
		return fmt.Errorf("failed to gather dashboard data: %w", err)
	}

	if verbose {
		displayVerboseDashboard(data)
	} else {
		displayConciseDashboard(data)
	}

	return nil
}

func gatherDashboardData(tm *TaskManager, project *Project) (*DashboardData, error) {
	data := &DashboardData{Project: project}

	// Get tasks for current project
	tasks, err := tm.ListTasks(nil, nil, nil, &project.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	data.Tasks = tasks

	// Get agents
	agents, err := tm.ListAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	data.Agents = agents

	// Get agent types for current project
	agentTypes, err := tm.ListAgentTypes(&project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent types: %w", err)
	}
	data.AgentTypes = agentTypes

	// Get recent notes (last 5)
	recentNotes, err := getRecentNotes(tm, project.ID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent notes: %w", err)
	}
	data.RecentNotes = recentNotes

	// Get recent TILs (last 3)
	recentTILs, err := getRecentTILs(tm, project.ID, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent TILs: %w", err)
	}
	data.RecentTILs = recentTILs

	return data, nil
}

func displayConciseDashboard(data *DashboardData) {
	fmt.Printf("# 📊 Project Dashboard\n\n")

	// Project header
	fmt.Printf("## 🏗️ **%s**\n", data.Project.Name)
	if data.Project.Description != "" {
		fmt.Printf("*%s*\n\n", data.Project.Description)
	}

	// Task status summary
	taskStats := getTaskStatusCounts(data.Tasks)
	fmt.Printf("## 📋 Task Overview\n")
	fmt.Printf("**Total:** %d | ", len(data.Tasks))
	fmt.Printf("✅ **%d** Completed | ", taskStats["completed"])
	fmt.Printf("🔄 **%d** In Progress | ", taskStats["in_progress"])
	fmt.Printf("⏳ **%d** Pending", taskStats["pending"])
	if taskStats["failed"] > 0 {
		fmt.Printf(" | ❌ **%d** Failed", taskStats["failed"])
	}
	fmt.Printf("\n\n")

	// Agent summary
	fmt.Printf("## 👥 Team Status\n")
	fmt.Printf("**%d** Agents across **%d** Types\n\n", len(data.Agents), len(data.AgentTypes))

	// Active work
	activeWork := getActiveWork(data.Tasks)
	if len(activeWork) > 0 {
		fmt.Printf("## 🚀 Current Work\n")
		for _, task := range activeWork {
			agentName := "Unassigned"
			if task.AgentID != nil {
				if agent := findAgent(data.Agents, *task.AgentID); agent != nil {
					agentName = agent.Name
				}
			}
			fmt.Printf("- **%s** (%s)\n", task.Title, agentName)
		}
		fmt.Printf("\n")
	}

	// Recent activity
	if len(data.RecentNotes) > 0 {
		fmt.Printf("## 📝 Recent Activity\n")
		for i, note := range data.RecentNotes {
			if i >= 3 { // Limit to 3 in concise view
				break
			}
			fmt.Printf("- %s *(%s)*\n",
				truncateToWords(note.Content, 8),
				formatTimeAgo(note.CreatedAt))
		}
		fmt.Printf("\n")
	}

	// Project guidelines
	if data.Project.Guidelines != "" {
		fmt.Printf("## 📜 Guidelines\n")
		fmt.Printf("*%s*\n\n", data.Project.Guidelines)
	}
}

func displayVerboseDashboard(data *DashboardData) {
	fmt.Printf("# 📊 Project Dashboard - Detailed View\n\n")

	// Project header with more details
	fmt.Printf("## 🏗️ Project: **%s**\n", data.Project.Name)
	fmt.Printf("**ID:** `%s`\n", data.Project.ID)
	if data.Project.Description != "" {
		fmt.Printf("**Description:** %s\n", data.Project.Description)
	}
	fmt.Printf("**Created:** %s\n\n", data.Project.CreatedAt.Format("2006-01-02 15:04"))

	// Detailed task breakdown
	fmt.Printf("## 📋 Task Breakdown\n")
	taskStats := getTaskStatusCounts(data.Tasks)
	totalTasks := len(data.Tasks)

	if totalTasks > 0 {
		fmt.Printf("| Status | Count | Percentage |\n")
		fmt.Printf("|--------|-------|------------|\n")
		fmt.Printf("| ✅ Completed | %d | %.1f%% |\n", taskStats["completed"], float64(taskStats["completed"])/float64(totalTasks)*100)
		fmt.Printf("| 🔄 In Progress | %d | %.1f%% |\n", taskStats["in_progress"], float64(taskStats["in_progress"])/float64(totalTasks)*100)
		fmt.Printf("| ⏳ Pending | %d | %.1f%% |\n", taskStats["pending"], float64(taskStats["pending"])/float64(totalTasks)*100)
		if taskStats["failed"] > 0 {
			fmt.Printf("| ❌ Failed | %d | %.1f%% |\n", taskStats["failed"], float64(taskStats["failed"])/float64(totalTasks)*100)
		}
		fmt.Printf("| **Total** | **%d** | **100.0%%** |\n\n", totalTasks)
	} else {
		fmt.Printf("*No tasks found for this project.*\n\n")
	}

	// Agent types and agents
	fmt.Printf("## 👥 Team Structure\n")
	if len(data.AgentTypes) > 0 {
		for _, agentType := range data.AgentTypes {
			scope := "Project"
			if agentType.Global {
				scope = "Global"
			}
			fmt.Printf("### %s (%s)\n", agentType.Name, scope)
			fmt.Printf("*%s*\n", agentType.Description)

			// Find agents of this type
			typeAgents := getAgentsOfType(data.Agents, agentType.ID)
			if len(typeAgents) > 0 {
				fmt.Printf("**Agents:** ")
				for i, agent := range typeAgents {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", agent.Name)
				}
				fmt.Printf("\n")
			} else {
				fmt.Printf("*No agents assigned to this type*\n")
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("*No agent types defined for this project.*\n\n")
	}

	// Detailed current work
	activeWork := getActiveWork(data.Tasks)
	if len(activeWork) > 0 {
		fmt.Printf("## 🚀 Work In Progress\n")
		for _, task := range activeWork {
			fmt.Printf("### %s\n", task.Title)
			fmt.Printf("**ID:** `%s` | **Status:** %s\n", task.ID, getStatusEmoji(string(task.Status)))
			if task.Description != "" {
				fmt.Printf("**Description:** %s\n", task.Description)
			}
			if task.AgentID != nil {
				if agent := findAgent(data.Agents, *task.AgentID); agent != nil {
					fmt.Printf("**Assigned to:** %s\n", agent.Name)
				}
			}
			fmt.Printf("**Started:** %s\n\n", task.UpdatedAt.Format("2006-01-02 15:04"))
		}
	}

	// Recent completions
	recentCompletions := getRecentCompletions(data.Tasks, 5)
	if len(recentCompletions) > 0 {
		fmt.Printf("## ✅ Recent Completions\n")
		for _, task := range recentCompletions {
			agentName := "Unknown"
			if task.AgentID != nil {
				if agent := findAgent(data.Agents, *task.AgentID); agent != nil {
					agentName = agent.Name
				}
			}
			fmt.Printf("- **%s** by %s (%s)\n",
				task.Title,
				agentName,
				formatTimeAgo(task.UpdatedAt))
		}
		fmt.Printf("\n")
	}

	// Recent notes
	if len(data.RecentNotes) > 0 {
		fmt.Printf("## 📝 Recent Notes\n")
		for _, note := range data.RecentNotes {
			fmt.Printf("- %s *(%s)*\n",
				note.Content,
				formatTimeAgo(note.CreatedAt))
		}
		fmt.Printf("\n")
	}

	// Recent TILs
	if len(data.RecentTILs) > 0 {
		fmt.Printf("## 💡 Recent Insights (TIL)\n")
		for _, til := range data.RecentTILs {
			fmt.Printf("### %s\n", til.Title)
			fmt.Printf("%s\n", til.Content)
			fmt.Printf("*Shared %s*\n\n", formatTimeAgo(til.CreatedAt))
		}
	}

	// Project guidelines
	if data.Project.Guidelines != "" {
		fmt.Printf("## 📜 Project Guidelines\n")
		fmt.Printf("%s\n\n", data.Project.Guidelines)
	}
}

// Helper functions

func getTaskStatusCounts(tasks []Task) map[string]int {
	counts := map[string]int{
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
		"failed":      0,
	}

	for _, task := range tasks {
		counts[string(task.Status)]++
	}

	return counts
}

func getActiveWork(tasks []Task) []Task {
	var active []Task
	for _, task := range tasks {
		if string(task.Status) == "in_progress" {
			active = append(active, task)
		}
	}

	// Sort by update time (most recent first)
	sort.Slice(active, func(i, j int) bool {
		return active[i].UpdatedAt.After(active[j].UpdatedAt)
	})

	return active
}

func getRecentCompletions(tasks []Task, limit int) []Task {
	var completed []Task
	for _, task := range tasks {
		if string(task.Status) == "completed" {
			completed = append(completed, task)
		}
	}

	// Sort by update time (most recent first)
	sort.Slice(completed, func(i, j int) bool {
		return completed[i].UpdatedAt.After(completed[j].UpdatedAt)
	})

	if len(completed) > limit {
		completed = completed[:limit]
	}

	return completed
}

func findAgent(agents []Agent, agentID string) *Agent {
	for _, agent := range agents {
		if agent.ID == agentID {
			return &agent
		}
	}
	return nil
}

func getAgentsOfType(agents []Agent, agentTypeID string) []Agent {
	var typeAgents []Agent
	for _, agent := range agents {
		if agent.AgentTypeSlug != nil && *agent.AgentTypeSlug == agentTypeID {
			typeAgents = append(typeAgents, agent)
		}
	}
	return typeAgents
}

func getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳ Pending"
	case "in_progress":
		return "🔄 In Progress"
	case "completed":
		return "✅ Completed"
	case "failed":
		return "❌ Failed"
	default:
		return status
	}
}

func truncateToWords(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}
	return strings.Join(words[:maxWords], " ") + "..."
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func getRecentNotes(tm *TaskManager, projectID string, limit int) ([]Note, error) {
	// Get all tasks for the project first
	tasks, err := tm.ListTasks(nil, nil, nil, &projectID, nil)
	if err != nil {
		return nil, err
	}

	var allNotes []Note
	for _, task := range tasks {
		notes, err := tm.ListNotes(&task.ID, nil)
		if err != nil {
			continue // Skip this task's notes if there's an error
		}
		allNotes = append(allNotes, notes...)
	}

	// Sort by creation time (most recent first)
	sort.Slice(allNotes, func(i, j int) bool {
		return allNotes[i].CreatedAt.After(allNotes[j].CreatedAt)
	})

	if len(allNotes) > limit {
		allNotes = allNotes[:limit]
	}

	return allNotes, nil
}

func getRecentTILs(tm *TaskManager, projectID string, limit int) ([]TIL, error) {
	tils, err := tm.ListTILs(&projectID, nil, nil)
	if err != nil {
		return nil, err
	}

	// Sort by creation time (most recent first)
	sort.Slice(tils, func(i, j int) bool {
		return tils[i].CreatedAt.After(tils[j].CreatedAt)
	})

	if len(tils) > limit {
		tils = tils[:limit]
	}

	return tils, nil
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().Bool("verbose", false, "Show detailed dashboard view")
}
