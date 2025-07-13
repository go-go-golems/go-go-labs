package agent

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// updateStatus updates the agent's status in the backend
func (a *Agent) updateStatus() error {
	if !a.registered {
		return nil
	}

	updateReq := models.UpdateAgentRequest{
		Status:       stringPtr(string(a.state)),
		Progress:     &a.progress,
		FilesChanged: &a.filesChanged,
		LinesAdded:   &a.linesAdded,
		LinesRemoved: &a.linesRemoved,
	}

	if a.currentTask != "" {
		updateReq.CurrentTask = &a.currentTask
	}

	if a.pendingQuestion != "" {
		updateReq.PendingQuestion = &a.pendingQuestion
	} else {
		updateReq.PendingQuestion = stringPtr("") // Clear pending question
	}

	_, err := a.client.UpdateAgent(a.id, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	log.Debug().
		Str("agent", a.id).
		Str("state", string(a.state)).
		Int("progress", a.progress).
		Str("task", a.currentTask).
		Msg("Updated agent status")

	return nil
}

// handleError handles agent errors and transitions to error state
func (a *Agent) handleError(err error) {
	log.Error().Err(err).Str("agent", a.id).Msg("Agent encountered error")

	a.state = StateError

	// Log error event
	_, eventErr := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeError),
		Message: fmt.Sprintf("Agent error: %s", err.Error()),
		Metadata: map[string]interface{}{
			"error_type": "agent_internal",
			"context":    a.currentTask,
			"progress":   a.progress,
		},
	})

	if eventErr != nil {
		log.Error().Err(eventErr).Msg("Failed to log error event")
	}

	// Update status to reflect error state
	a.updateStatus()
}

// shutdown gracefully shuts down the agent
func (a *Agent) shutdown() error {
	if !a.registered {
		return nil
	}

	log.Info().Str("agent", a.id).Msg("Agent shutting down")

	a.state = StateShuttingDown

	// Log shutdown event
	_, err := a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: "Agent shutting down gracefully",
		Metadata: map[string]interface{}{
			"shutdown_reason": "requested",
			"uptime":          time.Since(a.workStartTime).String(),
			"final_progress":  a.progress,
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log shutdown event")
	}

	// Update final status
	if err := a.updateStatus(); err != nil {
		log.Warn().Err(err).Msg("Failed to update final status")
	}

	// Clean up - mark incomplete todos, etc.
	a.cleanupOnShutdown()

	// Transition to finished state
	a.state = StateFinished

	// Log finished event
	_, err = a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeInfo),
		Message: "Agent shutdown completed",
		Metadata: map[string]interface{}{
			"final_state": string(a.state),
			"work_summary": map[string]interface{}{
				"files_changed":  a.filesChanged,
				"lines_added":    a.linesAdded,
				"lines_removed":  a.linesRemoved,
				"final_progress": a.progress,
			},
		},
	})

	if err != nil {
		log.Warn().Err(err).Msg("Failed to log finished event")
	}

	// Update final status to finished
	if err := a.updateStatus(); err != nil {
		log.Warn().Err(err).Msg("Failed to update finished status")
	}

	log.Info().Str("agent", a.id).Str("state", string(a.state)).Msg("Agent finished")

	// Optionally delete agent from backend (or leave for historical purposes)
	// Uncomment the next lines if you want to remove agent on shutdown
	// if err := a.client.DeleteAgent(a.id); err != nil {
	//     log.Warn().Err(err).Msg("Failed to delete agent on shutdown")
	// }

	return nil
}

// cleanupOnShutdown performs cleanup tasks before shutdown
func (a *Agent) cleanupOnShutdown() {
	// Mark any in-progress todos as not current
	todos, err := a.client.ListTodos(a.id)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list todos for cleanup")
		return
	}

	for _, todo := range todos {
		if todo.Current && !todo.Completed {
			current := false
			_, err := a.client.UpdateTodo(a.id, todo.ID, models.UpdateTodoRequest{
				Current: &current,
			})
			if err != nil {
				log.Warn().Err(err).Str("todo", todo.ID).Msg("Failed to update todo on shutdown")
			}
		}
	}

	log.Debug().Str("agent", a.id).Msg("Cleanup completed")
}

// getDetailedStatus returns detailed status information
func (a *Agent) getDetailedStatus() map[string]interface{} {
	status := map[string]interface{}{
		"id":                 a.id,
		"name":               a.config.Name,
		"state":              string(a.state),
		"worktree":           a.config.Worktree,
		"current_task":       a.currentTask,
		"progress":           a.progress,
		"files_changed":      a.filesChanged,
		"lines_added":        a.linesAdded,
		"lines_removed":      a.linesRemoved,
		"question_posted":    a.questionPosted,
		"pending_question":   a.pendingQuestion,
		"registered":         a.registered,
		"randomized":         a.config.Randomized,
		"last_tick":          a.lastTick,
		"last_command_check": a.lastCommandCheck,
	}

	if !a.workStartTime.IsZero() {
		status["work_start_time"] = a.workStartTime
		status["work_duration"] = time.Since(a.workStartTime).String()
	}

	if !a.lastCommitTime.IsZero() {
		status["last_commit_time"] = a.lastCommitTime
		status["time_since_commit"] = time.Since(a.lastCommitTime).String()
	}

	if a.scenario != nil {
		status["scenario"] = map[string]interface{}{
			"name":                 a.scenario.Name,
			"description":          a.scenario.Description,
			"estimated_duration":   a.scenario.EstimatedDuration.String(),
			"error_probability":    a.scenario.ErrorProbability,
			"question_probability": a.scenario.QuestionProbability,
			"steps_count":          len(a.scenario.Steps),
		}
	}

	return status
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
