package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

type DB struct {
	*sqlx.DB
}

// Initialize creates a new database connection and runs migrations
func Initialize(dbPath string) (*DB, error) {
	sqlDB, err := sqlx.Connect("sqlite3", dbPath+"?_foreign_keys=1&_journal_mode=WAL")
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	db := &DB{sqlDB}

	if err := db.migrate(); err != nil {
		return nil, errors.Wrap(err, "failed to run migrations")
	}

	return db, nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'idle',
		current_task TEXT,
		worktree TEXT NOT NULL,
		files_changed INTEGER DEFAULT 0,
		lines_added INTEGER DEFAULT 0,
		lines_removed INTEGER DEFAULT 0,
		last_commit DATETIME,
		progress INTEGER DEFAULT 0,
		pending_question TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		metadata TEXT, -- JSON
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS todo_items (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		text TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		current BOOLEAN DEFAULT FALSE,
		order_num INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		assigned_agent_id TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		priority TEXT NOT NULL DEFAULT 'medium',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		assigned_at DATETIME,
		completed_at DATETIME,
		FOREIGN KEY (assigned_agent_id) REFERENCES agents(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS commands (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		content TEXT NOT NULL,
		type TEXT NOT NULL,
		response TEXT,
		status TEXT NOT NULL DEFAULT 'sent',
		sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		responded_at DATETIME,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_events_agent_id ON events(agent_id);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
	CREATE INDEX IF NOT EXISTS idx_todo_items_agent_id ON todo_items(agent_id);
	CREATE INDEX IF NOT EXISTS idx_todo_items_completed ON todo_items(completed);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_assigned_agent_id ON tasks(assigned_agent_id);
	CREATE INDEX IF NOT EXISTS idx_commands_agent_id ON commands(agent_id);
	CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);

	-- Trigger to update updated_at on agents
	CREATE TRIGGER IF NOT EXISTS update_agents_updated_at 
		AFTER UPDATE ON agents
		BEGIN
			UPDATE agents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`

	_, err := db.Exec(schema)
	return err
}

// Agent operations

func (db *DB) CreateAgent(req models.CreateAgentRequest) (*models.Agent, error) {
	agent := &models.Agent{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Status:       string(models.AgentStatusIdle),
		Worktree:     req.Worktree,
		FilesChanged: 0,
		LinesAdded:   0,
		LinesRemoved: 0,
		Progress:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO agents (id, name, status, worktree, files_changed, lines_added, lines_removed, progress, created_at, updated_at)
		VALUES (:id, :name, :status, :worktree, :files_changed, :lines_added, :lines_removed, :progress, :created_at, :updated_at)
	`
	_, err := db.NamedExec(query, agent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create agent")
	}

	return agent, nil
}

func (db *DB) GetAgent(id string) (*models.Agent, error) {
	var agent models.Agent
	query := "SELECT * FROM agents WHERE id = ?"
	err := db.Get(&agent, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get agent")
	}
	return &agent, nil
}

func (db *DB) ListAgents(status string, limit, offset int) ([]models.Agent, int, error) {
	var agents []models.Agent
	var total int

	// Count query
	countQuery := "SELECT COUNT(*) FROM agents"
	var countArgs []interface{}
	if status != "" {
		countQuery += " WHERE status = ?"
		countArgs = append(countArgs, status)
	}

	err := db.Get(&total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count agents")
	}

	// List query
	query := "SELECT * FROM agents"
	var args []interface{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err = db.Select(&agents, query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list agents")
	}

	return agents, total, nil
}

func (db *DB) UpdateAgent(id string, req models.UpdateAgentRequest) (*models.Agent, error) {
	// Build dynamic update query
	setParts := []string{}
	args := map[string]interface{}{"id": id}

	if req.Name != nil {
		setParts = append(setParts, "name = :name")
		args["name"] = *req.Name
	}
	if req.Status != nil {
		setParts = append(setParts, "status = :status")
		args["status"] = *req.Status
	}
	if req.CurrentTask != nil {
		setParts = append(setParts, "current_task = :current_task")
		args["current_task"] = *req.CurrentTask
	}
	if req.Worktree != nil {
		setParts = append(setParts, "worktree = :worktree")
		args["worktree"] = *req.Worktree
	}
	if req.FilesChanged != nil {
		setParts = append(setParts, "files_changed = :files_changed")
		args["files_changed"] = *req.FilesChanged
	}
	if req.LinesAdded != nil {
		setParts = append(setParts, "lines_added = :lines_added")
		args["lines_added"] = *req.LinesAdded
	}
	if req.LinesRemoved != nil {
		setParts = append(setParts, "lines_removed = :lines_removed")
		args["lines_removed"] = *req.LinesRemoved
	}
	if req.Progress != nil {
		setParts = append(setParts, "progress = :progress")
		args["progress"] = *req.Progress
	}
	if req.PendingQuestion != nil {
		setParts = append(setParts, "pending_question = :pending_question")
		args["pending_question"] = *req.PendingQuestion
	}

	if len(setParts) == 0 {
		return db.GetAgent(id)
	}

	query := fmt.Sprintf("UPDATE agents SET %s WHERE id = :id", joinStrings(setParts, ", "))
	_, err := db.NamedExec(query, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update agent")
	}

	return db.GetAgent(id)
}

func (db *DB) DeleteAgent(id string) error {
	query := "DELETE FROM agents WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete agent")
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

// Event operations

func (db *DB) CreateEvent(agentID string, req models.CreateEventRequest) (*models.Event, error) {
	event := &models.Event{
		ID        : uuid.New().String(),
		AgentID   : agentID,
		Type      : req.Type,
		Message   : req.Message,
		Metadata  : req.Metadata,
		Timestamp : time.Now(),
	}

	var metadataJSON []byte
	var err error
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal metadata")
		}
	}

	query := `
		INSERT INTO events (id, agent_id, type, message, metadata, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = db.Exec(query, event.ID, event.AgentID, event.Type, event.Message, metadataJSON, event.Timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create event")
	}

	return event, nil
}

func (db *DB) ListEvents(agentID, eventType string, since *time.Time, limit, offset int) ([]models.Event, int, error) {
	var events []models.Event
	var total int

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	if agentID != "" {
		whereClauses = append(whereClauses, "agent_id = ?")
		args = append(args, agentID)
	}
	if eventType != "" {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, eventType)
	}
	if since != nil {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, since)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + joinStrings(whereClauses, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM events" + whereClause
	err := db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count events")
	}

	// List query
	query := "SELECT id, agent_id, type, message, metadata, timestamp FROM events" + whereClause + " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, offset)

	rows, err := db.Query(query, listArgs...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list events")
	}
	defer rows.Close()

	for rows.Next() {
		var event models.Event
		var metadataJSON *string
		
		err := rows.Scan(&event.ID, &event.AgentID, &event.Type, &event.Message, &metadataJSON, &event.Timestamp)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan event")
		}

		if metadataJSON != nil {
			err := json.Unmarshal([]byte(*metadataJSON), &event.Metadata)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed to unmarshal metadata")
			}
		}

		events = append(events, event)
	}

	return events, total, nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
