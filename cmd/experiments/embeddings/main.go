package main

import (
	"context"
	"fmt"
	"github.com/milosgajdos/go-embeddings/openai"
	"github.com/philippgille/chromem-go"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/milosgajdos/go-embeddings"
)

// #cgo LDFLAGS: -Lvendor/sqlite-vss-libs/ -Wl,-undefined,dynamic_lookup
import "C"

func main() {
	testChromem()
}

func testEmbedding() {
	c := openai.NewEmbedder()
	ctx := context.Background()

	input := "The lazy brown dog jumped over the fox fence."

	req := &openai.EmbeddingRequest{
		Input: []string{input},
		//Model:          "text-embedding-ada-002",
		//Model:          "text-embedding-3-large",
		Model:          "text-embedding-3-small",
		User:           "",
		EncodingFormat: "float",
		Dims:           512,
	}

	embeddings, err := c.Embed(ctx, req)
	if err != nil {
		panic(err)
	}

	for _, emb := range embeddings {
		fmt.Printf("Embedding size: %d\n", len(emb.Vector))
	}
}

func testChromem() {
	ctx := context.Background()

	// Set up chromem-go in-memory, for easy prototyping. Can add persistence easily!
	// We call it DB instead of client because there's no client-server separation. The DB is embedded.
	db := chromem.NewDB()

	// Create collection. GetCollection, GetOrCreateCollection, DeleteCollection also available!
	collection, _ := db.CreateCollection(
		"all-my-documents",
		nil,
		nil,
	)

	// Add docs to the collection. Update and delete will be added in the future.

	// Can be multi-threaded with AddConcurrently()!
	// We're showing the Chroma-like method here, but more Go-idiomatic methods are also available!
	_ = collection.Add(ctx,
		[]string{"doc1", "doc2"}, // unique ID for each doc
		nil,                      // We handle embedding automatically. You can skip that and add your own embeddings as well.
		[]map[string]string{{"source": "notion"}, {"source": "google-docs"}}, // Filter on these!
		[]string{"This is document1", "This is document2"},
	)

	// Query/search 2 most similar results. Getting by ID will be added in the future.
	results, _ := collection.Query(ctx,
		"This is a query document",
		2,
		map[string]string{"metadata_field": "is_equal_to_this"}, // optional filter
		map[string]string{"$contains": "search_string"},         // optional filter
	)

	for _, result := range results {
		fmt.Println(result)
	}
}
