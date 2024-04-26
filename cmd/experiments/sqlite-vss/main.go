package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/milosgajdos/go-embeddings"
	"github.com/milosgajdos/go-embeddings/openai"
	"log"

	_ "github.com/asg017/sqlite-vss/bindings/go"
	_ "github.com/mattn/go-sqlite3"
)

// #cgo LDFLAGS: -L../../../thirdparty/sqlite-vss-libs/ -Wl,-undefined,dynamic_lookup
import "C"

type Embedder struct {
	db       *sql.DB
	embedder embeddings.Embedder[*openai.EmbeddingRequest]
}

func (e *Embedder) computeEmbeddings(t string) error {
	ctx := context.Background()

	req := &openai.EmbeddingRequest{
		Input: []string{t, t},
		Model: "text-embedding-3-small",
		Dims:  128,
	}

	embeddings, err := e.embedder.Embed(ctx, req)
	if err != nil {
		return err
	}

	_ = embeddings

	for _, emb := range embeddings {
		fmt.Printf("Embedding size: %d\n", len(emb.Vector))
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
	// create a virtual table with embedding(128)
	_, err := e.db.Exec(`
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
		log.Fatal(err)
	}

	defer e.Close()

	fmt.Println(e.VSSVersion())
	err = e.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", "file:test.db?mode=memory")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	return
	// Load the sqlite-vss extension
	//err = sqlite_vss.Load(db)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Create a vss0 virtual table
	_, err = db.Exec(`
		CREATE VIRTUAL TABLE docs USING vss0(
			embedding(3)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert a document embedding
	embedding := []float32{0.1, 0.2, 0.3}
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO docs(embedding) VALUES (?)", string(embeddingJSON))
	if err != nil {
		log.Fatal(err)
	}

	// Perform a similarity search
	query := []float32{0.15, 0.25, 0.35}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(`
		SELECT rowid, distance 
		FROM docs
		WHERE vss_search(embedding, ?)
		LIMIT 5
	`, string(queryJSON))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Print the search results
	fmt.Println("Search Results:")
	for rows.Next() {
		var rowid int
		var distance float64
		err := rows.Scan(&rowid, &distance)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Document ID: %d, Distance: %.4f\n", rowid, distance)
	}
}

func main2() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create virtual table docs using vss0(embedding(3))")
	if err != nil {
		log.Fatal(err)
	}

	// Insert a document embedding
	embedding := []float32{0.1, 0.2, 0.3}
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("insert into docs(embedding) values(?)", string(embeddingJSON))
	if err != nil {
		log.Fatal(err)
	}

	// Perform a similarity search
	query := []float32{0.15, 0.25, 0.35}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("select rowid, distance from docs where vss_search(embedding, ?)", string(queryJSON))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rowid int
		var distance float32
		err := rows.Scan(&rowid, &distance)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("rowid=%d, distance=%f\n", rowid, distance)
	}
}
