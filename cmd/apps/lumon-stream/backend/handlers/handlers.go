package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/models"
)

// Response is a generic API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// StreamInfoResponse combines stream info and steps
type StreamInfoResponse struct {
	models.StreamInfo
	CompletedSteps []models.Step `json:"completedSteps"`
	ActiveStep     *models.Step  `json:"activeStep"`
	UpcomingSteps  []models.Step `json:"upcomingSteps"`
}

// GetStreamInfo handles GET requests for stream information
func GetStreamInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get stream info
	info, err := database.GetStreamInfo()
	if err != nil {
		log.Printf("Error getting stream info: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to retrieve stream information",
		})
		return
	}

	// Get steps
	completedSteps, activeSteps, upcomingSteps, err := database.GetSteps()
	if err != nil {
		log.Printf("Error getting steps: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to retrieve steps",
		})
		return
	}

	// Prepare response
	response := StreamInfoResponse{
		StreamInfo:     info,
		CompletedSteps: completedSteps,
		UpcomingSteps:  upcomingSteps,
	}

	// Set active step if available
	if len(activeSteps) > 0 {
		response.ActiveStep = &activeSteps[0]
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    response,
	})
}

// UpdateStreamInfo handles POST requests to update stream information
func UpdateStreamInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var info models.StreamInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if info.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Title is required",
		})
		return
	}

	// Ensure StartTime is set
	if info.StartTime.IsZero() {
		info.StartTime = time.Now()
	}

	// Update stream info
	err = database.UpdateStreamInfo(info)
	if err != nil {
		log.Printf("Error updating stream info: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to update stream information",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Stream information updated successfully",
	})
}

// AddStep handles POST requests to add a new step
func AddStep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var step struct {
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&step)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if step.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Step content is required",
		})
		return
	}

	// Validate status
	if step.Status != "completed" && step.Status != "active" && step.Status != "upcoming" {
		step.Status = "upcoming" // Default to upcoming if invalid
	}

	// Add step
	err = database.AddStep(step.Content, step.Status)
	if err != nil {
		log.Printf("Error adding step: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to add step",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Step added successfully",
	})
}

// UpdateStepStatus handles POST requests to update a step's status
func UpdateStepStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if request.ID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Valid step ID is required",
		})
		return
	}

	// Validate status
	if request.Status != "completed" && request.Status != "active" && request.Status != "upcoming" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Valid status is required (completed, active, or upcoming)",
		})
		return
	}

	// Update step status
	err = database.UpdateStepStatus(request.ID, request.Status)
	if err != nil {
		log.Printf("Error updating step status: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Failed to update step status",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Step status updated successfully",
	})
}
