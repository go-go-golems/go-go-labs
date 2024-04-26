package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/milosgajdos/go-embeddings"
	"github.com/milosgajdos/go-embeddings/openai"
	"github.com/rs/zerolog/log"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
)

// #cgo LDFLAGS: -L../../../thirdparty/sqlite-vss-libs/ -Wl,-undefined,dynamic_lookup
import "C"

type Embedder struct {
	db       *sql.DB
	embedder embeddings.Embedder[*openai.EmbeddingRequest]
}

func (e *Embedder) indexDocument(title string, body string) error {
	ctx := context.Background()

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
		Model:          "text-embedding-3-small",
		EncodingFormat: "float",
		Dims:           128,
	}

	embeddings_, err := e.embedder.Embed(ctx, req)
	if err != nil {
		return err
	}

	for _, emb := range embeddings_ {
		fmt.Printf("Embedding size: %d\n", len(emb.Vector))
	}

	if len(embeddings_) != 1 {
		return fmt.Errorf("expected 1 embedding, got %d", len(embeddings_))
	}

	// insert document and first embedding as json
	jsonVector, err := json.Marshal(embeddings_[0].Vector)
	_, err = e.db.Exec("INSERT INTO documents (body, title, embedding) VALUES (?, ?, ?)", body, title, string(jsonVector))

	// use the inserted row id to insertt he embeddings into the docs virtual table
	_, err = e.db.Exec("INSERT INTO docs(rowid, embedding) VALUES (last_insert_rowid(), ?)", string(jsonVector))

	return nil
}

func (e *Embedder) search(question string) error {
	ctx := context.Background()

	req := &openai.EmbeddingRequest{
		Input:          []string{question},
		Model:          "text-embedding-3-small",
		EncodingFormat: "float",
		Dims:           128,
	}

	embeddings_, err := e.embedder.Embed(ctx, req)
	if err != nil {
		return err
	}

	if len(embeddings_) != 1 {
		return fmt.Errorf("expected 1 embedding, got %d", len(embeddings_))
	}

	// now use the embedding to do a similarity search
	jsonVector, err := json.Marshal(embeddings_[0].Vector)
	rows, err := e.db.Query(`
		SELECT rowid, distance
		FROM docs
		WHERE vss_search(embedding, ?)
		LIMIT 5
	`, string(jsonVector))

	if err != nil {
		return err
	}

	defer rows.Close()

	// Print the search results
	fmt.Println("Search Results:")
	for rows.Next() {
		var rowid int
		var distance float64
		err := rows.Scan(&rowid, &distance)
		if err != nil {
			return err
		}
		fmt.Printf("Document ID: %d, Distance: %.4f\n", rowid, distance)
	}

	return nil
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
			embedding BLOB
		)
	`)
	if err != nil {
		return err
	}

	// create a virtual table with embedding(128)
	_, err = e.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS docs USING vss0(
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

func main() {
	e, err := NewEmbedder("file:test.db")
	if err != nil {
		log.Fatal().Err(err).Msg("could not create embedder")
	}

	defer e.Close()

	fmt.Println(e.VSSVersion())
	err = e.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize embedder")
	}

	err = e.indexDocument("A big headline", "Today's news is horrible.")
	if err != nil {
		log.Fatal().Err(err).Msg("could not index document")
	}

	err = e.search("what is going on?")
	if err != nil {
		log.Fatal().Err(err).Msg("could not search")
	}

	//// Insert a document embedding
	//embedding := []float32{0.1, 0.2, 0.3}
	//embeddingJSON, err := json.Marshal(embedding)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//_, err = db.Exec("INSERT INTO docs(embedding) VALUES (?)", string(embeddingJSON))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Perform a similarity search
	//query := []float32{0.15, 0.25, 0.35}
	//queryJSON, err := json.Marshal(query)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//rows, err := db.Query(`
	//	SELECT rowid, distance
	//	FROM docs
	//	WHERE vss_search(embedding, ?)
	//	LIMIT 5
	//`, string(queryJSON))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer rows.Close()
	//
	//// Print the search results
	//fmt.Println("Search Results:")
	//for rows.Next() {
	//	var rowid int
	//	var distance float64
	//	err := rows.Scan(&rowid, &distance)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	fmt.Printf("Document ID: %d, Distance: %.4f\n", rowid, distance)
	//}
}
