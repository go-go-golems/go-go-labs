package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate comprehensive project reports",
	Long: `Generate detailed reports on project development including tasks, agents, timeline, and insights.

Examples:
  amp-tasks report                     # Generate default markdown report
  amp-tasks report --format json      # Generate JSON report  
  amp-tasks report --format text      # Generate text report
  amp-tasks report --output report.md # Save to file
  amp-tasks report --verbose          # Include all details
  amp-tasks report --summary          # Condensed version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tm, err := NewTaskManager(dbPath, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize task manager: %w", err)
		}
		defer tm.Close()

		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")
		summary, _ := cmd.Flags().GetBool("summary")

		report, err := generateReport(tm, verbose, summary)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		var content string
		switch format {
		case "json":
			jsonData, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			content = string(jsonData)
		case "text":
			content = formatReportText(report)
		case "markdown":
			content = formatReportMarkdown(report)
		default:
			content = formatReportMarkdown(report)
		}

		if output != "" {
			return os.WriteFile(output, []byte(content), 0644)
		}

		fmt.Print(content)
		return nil
	},
}

type ProjectReport struct {
	Project            Project                `json:"project"`
	Summary            ProjectSummary         `json:"summary"`
	Timeline           []TimelineEvent        `json:"timeline"`
	AgentContributions []AgentContribution    `json:"agent_contributions"`
	TaskAnalysis       TaskAnalysis           `json:"task_analysis"`
	Knowledge          KnowledgeDocumentation `json:"knowledge"`
	Insights           ProjectInsights        `json:"insights"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

type ProjectSummary struct {
	TotalTasks      int           `json:"total_tasks"`
	CompletedTasks  int           `json:"completed_tasks"`
	PendingTasks    int           `json:"pending_tasks"`
	FailedTasks     int           `json:"failed_tasks"`
	TotalAgents     int           `json:"total_agents"`
	TotalNotes      int           `json:"total_notes"`
	TotalTILs       int           `json:"total_tils"`
	ProjectDuration time.Duration `json:"project_duration"`
	CompletionRate  float64       `json:"completion_rate"`
	AvgTaskDuration time.Duration `json:"avg_task_duration"`
}

type TimelineEvent struct {
	Type        string    `json:"type"` // task_created, task_completed, task_assigned, note_added, til_created
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	TaskID      *string   `json:"task_id,omitempty"`
	AgentID     *string   `json:"agent_id,omitempty"`
	AgentName   *string   `json:"agent_name,omitempty"`
}

type AgentContribution struct {
	Agent             Agent         `json:"agent"`
	TasksCompleted    int           `json:"tasks_completed"`
	TasksFailed       int           `json:"tasks_failed"`
	TasksPending      int           `json:"tasks_pending"`
	NotesWritten      int           `json:"notes_written"`
	TILsShared        int           `json:"tils_shared"`
	SuccessRate       float64       `json:"success_rate"`
	AvgCompletionTime time.Duration `json:"avg_completion_time"`
	FirstActivity     *time.Time    `json:"first_activity,omitempty"`
	LastActivity      *time.Time    `json:"last_activity,omitempty"`
}

type TaskAnalysis struct {
	TaskHierarchy    []TaskNode          `json:"task_hierarchy"`
	Dependencies     []TaskDependency    `json:"dependencies"`
	CompletionTimes  map[string]float64  `json:"completion_times"` // taskID -> hours
	BlockingPatterns []string            `json:"blocking_patterns"`
	VelocityTrends   []VelocityDataPoint `json:"velocity_trends"`
}

type TaskNode struct {
	Task     Task       `json:"task"`
	Children []TaskNode `json:"children,omitempty"`
	Notes    []Note     `json:"notes,omitempty"`
	TILs     []TIL      `json:"tils,omitempty"`
}

type VelocityDataPoint struct {
	Date           time.Time `json:"date"`
	TasksCompleted int       `json:"tasks_completed"`
	TasksCreated   int       `json:"tasks_created"`
}

type KnowledgeDocumentation struct {
	NotesByTask  map[string][]Note `json:"notes_by_task"`
	NotesByAgent map[string][]Note `json:"notes_by_agent"`
	TILsByTask   map[string][]TIL  `json:"tils_by_task"`
	TILsByAgent  map[string][]TIL  `json:"tils_by_agent"`
	TopLearnings []TIL             `json:"top_learnings"`
}

type ProjectInsights struct {
	MostProductiveAgent   *Agent   `json:"most_productive_agent,omitempty"`
	LongestRunningTask    *Task    `json:"longest_running_task,omitempty"`
	MostCollaborativeTask *Task    `json:"most_collaborative_task,omitempty"`
	CommonBlockers        []string `json:"common_blockers"`
	KnowledgeSharing      float64  `json:"knowledge_sharing_rate"` // TILs per completed task
	TaskVelocity          float64  `json:"task_velocity"`          // tasks per day
	AgentUtilization      float64  `json:"agent_utilization"`      // % of agents actively working
}

func generateReport(tm *TaskManager, verbose, summary bool) (*ProjectReport, error) {
	project, err := tm.GetDefaultProject()
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Gather all data
	tasks, err := tm.ListTasks(nil, nil, nil, &project.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	agents, err := tm.ListAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	notes, err := tm.ListNotes(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	tils, err := tm.ListTILs(&project.ID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list TILs: %w", err)
	}

	dependencies, err := getAllDependencies(tm, tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	report := &ProjectReport{
		Project:     *project,
		GeneratedAt: time.Now(),
	}

	// Generate summary
	report.Summary = generateSummary(tasks, agents, notes, tils, project.CreatedAt)

	// Generate timeline
	report.Timeline = generateTimeline(tasks, notes, tils, agents)

	// Generate agent contributions
	report.AgentContributions = generateAgentContributions(tasks, notes, tils, agents)

	// Generate task analysis
	report.TaskAnalysis = generateTaskAnalysis(tasks, dependencies, notes, tils)

	// Generate knowledge documentation
	report.Knowledge = generateKnowledgeDocumentation(notes, tils, tasks, agents)

	// Generate insights
	report.Insights = generateInsights(tasks, agents, notes, tils, report.Summary, report.AgentContributions)

	return report, nil
}

func getAllDependencies(tm *TaskManager, tasks []Task) ([]TaskDependency, error) {
	var allDeps []TaskDependency
	for _, task := range tasks {
		deps, err := tm.GetTaskDependencies(task.ID)
		if err != nil {
			return nil, err
		}
		allDeps = append(allDeps, deps...)
	}
	return allDeps, nil
}

func generateSummary(tasks []Task, agents []Agent, notes []Note, tils []TIL, projectStart time.Time) ProjectSummary {
	summary := ProjectSummary{
		TotalTasks:  len(tasks),
		TotalAgents: len(agents),
		TotalNotes:  len(notes),
		TotalTILs:   len(tils),
	}

	var completedTasks []Task
	for _, task := range tasks {
		switch task.Status {
		case TaskStatusCompleted:
			summary.CompletedTasks++
			completedTasks = append(completedTasks, task)
		case TaskStatusPending:
			summary.PendingTasks++
		case TaskStatusFailed:
			summary.FailedTasks++
		}
	}

	if summary.TotalTasks > 0 {
		summary.CompletionRate = float64(summary.CompletedTasks) / float64(summary.TotalTasks) * 100
	}

	summary.ProjectDuration = time.Since(projectStart)

	// Calculate average task duration for completed tasks
	if len(completedTasks) > 0 {
		var totalDuration time.Duration
		for _, task := range completedTasks {
			totalDuration += task.UpdatedAt.Sub(task.CreatedAt)
		}
		summary.AvgTaskDuration = totalDuration / time.Duration(len(completedTasks))
	}

	return summary
}

func generateTimeline(tasks []Task, notes []Note, tils []TIL, agents []Agent) []TimelineEvent {
	var events []TimelineEvent

	agentNames := make(map[string]string)
	for _, agent := range agents {
		agentNames[agent.ID] = agent.Name
	}

	// Task events
	for _, task := range tasks {
		events = append(events, TimelineEvent{
			Type:        "task_created",
			Time:        task.CreatedAt,
			Description: fmt.Sprintf("Task created: %s", task.Title),
			TaskID:      &task.ID,
		})

		if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
			events = append(events, TimelineEvent{
				Type:        "task_" + string(task.Status),
				Time:        task.UpdatedAt,
				Description: fmt.Sprintf("Task %s: %s", task.Status, task.Title),
				TaskID:      &task.ID,
				AgentID:     task.AgentID,
				AgentName:   getAgentName(task.AgentID, agentNames),
			})
		}

		if task.AgentID != nil {
			agentName := getAgentName(task.AgentID, agentNames)
			events = append(events, TimelineEvent{
				Type:        "task_assigned",
				Time:        task.UpdatedAt,
				Description: fmt.Sprintf("Task assigned to %s: %s", *agentName, task.Title),
				TaskID:      &task.ID,
				AgentID:     task.AgentID,
				AgentName:   agentName,
			})
		}
	}

	// Note events
	for _, note := range notes {
		agentName := getAgentName(&note.AgentID, agentNames)
		events = append(events, TimelineEvent{
			Type:        "note_added",
			Time:        note.CreatedAt,
			Description: fmt.Sprintf("Note added by %s", *agentName),
			TaskID:      &note.TaskID,
			AgentID:     &note.AgentID,
			AgentName:   agentName,
		})
	}

	// TIL events
	for _, til := range tils {
		agentName := getAgentName(&til.AgentID, agentNames)
		events = append(events, TimelineEvent{
			Type:        "til_created",
			Time:        til.CreatedAt,
			Description: fmt.Sprintf("TIL shared by %s: %s", *agentName, til.Title),
			TaskID:      til.TaskID,
			AgentID:     &til.AgentID,
			AgentName:   agentName,
		})
	}

	// Sort by time
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.Before(events[j].Time)
	})

	return events
}

func getAgentName(agentID *string, agentNames map[string]string) *string {
	if agentID == nil {
		return nil
	}
	if name, ok := agentNames[*agentID]; ok {
		return &name
	}
	return agentID
}

func generateAgentContributions(tasks []Task, notes []Note, tils []TIL, agents []Agent) []AgentContribution {
	contributions := make(map[string]*AgentContribution)

	// Initialize contributions for all agents
	for _, agent := range agents {
		contributions[agent.ID] = &AgentContribution{
			Agent: agent,
		}
	}

	// Count task contributions
	var completionTimes []time.Duration
	for _, task := range tasks {
		if task.AgentID != nil {
			contrib, exists := contributions[*task.AgentID]
			if !exists {
				continue // Skip tasks assigned to agents not in our list
			}
			switch task.Status {
			case TaskStatusCompleted:
				contrib.TasksCompleted++
				duration := task.UpdatedAt.Sub(task.CreatedAt)
				completionTimes = append(completionTimes, duration)
				if contrib.FirstActivity == nil || task.CreatedAt.Before(*contrib.FirstActivity) {
					contrib.FirstActivity = &task.CreatedAt
				}
				if contrib.LastActivity == nil || task.UpdatedAt.After(*contrib.LastActivity) {
					contrib.LastActivity = &task.UpdatedAt
				}
			case TaskStatusFailed:
				contrib.TasksFailed++
			case TaskStatusPending, TaskStatusInProgress:
				contrib.TasksPending++
			}
		}
	}

	// Calculate average completion time per agent
	agentCompletionTimes := make(map[string][]time.Duration)
	for _, task := range tasks {
		if task.AgentID != nil && task.Status == TaskStatusCompleted {
			duration := task.UpdatedAt.Sub(task.CreatedAt)
			agentCompletionTimes[*task.AgentID] = append(agentCompletionTimes[*task.AgentID], duration)
		}
	}

	for agentID, durations := range agentCompletionTimes {
		if len(durations) > 0 {
			if contrib, exists := contributions[agentID]; exists {
				var total time.Duration
				for _, d := range durations {
					total += d
				}
				contrib.AvgCompletionTime = total / time.Duration(len(durations))
			}
		}
	}

	// Count notes
	for _, note := range notes {
		if contrib, ok := contributions[note.AgentID]; ok {
			contrib.NotesWritten++
			if contrib.FirstActivity == nil || note.CreatedAt.Before(*contrib.FirstActivity) {
				contrib.FirstActivity = &note.CreatedAt
			}
			if contrib.LastActivity == nil || note.CreatedAt.After(*contrib.LastActivity) {
				contrib.LastActivity = &note.CreatedAt
			}
		}
	}

	// Count TILs
	for _, til := range tils {
		if contrib, ok := contributions[til.AgentID]; ok {
			contrib.TILsShared++
			if contrib.FirstActivity == nil || til.CreatedAt.Before(*contrib.FirstActivity) {
				contrib.FirstActivity = &til.CreatedAt
			}
			if contrib.LastActivity == nil || til.CreatedAt.After(*contrib.LastActivity) {
				contrib.LastActivity = &til.CreatedAt
			}
		}
	}

	// Calculate success rates
	for _, contrib := range contributions {
		total := contrib.TasksCompleted + contrib.TasksFailed
		if total > 0 {
			contrib.SuccessRate = float64(contrib.TasksCompleted) / float64(total) * 100
		}
	}

	// Convert to slice and sort by productivity
	var result []AgentContribution
	for _, contrib := range contributions {
		result = append(result, *contrib)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TasksCompleted > result[j].TasksCompleted
	})

	return result
}

func generateTaskAnalysis(tasks []Task, dependencies []TaskDependency, notes []Note, tils []TIL) TaskAnalysis {
	analysis := TaskAnalysis{
		Dependencies:    dependencies,
		CompletionTimes: make(map[string]float64),
	}

	// Build task hierarchy
	analysis.TaskHierarchy = buildTaskHierarchy(tasks, notes, tils)

	// Calculate completion times
	for _, task := range tasks {
		if task.Status == TaskStatusCompleted {
			duration := task.UpdatedAt.Sub(task.CreatedAt)
			analysis.CompletionTimes[task.ID] = duration.Hours()
		}
	}

	// Identify blocking patterns
	analysis.BlockingPatterns = identifyBlockingPatterns(tasks, dependencies)

	// Generate velocity trends
	analysis.VelocityTrends = generateVelocityTrends(tasks)

	return analysis
}

func buildTaskHierarchy(tasks []Task, notes []Note, tils []TIL) []TaskNode {
	taskMap := make(map[string]*TaskNode)
	notesByTask := make(map[string][]Note)
	tilsByTask := make(map[string][]TIL)

	// Group notes and TILs by task
	for _, note := range notes {
		notesByTask[note.TaskID] = append(notesByTask[note.TaskID], note)
	}
	for _, til := range tils {
		if til.TaskID != nil {
			tilsByTask[*til.TaskID] = append(tilsByTask[*til.TaskID], til)
		}
	}

	// Create nodes for all tasks
	for _, task := range tasks {
		taskMap[task.ID] = &TaskNode{
			Task:     task,
			Notes:    notesByTask[task.ID],
			TILs:     tilsByTask[task.ID],
			Children: []TaskNode{},
		}
	}

	// Build hierarchy
	var rootNodes []TaskNode
	for _, task := range tasks {
		node := taskMap[task.ID]
		if task.ParentID == nil {
			rootNodes = append(rootNodes, *node)
		} else {
			if parent, ok := taskMap[*task.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	return rootNodes
}

func identifyBlockingPatterns(tasks []Task, dependencies []TaskDependency) []string {
	var patterns []string

	// Look for tasks with many dependencies
	depCount := make(map[string]int)
	for _, dep := range dependencies {
		depCount[dep.TaskID]++
	}

	for taskID, count := range depCount {
		if count > 3 {
			patterns = append(patterns, fmt.Sprintf("Task %s has %d dependencies", taskID, count))
		}
	}

	// Look for long-running pending tasks
	for _, task := range tasks {
		if task.Status == TaskStatusPending {
			age := time.Since(task.CreatedAt)
			if age > 7*24*time.Hour {
				patterns = append(patterns, fmt.Sprintf("Task %s pending for %s", task.ID, age.Round(time.Hour)))
			}
		}
	}

	return patterns
}

func generateVelocityTrends(tasks []Task) []VelocityDataPoint {
	dailyStats := make(map[string]*VelocityDataPoint)

	for _, task := range tasks {
		createDate := task.CreatedAt.Format("2006-01-02")
		if _, ok := dailyStats[createDate]; !ok {
			date, _ := time.Parse("2006-01-02", createDate)
			dailyStats[createDate] = &VelocityDataPoint{Date: date}
		}
		dailyStats[createDate].TasksCreated++

		if task.Status == TaskStatusCompleted {
			completeDate := task.UpdatedAt.Format("2006-01-02")
			if _, ok := dailyStats[completeDate]; !ok {
				date, _ := time.Parse("2006-01-02", completeDate)
				dailyStats[completeDate] = &VelocityDataPoint{Date: date}
			}
			dailyStats[completeDate].TasksCompleted++
		}
	}

	var trends []VelocityDataPoint
	for _, point := range dailyStats {
		trends = append(trends, *point)
	}

	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Date.Before(trends[j].Date)
	})

	return trends
}

func generateKnowledgeDocumentation(notes []Note, tils []TIL, tasks []Task, agents []Agent) KnowledgeDocumentation {
	doc := KnowledgeDocumentation{
		NotesByTask:  make(map[string][]Note),
		NotesByAgent: make(map[string][]Note),
		TILsByTask:   make(map[string][]TIL),
		TILsByAgent:  make(map[string][]TIL),
	}

	// Group notes
	for _, note := range notes {
		doc.NotesByTask[note.TaskID] = append(doc.NotesByTask[note.TaskID], note)
		doc.NotesByAgent[note.AgentID] = append(doc.NotesByAgent[note.AgentID], note)
	}

	// Group TILs
	for _, til := range tils {
		if til.TaskID != nil {
			doc.TILsByTask[*til.TaskID] = append(doc.TILsByTask[*til.TaskID], til)
		}
		doc.TILsByAgent[til.AgentID] = append(doc.TILsByAgent[til.AgentID], til)
	}

	// Identify top learnings (most recent TILs)
	sortedTILs := make([]TIL, len(tils))
	copy(sortedTILs, tils)
	sort.Slice(sortedTILs, func(i, j int) bool {
		return sortedTILs[i].CreatedAt.After(sortedTILs[j].CreatedAt)
	})

	maxTop := 10
	if len(sortedTILs) < maxTop {
		maxTop = len(sortedTILs)
	}
	doc.TopLearnings = sortedTILs[:maxTop]

	return doc
}

func generateInsights(tasks []Task, agents []Agent, notes []Note, tils []TIL, summary ProjectSummary, contributions []AgentContribution) ProjectInsights {
	insights := ProjectInsights{
		CommonBlockers: []string{},
	}

	// Most productive agent
	if len(contributions) > 0 && contributions[0].TasksCompleted > 0 {
		insights.MostProductiveAgent = &contributions[0].Agent
	}

	// Longest running task
	var longestTask *Task
	var maxDuration time.Duration
	for _, task := range tasks {
		duration := time.Since(task.CreatedAt)
		if task.Status != TaskStatusCompleted && duration > maxDuration {
			maxDuration = duration
			longestTask = &task
		}
	}
	insights.LongestRunningTask = longestTask

	// Most collaborative task (most notes)
	taskNoteCount := make(map[string]int)
	for _, note := range notes {
		taskNoteCount[note.TaskID]++
	}

	var mostCollaborativeTask *Task
	maxNotes := 0
	for _, task := range tasks {
		if count := taskNoteCount[task.ID]; count > maxNotes {
			maxNotes = count
			mostCollaborativeTask = &task
		}
	}
	insights.MostCollaborativeTask = mostCollaborativeTask

	// Knowledge sharing rate
	if summary.CompletedTasks > 0 {
		insights.KnowledgeSharing = float64(summary.TotalTILs) / float64(summary.CompletedTasks)
	}

	// Task velocity (tasks completed per day)
	if summary.ProjectDuration.Hours() > 0 {
		days := summary.ProjectDuration.Hours() / 24
		insights.TaskVelocity = float64(summary.CompletedTasks) / days
	}

	// Agent utilization
	activeAgents := 0
	for _, contrib := range contributions {
		if contrib.TasksCompleted > 0 || contrib.TasksPending > 0 {
			activeAgents++
		}
	}
	if len(agents) > 0 {
		insights.AgentUtilization = float64(activeAgents) / float64(len(agents)) * 100
	}

	return insights
}

func formatReportMarkdown(report *ProjectReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# 📋 Project Report: %s\n\n", report.Project.Name))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	if report.Project.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", report.Project.Description))
	}

	if report.Project.Guidelines != "" {
		sb.WriteString(fmt.Sprintf("**Guidelines:** %s\n\n", report.Project.Guidelines))
	}

	// Project Overview
	sb.WriteString("## 📊 Project Overview\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Total Tasks** | %d |\n", report.Summary.TotalTasks))
	sb.WriteString(fmt.Sprintf("| **Completed Tasks** | %d |\n", report.Summary.CompletedTasks))
	sb.WriteString(fmt.Sprintf("| **Pending Tasks** | %d |\n", report.Summary.PendingTasks))
	sb.WriteString(fmt.Sprintf("| **Failed Tasks** | %d |\n", report.Summary.FailedTasks))
	sb.WriteString(fmt.Sprintf("| **Completion Rate** | %.1f%% |\n", report.Summary.CompletionRate))
	sb.WriteString(fmt.Sprintf("| **Total Agents** | %d |\n", report.Summary.TotalAgents))
	sb.WriteString(fmt.Sprintf("| **Notes Written** | %d |\n", report.Summary.TotalNotes))
	sb.WriteString(fmt.Sprintf("| **TILs Shared** | %d |\n", report.Summary.TotalTILs))
	sb.WriteString(fmt.Sprintf("| **Project Duration** | %s |\n", report.Summary.ProjectDuration.Round(time.Hour)))
	sb.WriteString(fmt.Sprintf("| **Avg Task Duration** | %s |\n", report.Summary.AvgTaskDuration.Round(time.Hour)))
	sb.WriteString("\n")

	// Timeline
	sb.WriteString("## ⏰ Project Timeline\n\n")
	for i, event := range report.Timeline {
		if i >= 20 { // Limit timeline display
			sb.WriteString(fmt.Sprintf("... and %d more events\n\n", len(report.Timeline)-20))
			break
		}

		icon := getEventIcon(event.Type)
		agentInfo := ""
		if event.AgentName != nil {
			agentInfo = fmt.Sprintf(" (by %s)", *event.AgentName)
		}

		sb.WriteString(fmt.Sprintf("- **%s** %s %s%s\n",
			event.Time.Format("2006-01-02 15:04"),
			icon,
			event.Description,
			agentInfo))
	}
	sb.WriteString("\n")

	// Agent Contributions
	sb.WriteString("## 👥 Agent Contributions\n\n")
	if len(report.AgentContributions) > 0 {
		sb.WriteString("| Agent | Completed | Failed | Pending | Success Rate | Notes | TILs | Avg Time |\n")
		sb.WriteString("|-------|-----------|--------|---------|--------------|-------|------|----------|\n")

		for _, contrib := range report.AgentContributions {
			avgTime := "N/A"
			if contrib.AvgCompletionTime > 0 {
				avgTime = contrib.AvgCompletionTime.Round(time.Hour).String()
			}

			sb.WriteString(fmt.Sprintf("| **%s** | %d | %d | %d | %.1f%% | %d | %d | %s |\n",
				contrib.Agent.Name,
				contrib.TasksCompleted,
				contrib.TasksFailed,
				contrib.TasksPending,
				contrib.SuccessRate,
				contrib.NotesWritten,
				contrib.TILsShared,
				avgTime))
		}
	} else {
		sb.WriteString("*No agent activity recorded.*\n")
	}
	sb.WriteString("\n")

	// Task Analysis
	sb.WriteString("## 📈 Task Analysis\n\n")
	sb.WriteString("### Task Hierarchy\n\n")
	formatTaskHierarchy(&sb, report.TaskAnalysis.TaskHierarchy, 0)
	sb.WriteString("\n")

	if len(report.TaskAnalysis.BlockingPatterns) > 0 {
		sb.WriteString("### 🚫 Blocking Patterns\n\n")
		for _, pattern := range report.TaskAnalysis.BlockingPatterns {
			sb.WriteString(fmt.Sprintf("- %s\n", pattern))
		}
		sb.WriteString("\n")
	}

	// Knowledge Documentation
	sb.WriteString("## 📚 Knowledge Documentation\n\n")
	sb.WriteString(fmt.Sprintf("**Total Notes:** %d across %d tasks\n\n",
		report.Summary.TotalNotes, len(report.Knowledge.NotesByTask)))
	sb.WriteString(fmt.Sprintf("**Total TILs:** %d across %d tasks\n\n",
		report.Summary.TotalTILs, len(report.Knowledge.TILsByTask)))

	if len(report.Knowledge.TopLearnings) > 0 {
		sb.WriteString("### 🎓 Recent Learnings\n\n")
		for _, til := range report.Knowledge.TopLearnings[:min(5, len(report.Knowledge.TopLearnings))] {
			sb.WriteString(fmt.Sprintf("**%s** (%s)\n", til.Title, til.CreatedAt.Format("2006-01-02")))
			sb.WriteString(fmt.Sprintf("> %s\n\n", til.Content))
		}
	}

	// Insights & Analytics
	sb.WriteString("## 💡 Insights & Analytics\n\n")

	if report.Insights.MostProductiveAgent != nil {
		sb.WriteString(fmt.Sprintf("**🏆 Most Productive Agent:** %s\n\n", report.Insights.MostProductiveAgent.Name))
	}

	if report.Insights.LongestRunningTask != nil {
		duration := time.Since(report.Insights.LongestRunningTask.CreatedAt)
		sb.WriteString(fmt.Sprintf("**⏳ Longest Running Task:** %s (running for %s)\n\n",
			report.Insights.LongestRunningTask.Title, duration.Round(time.Hour)))
	}

	if report.Insights.MostCollaborativeTask != nil {
		sb.WriteString(fmt.Sprintf("**🤝 Most Collaborative Task:** %s\n\n", report.Insights.MostCollaborativeTask.Title))
	}

	sb.WriteString("### Key Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Knowledge Sharing Rate** | %.2f TILs per completed task |\n", report.Insights.KnowledgeSharing))
	sb.WriteString(fmt.Sprintf("| **Task Velocity** | %.2f tasks per day |\n", report.Insights.TaskVelocity))
	sb.WriteString(fmt.Sprintf("| **Agent Utilization** | %.1f%% |\n", report.Insights.AgentUtilization))
	sb.WriteString("\n")

	// Velocity Chart (Simple ASCII)
	if len(report.TaskAnalysis.VelocityTrends) > 0 {
		sb.WriteString("### 📊 Velocity Trends\n\n")
		sb.WriteString("```\n")
		for _, point := range report.TaskAnalysis.VelocityTrends {
			bars := strings.Repeat("█", point.TasksCompleted)
			if bars == "" {
				bars = "·"
			}
			sb.WriteString(fmt.Sprintf("%s │ %s (%d completed, %d created)\n",
				point.Date.Format("01-02"), bars, point.TasksCompleted, point.TasksCreated))
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("---\n")
	sb.WriteString("*Report generated by amp-tasks*\n")

	return sb.String()
}

func formatReportText(report *ProjectReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PROJECT REPORT: %s\n", strings.ToUpper(report.Project.Name)))
	sb.WriteString(strings.Repeat("=", 80) + "\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	// Summary
	sb.WriteString("PROJECT SUMMARY\n")
	sb.WriteString(strings.Repeat("-", 20) + "\n")
	sb.WriteString(fmt.Sprintf("Tasks: %d total (%d completed, %d pending, %d failed)\n",
		report.Summary.TotalTasks, report.Summary.CompletedTasks,
		report.Summary.PendingTasks, report.Summary.FailedTasks))
	sb.WriteString(fmt.Sprintf("Completion Rate: %.1f%%\n", report.Summary.CompletionRate))
	sb.WriteString(fmt.Sprintf("Agents: %d\n", report.Summary.TotalAgents))
	sb.WriteString(fmt.Sprintf("Notes: %d\n", report.Summary.TotalNotes))
	sb.WriteString(fmt.Sprintf("TILs: %d\n", report.Summary.TotalTILs))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", report.Summary.ProjectDuration.Round(time.Hour)))
	sb.WriteString("\n")

	// Agent contributions
	sb.WriteString("AGENT CONTRIBUTIONS\n")
	sb.WriteString(strings.Repeat("-", 20) + "\n")
	for _, contrib := range report.AgentContributions {
		if contrib.TasksCompleted > 0 || contrib.TasksPending > 0 {
			sb.WriteString(fmt.Sprintf("%s: %d completed, %.1f%% success rate\n",
				contrib.Agent.Name, contrib.TasksCompleted, contrib.SuccessRate))
		}
	}
	sb.WriteString("\n")

	// Key insights
	sb.WriteString("KEY INSIGHTS\n")
	sb.WriteString(strings.Repeat("-", 20) + "\n")
	if report.Insights.MostProductiveAgent != nil {
		sb.WriteString(fmt.Sprintf("Most productive: %s\n", report.Insights.MostProductiveAgent.Name))
	}
	sb.WriteString(fmt.Sprintf("Task velocity: %.2f tasks/day\n", report.Insights.TaskVelocity))
	sb.WriteString(fmt.Sprintf("Knowledge sharing: %.2f TILs per task\n", report.Insights.KnowledgeSharing))
	sb.WriteString(fmt.Sprintf("Agent utilization: %.1f%%\n", report.Insights.AgentUtilization))

	return sb.String()
}

func formatTaskHierarchy(sb *strings.Builder, nodes []TaskNode, depth int) {
	for _, node := range nodes {
		indent := strings.Repeat("  ", depth)
		status := getTaskStatusEmoji(node.Task.Status)

		sb.WriteString(fmt.Sprintf("%s- %s **%s** (%s)\n",
			indent, status, node.Task.Title, string(node.Task.Status)))

		if len(node.Notes) > 0 {
			sb.WriteString(fmt.Sprintf("%s  📝 %d notes\n", indent, len(node.Notes)))
		}
		if len(node.TILs) > 0 {
			sb.WriteString(fmt.Sprintf("%s  🎓 %d TILs\n", indent, len(node.TILs)))
		}

		if len(node.Children) > 0 {
			formatTaskHierarchy(sb, node.Children, depth+1)
		}
	}
}

func getEventIcon(eventType string) string {
	switch eventType {
	case "task_created":
		return "📋"
	case "task_completed":
		return "✅"
	case "task_failed":
		return "❌"
	case "task_assigned":
		return "👤"
	case "note_added":
		return "📝"
	case "til_created":
		return "🎓"
	default:
		return "•"
	}
}

func getTaskStatusEmoji(status TaskStatus) string {
	switch status {
	case TaskStatusCompleted:
		return "✅"
	case TaskStatusFailed:
		return "❌"
	case TaskStatusInProgress:
		return "🔄"
	case TaskStatusPending:
		return "⏳"
	default:
		return "❓"
	}
}



func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringP("format", "f", "markdown", "Output format (markdown, text, json)")
	reportCmd.Flags().StringP("output", "o", "", "Output file path")
	reportCmd.Flags().BoolP("verbose", "v", false, "Include all details")
	reportCmd.Flags().BoolP("summary", "s", false, "Generate condensed version")
}
