package agent

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/mock-agent/scenarios"
)

// handleRandomizedBehavior introduces randomized state changes and behaviors
func (a *Agent) handleRandomizedBehavior() {
	now := time.Now()
	
	// Check if it's time for a potential state change
	if now.After(a.nextStateChange) {
		a.maybeChangeState()
		a.scheduleNextStateChange()
	}
	
	// Random events
	a.maybeGenerateRandomEvent()
	
	// Random todo updates
	a.maybeUpdateTodos()
	
	// Random metric fluctuations
	a.maybeUpdateMetrics()
}

// maybeChangeState randomly changes agent state
func (a *Agent) maybeChangeState() {
	if rand.Float64() > a.stateChangeProbability {
		return
	}
	
	// Don't randomly change state if we're actively working on a scenario
	if a.scenario != nil && a.state == StateActive {
		log.Debug().Str("agent", a.id).Msg("Skipping random state change - actively working on scenario")
		return
	}
	
	oldState := a.state
	newState := a.selectRandomState()
	
	if newState == oldState {
		return
	}
	
	log.Info().
		Str("agent", a.id).
		Str("old_state", string(oldState)).
		Str("new_state", string(newState)).
		Msg("Random state change")
	
	a.state = newState
	
	// Handle state-specific logic
	switch newState {
	case StateError:
		a.handleRandomError()
	case StateWaitingFeedback:
		if !a.questionPosted {
			a.askRandomQuestion()
		}
	case StateActive:
		if a.currentTask == "" {
			a.startRandomWork()
		}
	case StateIdle:
		a.currentTask = ""
		a.progress = 0
		a.questionPosted = false
		a.pendingQuestion = ""
	}
	
	a.updateStatus()
}

// selectRandomState selects a random valid state based on current state
func (a *Agent) selectRandomState() AgentState {
	possibleStates := []AgentState{}
	
	switch a.state {
	case StateIdle:
		possibleStates = []AgentState{StateActive, StateError}
	case StateActive:
		possibleStates = []AgentState{StateIdle, StateWaitingFeedback, StateError}
	case StateWaitingFeedback:
		possibleStates = []AgentState{StateActive, StateIdle, StateError}
	case StateError:
		possibleStates = []AgentState{StateIdle, StateActive}
	default:
		possibleStates = []AgentState{StateIdle}
	}
	
	if len(possibleStates) == 0 {
		return StateIdle
	}
	
	return possibleStates[rand.Intn(len(possibleStates))]
}

// scheduleNextStateChange schedules the next potential state change
func (a *Agent) scheduleNextStateChange() {
	// Random interval between 30 seconds and 5 minutes
	interval := time.Duration(rand.Intn(270)+30) * time.Second
	a.nextStateChange = time.Now().Add(interval)
}

// handleRandomError simulates entering an error state
func (a *Agent) handleRandomError() error {
	errors := []string{
		"Network connection timeout",
		"Dependency conflict detected",
		"Build failure in CI pipeline",
		"Permission denied on file operation",
		"Database connection lost",
		"Memory allocation error",
		"Invalid configuration detected",
		"Rate limit exceeded on API call",
		"Disk space insufficient",
		"Authentication token expired",
	}
	
	errorMsg := errors[rand.Intn(len(errors))]
	log.Warn().Str("agent", a.id).Str("error", errorMsg).Msg("Random error occurred")
	
	// Log error event
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeError),
		Message: errorMsg,
		Metadata: map[string]interface{}{
			"error_type": "random_simulation",
			"context":    a.currentTask,
			"progress":   a.progress,
		},
	})
	
	return err
}

// askRandomQuestion poses a random question
func (a *Agent) askRandomQuestion() error {
	questions := []string{
		"I've encountered an unexpected edge case. How should I handle it?",
		"The current implementation uses a deprecated API. Should I upgrade?",
		"I found inconsistent coding styles in the codebase. Should I standardize them?",
		"There's a trade-off between performance and memory usage here. What's preferred?",
		"I notice this functionality could be extracted into a reusable component. Proceed?",
		"The external service is returning unexpected data format. Should I add validation?",
		"I can implement this feature in two ways - simple or more flexible. Which approach?",
		"Found potential security vulnerability in third-party dependency. How to proceed?",
	}
	
	question := questions[rand.Intn(len(questions))]
	a.pendingQuestion = question
	a.questionPosted = true
	
	log.Info().Str("agent", a.id).Str("question", question).Msg("Random question posted")
	
	// Log question event
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeQuestion),
		Message: question,
		Metadata: map[string]interface{}{
			"question_type": "random_inquiry",
			"context":       a.currentTask,
			"progress":      a.progress,
		},
	})
	
	return err
}

// startRandomWork starts working on a random task
func (a *Agent) startRandomWork() {
	tasks := []string{
		"Refactoring legacy authentication module",
		"Implementing user preference system",
		"Optimizing database query performance",
		"Adding comprehensive error logging",
		"Creating automated deployment pipeline",
		"Updating documentation and examples",
		"Fixing memory leaks in background processes",
		"Implementing data validation framework",
		"Adding internationalization support",
		"Enhancing security with rate limiting",
	}
	
	task := tasks[rand.Intn(len(tasks))]
	a.currentTask = task
	a.workStartTime = time.Now()
	a.progress = rand.Intn(30) // Start with some random progress
	
	log.Info().Str("agent", a.id).Str("task", task).Msg("Started random work")
	
	// Create random todos
	a.createRandomTodos()
	
	// Log work start
	a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: "Started working on: " + task,
		Metadata: map[string]interface{}{
			"work_type":     "random_task",
			"initial_progress": a.progress,
		},
	})
}

// createRandomTodos creates random todo items for current work
func (a *Agent) createRandomTodos() {
	todoTemplates := []string{
		"Analyze current implementation and identify issues",
		"Research best practices and alternative approaches",
		"Create implementation plan and timeline",
		"Set up development environment and dependencies",
		"Implement core functionality with proper error handling",
		"Add comprehensive unit and integration tests",
		"Perform code review and optimization",
		"Update documentation and create examples",
		"Deploy to staging environment for testing",
		"Gather feedback and make final adjustments",
	}
	
	// Create 3-7 random todos
	numTodos := rand.Intn(5) + 3
	for i := 0; i < numTodos; i++ {
		todoText := todoTemplates[rand.Intn(len(todoTemplates))]
		a.client.CreateTodo(a.id, models.CreateTodoRequest{
			Text:  todoText,
			Order: i + 1,
		})
	}
}

// maybeGenerateRandomEvent occasionally generates random events
func (a *Agent) maybeGenerateRandomEvent() {
	if rand.Float64() > 0.02 { // 2% chance per tick
		return
	}
	
	eventTypes := []string{
		string(models.EventTypeInfo),
		string(models.EventTypeCommit),
		string(models.EventTypeSuccess),
	}
	
	messages := map[string][]string{
		string(models.EventTypeInfo): {
			"Reviewed code changes and found optimization opportunities",
			"Updated dependencies to latest stable versions",
			"Reorganized project structure for better maintainability",
			"Added automated code quality checks to pipeline",
			"Improved error messages for better debugging",
		},
		string(models.EventTypeCommit): {
			"Refactor: simplify authentication logic",
			"Feature: add user session management",
			"Fix: resolve memory leak in background task",
			"Update: improve API response times",
			"Security: patch vulnerability in validation",
		},
		string(models.EventTypeSuccess): {
			"Successfully completed performance optimization",
			"All integration tests are now passing",
			"Successfully deployed to staging environment",
			"Code coverage increased to 95%",
			"Performance benchmarks show 40% improvement",
		},
	}
	
	eventType := eventTypes[rand.Intn(len(eventTypes))]
	messageList := messages[eventType]
	message := messageList[rand.Intn(len(messageList))]
	
	a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    eventType,
		Message: message,
		Metadata: map[string]interface{}{
			"event_source": "random_generation",
			"context":      a.currentTask,
		},
	})
	
	log.Debug().Str("agent", a.id).Str("event", eventType).Str("message", message).Msg("Generated random event")
}

// maybeUpdateTodos randomly updates todo status
func (a *Agent) maybeUpdateTodos() {
	if rand.Float64() > 0.05 { // 5% chance per tick
		return
	}
	
	todos, err := a.client.ListTodos(a.id)
	if err != nil || len(todos) == 0 {
		return
	}
	
	// Find an incomplete todo to complete
	for _, todo := range todos {
		if !todo.Completed && rand.Float64() < 0.3 { // 30% chance to complete each todo
			completed := true
			current := false
			
			_, err := a.client.UpdateTodo(a.id, todo.ID, models.UpdateTodoRequest{
				Completed: &completed,
				Current:   &current,
			})
			
			if err == nil {
				log.Debug().Str("agent", a.id).Str("todo", todo.Text).Msg("Randomly completed todo")
			}
			break
		}
	}
	
	// Maybe set a random todo as current
	if rand.Float64() < 0.2 { // 20% chance
		incompleteTodos := []models.TodoItem{}
		for _, todo := range todos {
			if !todo.Completed {
				incompleteTodos = append(incompleteTodos, todo)
			}
		}
		
		if len(incompleteTodos) > 0 {
			todo := incompleteTodos[rand.Intn(len(incompleteTodos))]
			current := true
			
			a.client.UpdateTodo(a.id, todo.ID, models.UpdateTodoRequest{
				Current: &current,
			})
		}
	}
}

// maybeUpdateMetrics randomly updates agent metrics
func (a *Agent) maybeUpdateMetrics() {
	if rand.Float64() > 0.1 { // 10% chance per tick
		return
	}
	
	// Simulate incremental work
	if a.state == StateActive {
		if rand.Float64() < 0.5 {
			a.filesChanged += rand.Intn(3)
		}
		if rand.Float64() < 0.4 {
			a.linesAdded += rand.Intn(20) + 1
		}
		if rand.Float64() < 0.3 {
			a.linesRemoved += rand.Intn(10)
		}
		if rand.Float64() < 0.2 {
			a.progress = min(100, a.progress+rand.Intn(10)+1)
		}
	}
}

// selectNewScenario selects a new scenario for the agent
func (a *Agent) selectNewScenario() {
	scenario := scenarios.GetRandomScenario()
	a.scenario = &scenario
	
	log.Info().Str("agent", a.id).Str("scenario", scenario.Name).Msg("Selected new scenario")
}

// startScenario starts executing the current scenario
func (a *Agent) startScenario() error {
	if a.scenario == nil {
		log.Debug().Str("agent", a.id).Msg("Cannot start scenario: scenario is nil")
		return nil
	}
	
	log.Info().
		Str("agent", a.id).
		Str("scenario", a.scenario.Name).
		Str("description", a.scenario.Description).
		Dur("estimated_duration", a.scenario.EstimatedDuration).
		Msg("Starting scenario")
	
	a.currentTask = a.scenario.Name
	a.workStartTime = time.Now()
	a.progress = 0
	a.state = StateActive
	
	// Create todos from scenario
	for i, step := range a.scenario.Steps {
		_, err := a.client.CreateTodo(a.id, models.CreateTodoRequest{
			Text:  step,
			Order: i + 1,
		})
		if err != nil {
			log.Warn().Err(err).Str("step", step).Msg("Failed to create scenario todo")
		}
	}
	
	// Update status to reflect new state
	a.updateStatus()
	
	// Log scenario start
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: "Started scenario: " + a.scenario.Name,
		Metadata: map[string]interface{}{
			"scenario_name":        a.scenario.Name,
			"scenario_description": a.scenario.Description,
			"estimated_duration":   a.scenario.EstimatedDuration.String(),
		},
	})
	
	return err
}

// executeScenario continues executing the current scenario
func (a *Agent) executeScenario() error {
	if a.scenario == nil {
		return a.handleActiveState() // Fallback to default active behavior
	}
	
	// Calculate progress based on elapsed time
	elapsed := time.Since(a.workStartTime)
	progressPercent := int((elapsed.Seconds() / a.scenario.EstimatedDuration.Seconds()) * 100)
	oldProgress := a.progress
	a.progress = min(100, progressPercent)
	
	// Log detailed work progress
	if a.progress != oldProgress || rand.Float64() < 0.3 {
		a.logWorkActivity()
	}
	
	// Simulate realistic development work
	if rand.Float64() < 0.4 {
		a.simulateWorkActivity()
	}
	
	// Simulate scenario-specific behavior
	if rand.Float64() < a.scenario.ErrorProbability {
		return a.handleScenarioError()
	}
	
	if !a.questionPosted && rand.Float64() < a.scenario.QuestionProbability {
		return a.askScenarioQuestion()
	}
	
	// Complete scenario if done
	if a.progress >= 100 {
		return a.completeScenario()
	}
	
	return a.updateStatus()
}

// handleScenarioError handles scenario-specific errors
func (a *Agent) handleScenarioError() error {
	if a.scenario == nil {
		return nil
	}
	
	errors := a.scenario.PossibleErrors
	if len(errors) == 0 {
		return a.handleRandomError()
	}
	
	errorMsg := errors[rand.Intn(len(errors))]
	a.state = StateError
	
	log.Warn().Str("agent", a.id).Str("scenario", a.scenario.Name).Str("error", errorMsg).Msg("Scenario error occurred")
	
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeError),
		Message: errorMsg,
		Metadata: map[string]interface{}{
			"scenario_name": a.scenario.Name,
			"error_type":    "scenario_specific",
		},
	})
	
	return err
}

// askScenarioQuestion asks a scenario-specific question
func (a *Agent) askScenarioQuestion() error {
	if a.scenario == nil {
		return a.askQuestion()
	}
	
	questions := a.scenario.PossibleQuestions
	if len(questions) == 0 {
		return a.askRandomQuestion()
	}
	
	question := questions[rand.Intn(len(questions))]
	a.pendingQuestion = question
	a.questionPosted = true
	a.state = StateWaitingFeedback
	
	log.Info().Str("agent", a.id).Str("scenario", a.scenario.Name).Str("question", question).Msg("Scenario question posted")
	
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeQuestion),
		Message: question,
		Metadata: map[string]interface{}{
			"scenario_name":  a.scenario.Name,
			"question_type":  "scenario_specific",
		},
	})
	
	return err
}

// completeScenario completes the current scenario
func (a *Agent) completeScenario() error {
	if a.scenario == nil {
		return a.completeCurrentWork()
	}
	
	log.Info().Str("agent", a.id).Str("scenario", a.scenario.Name).Msg("Completing scenario")
	
	// Log scenario completion
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeSuccess),
		Message: "Completed scenario: " + a.scenario.Name,
		Metadata: map[string]interface{}{
			"scenario_name":     a.scenario.Name,
			"duration_taken":    time.Since(a.workStartTime).String(),
			"files_changed":     a.filesChanged,
			"lines_added":       a.linesAdded,
			"lines_removed":     a.linesRemoved,
		},
	})
	
	if err != nil {
		log.Warn().Err(err).Msg("Failed to log scenario completion")
	}
	
	// Apply scenario completion effects
	a.filesChanged += rand.Intn(10) + 1
	a.linesAdded += rand.Intn(200) + 50
	a.linesRemoved += rand.Intn(100) + 10
	
	return a.completeCurrentWork()
}

// logWorkActivity logs detailed information about what the agent is working on
func (a *Agent) logWorkActivity() {
	if a.scenario == nil || len(a.scenario.Steps) == 0 {
		return
	}
	
	// Calculate which step we should be working on based on progress
	stepIndex := int(float64(len(a.scenario.Steps)) * float64(a.progress) / 100.0)
	if stepIndex >= len(a.scenario.Steps) {
		stepIndex = len(a.scenario.Steps) - 1
	}
	
	currentStep := a.scenario.Steps[stepIndex]
	
	// Sometimes work on a nearby step for variety
	if rand.Float64() < 0.3 && len(a.scenario.Steps) > 1 {
		offset := rand.Intn(3) - 1 // -1, 0, or 1
		newIndex := stepIndex + offset
		if newIndex >= 0 && newIndex < len(a.scenario.Steps) {
			stepIndex = newIndex
			currentStep = a.scenario.Steps[stepIndex]
		}
	}
	
	log.Info().
		Str("agent", a.id).
		Str("current_step", currentStep).
		Int("step_number", stepIndex+1).
		Int("total_steps", len(a.scenario.Steps)).
		Int("progress", a.progress).
		Str("scenario", a.scenario.Name).
		Msg("Working on scenario step")
	
	// Log as an info event with scenario-specific step
	a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: fmt.Sprintf("🔧 Step %d/%d: %s (%d%% complete)", stepIndex+1, len(a.scenario.Steps), currentStep, a.progress),
		Metadata: map[string]interface{}{
			"activity_type":  "scenario_step",
			"step_number":    stepIndex + 1,
			"total_steps":    len(a.scenario.Steps),
			"current_step":   currentStep,
			"progress":       a.progress,
			"scenario":       a.scenario.Name,
			"scenario_type":  getScenarioType(a.scenario.Name),
		},
	})
}

// simulateWorkActivity simulates realistic development work with file changes and commits
func (a *Agent) simulateWorkActivity() {
	// Get scenario-specific work types
	scenarioType := getScenarioType(a.scenario.Name)
	
	var workTypes []struct {
		activity    string
		icon        string
		filesChange int
		linesAdd    int
		linesRemove int
		eventType   string
	}
	
	// Customize work types based on scenario type
	switch scenarioType {
	case "Bug Fix":
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Debugging issue reproduction", "🔍", 1, 5, 0, "debugging"},
			{"Analyzing root cause", "🕵️", 1, 10, 5, "analysis"},
			{"Implementing bug fix", "🐛", 1, 15, 20, "bugfix"},
			{"Adding regression tests", "🧪", 2, 50, 0, "testing"},
			{"Validating fix", "✅", 1, 5, 2, "validation"},
		}
	case "Feature Development":
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Designing new components", "🎨", 2, 80, 10, "design"},
			{"Implementing core logic", "✨", 3, 100, 5, "development"},
			{"Adding UI components", "🖼️", 2, 60, 15, "frontend"},
			{"Creating API endpoints", "🔌", 1, 40, 5, "backend"},
			{"Writing feature tests", "🧪", 2, 70, 0, "testing"},
		}
	case "Performance Optimization":
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Profiling performance bottlenecks", "📊", 1, 10, 0, "analysis"},
			{"Optimizing database queries", "🗄️", 2, 20, 30, "optimization"},
			{"Implementing caching layer", "⚡", 2, 50, 10, "performance"},
			{"Refactoring inefficient code", "♻️", 3, 40, 60, "refactor"},
			{"Adding performance monitoring", "📈", 1, 30, 5, "monitoring"},
		}
	case "Security Audit":
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Scanning for vulnerabilities", "🔒", 1, 5, 0, "security_scan"},
			{"Implementing input validation", "🛡️", 2, 40, 20, "security_fix"},
			{"Adding authentication checks", "🔐", 1, 30, 10, "auth"},
			{"Updating security policies", "📋", 1, 20, 5, "documentation"},
			{"Testing security measures", "🧪", 2, 35, 0, "security_test"},
		}
	case "Infrastructure":
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Configuring CI/CD pipeline", "⚙️", 3, 60, 20, "infrastructure"},
			{"Setting up monitoring", "📊", 2, 40, 5, "monitoring"},
			{"Implementing deployment scripts", "🚀", 2, 80, 30, "deployment"},
			{"Optimizing build process", "🔧", 1, 25, 40, "optimization"},
			{"Adding health checks", "❤️", 1, 20, 5, "monitoring"},
		}
	default:
		// Generic work types for other scenarios
		workTypes = []struct {
			activity    string
			icon        string
			filesChange int
			linesAdd    int
			linesRemove int
			eventType   string
		}{
			{"Writing code", "✨", 2, 50, 5, "development"},
			{"Fixing issues", "🐛", 1, 10, 15, "bugfix"},
			{"Refactoring", "♻️", 3, 30, 40, "refactor"},
			{"Adding tests", "🧪", 2, 80, 0, "testing"},
			{"Updating docs", "📝", 1, 20, 5, "documentation"},
		}
	}
	
	work := workTypes[rand.Intn(len(workTypes))]
	
	// Simulate file changes
	filesChanged := rand.Intn(work.filesChange) + 1
	linesAdded := rand.Intn(work.linesAdd) + 5
	linesRemoved := rand.Intn(work.linesRemove)
	
	a.filesChanged += filesChanged
	a.linesAdded += linesAdded
	a.linesRemoved += linesRemoved
	
	log.Debug().
		Str("agent", a.id).
		Str("work_type", work.activity).
		Int("files_changed", filesChanged).
		Int("lines_added", linesAdded).
		Int("lines_removed", linesRemoved).
		Msg("Simulating work activity")
	
	// Create detailed work event
	a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: fmt.Sprintf("%s %s", work.icon, work.activity),
		Metadata: map[string]interface{}{
			"work_type":      work.eventType,
			"files_changed":  filesChanged,
			"lines_added":    linesAdded,
			"lines_removed":  linesRemoved,
			"total_files":    a.filesChanged,
			"total_lines":    a.linesAdded - a.linesRemoved,
		},
	})
	
	// Sometimes create a commit
	if rand.Float64() < 0.3 {
		commitMessages := []string{
			fmt.Sprintf("feat: %s", work.activity),
			fmt.Sprintf("fix: resolve issue with %s", work.activity),
			fmt.Sprintf("refactor: improve %s", work.activity),
			fmt.Sprintf("test: add tests for %s", work.activity),
			fmt.Sprintf("docs: update documentation for %s", work.activity),
			"chore: minor improvements and cleanup",
			"style: fix code formatting issues",
			"perf: optimize performance bottlenecks",
		}
		
		commitMsg := commitMessages[rand.Intn(len(commitMessages))]
		
		log.Info().
			Str("agent", a.id).
			Str("commit_message", commitMsg).
			Msg("Simulating commit")
		
		a.client.CreateEvent(a.id, models.CreateEventRequest{
			Type:    string(models.EventTypeCommit),
			Message: fmt.Sprintf("📝 Committed: %s", commitMsg),
			Metadata: map[string]interface{}{
				"commit_message": commitMsg,
				"files_in_commit": filesChanged,
				"lines_changed":   linesAdded + linesRemoved,
			},
		})
		
		a.lastCommitTime = time.Now()
	}
}

// getScenarioType extracts the scenario type from the name
func getScenarioType(scenarioName string) string {
	if scenarioName == "" {
		return "unknown"
	}
	
	// Find the first " - " to extract the type
	for i := 0; i < len(scenarioName)-2; i++ {
		if scenarioName[i] == ' ' && scenarioName[i+1] == '-' && scenarioName[i+2] == ' ' {
			return scenarioName[:i]
		}
	}
	
	// Fallback to the whole name if no " - " found
	return scenarioName
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
