package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/templates"
)

// Web interface handlers

func (h *Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	// Get fleet status
	fleetStatus, err := h.db.GetFleetStatus()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get fleet status for index page")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get recent updates
	recentUpdates, err := h.db.GetRecentUpdates(10, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent updates for index page")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	component := templates.Index(fleetStatus, recentUpdates)
	err = component.Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render index template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) AgentsPage(w http.ResponseWriter, r *http.Request) {
	// Get all agents
	agents, _, err := h.db.ListAgents("", 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get agents for agents page")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	component := templates.Agents(agents)
	err = component.Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render agents template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
