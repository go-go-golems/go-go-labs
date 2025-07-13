package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Todo handlers

func (h *Handlers) ListAgentTodos(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	todos, err := h.db.ListTodos(agentID)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to list todos")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list todos")
		return
	}

	response := models.TodosListResponse{
		Todos: todos,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) CreateAgentTodo(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Text == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TEXT", "Todo text is required")
		return
	}

	todo, err := h.db.CreateTodo(agentID, req)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Interface("request", req).Msg("Failed to create todo")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create todo")
		return
	}

	log.Info().
		Str("agent_id", agentID).
		Str("todo_id", todo.ID).
		Str("todo_text", todo.Text).
		Bool("completed", todo.Completed).
		Bool("current", todo.Current).
		Msg("Todo created")

	// Broadcast todo creation
	h.sse.BroadcastTodoUpdated(agentID, todo, "created")

	writeJSONResponse(w, http.StatusCreated, todo)
}

func (h *Handlers) UpdateAgentTodo(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	todoID := chi.URLParam(r, "todoID")

	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}
	if todoID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TODO_ID", "Todo ID is required")
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	todo, err := h.db.UpdateTodo(agentID, todoID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "TODO_NOT_FOUND", "Todo not found")
			return
		}
		log.Error().Err(err).Str("agentID", agentID).Str("todoID", todoID).Interface("request", req).Msg("Failed to update todo")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update todo")
		return
	}

	// Log the update
	changes := []string{}
	if req.Completed != nil {
		changes = append(changes, fmt.Sprintf("completed: %t", *req.Completed))
	}
	if req.Current != nil {
		changes = append(changes, fmt.Sprintf("current: %t", *req.Current))
	}
	if req.Text != nil {
		changes = append(changes, fmt.Sprintf("text: %s", *req.Text))
	}

	log.Info().
		Str("agent_id", agentID).
		Str("todo_id", todoID).
		Strs("changes", changes).
		Msg("Todo updated")

	// Broadcast todo update
	h.sse.BroadcastTodoUpdated(agentID, todo, "updated")

	writeJSONResponse(w, http.StatusOK, todo)
}

func (h *Handlers) DeleteAgentTodo(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	todoID := chi.URLParam(r, "todoID")

	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}
	if todoID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TODO_ID", "Todo ID is required")
		return
	}

	// Get todo before deletion for broadcast
	todo, err := h.db.GetTodo(agentID, todoID)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Str("todoID", todoID).Msg("Failed to get todo for deletion")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete todo")
		return
	}

	err = h.db.DeleteTodo(agentID, todoID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "TODO_NOT_FOUND", "Todo not found")
			return
		}
		log.Error().Err(err).Str("agentID", agentID).Str("todoID", todoID).Msg("Failed to delete todo")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete todo")
		return
	}

	// Broadcast todo deletion
	if todo != nil {
		h.sse.BroadcastTodoUpdated(agentID, todo, "deleted")
	}

	w.WriteHeader(http.StatusNoContent)
}
