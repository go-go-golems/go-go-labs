package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Command operations

func (db *DB) CreateCommand(agentID string, req models.CreateCommandRequest) (*models.Command, error) {
	command := &models.Command{
		ID      : uuid.New().String(),
		AgentID : agentID,
		Content : req.Content,
		Type    : req.Type,
		Status  : string(models.CommandStatusSent),
		SentAt  : time.Now(),
	}

	query := `
		INSERT INTO commands (id, agent_id, content, type, status, sent_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, command.ID, command.AgentID, command.Content, command.Type, command.Status, command.SentAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}

	return command, nil
}

func (db *DB) ListCommands(agentID, status string, limit int) ([]models.Command, error) {
	var commands []models.Command
	
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if agentID != "" {
		whereClauses = append(whereClauses, "agent_id = ?")
		args = append(args, agentID)
	}
	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + joinStrings(whereClauses, " AND ")
	}

	query := "SELECT * FROM commands" + whereClause + " ORDER BY sent_at DESC LIMIT ?"
	args = append(args, limit)

	err := db.Select(&commands, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list commands")
	}

	return commands, nil
}

func (db *DB) GetCommand(agentID, commandID string) (*models.Command, error) {
	var command models.Command
	query := "SELECT * FROM commands WHERE id = ? AND agent_id = ?"
	err := db.Get(&command, query, commandID, agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get command")
	}
	return &command, nil
}

func (db *DB) UpdateCommand(agentID, commandID string, req models.UpdateCommandRequest) (*models.Command, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.Response != nil {
		setParts = append(setParts, "response = ?")
		args = append(args, *req.Response)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)
		
		// Set responded_at if status is acknowledged or completed
		if *req.Status == string(models.CommandStatusAcknowledged) || *req.Status == string(models.CommandStatusCompleted) {
			setParts = append(setParts, "responded_at = ?")
			args = append(args, time.Now())
		}
	}

	if len(setParts) == 0 {
		return db.GetCommand(agentID, commandID)
	}

	query := "UPDATE commands SET " + joinStrings(setParts, ", ") + " WHERE id = ? AND agent_id = ?"
	args = append(args, commandID, agentID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update command")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return db.GetCommand(agentID, commandID)
}
