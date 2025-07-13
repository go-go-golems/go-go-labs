package agent

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/mock-agent/client"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/mock-agent/scenarios"
)

// AgentState represents the current state of the agent
type AgentState string

const (
	StateIdle            AgentState = "idle"
	StateStarting        AgentState = "starting"
	StateActive          AgentState = "active"
	StateWaitingFeedback AgentState = "waiting_feedback"
	StateError           AgentState = "error"
	StateShuttingDown    AgentState = "shutting_down"
	StateFinished        AgentState = "finished"
)

// Config holds agent configuration
type Config struct {
	Name                 string
	Worktree             string
	Randomized           bool
	TickInterval         time.Duration
	CommandCheckInterval time.Duration
}

// Agent represents a mock agent instance
type Agent struct {
	id       string
	config   Config
	client   *client.Client
	state    AgentState
	scenario *scenarios.Scenario

	// Agent metrics
	filesChanged int
	linesAdded   int
	linesRemoved int
	progress     int
	currentTask  string

	// Internal state
	workStartTime   time.Time
	lastCommitTime  time.Time
	questionPosted  bool
	pendingQuestion string
	registered      bool
	todos           []models.TodoItem

	// Timing
	lastTick         time.Time
	lastCommandCheck time.Time

	// Randomization state
	nextStateChange        time.Time
	stateChangeProbability float64
}

// New creates a new mock agent
func New(apiClient *client.Client, config Config) *Agent {
	return &Agent{
		config:                 config,
		client:                 apiClient,
		state:                  StateStarting,
		stateChangeProbability: 0.3, // 30% chance of random state change per tick
		nextStateChange:        time.Now().Add(time.Duration(rand.Intn(60)) * time.Second),
	}
}

// Run starts the agent main loop
func (a *Agent) Run(ctx context.Context) error {
	log.Info().Str("name", a.config.Name).Msg("Agent starting up")

	// Register agent with the backend
	if err := a.register(); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	// Start main loop
	tickTicker := time.NewTicker(a.config.TickInterval)
	commandTicker := time.NewTicker(a.config.CommandCheckInterval)
	defer tickTicker.Stop()
	defer commandTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("agent", a.id).Msg("Agent received shutdown signal")
			return a.shutdown()
		case <-tickTicker.C:
			if err := a.tick(); err != nil {
				log.Error().Err(err).Str("agent", a.id).Msg("Error during agent tick")
				a.handleError(err)
			}
		case <-commandTicker.C:
			if err := a.checkCommands(); err != nil {
				log.Error().Err(err).Str("agent", a.id).Msg("Error checking commands")
			}
		}
	}
}

// register registers the agent with the backend
func (a *Agent) register() error {
	req := models.CreateAgentRequest{
		Name:     a.config.Name,
		Worktree: a.config.Worktree,
	}

	agent, err := a.client.CreateAgent(req)
	if err != nil {
		return err
	}

	a.id = agent.ID
	a.registered = true

	log.Info().Str("agent", a.id).Str("name", a.config.Name).Msg("Agent registered successfully")

	// Log startup event
	_, err = a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeStart),
		Message: fmt.Sprintf("Agent %s started up", a.config.Name),
		Metadata: map[string]interface{}{
			"worktree":   a.config.Worktree,
			"randomized": a.config.Randomized,
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log startup event")
	}

	// Initialize with a scenario if randomized mode is enabled
	if a.config.Randomized {
		a.selectNewScenario()
		if a.scenario != nil {
			log.Info().Str("agent", a.id).Str("scenario", a.scenario.Name).Msg("Starting initial scenario")
			return a.startScenario()
		}
	}

	a.state = StateIdle
	return a.updateStatus()
}

// tick performs one agent tick (main logic loop)
func (a *Agent) tick() error {
	a.lastTick = time.Now()

	// Handle randomized behavior
	if a.config.Randomized {
		a.handleRandomizedBehavior()
	}

	// Execute current state logic
	switch a.state {
	case StateIdle:
		return a.handleIdleState()
	case StateActive:
		return a.handleActiveState()
	case StateWaitingFeedback:
		return a.handleWaitingFeedbackState()
	case StateError:
		return a.handleErrorState()
	case StateShuttingDown, StateFinished:
		// Do nothing for shutdown/finished states
		return nil
	default:
		return nil
	}
}

// handleIdleState handles logic when agent is idle
func (a *Agent) handleIdleState() error {
	log.Info().Str("agent", a.id).
		Bool("randomized", a.config.Randomized).
		Bool("has_scenario", a.scenario != nil).
		Msg("Agent is idle")

	// Check for available tasks
	tasks, err := a.client.ListTasks("pending", "", 10, 0)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Pick up a task if available
	taskRoll := rand.Float64()
	log.Info().Str("agent", a.id).
		Int("available_tasks", len(tasks)).
		Float64("task_roll", taskRoll).
		Msg("Checking for available tasks")

	if len(tasks) > 0 && taskRoll < 0.3 { // 30% chance to pick up task
		task := tasks[rand.Intn(len(tasks))]
		log.Info().Str("agent", a.id).Str("task", task.Title).Msg("Picking up assigned task")
		return a.startTask(&task)
	}

	// Start a random scenario if in randomized mode
	scenarioRoll := rand.Float64()
	log.Info().Str("agent", a.id).
		Bool("randomized", a.config.Randomized).
		Bool("no_scenario", a.scenario == nil).
		Float64("scenario_roll", scenarioRoll).
		Msg("Checking scenario conditions")

	if a.config.Randomized && scenarioRoll < 0.7 {
		log.Info().Str("agent", a.id).Msg("Starting random scenario selection")
		a.selectNewScenario()
		if a.scenario != nil {
			log.Info().Str("agent", a.id).Str("scenario", a.scenario.Name).Msg("Starting scenario execution")
			return a.startScenario()
		} else {
			log.Warn().Str("agent", a.id).Msg("No scenario selected despite calling selectNewScenario")
		}
	} else {
		log.Info().Str("agent", a.id).Msg("Scenario conditions not met - staying idle")
	}

	return nil
}

// handleActiveState handles logic when agent is actively working
func (a *Agent) handleActiveState() error {
	if a.scenario != nil {
		return a.executeScenario()
	}

	// Default active behavior - simulate work progress
	elapsed := time.Since(a.workStartTime)

	// Update progress based on elapsed time
	totalDuration := 5 * time.Minute // Assume 5 minute tasks
	a.progress = int((elapsed.Seconds() / totalDuration.Seconds()) * 100)
	if a.progress > 100 {
		a.progress = 100
	}

	// Simulate file changes
	if rand.Float64() < 0.3 { // 30% chance per tick
		a.filesChanged += rand.Intn(3) + 1
		a.linesAdded += rand.Intn(50) + 5
		a.linesRemoved += rand.Intn(20) + 1
	}

	// Complete task if progress reaches 100%
	if a.progress >= 100 {
		return a.completeCurrentWork()
	}

	// Random chance to ask question
	if !a.questionPosted && rand.Float64() < 0.05 { // 5% chance per tick
		return a.askQuestion()
	}

	// Commit periodically
	if time.Since(a.lastCommitTime) > 2*time.Minute && rand.Float64() < 0.2 {
		return a.makeCommit()
	}

	return a.updateStatus()
}

// handleWaitingFeedbackState handles logic when agent is waiting for feedback
func (a *Agent) handleWaitingFeedbackState() error {
	log.Info().Str("agent", a.id).Msg("Agent waiting for feedback")

	// If we've been waiting too long, assume no response and continue
	if a.questionPosted && time.Since(a.workStartTime) > 10*time.Minute {
		log.Info().Str("agent", a.id).Msg("No response received, continuing work")
		a.questionPosted = false
		a.pendingQuestion = ""
		a.state = StateActive
		return a.updateStatus()
	}

	return nil
}

// handleErrorState handles logic when agent is in error state
func (a *Agent) handleErrorState() error {
	log.Info().Str("agent", a.id).Msg("Agent in error state, attempting recovery")

	// Random recovery after some time
	if rand.Float64() < 0.9 { // 10% chance per tick
		log.Info().Str("agent", a.id).Msg("Agent recovered from error")
		a.state = StateIdle
		a.progress = 0

		_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
			Type:    string(models.EventTypeInfo),
			Message: "Agent recovered from error state",
		})

		if err != nil {
			log.Warn().Err(err).Msg("Failed to log recovery event")
		}

		return a.updateStatus()
	}

	return nil
}

// startTask starts working on a task
func (a *Agent) startTask(task *models.Task) error {
	log.Info().Str("agent", a.id).Str("task", task.Title).Msg("Starting task")

	// Update task to assigned status
	status := string(models.TaskStatusInProgress)
	_, err := a.client.UpdateTask(task.ID, models.UpdateTaskRequest{
		Status:          &status,
		AssignedAgentID: &a.id,
	})
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	a.currentTask = task.Title
	a.workStartTime = time.Now()
	a.progress = 0
	a.state = StateActive

	// Create todos for this task
	todos := generateTodosForTask(task)
	for i, todoText := range todos {
		_, err := a.client.CreateTodo(a.id, models.CreateTodoRequest{
			Text:  todoText,
			Order: i + 1,
		})
		if err != nil {
			log.Warn().Err(err).Str("todo", todoText).Msg("Failed to create todo")
		}
	}

	// Log task start event
	_, err = a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: fmt.Sprintf("Started working on task: %s", task.Title),
		Metadata: map[string]interface{}{
			"task_id":    task.ID,
			"task_title": task.Title,
			"priority":   task.Priority,
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log task start event")
	}

	return a.updateStatus()
}

// completeCurrentWork completes the current work
func (a *Agent) completeCurrentWork() error {
	log.Info().Str("agent", a.id).Str("task", a.currentTask).Msg("Completing current work")

	// Mark all todos as completed
	todos, err := a.client.ListTodos(a.id)
	if err == nil {
		for _, todo := range todos {
			if !todo.Completed {
				completed := true
				_, err := a.client.UpdateTodo(a.id, todo.ID, models.UpdateTodoRequest{
					Completed: &completed,
				})
				if err != nil {
					log.Warn().Err(err).Str("todo", todo.ID).Msg("Failed to complete todo")
				}
			}
		}
	}

	// Log completion event
	_, err = a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeSuccess),
		Message: fmt.Sprintf("Completed work: %s", a.currentTask),
		Metadata: map[string]interface{}{
			"task":          a.currentTask,
			"files_changed": a.filesChanged,
			"lines_added":   a.linesAdded,
			"lines_removed": a.linesRemoved,
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log completion event")
	}

	// Reset state
	a.currentTask = ""
	a.progress = 0
	a.questionPosted = false
	a.pendingQuestion = ""
	a.state = StateIdle
	a.scenario = nil

	return a.updateStatus()
}

// askQuestion posts a question and waits for feedback
func (a *Agent) askQuestion() error {
	questions := []string{
		"Should I continue with the current approach or try a different strategy?",
		"I found a potential security issue. How should I handle it?",
		"The current implementation conflicts with existing code. What's the preferred solution?",
		"Should I add more comprehensive error handling here?",
		"I notice this could be optimized. Should I prioritize performance or readability?",
		"There are multiple ways to implement this feature. Which approach do you prefer?",
		"I found some deprecated dependencies. Should I upgrade them now?",
		"The test coverage is low in this area. Should I add more tests before proceeding?",
	}

	question := questions[rand.Intn(len(questions))]
	a.pendingQuestion = question
	a.questionPosted = true
	a.state = StateWaitingFeedback

	log.Info().Str("agent", a.id).Str("question", question).Msg("Agent asking question")

	// Log question event
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeQuestion),
		Message: question,
		Metadata: map[string]interface{}{
			"context":  a.currentTask,
			"progress": a.progress,
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log question event")
	}

	return a.updateStatus()
}

// makeCommit simulates making a commit
func (a *Agent) makeCommit() error {
	commitMessages := []string{
		"Fix bug in authentication module",
		"Add new feature for user management",
		"Improve error handling",
		"Update documentation",
		"Refactor code for better performance",
		"Add unit tests",
		"Fix security vulnerability",
		"Optimize database queries",
		"Update dependencies",
		"Improve logging",
	}

	message := commitMessages[rand.Intn(len(commitMessages))]
	a.lastCommitTime = time.Now()

	log.Info().Str("agent", a.id).Str("message", message).Msg("Making commit")

	// Log commit event
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeCommit),
		Message: fmt.Sprintf("Committed: %s", message),
		Metadata: map[string]interface{}{
			"commit_message": message,
			"files_changed":  rand.Intn(5) + 1,
			"lines_added":    rand.Intn(100) + 10,
			"lines_removed":  rand.Intn(50) + 1,
		},
	})

	return err
}

// generateTodosForTask generates realistic todos for a given task
func generateTodosForTask(task *models.Task) []string {
	baseTemplates := [][]string{
		{
			"Analyze requirements and create implementation plan",
			"Set up project structure and dependencies",
			"Implement core functionality",
			"Add error handling and validation",
			"Write unit tests",
			"Update documentation",
			"Perform code review and cleanup",
		},
		{
			"Review existing codebase for integration points",
			"Design API interfaces",
			"Implement backend logic",
			"Create frontend components",
			"Integrate with existing systems",
			"Test end-to-end functionality",
			"Deploy and monitor",
		},
		{
			"Identify and reproduce the issue",
			"Analyze root cause",
			"Design fix strategy",
			"Implement fix",
			"Test fix thoroughly",
			"Update related documentation",
			"Deploy fix to production",
		},
	}

	template := baseTemplates[rand.Intn(len(baseTemplates))]

	// Customize todos based on task priority
	if strings.Contains(strings.ToLower(task.Priority), "urgent") {
		template = append([]string{"URGENT: Prioritize immediate fix"}, template...)
	}

	// Add task-specific context
	for i, todo := range template {
		if strings.Contains(strings.ToLower(task.Title), "security") {
			template[i] = strings.Replace(todo, "functionality", "security features", 1)
		} else if strings.Contains(strings.ToLower(task.Title), "performance") {
			template[i] = strings.Replace(todo, "functionality", "performance optimizations", 1)
		}
	}

	return template
}
