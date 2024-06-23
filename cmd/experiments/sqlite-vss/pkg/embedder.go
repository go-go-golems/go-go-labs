package pkg

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/milosgajdos/go-embeddings"
	"github.com/milosgajdos/go-embeddings/openai"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"time"
)

type Embedder struct {
	db       *sql.DB
	embedder embeddings.Embedder[*openai.EmbeddingRequest]
}

const EmbeddingDimensions = 128
const EmbeddingModel = "text-embedding-3-small"

func (e *Embedder) IndexDocuments(filenames []string) error {
	for _, filename := range filenames {
		// Read the file content
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read file: %s", filename)
			continue
		}

		// Extract the title from the filename
		title := filepath.Base(filename)

		// Get the file modification timestamp
		fileInfo, err := os.Stat(filename)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get file info: %s", filename)
			continue
		}
		modifiedAt := fileInfo.ModTime()

		// Index the document
		err = e.IndexDocument(context.Background(), title, string(content), modifiedAt)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to index document: %s", filename)
			continue
		}

		log.Info().Msgf("Indexed document: %s", filename)
	}

	return nil
}

func (e *Embedder) IndexDocument(ctx context.Context, title string, body string, modifiedAt time.Time) error {
	// check if the title already exists
	var count int
	err := e.db.QueryRow("SELECT COUNT(*) FROM documents WHERE title = ?", title).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Info().Msgf("Document with title %s already exists", title)
		return nil
	}

	req := &openai.EmbeddingRequest{
		Input:          []string{title + body},
		Model:          EmbeddingModel,
		EncodingFormat: "float",
		Dims:           EmbeddingDimensions,
	}

	embeddings_, err := e.embedder.Embed(ctx, req)
	if err != nil {
		return err
	}

	if len(embeddings_) != 1 {
		return errors.Errorf("expected 1 embedding, got %d", len(embeddings_))
	}

	// insert document and first embedding as json
	jsonVector, err := json.Marshal(embeddings_[0].Vector)
	if err != nil {
		return err
	}
	_, err = e.db.Exec("INSERT INTO documents (body, title, embedding, dimensions, model, modified_at) VALUES (?, ?, ?, ?, ?, ?)", body, title, string(jsonVector), EmbeddingDimensions, EmbeddingModel, modifiedAt)
	if err != nil {
		return err
	}

	// use the inserted row id to insertt he embeddings into the embeddings virtual table
	_, err = e.db.Exec("INSERT INTO embeddings(rowid, embedding) VALUES (last_insert_rowid(), ?)", string(jsonVector))
	if err != nil {
		return err
	}

	return nil
}

func (e *Embedder) IndexHelpSystem(ctx context.Context, helpSystem *help.HelpSystem) error {
	for _, section := range helpSystem.Sections {
		title := section.Title
		if section.SubTitle != "" {
			title += " - " + section.SubTitle
		}
		err := e.IndexDocument(
			ctx,
			title,
			section.Short+"\n\n"+section.Content,
			time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

type SearchResult struct {
	ID       int
	Distance float64
	Title    string
	Body     string
}

func (e *Embedder) Search(ctx context.Context, question string) ([]SearchResult, error) {
	req := &openai.EmbeddingRequest{
		Input:          []string{question},
		Model:          EmbeddingModel,
		EncodingFormat: "float",
		Dims:           EmbeddingDimensions,
	}

	embeddings_, err := e.embedder.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(embeddings_) != 1 {
		return nil, errors.Errorf("expected 1 embedding, got %d", len(embeddings_))
	}

	jsonVector, err := json.Marshal(embeddings_[0].Vector)
	if err != nil {
		return nil, err
	}

	rows, err := e.db.Query(`
		WITH similar_documents AS (
			SELECT rowid, distance
			FROM embeddings
			WHERE vss_search(embedding, ?)
			LIMIT 5
		)
		SELECT d.id, similar_documents.distance, d.title, d.body
		FROM similar_documents
		JOIN documents d ON d.id = similar_documents.rowid
		ORDER BY similar_documents.distance ASC
	`, string(jsonVector))
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(&result.ID, &result.Distance, &result.Title, &result.Body)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func NewEmbedder(f string) (*Embedder, error) {
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		return nil, err
	}
	c := openai.NewEmbedder()
	return &Embedder{db: db, embedder: c}, nil
}

func (e *Embedder) Init() error {
	// create a documents table with a body and title and embedding column if not exists
	_, err := e.db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id INTEGER PRIMARY KEY,
			body TEXT,
			title TEXT,
			embedding BLOB,
			dimensions INTEGER,
			model TEXT,
			indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			modified_at TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// create a virtual table with embedding(128)
	_, err = e.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS embeddings USING vss0(
			embedding(128)
		)
	`)

	if err != nil {
		return err
	}

	return nil
}

func (e *Embedder) VSSVersion() string {
	var version string
	_ = e.db.QueryRow("SELECT vss_version()").Scan(&version)
	return version
}

func (e *Embedder) Close() {
	_ = e.db.Close()
}
