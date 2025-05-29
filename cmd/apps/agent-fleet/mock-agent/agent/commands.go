package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// checkCommands checks for and processes new commands
func (a *Agent) checkCommands() error {
	if !a.registered {
		return nil
	}
	
	a.lastCommandCheck = time.Now()
	
	// Get new commands (sent status)
	commands, err := a.client.ListCommands(a.id, "sent", 10)
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}
	
	// Process each command
	for _, command := range commands {
		if err := a.processCommand(&command); err != nil {
			log.Error().Err(err).Str("command", command.ID).Msg("Failed to process command")
		}
	}
	
	return nil
}

// processCommand processes a single command
func (a *Agent) processCommand(command *models.Command) error {
	log.Info().
		Str("agent", a.id).
		Str("command", command.ID).
		Str("type", command.Type).
		Str("content", command.Content).
		Msg("Processing command")
	
	// First acknowledge the command
	status := string(models.CommandStatusAcknowledged)
	_, err := a.client.UpdateCommand(a.id, command.ID, models.UpdateCommandRequest{
		Status: &status,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge command: %w", err)
	}
	
	// Process based on command type
	var response string
	switch command.Type {
	case string(models.CommandTypeInstruction):
		response = a.handleInstruction(command.Content)
	case string(models.CommandTypeFeedback):
		response = a.handleFeedback(command.Content)
	case string(models.CommandTypeQuestion):
		response = a.handleQuestion(command.Content)
	default:
		response = fmt.Sprintf("Unknown command type: %s", command.Type)
	}
	
	// Update command with response
	completedStatus := string(models.CommandStatusCompleted)
	_, err = a.client.UpdateCommand(a.id, command.ID, models.UpdateCommandRequest{
		Status:   &completedStatus,
		Response: &response,
	})
	if err != nil {
		return fmt.Errorf("failed to complete command: %w", err)
	}
	
	// Log command processing
	_, err = a.client.CreateEvent(a.id, models.CreateEventRequest{
		Type:    string(models.EventTypeCommand),
		Message: fmt.Sprintf("Processed %s command: %s", command.Type, truncateString(command.Content, 50)),
		Metadata: map[string]interface{}{
			"command_id":   command.ID,
			"command_type": command.Type,
			"response":     truncateString(response, 100),
		},
	})
	
	if err != nil {
		log.Warn().Err(err).Msg("Failed to log command processing event")
	}
	
	return nil
}

// handleInstruction processes an instruction command
func (a *Agent) handleInstruction(content string) string {
	content = strings.ToLower(content)
	
	// Handle common instructions
	if strings.Contains(content, "stop") || strings.Contains(content, "pause") {
		log.Info().Str("agent", a.id).Msg("Received stop instruction")
		a.state = StateIdle
		a.currentTask = ""
		a.progress = 0
		a.updateStatus()
		return "Acknowledged. Stopping current work and returning to idle state."
	}
	
	if strings.Contains(content, "continue") || strings.Contains(content, "proceed") {
		log.Info().Str("agent", a.id).Msg("Received continue instruction")
		if a.state == StateWaitingFeedback {
			a.state = StateActive
			a.questionPosted = false
			a.pendingQuestion = ""
			a.updateStatus()
		}
		return "Acknowledged. Continuing with current work."
	}
	
	if strings.Contains(content, "restart") || strings.Contains(content, "reset") {
		log.Info().Str("agent", a.id).Msg("Received restart instruction")
		a.state = StateIdle
		a.currentTask = ""
		a.progress = 0
		a.filesChanged = 0
		a.linesAdded = 0
		a.linesRemoved = 0
		a.questionPosted = false
		a.pendingQuestion = ""
		a.updateStatus()
		return "Acknowledged. Resetting agent state and returning to idle."
	}
	
	if strings.Contains(content, "focus") || strings.Contains(content, "prioritize") {
		log.Info().Str("agent", a.id).Msg("Received focus instruction")
		// Extract what to focus on
		var focusArea string
		if strings.Contains(content, "security") {
			focusArea = "security"
		} else if strings.Contains(content, "performance") {
			focusArea = "performance"
		} else if strings.Contains(content, "testing") {
			focusArea = "testing"
		} else {
			focusArea = "quality"
		}
		
		return fmt.Sprintf("Acknowledged. Focusing on %s aspects of the current work.", focusArea)
	}
	
	// Generic response for other instructions
	return fmt.Sprintf("Acknowledged instruction: %s. Adjusting work approach accordingly.", truncateString(content, 50))
}

// handleFeedback processes feedback command
func (a *Agent) handleFeedback(content string) string {
	log.Info().Str("agent", a.id).Str("feedback", content).Msg("Received feedback")
	
	// If we were waiting for feedback, resume work
	if a.state == StateWaitingFeedback {
		a.state = StateActive
		a.questionPosted = false
		a.pendingQuestion = ""
		a.updateStatus()
	}
	
	content = strings.ToLower(content)
	
	// Analyze feedback tone and respond accordingly
	if strings.Contains(content, "good") || strings.Contains(content, "excellent") || strings.Contains(content, "approve") {
		return "Thank you for the positive feedback! I'll continue with the current approach."
	}
	
	if strings.Contains(content, "change") || strings.Contains(content, "different") || strings.Contains(content, "modify") {
		// Simulate making changes based on feedback
		a.progress = max(0, a.progress-10) // Rollback progress slightly for changes
		return "Understood. I'll modify the approach based on your feedback and implement the suggested changes."
	}
	
	if strings.Contains(content, "error") || strings.Contains(content, "wrong") || strings.Contains(content, "issue") {
		a.progress = max(0, a.progress-20) // Rollback more for errors
		return "Thank you for catching that. I'll review and fix the identified issues before proceeding."
	}
	
	return fmt.Sprintf("Thank you for the feedback: %s. I'll incorporate this into my work.", truncateString(content, 50))
}

// handleQuestion processes question command
func (a *Agent) handleQuestion(content string) string {
	log.Info().Str("agent", a.id).Str("question", content).Msg("Received question")
	
	content = strings.ToLower(content)
	
	// Provide contextual answers based on current state
	if strings.Contains(content, "status") || strings.Contains(content, "progress") {
		return fmt.Sprintf("Current status: %s. Progress: %d%%. Working on: %s. Files changed: %d, Lines added: %d, Lines removed: %d.",
			a.state, a.progress, a.currentTask, a.filesChanged, a.linesAdded, a.linesRemoved)
	}
	
	if strings.Contains(content, "todo") || strings.Contains(content, "task") {
		todos, err := a.client.ListTodos(a.id)
		if err != nil || len(todos) == 0 {
			return "No current todos available."
		}
		
		completed := 0
		for _, todo := range todos {
			if todo.Completed {
				completed++
			}
		}
		
		return fmt.Sprintf("Current todos: %d total, %d completed, %d remaining. Current task: %s",
			len(todos), completed, len(todos)-completed, a.currentTask)
	}
	
	if strings.Contains(content, "problem") || strings.Contains(content, "issue") || strings.Contains(content, "stuck") {
		if a.state == StateWaitingFeedback {
			return fmt.Sprintf("I'm currently waiting for feedback on: %s", a.pendingQuestion)
		} else if a.state == StateError {
			return "I'm currently in an error state and attempting to recover."
		} else {
			return "No significant problems at the moment. Work is progressing normally."
		}
	}
	
	if strings.Contains(content, "next") || strings.Contains(content, "plan") {
		todos, err := a.client.ListTodos(a.id)
		if err != nil || len(todos) == 0 {
			return "No specific next steps planned. Will continue with current work or look for new tasks."
		}
		
		// Find next incomplete todo
		for _, todo := range todos {
			if !todo.Completed {
				return fmt.Sprintf("Next step: %s", todo.Text)
			}
		}
		
		return "All current todos completed. Ready for new tasks."
	}
	
	// Generic response
	return fmt.Sprintf("Regarding your question about '%s': I'll need to analyze this further and provide a detailed response based on current context.",
		truncateString(content, 30))
}

// truncateString truncates a string to specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
