package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Event handlers

func (h *Handlers) ListAgentEvents(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	eventType := r.URL.Query().Get("type")
	since := parseQueryTime(r, "since")
	limit := parseQueryInt(r, "limit", 100)
	offset := parseQueryInt(r, "offset", 0)

	events, total, err := h.db.ListEvents(agentID, eventType, since, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Msg("Failed to list agent events")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list events")
		return
	}

	response := models.EventsListResponse{
		Events: events,
		ListResponse: models.ListResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) CreateAgentEvent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_AGENT_ID", "Agent ID is required")
		return
	}

	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Type == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TYPE", "Event type is required")
		return
	}
	if req.Message == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_MESSAGE", "Event message is required")
		return
	}

	event, err := h.db.CreateEvent(agentID, req)
	if err != nil {
		log.Error().Err(err).Str("agentID", agentID).Interface("request", req).Msg("Failed to create event")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create event")
		return
	}

	log.Info().
		Str("agent_id", agentID).
		Str("event_id", event.ID).
		Str("event_type", event.Type).
		Str("message", event.Message).
		Msg("Event created")

	// Broadcast event creation
	h.sse.BroadcastAgentEventCreated(agentID, event)

	// Also broadcast specific events for errors and warnings
	switch event.Type {
	case "error":
		h.sse.BroadcastAgentErrorPosted(agentID, event.Message, nil)
	case "warning":
		h.sse.BroadcastAgentWarningPosted(agentID, event.Message, nil)
	}

	writeJSONResponse(w, http.StatusCreated, event)
}
