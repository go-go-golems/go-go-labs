package handlers

import (
	"net/http"
	"time"

	"github.com/wesen/friday-talks/internal/auth"
	"github.com/wesen/friday-talks/internal/models"
	"github.com/wesen/friday-talks/internal/templates"
)

// HomeHandler handles the home page
type HomeHandler struct {
	talkRepo models.TalkRepository
}

// NewHomeHandler creates a new HomeHandler
func NewHomeHandler(talkRepo models.TalkRepository) *HomeHandler {
	return &HomeHandler{
		talkRepo: talkRepo,
	}
}

// HandleHome renders the home page
func (h *HomeHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Redirect to index if not on root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())

	// Get upcoming (scheduled) talks
	upcomingTalks, err := h.getUpcomingTalks(r)
	if err != nil {
		http.Error(w, "Failed to get upcoming talks", http.StatusInternalServerError)
		return
	}

	// Get recent (completed) talks
	recentTalks, err := h.getRecentTalks(r)
	if err != nil {
		http.Error(w, "Failed to get recent talks", http.StatusInternalServerError)
		return
	}

	// Get proposed talks
	proposedTalks, err := h.getProposedTalks(r)
	if err != nil {
		http.Error(w, "Failed to get proposed talks", http.StatusInternalServerError)
		return
	}

	// Render home page
	templates.Home(user, upcomingTalks, recentTalks, proposedTalks).Render(r.Context(), w)
}

// getUpcomingTalks returns upcoming talks that are scheduled
func (h *HomeHandler) getUpcomingTalks(r *http.Request) ([]*models.Talk, error) {
	// Get scheduled talks
	talks, err := h.talkRepo.ListByStatus(r.Context(), models.TalkStatusScheduled)
	if err != nil {
		return nil, err
	}

	// Filter to include only future talks
	now := time.Now()
	var upcomingTalks []*models.Talk
	for _, talk := range talks {
		if talk.ScheduledDate != nil && talk.ScheduledDate.After(now) {
			upcomingTalks = append(upcomingTalks, talk)
		}
	}

	// Sort by date (should already be sorted by the repository)
	// Limit to 3 for display
	if len(upcomingTalks) > 3 {
		upcomingTalks = upcomingTalks[:3]
	}

	return upcomingTalks, nil
}

// getRecentTalks returns recent talks that have been completed
func (h *HomeHandler) getRecentTalks(r *http.Request) ([]*models.Talk, error) {
	// Get completed talks
	talks, err := h.talkRepo.ListByStatus(r.Context(), models.TalkStatusCompleted)
	if err != nil {
		return nil, err
	}

	// Limit to 3 for display (should already be sorted by date by the repository)
	if len(talks) > 3 {
		talks = talks[:3]
	}

	return talks, nil
}

// getProposedTalks returns talks that have been proposed but not yet scheduled
func (h *HomeHandler) getProposedTalks(r *http.Request) ([]*models.Talk, error) {
	// Get proposed talks
	talks, err := h.talkRepo.ListByStatus(r.Context(), models.TalkStatusProposed)
	if err != nil {
		return nil, err
	}

	// Limit to 3 for display
	if len(talks) > 3 {
		talks = talks[:3]
	}

	return talks, nil
}

// HandleNotFound renders the 404 page
func (h *HomeHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())

	w.WriteHeader(http.StatusNotFound)
	templates.NotFound(user).Render(r.Context(), w)
}

// HandleInternalServerError renders the 500 page
func (h *HomeHandler) HandleInternalServerError(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())

	w.WriteHeader(http.StatusInternalServerError)
	templates.InternalServerError(user).Render(r.Context(), w)
}
