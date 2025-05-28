package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Fleet handlers

func (h *Handlers) GetFleetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.db.GetFleetStatus()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get fleet status")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get fleet status")
		return
	}

	writeJSONResponse(w, http.StatusOK, status)
}

func (h *Handlers) GetRecentUpdates(w http.ResponseWriter, r *http.Request) {
	limit := parseQueryInt(r, "limit", 20)
	since := parseQueryTime(r, "since")

	events, err := h.db.GetRecentUpdates(limit, since)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent updates")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get recent updates")
		return
	}

	response := models.RecentUpdatesResponse{
		Updates: events,
	}

	writeJSONResponse(w, http.StatusOK, response)
}
