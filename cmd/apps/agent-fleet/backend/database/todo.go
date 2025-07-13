package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Todo operations

func (db *DB) CreateTodo(agentID string, req models.CreateTodoRequest) (*models.TodoItem, error) {
	todo := &models.TodoItem{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		Text:      req.Text,
		Completed: false,
		Current:   false,
		Order:     req.Order,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO todo_items (id, agent_id, text, completed, current, order_num, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, todo.ID, todo.AgentID, todo.Text, todo.Completed, todo.Current, todo.Order, todo.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create todo")
	}

	return todo, nil
}

func (db *DB) ListTodos(agentID string) ([]models.TodoItem, error) {
	var todos []models.TodoItem
	query := "SELECT * FROM todo_items WHERE agent_id = ? ORDER BY order_num ASC, created_at ASC"
	err := db.Select(&todos, query, agentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list todos")
	}
	return todos, nil
}

func (db *DB) GetTodo(agentID, todoID string) (*models.TodoItem, error) {
	var todo models.TodoItem
	query := "SELECT * FROM todo_items WHERE id = ? AND agent_id = ?"
	err := db.Get(&todo, query, todoID, agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get todo")
	}
	return &todo, nil
}

func (db *DB) UpdateTodo(agentID, todoID string, req models.UpdateTodoRequest) (*models.TodoItem, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.Text != nil {
		setParts = append(setParts, "text = ?")
		args = append(args, *req.Text)
	}
	if req.Completed != nil {
		setParts = append(setParts, "completed = ?")
		args = append(args, *req.Completed)
		if *req.Completed {
			setParts = append(setParts, "completed_at = ?")
			args = append(args, time.Now())
		} else {
			setParts = append(setParts, "completed_at = NULL")
		}
	}
	if req.Current != nil {
		// If setting this todo as current, unset all others for this agent
		if *req.Current {
			_, err := db.Exec("UPDATE todo_items SET current = FALSE WHERE agent_id = ?", agentID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unset other current todos")
			}
		}
		setParts = append(setParts, "current = ?")
		args = append(args, *req.Current)
	}

	if len(setParts) == 0 {
		return db.GetTodo(agentID, todoID)
	}

	query := "UPDATE todo_items SET " + joinStrings(setParts, ", ") + " WHERE id = ? AND agent_id = ?"
	args = append(args, todoID, agentID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update todo")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return db.GetTodo(agentID, todoID)
}

func (db *DB) DeleteTodo(agentID, todoID string) error {
	query := "DELETE FROM todo_items WHERE id = ? AND agent_id = ?"
	result, err := db.Exec(query, todoID, agentID)
	if err != nil {
		return errors.Wrap(err, "failed to delete todo")
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
