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

// AnalysisNoteWithRefs combines note with referenced commit and file
type AnalysisNoteWithRefs struct {
	AnalysisNote
	Commit *Commit `json:"commit,omitempty"`
	File   *File   `json:"file,omitempty"`
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

// PRChangelogWithRefs combines changelog entry with referenced commit and file
type PRChangelogWithRefs struct {
	PRChangelog
	Commit *Commit `json:"commit,omitempty"`
	File   *File   `json:"file,omitempty"`
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
	Changelog []PRChangelogWithRefs  `json:"changelog"`
	Notes     []AnalysisNoteWithRefs `json:"notes"`
}

// CommitWithRefs includes PR associations and related info
type CommitWithRefsAndPRs struct {
	CommitWithFiles
	PRAssociations []PRAssociation `json:"pr_associations,omitempty"`
	Notes          []AnalysisNote  `json:"notes,omitempty"`
}

// PRAssociation links a commit to PRs
type PRAssociation struct {
	PRID   int64  `json:"pr_id"`
	PRName string `json:"pr_name"`
	Action string `json:"action"`
}

// SymbolHistory tracks when a symbol was modified
type SymbolHistory struct {
	SymbolName string   `json:"symbol_name"`
	SymbolKind string   `json:"symbol_kind"`
	FilePath   string   `json:"file_path"`
	Commits    []Commit `json:"commits"`
}

// FileWithHistory combines file with its commit history
type FileWithHistory struct {
	File
	CommitCount    int               `json:"commit_count"`
	RecentCommits  []Commit          `json:"recent_commits"`
	RelatedFiles   []RelatedFile     `json:"related_files,omitempty"`
	PRReferences   []PRReference     `json:"pr_references,omitempty"`
	Notes          []AnalysisNote    `json:"notes,omitempty"`
}

// PRReference represents a PR that references a file
type PRReference struct {
	PRID      int64  `json:"pr_id"`
	PRName    string `json:"pr_name"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	CreatedAt string `json:"created_at"`
}

// RelatedFile represents files often changed together
type RelatedFile struct {
	File
	ChangeCount int `json:"change_count"`
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

// GetPRByID retrieves a PR with its changelog and notes (with cross-references)
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

	// Get changelog with commit and file references
	changelogQuery := `
		SELECT pcl.id, pcl.pr_id, pcl.commit_id, pcl.file_id, pcl.action, pcl.details, pcl.created_at,
		       c.id, c.hash, c.subject, c.author_name, c.committed_at,
		       f.id, f.path
		FROM pr_changelog pcl
		LEFT JOIN commits c ON pcl.commit_id = c.id
		LEFT JOIN files f ON pcl.file_id = f.id
		WHERE pcl.pr_id = ?
		ORDER BY pcl.created_at DESC
	`
	rows, err := db.conn.Query(changelogQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changelog []PRChangelogWithRefs
	for rows.Next() {
		var entry PRChangelogWithRefs
		var prID, commitID, fileID sql.NullInt64
		var details sql.NullString
		var cID sql.NullInt64
		var cHash, cSubject, cAuthor, cCommittedAt sql.NullString
		var fID sql.NullInt64
		var fPath sql.NullString
		
		err := rows.Scan(&entry.ID, &prID, &commitID, &fileID, &entry.Action, &details, &entry.CreatedAt,
			&cID, &cHash, &cSubject, &cAuthor, &cCommittedAt,
			&fID, &fPath)
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
		
		// Add commit reference
		if cID.Valid {
			entry.Commit = &Commit{
				ID:          cID.Int64,
				Hash:        cHash.String,
				Subject:     cSubject.String,
				AuthorName:  cAuthor.String,
				CommittedAt: cCommittedAt.String,
			}
		}
		
		// Add file reference
		if fID.Valid {
			entry.File = &File{
				ID:   fID.Int64,
				Path: fPath.String,
			}
		}
		
		changelog = append(changelog, entry)
	}

	// Get related notes with commit and file references
	notesQuery := `
		SELECT an.id, an.commit_id, an.file_id, an.note_type, an.note, an.tags, an.created_at,
		       c.id, c.hash, c.subject, c.author_name, c.committed_at,
		       f.id, f.path
		FROM analysis_notes an
		LEFT JOIN commits c ON an.commit_id = c.id
		LEFT JOIN files f ON an.file_id = f.id
		WHERE an.tags LIKE ?
		ORDER BY an.created_at DESC
	`
	notesRows, err := db.conn.Query(notesQuery, "%"+pr.Name+"%")
	if err != nil {
		return nil, err
	}
	defer notesRows.Close()

	var notes []AnalysisNoteWithRefs
	for notesRows.Next() {
		var note AnalysisNoteWithRefs
		var commitID, fileID sql.NullInt64
		var tags sql.NullString
		var cID sql.NullInt64
		var cHash, cSubject, cAuthor, cCommittedAt sql.NullString
		var fID sql.NullInt64
		var fPath sql.NullString
		
		err := notesRows.Scan(&note.ID, &commitID, &fileID, &note.NoteType, &note.Note, &tags, &note.CreatedAt,
			&cID, &cHash, &cSubject, &cAuthor, &cCommittedAt,
			&fID, &fPath)
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
		
		// Add commit reference
		if cID.Valid {
			note.Commit = &Commit{
				ID:          cID.Int64,
				Hash:        cHash.String,
				Subject:     cSubject.String,
				AuthorName:  cAuthor.String,
				CommittedAt: cCommittedAt.String,
			}
		}
		
		// Add file reference
		if fID.Valid {
			note.File = &File{
				ID:   fID.Int64,
				Path: fPath.String,
			}
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

// GetCommitWithPRAssociations retrieves a commit with PR associations and notes
func (db *DB) GetCommitWithPRAssociations(hash string) (*CommitWithRefsAndPRs, error) {
	commit, err := db.GetCommitByHash(hash)
	if err != nil {
		return nil, err
	}

	files, err := db.GetCommitFiles(commit.ID)
	if err != nil {
		return nil, err
	}

	// Get PR associations
	prQuery := `
		SELECT DISTINCT p.id, p.name, pcl.action
		FROM pr_changelog pcl
		JOIN prs p ON pcl.pr_id = p.id
		WHERE pcl.commit_id = ?
		ORDER BY p.name
	`
	prRows, err := db.conn.Query(prQuery, commit.ID)
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	var prAssociations []PRAssociation
	for prRows.Next() {
		var assoc PRAssociation
		err := prRows.Scan(&assoc.PRID, &assoc.PRName, &assoc.Action)
		if err != nil {
			return nil, err
		}
		prAssociations = append(prAssociations, assoc)
	}

	// Get notes for this commit
	notesQuery := `
		SELECT id, commit_id, file_id, note_type, note, tags, created_at
		FROM analysis_notes
		WHERE commit_id = ?
		ORDER BY created_at DESC
	`
	notesRows, err := db.conn.Query(notesQuery, commit.ID)
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

	return &CommitWithRefsAndPRs{
		CommitWithFiles: CommitWithFiles{
			Commit:  *commit,
			Files:   files,
		},
		PRAssociations: prAssociations,
		Notes:          notes,
	}, nil
}

// GetFileWithDetails retrieves file with history, related files, and notes
func (db *DB) GetFileWithDetails(fileID int64, limit int) (*FileWithHistory, error) {
	// Get file info
	var file File
	err := db.conn.QueryRow("SELECT id, path FROM files WHERE id = ?", fileID).Scan(&file.ID, &file.Path)
	if err != nil {
		return nil, err
	}

	// Get commit count
	var commitCount int
	err = db.conn.QueryRow(`
		SELECT COUNT(DISTINCT cf.commit_id)
		FROM commit_files cf
		WHERE cf.file_id = ?
	`, fileID).Scan(&commitCount)
	if err != nil {
		return nil, err
	}

	// Get recent commits
	commits, err := db.GetFileHistory(fileID, limit)
	if err != nil {
		return nil, err
	}

	// Extract just commits
	var recentCommits []Commit
	for _, cwf := range commits {
		recentCommits = append(recentCommits, cwf.Commit)
	}

	// Get files often changed together
	relatedQuery := `
		SELECT f.id, f.path, COUNT(DISTINCT cf1.commit_id) as change_count
		FROM commit_files cf1
		JOIN commit_files cf2 ON cf1.commit_id = cf2.commit_id AND cf2.file_id = ?
		JOIN files f ON cf1.file_id = f.id
		WHERE cf1.file_id != ?
		GROUP BY f.id, f.path
		ORDER BY change_count DESC
		LIMIT 10
	`
	relatedRows, err := db.conn.Query(relatedQuery, fileID, fileID)
	if err != nil {
		return nil, err
	}
	defer relatedRows.Close()

	var relatedFiles []RelatedFile
	for relatedRows.Next() {
		var rf RelatedFile
		err := relatedRows.Scan(&rf.ID, &rf.Path, &rf.ChangeCount)
		if err != nil {
			return nil, err
		}
		relatedFiles = append(relatedFiles, rf)
	}

	// Get PR references for this file
	prQuery := `
		SELECT DISTINCT p.id, p.name, pcl.action, pcl.details, pcl.created_at
		FROM pr_changelog pcl
		JOIN prs p ON pcl.pr_id = p.id
		WHERE pcl.file_id = ?
		ORDER BY pcl.created_at DESC
		LIMIT 20
	`
	prRows, err := db.conn.Query(prQuery, fileID)
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	var prReferences []PRReference
	for prRows.Next() {
		var pr PRReference
		err := prRows.Scan(&pr.PRID, &pr.PRName, &pr.Action, &pr.Details, &pr.CreatedAt)
		if err != nil {
			return nil, err
		}
		prReferences = append(prReferences, pr)
	}

	// Get notes for this file
	notesQuery := `
		SELECT id, commit_id, file_id, note_type, note, tags, created_at
		FROM analysis_notes
		WHERE file_id = ?
		ORDER BY created_at DESC
		LIMIT 20
	`
	notesRows, err := db.conn.Query(notesQuery, fileID)
	if err != nil {
		return nil, err
	}
	defer notesRows.Close()

	var notes []AnalysisNote
	for notesRows.Next() {
		var note AnalysisNote
		var commitID, fileIDNull sql.NullInt64
		var tags sql.NullString
		err := notesRows.Scan(&note.ID, &commitID, &fileIDNull, &note.NoteType, &note.Note, &tags, &note.CreatedAt)
		if err != nil {
			return nil, err
		}
		if commitID.Valid {
			val := commitID.Int64
			note.CommitID = &val
		}
		if fileIDNull.Valid {
			val := fileIDNull.Int64
			note.FileID = &val
		}
		if tags.Valid {
			note.Tags = tags.String
		}
		notes = append(notes, note)
	}

	return &FileWithHistory{
		File:          file,
		CommitCount:   commitCount,
		RecentCommits: recentCommits,
		RelatedFiles:  relatedFiles,
		PRReferences:  prReferences,
		Notes:         notes,
	}, nil
}

// GetSymbolHistory retrieves the history of a specific symbol
func (db *DB) GetSymbolHistory(symbolName string, limit int) ([]SymbolHistory, error) {
	query := `
		SELECT DISTINCT cs.symbol_name, cs.symbol_kind, f.path
		FROM commit_symbols cs
		JOIN files f ON cs.file_id = f.id
		WHERE cs.symbol_name = ?
		ORDER BY f.path
	`
	rows, err := db.conn.Query(query, symbolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []SymbolHistory
	for rows.Next() {
		var sh SymbolHistory
		err := rows.Scan(&sh.SymbolName, &sh.SymbolKind, &sh.FilePath)
		if err != nil {
			return nil, err
		}

		// Get commits for this symbol in this file
		commitQuery := `
			SELECT c.id, c.hash, c.parents, c.author_name, c.author_email, c.authored_at,
			       c.committer_name, c.committer_email, c.committed_at, c.subject, c.body, c.document_summary
			FROM commits c
			JOIN commit_symbols cs ON c.id = cs.commit_id
			JOIN files f ON cs.file_id = f.id
			WHERE cs.symbol_name = ? AND f.path = ?
			ORDER BY c.committed_at DESC
			LIMIT ?
		`
		commitRows, err := db.conn.Query(commitQuery, symbolName, sh.FilePath, limit)
		if err != nil {
			return nil, err
		}

		var commits []Commit
		for commitRows.Next() {
			var c Commit
			var docSummary sql.NullString
			err := commitRows.Scan(
				&c.ID, &c.Hash, &c.Parents, &c.AuthorName, &c.AuthorEmail, &c.AuthoredAt,
				&c.CommitterName, &c.CommitterEmail, &c.CommittedAt, &c.Subject, &c.Body, &docSummary,
			)
			if err != nil {
				commitRows.Close()
				return nil, err
			}
			if docSummary.Valid {
				c.DocumentSummary = json.RawMessage(docSummary.String)
			}
			commits = append(commits, c)
		}
		commitRows.Close()

		sh.Commits = commits
		histories = append(histories, sh)
	}

	return histories, nil
}

// SearchSymbols searches for symbols by name pattern
func (db *DB) SearchSymbols(pattern string, limit int) ([]CommitSymbol, error) {
	query := `
		SELECT DISTINCT symbol_name, symbol_kind, file_id
		FROM commit_symbols
		WHERE symbol_name LIKE ?
		ORDER BY symbol_name
		LIMIT ?
	`
	rows, err := db.conn.Query(query, "%"+pattern+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []CommitSymbol
	for rows.Next() {
		var s CommitSymbol
		err := rows.Scan(&s.SymbolName, &s.SymbolKind, &s.FileID)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}

	return symbols, rows.Err()
}

