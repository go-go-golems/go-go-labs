package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wesen/friday-talks/internal/auth"
	"github.com/wesen/friday-talks/internal/models"
)

// HandleVoteOnTalk handles voting on a talk
func (h *TalkHandler) HandleVoteOnTalk(w http.ResponseWriter, r *http.Request) {
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

	// Check if talk is still in proposed state
	if talk.Status != models.TalkStatusProposed {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Can only vote on proposed talks", http.StatusSeeOther)
		return
	}

	// Check if user is not the speaker (can't vote on own talk)
	if talk.SpeakerID == user.ID {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=You cannot vote on your own talk", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get vote data
	interestLevelStr := r.FormValue("interest_level")
	interestLevel, err := strconv.Atoi(interestLevelStr)
	if err != nil || interestLevel < 1 || interestLevel > 5 {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Invalid interest level", http.StatusSeeOther)
		return
	}

	// Create availability map from form data
	availability := make(map[string]bool)
	for _, date := range talk.PreferredDates {
		// Check if the checkbox for this date was checked
		value := r.FormValue("availability_" + date)
		availability[date] = (value == "true")
	}

	// Check if we're updating an existing vote
	existingVote, err := h.voteRepo.FindByIDs(r.Context(), user.ID, talkID)
	if err == nil {
		// Update existing vote
		existingVote.InterestLevel = interestLevel
		existingVote.Availability = availability

		if err := h.voteRepo.Update(r.Context(), existingVote); err != nil {
			http.Error(w, "Failed to update vote", http.StatusInternalServerError)
			return
		}
	} else {
		// Create new vote
		vote := &models.Vote{
			UserID:        user.ID,
			TalkID:        talkID,
			InterestLevel: interestLevel,
			Availability:  availability,
		}

		if err := h.voteRepo.Create(r.Context(), vote); err != nil {
			http.Error(w, "Failed to create vote", http.StatusInternalServerError)
			return
		}
	}

	// Redirect back to talk page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Vote submitted successfully", http.StatusSeeOther)
}

// HandleManageAttendance handles managing attendance for a talk
func (h *TalkHandler) HandleManageAttendance(w http.ResponseWriter, r *http.Request) {
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

	// Check if talk is scheduled
	if talk.Status != models.TalkStatusScheduled {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Can only manage attendance for scheduled talks", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get status from form
	status := models.AttendanceStatus(r.FormValue("status"))
	if status != models.AttendanceStatusConfirmed && status != models.AttendanceStatusDeclined {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Invalid attendance status", http.StatusSeeOther)
		return
	}

	// Check if we're updating an existing attendance record
	existingAttendance, err := h.attendanceRepo.FindByIDs(r.Context(), talkID, user.ID)
	if err == nil {
		// Update existing attendance
		existingAttendance.Status = status

		if err := h.attendanceRepo.Update(r.Context(), existingAttendance); err != nil {
			http.Error(w, "Failed to update attendance", http.StatusInternalServerError)
			return
		}
	} else {
		// Create new attendance record
		attendance := &models.Attendance{
			TalkID: talkID,
			UserID: user.ID,
			Status: status,
		}

		if err := h.attendanceRepo.Create(r.Context(), attendance); err != nil {
			http.Error(w, "Failed to create attendance record", http.StatusInternalServerError)
			return
		}
	}

	// Redirect back to talk page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Attendance updated successfully", http.StatusSeeOther)
}

// HandleProvideFeedback handles providing feedback for a talk
func (h *TalkHandler) HandleProvideFeedback(w http.ResponseWriter, r *http.Request) {
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

	// Check if talk is completed
	if talk.Status != models.TalkStatusCompleted {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Can only provide feedback for completed talks", http.StatusSeeOther)
		return
	}

	// Check if user attended the talk
	attendance, err := h.attendanceRepo.FindByIDs(r.Context(), talkID, user.ID)
	if err != nil || attendance.Status != models.AttendanceStatusAttended {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=You must have attended the talk to provide feedback", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get feedback from form
	feedback := r.FormValue("feedback")
	if feedback == "" {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Feedback cannot be empty", http.StatusSeeOther)
		return
	}

	// Update attendance record with feedback
	attendance.Feedback = feedback

	if err := h.attendanceRepo.Update(r.Context(), attendance); err != nil {
		http.Error(w, "Failed to save feedback", http.StatusInternalServerError)
		return
	}

	// Redirect back to talk page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Feedback submitted successfully", http.StatusSeeOther)
}

// HandleAddResource handles adding a resource to a talk
func (h *TalkHandler) HandleAddResource(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is the speaker
	if talk.SpeakerID != user.ID {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Only the speaker can add resources to a talk", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get resource data from form
	title := r.FormValue("title")
	url := r.FormValue("url")
	resourceType := models.ResourceType(r.FormValue("type"))

	// Validate input
	if title == "" || url == "" {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Title and URL are required", http.StatusSeeOther)
		return
	}

	// Create new resource
	resource := &models.Resource{
		TalkID: talkID,
		Title:  title,
		URL:    url,
		Type:   resourceType,
	}

	if err := h.resourceRepo.Create(r.Context(), resource); err != nil {
		http.Error(w, "Failed to create resource", http.StatusInternalServerError)
		return
	}

	// Redirect back to talk page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Resource added successfully", http.StatusSeeOther)
}

// HandleDeleteResource handles deleting a resource from a talk
func (h *TalkHandler) HandleDeleteResource(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get talk and resource IDs from URL parameters
	talkIDStr := chi.URLParam(r, "id")
	talkID, err := strconv.Atoi(talkIDStr)
	if err != nil {
		http.Error(w, "Invalid talk ID", http.StatusBadRequest)
		return
	}

	resourceIDStr := chi.URLParam(r, "resourceId")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
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

	// Check if user is the speaker
	if talk.SpeakerID != user.ID {
		http.Redirect(w, r, "/talks/"+talkIDStr+"?error=Only the speaker can delete resources from a talk", http.StatusSeeOther)
		return
	}

	// Delete the resource
	if err := h.resourceRepo.Delete(r.Context(), resourceID); err != nil {
		http.Error(w, "Failed to delete resource", http.StatusInternalServerError)
		return
	}

	// Redirect back to talk page
	http.Redirect(w, r, "/talks/"+talkIDStr+"?success=Resource deleted successfully", http.StatusSeeOther)
}
