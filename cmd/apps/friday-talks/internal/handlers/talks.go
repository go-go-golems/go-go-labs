package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/auth"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/models"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/services"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/templates"
)

// TalkHandler handles talk-related routes
type TalkHandler struct {
	talkRepo       models.TalkRepository
	userRepo       models.UserRepository
	voteRepo       models.VoteRepository
	attendanceRepo models.AttendanceRepository
	resourceRepo   models.ResourceRepository
	scheduler      *services.SchedulerService
}

// NewTalkHandler creates a new TalkHandler
func NewTalkHandler(
	talkRepo models.TalkRepository,
	userRepo models.UserRepository,
	voteRepo models.VoteRepository,
	attendanceRepo models.AttendanceRepository,
	resourceRepo models.ResourceRepository,
	scheduler *services.SchedulerService,
) *TalkHandler {
	return &TalkHandler{
		talkRepo:       talkRepo,
		userRepo:       userRepo,
		voteRepo:       voteRepo,
		attendanceRepo: attendanceRepo,
		resourceRepo:   resourceRepo,
		scheduler:      scheduler,
	}
}

// HandleListTalks handles listing all talks with optional filtering
func (h *TalkHandler) HandleListTalks(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())

	// Get status filter from query parameters
	status := r.URL.Query().Get("status")
	var talks []*models.Talk
	var err error

	if status != "" {
		// Filter by status
		talks, err = h.talkRepo.ListByStatus(r.Context(), models.TalkStatus(status))
	} else {
		// Get all talks
		talks, err = h.talkRepo.List(r.Context())
	}

	if err != nil {
		http.Error(w, "Failed to get talks", http.StatusInternalServerError)
		return
	}

	// Load speaker information for each talk
	for _, talk := range talks {
		if talk.Speaker == nil {
			speaker, err := h.userRepo.FindByID(r.Context(), talk.SpeakerID)
			if err != nil {
				// Log error but continue
				continue
			}
			talk.Speaker = speaker
		}
	}

	// Render talks list page
	templates.TalksList(user, talks, status).Render(r.Context(), w)
}

// HandleGetTalk handles displaying a specific talk
func (h *TalkHandler) HandleGetTalk(w http.ResponseWriter, r *http.Request) {
	// Get talk ID from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())

	// Get talk from repository
	talk, err := h.talkRepo.FindByID(r.Context(), talkID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get resources for the talk
	resources, err := h.resourceRepo.ListByTalk(r.Context(), talk.ID)
	if err != nil {
		// Log error but continue
		resources = []*models.Resource{}
	}

	// Check if user has voted for this talk
	var voted bool
	var attendance *models.Attendance
	if user != nil {
		// Check if user has voted
		_, err := h.voteRepo.FindByIDs(r.Context(), user.ID, talk.ID)
		voted = err == nil

		// Check attendance
		attendance, _ = h.attendanceRepo.FindByIDs(r.Context(), talk.ID, user.ID)
	}

	// Extract success/error messages from query params
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	// Render talk details page
	templates.TalkDetail(user, talk, voted, attendance, resources, errorMsg, success).Render(r.Context(), w)
}

// HandleMyTalks handles displaying the user's talks
func (h *TalkHandler) HandleMyTalks(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get talks by user
	allTalks, err := h.talkRepo.ListBySpeaker(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get talks", http.StatusInternalServerError)
		return
	}

	// Separate talks by status
	var proposedTalks, scheduledTalks, completedTalks []*models.Talk
	for _, talk := range allTalks {
		switch talk.Status {
		case models.TalkStatusProposed:
			proposedTalks = append(proposedTalks, talk)
		case models.TalkStatusScheduled:
			scheduledTalks = append(scheduledTalks, talk)
		case models.TalkStatusCompleted:
			completedTalks = append(completedTalks, talk)
		}
	}

	// Render my talks page
	templates.MyTalks(user, proposedTalks, scheduledTalks, completedTalks).Render(r.Context(), w)
}
