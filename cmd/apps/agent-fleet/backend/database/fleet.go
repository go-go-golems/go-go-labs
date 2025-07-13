package database

import (
	"time"

	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Fleet operations

func (db *DB) GetFleetStatus() (*models.FleetStatus, error) {
	status := &models.FleetStatus{}

	// Total agents
	err := db.Get(&status.TotalAgents, "SELECT COUNT(*) FROM agents")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total agents")
	}

	// Active agents
	err = db.Get(&status.ActiveAgents, "SELECT COUNT(*) FROM agents WHERE status = 'active'")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active agents")
	}

	// Pending tasks
	err = db.Get(&status.PendingTasks, "SELECT COUNT(*) FROM tasks WHERE status = 'pending'")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pending tasks")
	}

	// Agents needing feedback
	err = db.Get(&status.AgentsNeedingFeedback, "SELECT COUNT(*) FROM agents WHERE status = 'waiting_feedback'")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get agents needing feedback")
	}

	// Total files changed
	err = db.Get(&status.TotalFilesChanged, "SELECT COALESCE(SUM(files_changed), 0) FROM agents")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total files changed")
	}

	// Total commits today
	today := time.Now().Truncate(24 * time.Hour)
	err = db.Get(&status.TotalCommitsToday, "SELECT COUNT(*) FROM events WHERE type = 'commit' AND timestamp >= ?", today)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total commits today")
	}

	return status, nil
}

func (db *DB) GetRecentUpdates(limit int, since *time.Time) ([]models.Event, error) {
	query := "SELECT id, agent_id, type, message, metadata, timestamp FROM events"
	args := []interface{}{}

	if since != nil {
		query += " WHERE timestamp >= ?"
		args = append(args, since)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	events, _, err := db.ListEvents("", "", since, limit, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get recent updates")
	}

	return events, nil
}
