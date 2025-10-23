package models

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

// Commit represents a git commit record
type Commit struct {
	ID              int64           `json:"id" db:"id"`
	Hash            string          `json:"hash" db:"hash"`
	Parents         string          `json:"parents" db:"parents"`
	AuthorName      string          `json:"author_name" db:"author_name"`
	AuthorEmail     string          `json:"author_email" db:"author_email"`
	AuthoredAt      string          `json:"authored_at" db:"authored_at"`
	CommitterName   string          `json:"committer_name" db:"committer_name"`
	CommitterEmail  string          `json:"committer_email" db:"committer_email"`
	CommittedAt     string          `json:"committed_at" db:"committed_at"`
	Subject         string          `json:"subject" db:"subject"`
	Body            string          `json:"body" db:"body"`
	DocumentSummary json.RawMessage `json:"document_summary" db:"document_summary"`
}

// File represents a file tracked in the repository
type File struct {
	ID   int64  `json:"id" db:"id"`
	Path string `json:"path" db:"path"`
}

// CommitFile represents the relationship between commits and files
type CommitFile struct {
	CommitID   int64  `json:"commit_id" db:"commit_id"`
	FileID     int64  `json:"file_id" db:"file_id"`
	ChangeType string `json:"change_type" db:"change_type"`
	OldPath    string `json:"old_path" db:"old_path"`
	Additions  int    `json:"additions" db:"additions"`
	Deletions  int    `json:"deletions" db:"deletions"`
}

// CommitSymbol represents symbols found in commit files
type CommitSymbol struct {
	CommitID   int64  `json:"commit_id" db:"commit_id"`
	FileID     int64  `json:"file_id" db:"file_id"`
	SymbolName string `json:"symbol_name" db:"symbol_name"`
	SymbolKind string `json:"symbol_kind" db:"symbol_kind"`
}

// AnalysisNote represents manual analysis notes
type AnalysisNote struct {
	ID        int64  `json:"id" db:"id"`
	CommitID  *int64 `json:"commit_id" db:"commit_id"`
	FileID    *int64 `json:"file_id" db:"file_id"`
	NoteType  string `json:"note_type" db:"note_type"`
	Note      string `json:"note" db:"note"`
	Tags      string `json:"tags" db:"tags"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

// PR represents a pull request slice
type PR struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Status      string `json:"status" db:"status"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}

// PRChangelog represents actions taken for a PR
type PRChangelog struct {
	ID        int64  `json:"id" db:"id"`
	PRID      *int64 `json:"pr_id" db:"pr_id"`
	CommitID  *int64 `json:"commit_id" db:"commit_id"`
	FileID    *int64 `json:"file_id" db:"file_id"`
	Action    string `json:"action" db:"action"`
	Details   string `json:"details" db:"details"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

// CommitWithFiles combines commit info with its changed files
type CommitWithFiles struct {
	Commit
	Files []FileChange `json:"files"`
}

// FileChange combines file and commit_file information
type FileChange struct {
	FileID     int64  `json:"file_id"`
	Path       string `json:"path"`
	ChangeType string `json:"change_type"`
	OldPath    string `json:"old_path,omitempty"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
}

// PRWithDetails combines PR info with related commits and notes
type PRWithDetails struct {
	PR
	Changelog []PRChangelog  `json:"changelog"`
	Notes     []AnalysisNote `json:"notes"`
}

// DB wraps the sqlite database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetCommits retrieves commits with pagination and optional filtering
func (db *DB) GetCommits(limit, offset int, search string) ([]Commit, error) {
	query := `
		SELECT id, hash, parents, author_name, author_email, authored_at,
		       committer_name, committer_email, committed_at, subject, body, document_summary
		FROM commits
		WHERE 1=1
	`
	args := []interface{}{}

	if search != "" {
		query += ` AND (subject LIKE ? OR body LIKE ? OR hash LIKE ?)`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query += ` ORDER BY committed_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []Commit
	for rows.Next() {
		var c Commit
		var docSummary sql.NullString
		err := rows.Scan(
			&c.ID, &c.Hash, &c.Parents, &c.AuthorName, &c.AuthorEmail, &c.AuthoredAt,
			&c.CommitterName, &c.CommitterEmail, &c.CommittedAt, &c.Subject, &c.Body, &docSummary,
		)
		if err != nil {
			return nil, err
		}
		if docSummary.Valid {
			c.DocumentSummary = json.RawMessage(docSummary.String)
		}
		commits = append(commits, c)
	}

	return commits, rows.Err()
}

// GetCommitByHash retrieves a single commit by its hash
func (db *DB) GetCommitByHash(hash string) (*Commit, error) {
	query := `
		SELECT id, hash, parents, author_name, author_email, authored_at,
		       committer_name, committer_email, committed_at, subject, body, document_summary
		FROM commits
		WHERE hash LIKE ?
	`

	var c Commit
	var docSummary sql.NullString
	err := db.conn.QueryRow(query, hash+"%").Scan(
		&c.ID, &c.Hash, &c.Parents, &c.AuthorName, &c.AuthorEmail, &c.AuthoredAt,
		&c.CommitterName, &c.CommitterEmail, &c.CommittedAt, &c.Subject, &c.Body, &docSummary,
	)
	if err != nil {
		return nil, err
	}
	if docSummary.Valid {
		c.DocumentSummary = json.RawMessage(docSummary.String)
	}

	return &c, nil
}

// GetCommitFiles retrieves files changed in a commit
func (db *DB) GetCommitFiles(commitID int64) ([]FileChange, error) {
	query := `
		SELECT f.id, f.path, cf.change_type, cf.old_path, cf.additions, cf.deletions
		FROM commit_files cf
		JOIN files f ON cf.file_id = f.id
		WHERE cf.commit_id = ?
		ORDER BY f.path
	`

	rows, err := db.conn.Query(query, commitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileChange
	for rows.Next() {
		var fc FileChange
		var oldPath sql.NullString
		err := rows.Scan(&fc.FileID, &fc.Path, &fc.ChangeType, &oldPath, &fc.Additions, &fc.Deletions)
		if err != nil {
			return nil, err
		}
		if oldPath.Valid {
			fc.OldPath = oldPath.String
		}
		files = append(files, fc)
	}

	return files, rows.Err()
}

// GetCommitSymbols retrieves symbols in a commit
func (db *DB) GetCommitSymbols(commitID int64) ([]CommitSymbol, error) {
	query := `
		SELECT commit_id, file_id, symbol_name, symbol_kind
		FROM commit_symbols
		WHERE commit_id = ?
		ORDER BY symbol_name
	`

	rows, err := db.conn.Query(query, commitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []CommitSymbol
	for rows.Next() {
		var s CommitSymbol
		err := rows.Scan(&s.CommitID, &s.FileID, &s.SymbolName, &s.SymbolKind)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}

	return symbols, rows.Err()
}

// GetPRs retrieves all PRs
func (db *DB) GetPRs() ([]PR, error) {
	query := `
		SELECT id, name, description, status, created_at, updated_at
		FROM prs
		ORDER BY updated_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []PR
	for rows.Next() {
		var pr PR
		var updatedAt sql.NullString
		err := rows.Scan(&pr.ID, &pr.Name, &pr.Description, &pr.Status, &pr.CreatedAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			pr.UpdatedAt = updatedAt.String
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

// GetPRByID retrieves a PR with its changelog and notes
func (db *DB) GetPRByID(id int64) (*PRWithDetails, error) {
	// Get PR
	var pr PR
	var updatedAt sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, name, description, status, created_at, updated_at
		FROM prs WHERE id = ?
	`, id).Scan(&pr.ID, &pr.Name, &pr.Description, &pr.Status, &pr.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if updatedAt.Valid {
		pr.UpdatedAt = updatedAt.String
	}

	// Get changelog
	changelogQuery := `
		SELECT id, pr_id, commit_id, file_id, action, details, created_at
		FROM pr_changelog
		WHERE pr_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.conn.Query(changelogQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changelog []PRChangelog
	for rows.Next() {
		var entry PRChangelog
		var prID, commitID, fileID sql.NullInt64
		var details sql.NullString
		err := rows.Scan(&entry.ID, &prID, &commitID, &fileID, &entry.Action, &details, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		if prID.Valid {
			val := prID.Int64
			entry.PRID = &val
		}
		if commitID.Valid {
			val := commitID.Int64
			entry.CommitID = &val
		}
		if fileID.Valid {
			val := fileID.Int64
			entry.FileID = &val
		}
		if details.Valid {
			entry.Details = details.String
		}
		changelog = append(changelog, entry)
	}

	// Get related notes (tagged with PR name or mentioned in details)
	notesQuery := `
		SELECT id, commit_id, file_id, note_type, note, tags, created_at
		FROM analysis_notes
		WHERE tags LIKE ?
		ORDER BY created_at DESC
	`
	notesRows, err := db.conn.Query(notesQuery, "%"+pr.Name+"%")
	if err != nil {
		return nil, err
	}
	defer notesRows.Close()

	var notes []AnalysisNote
	for notesRows.Next() {
		var note AnalysisNote
		var commitID, fileID sql.NullInt64
		var tags sql.NullString
		err := notesRows.Scan(&note.ID, &commitID, &fileID, &note.NoteType, &note.Note, &tags, &note.CreatedAt)
		if err != nil {
			return nil, err
		}
		if commitID.Valid {
			val := commitID.Int64
			note.CommitID = &val
		}
		if fileID.Valid {
			val := fileID.Int64
			note.FileID = &val
		}
		if tags.Valid {
			note.Tags = tags.String
		}
		notes = append(notes, note)
	}

	return &PRWithDetails{
		PR:        pr,
		Changelog: changelog,
		Notes:     notes,
	}, nil
}

// GetFiles retrieves files with optional path filtering
func (db *DB) GetFiles(pathPrefix string, limit, offset int) ([]File, error) {
	query := `SELECT id, path FROM files WHERE 1=1`
	args := []interface{}{}

	if pathPrefix != "" {
		query += ` AND path LIKE ?`
		args = append(args, pathPrefix+"%")
	}

	query += ` ORDER BY path LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		err := rows.Scan(&f.ID, &f.Path)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, rows.Err()
}

// GetFileHistory retrieves commits that modified a specific file
func (db *DB) GetFileHistory(fileID int64, limit int) ([]CommitWithFiles, error) {
	query := `
		SELECT c.id, c.hash, c.parents, c.author_name, c.author_email, c.authored_at,
		       c.committer_name, c.committer_email, c.committed_at, c.subject, c.body, c.document_summary
		FROM commits c
		JOIN commit_files cf ON c.id = cf.commit_id
		WHERE cf.file_id = ?
		ORDER BY c.committed_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, fileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []CommitWithFiles
	for rows.Next() {
		var c Commit
		var docSummary sql.NullString
		err := rows.Scan(
			&c.ID, &c.Hash, &c.Parents, &c.AuthorName, &c.AuthorEmail, &c.AuthoredAt,
			&c.CommitterName, &c.CommitterEmail, &c.CommittedAt, &c.Subject, &c.Body, &docSummary,
		)
		if err != nil {
			return nil, err
		}
		if docSummary.Valid {
			c.DocumentSummary = json.RawMessage(docSummary.String)
		}

		files, err := db.GetCommitFiles(c.ID)
		if err != nil {
			return nil, err
		}

		commits = append(commits, CommitWithFiles{
			Commit: c,
			Files:  files,
		})
	}

	return commits, rows.Err()
}

// GetAnalysisNotes retrieves analysis notes with optional filtering
func (db *DB) GetAnalysisNotes(noteType string, tags string, limit, offset int) ([]AnalysisNote, error) {
	query := `
		SELECT id, commit_id, file_id, note_type, note, tags, created_at
		FROM analysis_notes
		WHERE 1=1
	`
	args := []interface{}{}

	if noteType != "" {
		query += ` AND note_type = ?`
		args = append(args, noteType)
	}

	if tags != "" {
		query += ` AND tags LIKE ?`
		args = append(args, "%"+tags+"%")
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []AnalysisNote
	for rows.Next() {
		var note AnalysisNote
		var commitID, fileID sql.NullInt64
		var tags sql.NullString
		err := rows.Scan(&note.ID, &commitID, &fileID, &note.NoteType, &note.Note, &tags, &note.CreatedAt)
		if err != nil {
			return nil, err
		}
		if commitID.Valid {
			val := commitID.Int64
			note.CommitID = &val
		}
		if fileID.Valid {
			val := fileID.Int64
			note.FileID = &val
		}
		if tags.Valid {
			note.Tags = tags.String
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// GetStats returns summary statistics about the database
func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count commits
	var commitCount int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM commits").Scan(&commitCount)
	if err != nil {
		return nil, err
	}
	stats["commit_count"] = commitCount

	// Count files
	var fileCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM files").Scan(&fileCount)
	if err != nil {
		return nil, err
	}
	stats["file_count"] = fileCount

	// Count PRs
	var prCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM prs").Scan(&prCount)
	if err != nil {
		return nil, err
	}
	stats["pr_count"] = prCount

	// Count analysis notes
	var noteCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM analysis_notes").Scan(&noteCount)
	if err != nil {
		return nil, err
	}
	stats["analysis_note_count"] = noteCount

	// Get date range
	var minDate, maxDate string
	err = db.conn.QueryRow("SELECT MIN(committed_at), MAX(committed_at) FROM commits").Scan(&minDate, &maxDate)
	if err != nil {
		return nil, err
	}
	stats["earliest_commit"] = minDate
	stats["latest_commit"] = maxDate

	// Get PR status breakdown
	rows, err := db.conn.Query("SELECT status, COUNT(*) as count FROM prs GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prStatusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}
		prStatusCounts[status] = count
	}
	stats["pr_status_counts"] = prStatusCounts

	return stats, nil
}

