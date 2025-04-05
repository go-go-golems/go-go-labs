package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/models"
	_ "github.com/mattn/go-sqlite3"
)

// DB is the database connection
var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return err
	}

	// Initialize with default data if empty
	err = initializeDefaultData()
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary tables in the database
func createTables() error {
	streamInfoTable := `
	CREATE TABLE IF NOT EXISTS stream_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		start_time DATETIME NOT NULL,
		language TEXT,
		github_repo TEXT,
		viewer_count INTEGER DEFAULT 0
	);`

	stepsTable := `
	CREATE TABLE IF NOT EXISTS steps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	_, err := DB.Exec(streamInfoTable)
	if err != nil {
		return fmt.Errorf("error creating stream_info table: %w", err)
	}

	_, err = DB.Exec(stepsTable)
	if err != nil {
		return fmt.Errorf("error creating steps table: %w", err)
	}

	return nil
}

// initializeDefaultData adds default data if the database is empty
func initializeDefaultData() error {
	// Check if stream_info table is empty
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM stream_info").Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking stream_info table: %w", err)
	}

	if count == 0 {
		// Insert default stream info
		_, err = DB.Exec(`
			INSERT INTO stream_info (title, description, start_time, language, github_repo, viewer_count)
			VALUES (?, ?, ?, ?, ?, ?)`,
			"Building a React Component Library",
			"Creating reusable UI components with TailwindCSS",
			time.Now(),
			"JavaScript/React",
			"https://github.com/yourusername/component-library",
			42,
		)
		if err != nil {
			return fmt.Errorf("error inserting default stream info: %w", err)
		}

		// Insert default steps
		completedSteps := []string{
			"Project setup and initialization",
			"Design system planning",
		}
		for _, step := range completedSteps {
			_, err = DB.Exec(`
				INSERT INTO steps (content, status, created_at)
				VALUES (?, ?, ?)`,
				step,
				"completed",
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("error inserting completed step: %w", err)
			}
		}

		// Insert active step
		_, err = DB.Exec(`
			INSERT INTO steps (content, status, created_at)
			VALUES (?, ?, ?)`,
			"Setting up component architecture",
			"active",
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("error inserting active step: %w", err)
		}

		// Insert upcoming steps
		upcomingSteps := []string{
			"Implement Button component",
			"Create Card component",
			"Build Form elements",
			"Add dark mode toggle",
		}
		for _, step := range upcomingSteps {
			_, err = DB.Exec(`
				INSERT INTO steps (content, status, created_at)
				VALUES (?, ?, ?)`,
				step,
				"upcoming",
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("error inserting upcoming step: %w", err)
			}
		}
	}

	return nil
}

// GetStreamInfo retrieves the stream information from the database
func GetStreamInfo() (models.StreamInfo, error) {
	var info models.StreamInfo
	err := DB.QueryRow(`
		SELECT id, title, description, start_time, language, github_repo, viewer_count
		FROM stream_info
		ORDER BY id DESC
		LIMIT 1
	`).Scan(
		&info.ID,
		&info.Title,
		&info.Description,
		&info.StartTime,
		&info.Language,
		&info.GithubRepo,
		&info.ViewerCount,
	)
	if err != nil {
		return models.StreamInfo{}, fmt.Errorf("error getting stream info: %w", err)
	}
	return info, nil
}

// UpdateStreamInfo updates the stream information in the database
func UpdateStreamInfo(info models.StreamInfo) error {
	_, err := DB.Exec(`
		UPDATE stream_info
		SET title = ?, description = ?, start_time = ?, language = ?, github_repo = ?, viewer_count = ?
		WHERE id = ?
	`,
		info.Title,
		info.Description,
		info.StartTime,
		info.Language,
		info.GithubRepo,
		info.ViewerCount,
		info.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating stream info: %w", err)
	}
	return nil
}

// GetSteps retrieves all steps from the database grouped by status
func GetSteps() ([]models.Step, []models.Step, []models.Step, error) {
	var completedSteps, activeSteps, upcomingSteps []models.Step

	// Get completed steps
	rows, err := DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'completed'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting completed steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning completed step: %w", err)
		}
		completedSteps = append(completedSteps, step)
	}

	// Get active step
	rows, err = DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'active'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting active step: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning active step: %w", err)
		}
		activeSteps = append(activeSteps, step)
	}

	// Get upcoming steps
	rows, err = DB.Query(`
		SELECT id, content, status, created_at
		FROM steps
		WHERE status = 'upcoming'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting upcoming steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step models.Step
		err := rows.Scan(&step.ID, &step.Content, &step.Status, &step.CreatedAt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error scanning upcoming step: %w", err)
		}
		upcomingSteps = append(upcomingSteps, step)
	}

	return completedSteps, activeSteps, upcomingSteps, nil
}

// UpdateStepStatus updates the status of a step
func UpdateStepStatus(id int, status string) error {
	_, err := DB.Exec(`
		UPDATE steps
		SET status = ?
		WHERE id = ?
	`, status, id)
	if err != nil {
		return fmt.Errorf("error updating step status: %w", err)
	}
	return nil
}

// AddStep adds a new step to the database
func AddStep(content, status string) error {
	_, err := DB.Exec(`
		INSERT INTO steps (content, status, created_at)
		VALUES (?, ?, ?)
	`, content, status, time.Now())
	if err != nil {
		return fmt.Errorf("error adding step: %w", err)
	}
	return nil
}
