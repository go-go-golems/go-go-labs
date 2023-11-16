package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-go-golems/clay/pkg/watcher"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// First version
// https://chat.openai.com/share/bb5425fe-938e-4fcc-b00e-e4564426eeaf

type ChangeLog struct {
	ID              int64
	Path            string
	Action          string
	DateTime        string
	PreviousContent string
	CurrentContent  string
}

func initDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while opening SQLite database")
	}

	// Update table schema to include previous and current content columns
	statement, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS changelog (
			id INTEGER PRIMARY KEY,
			path TEXT,
			action TEXT,
			date_time TEXT,
			previous_content TEXT,
			current_content TEXT
		)
	`)
	_, err = statement.Exec()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func logChange(db *sql.DB, change ChangeLog) error {
	statement, _ := db.Prepare("INSERT INTO changelog (path, action, date_time, previous_content, current_content) VALUES (?, ?, ?, ?, ?)")
	_, err := statement.Exec(change.Path, change.Action, change.DateTime, change.PreviousContent, change.CurrentContent)
	if err != nil {
		return err
	}

	return nil
}

func getRelativePath(basePath, filePath string) (string, error) {
	relativePath, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return "", err
	}
	return relativePath, nil
}

func getLatestContent(db *sql.DB, path string) string {
	var content string
	query := `SELECT current_content FROM changelog WHERE path = ? ORDER BY id DESC LIMIT 1`
	err := db.QueryRow(query, path).Scan(&content)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed querying latest content for %s: %v\n", path, err)
	}
	return content
}

//nolint:unused
func traverseAndCapture(path string, db *sql.DB) error {
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", currentPath, err)
			return err
		}
		// TODO(manuel, 2023-11-03) Doesn't capture the initial state of the file into .history dir
		if !info.IsDir() {
			content, err := os.ReadFile(currentPath)
			if err != nil {
				log.Printf("Failed reading file %s: %v\n", currentPath, err)
				return nil
			}
			_ = logChange(db, ChangeLog{
				Path:            currentPath,
				Action:          "startup",
				DateTime:        time.Now().Format(time.RFC3339),
				PreviousContent: "", // On startup, previous content can be empty
				CurrentContent:  string(content),
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func startWatcher(paths []string, useSQLite bool, dbPath string) {
	var db *sql.DB
	var err error
	if useSQLite {
		db, err = initDB(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Error while initializing SQLite database")
		}
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
	}

	watcher := watcher.NewWatcher(
		watcher.WithPaths(paths...),
		watcher.WithWriteCallback(func(path string) error {
			currentContent, err := os.ReadFile(path)
			if err != nil {
				log.Error().Str("path", path).Err(err).Msg("Failed reading file")
				return err
			}

			// Determine destination within .history
			for _, watchPath := range paths {
				if strings.HasPrefix(path, watchPath) {
					relativePath, err := getRelativePath(watchPath, path)
					if err != nil {
						log.Error().Str("path", path).Err(err).Msg("Failed determining relative path")
						return err
					}

					historyPath := filepath.Join(".history", relativePath)
					historyDir := filepath.Dir(historyPath)
					if _, err := os.Stat(historyDir); os.IsNotExist(err) {
						err := os.MkdirAll(historyDir, 0755)
						if err != nil {
							log.Error().Str("path", historyDir).Err(err).Msg("Failed creating directory")
							return err
						}
						log.Info().Str("path", historyDir).Msg("Created directory")
					}

					timestamp := time.Now().Format("20060102150405")
					historyFile := filepath.Join(historyDir, fmt.Sprintf("%s.%s", filepath.Base(path), timestamp))
					if err := os.WriteFile(historyFile, currentContent, 0644); err != nil {
						log.Error().Str("path", historyFile).Err(err).Msg("Failed writing file")
						return err
					}
					log.Info().Str("path", historyFile).Msg("Wrote file")
					break
				}
			}

			// Log to SQLite
			if useSQLite {
				previousContent := getLatestContent(db, path)

				_ = logChange(db, ChangeLog{
					Path:            path,
					Action:          "modified",
					DateTime:        time.Now().Format(time.RFC3339),
					PreviousContent: previousContent,
					CurrentContent:  string(currentContent),
				})
			}

			return nil
		}),
		watcher.WithRemoveCallback(func(path string) error {
			// Log to SQLite
			if useSQLite {
				err := logChange(db, ChangeLog{
					Path:     path,
					Action:   "removed",
					DateTime: time.Now().Format(time.RFC3339),
				})
				if err != nil {
					return err
				}
			}

			return nil
		}),
	)

	if err := watcher.Run(context.Background()); err != nil {
		fmt.Println("Error:", err)
	}
}

var (
	pathsToWatch []string
	useSQLite    bool
	sqliteDBPath string
)

var rootCmd = &cobra.Command{
	Use:   "yourApp",
	Short: "Description of your app",
	Long:  `Detailed description of your app`,
	Run: func(cmd *cobra.Command, args []string) {
		startWatcher(pathsToWatch, useSQLite, sqliteDBPath)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}

func init() {
	rootCmd.Flags().StringArrayVarP(&pathsToWatch, "watch", "w", []string{}, "Paths to watch")
	rootCmd.Flags().BoolVarP(&useSQLite, "sqlite", "s", false, "Enable SQLite logging")
	rootCmd.Flags().StringVarP(&sqliteDBPath, "dbpath", "d", "./changes.db", "Path to SQLite database")
}
