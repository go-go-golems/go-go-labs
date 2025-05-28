package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Command handlers

func (h *Handlers) ListAgentCommands(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	status := r.URL.Query().Get("status")
	limit := parseQueryInt(r, "limit", 50)

	commands, err := h.db.ListCommands(agentID, status, limit)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to list commands")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list commands")
		return
	}

	response := models.CommandsListResponse{
		Commands: commands,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) CreateAgentCommand(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	var req models.CreateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Content == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_CONTENT", "Command content is required")
		return
	}
	if req.Type == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TYPE", "Command type is required")
		return
	}

	// Validate command type
	switch req.Type {
	case string(models.CommandTypeInstruction), string(models.CommandTypeFeedback), string(models.CommandTypeQuestion):
		// Valid
	default:
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid command type")
		return
	}

	command, err := h.db.CreateCommand(agentID, req)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Interface("request", req).Msg("Failed to create command")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create command")
		return
	}

	// Broadcast command creation
	h.sse.BroadcastCommandReceived(agentID, command)

	writeJSONResponse(w, http.StatusCreated, command)
}

func (h *Handlers) UpdateAgentCommand(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	commandID := chi.URLParam(r, "commandID")
	
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}
	if commandID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_COMMAND_ID", "Command ID is required")
		return
	}

	var req models.UpdateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Validate status if provided
	if req.Status != nil {
		switch *req.Status {
		case string(models.CommandStatusSent), string(models.CommandStatusAcknowledged), string(models.CommandStatusCompleted):
			// Valid
		default:
			writeErrorResponse(w, http.StatusBadRequest, "INVALID_STATUS", "Invalid command status")
			return
		}
	}

	command, err := h.db.UpdateCommand(agentID, commandID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "COMMAND_NOT_FOUND", "Command not found")
			return
		}
		log.Error().Err(err).Str("agentID", agentID).Str("commandID", commandID).Interface("request", req).Msg("Failed to update command")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update command")
		return
	}

	writeJSONResponse(w, http.StatusOK, command)
}
