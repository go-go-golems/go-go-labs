package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/auth"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/models"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/services"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/templates"
)

// HandleProposeTalk handles proposing a new talk
func (h *TalkHandler) HandleProposeTalk(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get upcoming Fridays for selection
	fridays := h.scheduler.GetUpcomingFridays(8) // Next 8 weeks

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		preferredDates := r.Form["preferred_dates[]"]

		// Validate input
		if title == "" || description == "" {
			templates.ProposeTalk(user, "Title and description are required", fridays).Render(r.Context(), w)
			return
		}

		if len(preferredDates) == 0 {
			templates.ProposeTalk(user, "Please select at least one preferred date", fridays).Render(r.Context(), w)
			return
		}

		// Create the talk
		talk := &models.Talk{
			Title:          title,
			Description:    description,
			SpeakerID:      user.ID,
			PreferredDates: preferredDates,
			Status:         models.TalkStatusProposed,
		}

		if err := h.talkRepo.Create(r.Context(), talk); err != nil {
			http.Error(w, "Failed to create talk proposal", http.StatusInternalServerError)
			return
		}

		// Redirect to the talk detail page
		http.Redirect(w, r, "/talks/"+strconv.Itoa(talk.ID)+"?success=Talk proposed successfully", http.StatusSeeOther)
		return
	}

	// Render propose talk form
	templates.ProposeTalk(user, "", fridays).Render(r.Context(), w)
}

// HandleEditTalk handles editing an existing talk
func (h *TalkHandler) HandleEditTalk(w http.ResponseWriter, r *http.Request) {
	// Get talk ID from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get talk from repository
	talk, err := h.talkRepo.FindByID(r.Context(), talkID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if user is the speaker or admin
	if talk.SpeakerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Check if talk can be edited (only proposed talks can be edited)
	if talk.Status != models.TalkStatusProposed {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Only proposed talks can be edited", http.StatusSeeOther)
		return
	}

	// Get upcoming Fridays for selection
	fridays := h.scheduler.GetUpcomingFridays(8) // Next 8 weeks

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		preferredDates := r.Form["preferred_dates[]"]

		// Validate input
		if title == "" || description == "" {
			templates.EditTalk(user, talk, "Title and description are required", fridays).Render(r.Context(), w)
			return
		}

		if len(preferredDates) == 0 {
			templates.EditTalk(user, talk, "Please select at least one preferred date", fridays).Render(r.Context(), w)
			return
		}

		// Update the talk
		talk.Title = title
		talk.Description = description
		talk.PreferredDates = preferredDates

		if err := h.talkRepo.Update(r.Context(), talk); err != nil {
			http.Error(w, "Failed to update talk", http.StatusInternalServerError)
			return
		}

		// Redirect to the talk detail page
		http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Talk updated successfully", http.StatusSeeOther)
		return
	}

	// Render edit talk form
	templates.EditTalk(user, talk, "", fridays).Render(r.Context(), w)
}

// HandleCancelTalk handles canceling a talk
func (h *TalkHandler) HandleCancelTalk(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get talk ID from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get talk from repository
	talk, err := h.talkRepo.FindByID(r.Context(), talkID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check if user is the speaker or admin
	if talk.SpeakerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Cancel the talk
	talk.Status = models.TalkStatusCanceled

	if err := h.talkRepo.Update(r.Context(), talk); err != nil {
		http.Error(w, "Failed to cancel talk", http.StatusInternalServerError)
		return
	}

	// Redirect to the talk detail page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Talk canceled successfully", http.StatusSeeOther)
}

// HandleScheduleTalk handles scheduling a talk
func (h *TalkHandler) HandleScheduleTalk(w http.ResponseWriter, r *http.Request) {
	// Get talk ID from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get talk from repository
	talk, err := h.talkRepo.FindByID(r.Context(), talkID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Only allow scheduling proposed talks
	if talk.Status != models.TalkStatusProposed {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Only proposed talks can be scheduled", http.StatusSeeOther)
		return
	}

	// Load speaker info if not already loaded
	if talk.Speaker == nil {
		speaker, err := h.userRepo.FindByID(r.Context(), talk.SpeakerID)
		if err == nil {
			talk.Speaker = speaker
		}
	}

	// Get upcoming Fridays for selection
	fridays := h.scheduler.GetUpcomingFridays(8) // Next 8 weeks

	// Find other talks for ranking
	rankings, err := h.scheduler.FindBestTalksForDate(r.Context(), fridays[0])
	if err != nil {
		rankings = []*services.TalkRanking{}
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		dateStr := r.FormValue("scheduled_date")
		if dateStr == "" {
			http.Error(w, "Date is required", http.StatusBadRequest)
			return
		}

		// Parse the date
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		// Check if date is already scheduled
		scheduled, err := h.scheduler.IsDateScheduled(r.Context(), date)
		if err != nil {
			http.Error(w, "Failed to check date availability", http.StatusInternalServerError)
			return
		}

		if scheduled {
			http.Redirect(w, r, "/talks/"+talkIDStr+"/schedule?error=Date already has a scheduled talk", http.StatusSeeOther)
			return
		}

		// Update the talk status and scheduled date
		if err := h.scheduler.ScheduleTalk(r.Context(), talkID, date); err != nil {
			http.Error(w, "Failed to schedule talk", http.StatusInternalServerError)
			return
		}

		// Redirect to the talk detail page
		http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Talk scheduled successfully", http.StatusSeeOther)
		return
	}

	// Render schedule talk form
	templates.ScheduleTalk(user, talk, rankings, fridays).Render(r.Context(), w)
}

// HandleCompleteTalk handles marking a talk as completed
func (h *TalkHandler) HandleCompleteTalk(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get talk ID from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get talk from repository
	talk, err := h.talkRepo.FindByID(r.Context(), talkID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Only allow completing scheduled talks
	if talk.Status != models.TalkStatusScheduled {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Only scheduled talks can be marked as completed", http.StatusSeeOther)
		return
	}

	// Mark as completed
	talk.Status = models.TalkStatusCompleted

	if err := h.talkRepo.Update(r.Context(), talk); err != nil {
		http.Error(w, "Failed to complete talk", http.StatusInternalServerError)
		return
	}

	// Update attendance records for confirmed attendees
	attendances, err := h.attendanceRepo.ListByTalk(r.Context(), talkID)
	if err == nil {
		for _, attendance := range attendances {
			if attendance.Status == models.AttendanceStatusConfirmed {
				attendance.Status = models.AttendanceStatusAttended
				h.attendanceRepo.Update(r.Context(), attendance)
			}
		}
	}

	// Redirect to the talk detail page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Talk marked as completed", http.StatusSeeOther)
}
