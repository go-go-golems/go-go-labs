package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Task handlers

func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	assignedAgentID := r.URL.Query().Get("assigned_agent_id")
	priority := r.URL.Query().Get("priority")
	limit := parseQueryInt(r, "limit", 50)
	offset := parseQueryInt(r, "offset", 0)

	// Enforce maximum limit
	if limit > 100 {
		limit = 100
	}

	tasks, total, err := h.db.ListTasks(status, assignedAgentID, priority, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tasks")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list tasks")
		return
	}

	response := models.TasksListResponse{
		Tasks: tasks,
		ListResponse: models.ListResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required")
		return
	}

	task, err := h.db.GetTask(taskID)
	if err != nil {
		log.Error().Err(err).Str("taskID", taskID).Msg("Failed to get task")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get task")
		return
	}

	if task == nil {
		writeErrorResponse(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
		return
	}

	writeJSONResponse(w, http.StatusOK, task)
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	if req.Title == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TITLE", "Task title is required")
		return
	}
	if req.Description == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_DESCRIPTION", "Task description is required")
		return
	}
	if req.Priority == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_PRIORITY", "Task priority is required")
		return
	}

	// Validate priority
	switch req.Priority {
	case string(models.TaskPriorityLow), string(models.TaskPriorityMedium), string(models.TaskPriorityHigh), string(models.TaskPriorityUrgent):
		// Valid
	default:
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_PRIORITY", "Invalid task priority")
		return
	}

	task, err := h.db.CreateTask(req)
	if err != nil {
		log.Error().Err(err).Interface("request", req).Msg("Failed to create task")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create task")
		return
	}

	// Broadcast task assignment if assigned to an agent
	if task.AssignedAgentID != nil {
		h.sse.BroadcastTaskAssigned(task, *task.AssignedAgentID)
	}

	writeJSONResponse(w, http.StatusCreated, task)
}

func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Validate status if provided
	if req.Status != nil {
		switch *req.Status {
		case string(models.TaskStatusPending), string(models.TaskStatusAssigned), string(models.TaskStatusInProgress), string(models.TaskStatusCompleted), string(models.TaskStatusFailed):
			// Valid
		default:
			writeErrorResponse(w, http.StatusBadRequest, "INVALID_STATUS", "Invalid task status")
			return
		}
	}

	// Validate priority if provided
	if req.Priority != nil {
		switch *req.Priority {
		case string(models.TaskPriorityLow), string(models.TaskPriorityMedium), string(models.TaskPriorityHigh), string(models.TaskPriorityUrgent):
			// Valid
		default:
			writeErrorResponse(w, http.StatusBadRequest, "INVALID_PRIORITY", "Invalid task priority")
			return
		}
	}

	task, err := h.db.UpdateTask(taskID, req)
	if err != nil {
		log.Error().Err(err).Str("taskID", taskID).Interface("request", req).Msg("Failed to update task")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update task")
		return
	}

	// Broadcast task assignment if agent was assigned
	if req.AssignedAgentID != nil && *req.AssignedAgentID != "" {
		h.sse.BroadcastTaskAssigned(task, *req.AssignedAgentID)
	}

	writeJSONResponse(w, http.StatusOK, task)
}

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_TASK_ID", "Task ID is required")
		return
	}

	err := h.db.DeleteTask(taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
			return
		}
		log.Error().Err(err).Str("taskID", taskID).Msg("Failed to delete task")
		writeErrorResponse(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
