package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Task operations

func (db *DB) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	task := &models.Task{
		ID:              uuid.New().String(),
		Title:           req.Title,
		Description:     req.Description,
		AssignedAgentID: req.AssignedAgentID,
		Status:          string(models.TaskStatusPending),
		Priority:        req.Priority,
		CreatedAt:       time.Now(),
	}

	// If assigned to an agent, set assigned_at and status
	if task.AssignedAgentID != nil {
		now := time.Now()
		task.AssignedAt = &now
		task.Status = string(models.TaskStatusAssigned)
	}

	query := `
		INSERT INTO tasks (id, title, description, assigned_agent_id, status, priority, created_at, assigned_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, task.ID, task.Title, task.Description, task.AssignedAgentID, task.Status, task.Priority, task.CreatedAt, task.AssignedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create task")
	}

	return task, nil
}

func (db *DB) GetTask(id string) (*models.Task, error) {
	var task models.Task
	query := "SELECT * FROM tasks WHERE id = ?"
	err := db.Get(&task, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get task")
	}
	return &task, nil
}

func (db *DB) ListTasks(status, assignedAgentID, priority string, limit, offset int) ([]models.Task, int, error) {
	var tasks []models.Task
	var total int

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}
	if assignedAgentID != "" {
		whereClauses = append(whereClauses, "assigned_agent_id = ?")
		args = append(args, assignedAgentID)
	}
	if priority != "" {
		whereClauses = append(whereClauses, "priority = ?")
		args = append(args, priority)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + joinStrings(whereClauses, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM tasks" + whereClause
	err := db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count tasks")
	}

	// List query
	query := "SELECT * FROM tasks" + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, offset)

	err = db.Select(&tasks, query, listArgs...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list tasks")
	}

	return tasks, total, nil
}

func (db *DB) UpdateTask(id string, req models.UpdateTaskRequest) (*models.Task, error) {
	// Build dynamic update query
	setParts := []string{}
	args := map[string]interface{}{"id": id}

	if req.Title != nil {
		setParts = append(setParts, "title = :title")
		args["title"] = *req.Title
	}
	if req.Description != nil {
		setParts = append(setParts, "description = :description")
		args["description"] = *req.Description
	}
	if req.Status != nil {
		setParts = append(setParts, "status = :status")
		args["status"] = *req.Status

		// Set completed_at if status is completed
		if *req.Status == string(models.TaskStatusCompleted) {
			setParts = append(setParts, "completed_at = :completed_at")
			args["completed_at"] = time.Now()
		}
	}
	if req.Priority != nil {
		setParts = append(setParts, "priority = :priority")
		args["priority"] = *req.Priority
	}
	if req.AssignedAgentID != nil {
		setParts = append(setParts, "assigned_agent_id = :assigned_agent_id")
		args["assigned_agent_id"] = *req.AssignedAgentID

		// Set assigned_at if assigning to an agent
		if *req.AssignedAgentID != "" {
			setParts = append(setParts, "assigned_at = :assigned_at")
			args["assigned_at"] = time.Now()
		}
	}

	if len(setParts) == 0 {
		return db.GetTask(id)
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = :id", joinStrings(setParts, ", "))
	_, err := db.NamedExec(query, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update task")
	}

	return db.GetTask(id)
}

func (db *DB) DeleteTask(id string) error {
	query := "DELETE FROM tasks WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete task")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
