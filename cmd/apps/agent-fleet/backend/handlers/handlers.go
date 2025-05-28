package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

type Handlers struct {
	db  *database.DB
	sse *SSEManager
}

func New(db *database.DB) *Handlers {
	return &Handlers{
		db:  db,
		sse: NewSSEManager(),
	}
}

// Helper functions

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	errorResp := models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	writeJSONResponse(w, statusCode, errorResp)
}

func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	
	return parsed
}

func parseQueryTime(r *http.Request, key string) *time.Time {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	
	parsed, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil
	}
	
	return &parsed
}

// Agent handlers

func (h *Handlers) ListAgents(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseQueryInt(r, "limit", 50)
	offset := parseQueryInt(r, "offset", 0)
	
	// Enforce maximum limit
	if limit > 100 {
		limit = 100
	}

	agents, total, err := h.db.ListAgents(status, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list agents")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list agents")
		return
	}

	response := models.AgentsListResponse{
		Agents: agents,
		ListResponse: models.ListResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	agent, err := h.db.GetAgent(agentID)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to get agent")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get agent")
		return
	}

	if agent == nil {
		writeErrorResponse(w, http.StatusNotFound, "AGENT_NOT_FOUND", "Agent not found")
		return
	}

	writeJSONResponse(w, http.StatusOK, agent)
}

func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_NAME", "Agent name is required")
		return
	}
	if req.Worktree == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_WORKTREE", "Agent worktree is required")
		return
	}

	agent, err := h.db.CreateAgent(req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Failed to create agent")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create agent")
		return
	}

	// Broadcast agent creation
	h.sse.BroadcastAgentStatusChanged(agent.ID, "", agent.Status, agent)

	writeJSONResponse(w, http.StatusCreated, agent)
}

func (h *Handlers) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	// Get current agent for old status
	currentAgent, err := h.db.GetAgent(agentID)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to get current agent")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get agent")
		return
	}
	if currentAgent == nil {
		writeErrorResponse(w, http.StatusNotFound, "AGENT_NOT_FOUND", "Agent not found")
		return
	}

	var req models.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	agent, err := h.db.UpdateAgent(agentID, req)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Interface("request", req).Msg("Failed to update agent")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update agent")
		return
	}

	// Broadcast status change if status changed
	if req.Status != nil && *req.Status != currentAgent.Status {
		h.sse.BroadcastAgentStatusChanged(agent.ID, currentAgent.Status, agent.Status, agent)
	}

	// Broadcast progress update if progress fields changed
	if req.Progress != nil || req.FilesChanged != nil || req.LinesAdded != nil || req.LinesRemoved != nil {
		h.sse.BroadcastAgentProgressUpdated(agent.ID, agent.Progress, agent.FilesChanged, agent.LinesAdded, agent.LinesRemoved)
	}

	// Broadcast question if pending question was set
	if req.PendingQuestion != nil && *req.PendingQuestion != "" {
		h.sse.BroadcastAgentQuestionPosted(agent.ID, *req.PendingQuestion, agent)
	}

	writeJSONResponse(w, http.StatusOK, agent)
}

func (h *Handlers) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	err := h.db.DeleteAgent(agentID)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to delete agent")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete agent")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
