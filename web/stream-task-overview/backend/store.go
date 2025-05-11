package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// StreamStore manages stream data
type StreamStore struct {
	mutex sync.RWMutex
	db    *sqlx.DB
	sql   squirrel.StatementBuilderType
}

// NewStreamStore creates a new store with SQLite persistence
func NewStreamStore() *StreamStore {
	log.Debug().Msg("Creating new StreamStore")

	// Connect to SQLite database
	log.Debug().Str("db_path", "./stream.db").Msg("Connecting to SQLite database")
	db, err := sqlx.Connect("sqlite3", "./stream.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Database connection established")

	// Create SQL builder with SQLite placeholder
	sql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	// Create store
	store := &StreamStore{
		db:  db,
		sql: sql,
	}

	// Initialize database schema
	log.Debug().Msg("Initializing database schema")
	store.initSchema()

	// Create default data if none exists
	log.Debug().Msg("Initializing default data")
	store.initDefaultData()

	log.Info().Msg("StreamStore successfully initialized")
	return store
}

// initSchema creates database tables if they don't exist
func (s *StreamStore) initSchema() {
	// Create stream_info table
	log.Debug().Msg("Creating stream_info table if not exists")
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS stream_info (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		language TEXT NOT NULL,
		github_repo TEXT NOT NULL,
		viewer_count INTEGER NOT NULL
	);
	`)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create stream_info table")
	}

	// Create steps table
	log.Debug().Msg("Creating steps table if not exists")
	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS steps (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		completed TEXT NOT NULL,
		active TEXT NOT NULL,
		upcoming TEXT NOT NULL
	);
	`)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create steps table")
	}

	log.Info().Msg("Database schema setup complete")
}

// initDefaultData inserts default data if tables are empty
func (s *StreamStore) initDefaultData() {
	// Check if stream_info has data
	log.Debug().Msg("Checking if default data needs to be created")
	var count int
	err := s.db.Get(&count, "SELECT COUNT(*) FROM stream_info")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check stream_info data")
	}

	if count == 0 {
		log.Info().Msg("No existing data found, creating default data")
		// Insert default stream info
		log.Debug().Msg("Creating default stream info")
		query := s.sql.Insert("stream_info").Columns(
			"id", "title", "description", "start_time", 
			"language", "github_repo", "viewer_count",
		).Values(
			1,
			"Building a React Component Library",
			"Creating reusable UI components with TailwindCSS",
			time.Now(),
			"JavaScript/React",
			"https://github.com/yourusername/component-library",
			42,
		)

		sql, args, err := query.ToSql()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to build SQL for stream info")
		}

		_, err = s.db.Exec(sql, args...)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to insert default stream info")
		}
		log.Debug().Msg("Default stream info created successfully")

		// Insert default steps
		log.Debug().Msg("Creating default steps")
		completed, err := json.Marshal([]string{
			"Project setup and initialization",
			"Design system planning",
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal completed steps")
		}

		upcoming, err := json.Marshal([]string{
			"Implement Button component",
			"Create Card component",
			"Build Form elements",
			"Add dark mode toggle",
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to marshal upcoming steps")
		}

		query = s.sql.Insert("steps").Columns(
			"id", "completed", "active", "upcoming",
		).Values(
			1,
			string(completed),
			"Setting up component architecture",
			string(upcoming),
		)

		sql, args, err = query.ToSql()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to build SQL for steps")
		}

		_, err = s.db.Exec(sql, args...)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to insert default steps")
		}
		log.Debug().Msg("Default steps created successfully")
		log.Info().Msg("Default data creation complete")
	} else {
		log.Info().Msg("Existing data found, skipping default data creation")
	}
}

// GetStreamInfo returns the current stream info
func (s *StreamStore) GetStreamInfo() StreamInfo {
	log.Debug().Msg("Getting stream info")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := s.sql.Select(
		"title", "description", "start_time", 
		"language", "github_repo", "viewer_count",
	).From("stream_info").Where(squirrel.Eq{"id": 1})

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL for GetStreamInfo")
		return StreamInfo{}
	}

	var info StreamInfo
	err = s.db.Get(&info, sql, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stream info from database")
		return StreamInfo{}
	}

	log.Debug().Interface("info", info).Msg("Retrieved stream info")
	return info
}

// UpdateStreamInfo updates stream info
func (s *StreamStore) UpdateStreamInfo(info StreamInfo) {
	log.Debug().Interface("info", info).Msg("Updating stream info")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := s.sql.Update("stream_info").SetMap(map[string]interface{}{
		"title":        info.Title,
		"description":  info.Description,
		"start_time":   info.StartTime,
		"language":     info.Language,
		"github_repo":  info.GithubRepo,
		"viewer_count": info.ViewerCount,
	}).Where(squirrel.Eq{"id": 1})

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL for UpdateStreamInfo")
		return
	}

	_, err = s.db.Exec(sql, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update stream info in database")
		return
	}

	log.Info().Msg("Stream info updated successfully")
}

// GetSteps returns all steps
func (s *StreamStore) GetSteps() StepInfo {
	log.Debug().Msg("Getting steps")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := s.sql.Select("completed", "active", "upcoming").From("steps").Where(squirrel.Eq{"id": 1})

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL for GetSteps")
		return StepInfo{}
	}

	var row struct {
		Completed string `db:"completed"`
		Active    string `db:"active"`
		Upcoming  string `db:"upcoming"`
	}

	err = s.db.Get(&row, sql, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get steps from database")
		return StepInfo{}
	}

	// Parse JSON arrays
	var steps StepInfo
	steps.Active = row.Active

	// Parse completed steps
	err = json.Unmarshal([]byte(row.Completed), &steps.Completed)
	if err != nil {
		log.Error().Err(err).Str("json", row.Completed).Msg("Failed to unmarshal completed steps")
		steps.Completed = []string{}
	}

	// Parse upcoming steps
	err = json.Unmarshal([]byte(row.Upcoming), &steps.Upcoming)
	if err != nil {
		log.Error().Err(err).Str("json", row.Upcoming).Msg("Failed to unmarshal upcoming steps")
		steps.Upcoming = []string{}
	}

	log.Debug().Interface("steps", steps).Msg("Retrieved steps")
	return steps
}

// updateSteps updates all steps in database
func (s *StreamStore) updateSteps(steps StepInfo) error {
	log.Debug().Interface("steps", steps).Msg("Updating steps in database")

	// Marshal arrays to JSON
	completed, err := json.Marshal(steps.Completed)
	if err != nil {
		log.Error().Err(err).Interface("completed", steps.Completed).Msg("Failed to marshal completed steps")
		return errors.Wrap(err, "marshal completed steps")
	}

	upcoming, err := json.Marshal(steps.Upcoming)
	if err != nil {
		log.Error().Err(err).Interface("upcoming", steps.Upcoming).Msg("Failed to marshal upcoming steps")
		return errors.Wrap(err, "marshal upcoming steps")
	}

	// Update database
	query := s.sql.Update("steps").SetMap(map[string]interface{}{
		"completed": string(completed),
		"active":    steps.Active,
		"upcoming":  string(upcoming),
	}).Where(squirrel.Eq{"id": 1})

	sql, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL for updateSteps")
		return errors.Wrap(err, "build SQL")
	}

	_, err = s.db.Exec(sql, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute SQL for updateSteps")
		return errors.Wrap(err, "execute SQL")
	}

	log.Info().Msg("Steps updated successfully")
	return nil
}

// SetActiveStep sets a new active step
func (s *StreamStore) SetActiveStep(step string) {
	log.Debug().Str("step", step).Msg("Setting active step")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current steps
	steps := s.GetSteps()

	// Add current active to completed if it exists
	if steps.Active != "" {
		log.Debug().Str("previous_active", steps.Active).Msg("Moving previous active step to completed")
		steps.Completed = append(steps.Completed, steps.Active)
	}

	// Set new active step
	steps.Active = step

	// Update database
	err := s.updateSteps(steps)
	if err != nil {
		log.Error().Err(err).Str("step", step).Msg("Failed to set active step")
		return
	}

	log.Info().Str("step", step).Msg("Active step set successfully")
}

// AddUpcomingStep adds a new upcoming step
func (s *StreamStore) AddUpcomingStep(step string) {
	log.Debug().Str("step", step).Msg("Adding upcoming step")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current steps
	steps := s.GetSteps()

	// Add new upcoming step
	steps.Upcoming = append(steps.Upcoming, step)

	// Update database
	err := s.updateSteps(steps)
	if err != nil {
		log.Error().Err(err).Str("step", step).Msg("Failed to add upcoming step")
		return
	}

	log.Info().Str("step", step).Msg("Upcoming step added successfully")
}

// CompleteActiveStep completes the current active step
func (s *StreamStore) CompleteActiveStep() {
	log.Debug().Msg("Completing active step")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current steps
	steps := s.GetSteps()

	// Process only if there's an active step
	if steps.Active != "" {
		// Add to completed
		log.Debug().Str("active", steps.Active).Msg("Moving active step to completed")
		steps.Completed = append(steps.Completed, steps.Active)

		// Set next step as active if available
		if len(steps.Upcoming) > 0 {
			log.Debug().Str("next_step", steps.Upcoming[0]).Msg("Setting next step as active")
			steps.Active = steps.Upcoming[0]
			steps.Upcoming = steps.Upcoming[1:]
		} else {
			log.Debug().Msg("No upcoming steps, setting active to empty")
			steps.Active = ""
		}

		// Update database
		err := s.updateSteps(steps)
		if err != nil {
			log.Error().Err(err).Msg("Failed to complete active step")
			return
		}

		log.Info().Msg("Active step completed successfully")
	} else {
		log.Warn().Msg("No active step to complete")
	}
}

// ReactivateStep moves a step from completed/upcoming to active
func (s *StreamStore) ReactivateStep(step string, source string) {
	log.Debug().Str("step", step).Str("source", source).Msg("Reactivating step")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current steps
	steps := s.GetSteps()

	// Add current active to completed if it exists
	if steps.Active != "" {
		log.Debug().Str("previous_active", steps.Active).Msg("Moving previous active step to completed")
		steps.Completed = append(steps.Completed, steps.Active)
	}

	// Set step as active
	steps.Active = step

	// Remove from source list
	if source == "upcoming" {
		for i, s := range steps.Upcoming {
			if s == step {
				log.Debug().Int("index", i).Msg("Removing step from upcoming list")
				steps.Upcoming = append(steps.Upcoming[:i], steps.Upcoming[i+1:]...)
				break
			}
		}
	} else if source == "completed" {
		for i, s := range steps.Completed {
			if s == step {
				log.Debug().Int("index", i).Msg("Removing step from completed list")
				steps.Completed = append(steps.Completed[:i], steps.Completed[i+1:]...)
				break
			}
		}
	}

	// Update database
	err := s.updateSteps(steps)
	if err != nil {
		log.Error().Err(err).Str("step", step).Str("source", source).Msg("Failed to reactivate step")
		return
	}

	log.Info().Str("step", step).Str("source", source).Msg("Step reactivated successfully")
}