package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/rs/zerolog"
	_ "github.com/mattn/go-sqlite3"
)

// runEmbeddingDemo shows practical usage of the get_embedding SQL function
func runEmbeddingDemo() {
	// Setup logging
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	// Initialize Ollama client
	ollama := NewOllamaClient("http://127.0.0.1:11434", "all-minilm:latest", logger)

	// Open SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register custom functions including get_embedding
	if err := registerSQLiteFunctions(db, ollama); err != nil {
		log.Fatal(err)
	}

	// Setup database schema
	if err := setupDatabase(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("🚀 SQLite Embedding Function Demo")
	fmt.Println("==================================")

	// Demo 1: Insert documents with computed embeddings
	fmt.Println("\n📝 Demo 1: Inserting documents with computed embeddings")
	
	documents := []string{
		"Machine learning is revolutionizing technology",
		"SQLite is a lightweight database engine",
		"Vector databases enable semantic search",
		"Natural language processing understands human language",
		"Deep learning uses neural networks",
	}

	for i, doc := range documents {
		_, err := db.Exec(`
			INSERT INTO documents (content, embedding) 
			VALUES (?, get_embedding(?))
		`, doc, doc)
		
		if err != nil {
			fmt.Printf("❌ Failed to insert document %d: %v\n", i+1, err)
			continue
		}
		
		fmt.Printf("✅ Inserted: %s\n", doc)
	}

	// Demo 2: Real-time similarity search
	fmt.Println("\n🔍 Demo 2: Real-time similarity search using get_embedding")
	
	query := "artificial intelligence and neural networks"
	fmt.Printf("Searching for: %s\n", query)
	
	rows, err := db.Query(`
		SELECT 
			content,
			cosine_similarity(embedding, get_embedding(?)) as similarity
		FROM documents
		WHERE cosine_similarity(embedding, get_embedding(?)) > 0.1
		ORDER BY similarity DESC
		LIMIT 3
	`, query, query)
	
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Results:")
	for rows.Next() {
		var content string
		var similarity float64
		if err := rows.Scan(&content, &similarity); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %.3f | %s\n", similarity, content)
	}

	// Demo 3: Compare similarity between arbitrary texts
	fmt.Println("\n🔗 Demo 3: Compare similarity between any two texts")
	
	comparisons := []struct {
		text1, text2 string
	}{
		{"I love programming", "Coding is fun"},
		{"Database technology", "Information storage systems"},
		{"Machine learning", "Artificial intelligence"},
		{"Cat", "Automobile"},
	}

	for _, comp := range comparisons {
		var similarity float64
		err := db.QueryRow(`
			SELECT cosine_similarity(
				get_embedding(?), 
				get_embedding(?)
			)
		`, comp.text1, comp.text2).Scan(&similarity)
		
		if err != nil {
			fmt.Printf("❌ Failed to compare: %v\n", err)
			continue
		}
		
		fmt.Printf("%.3f | '%s' vs '%s'\n", similarity, comp.text1, comp.text2)
	}

	// Demo 4: Update existing documents with embeddings
	fmt.Println("\n🔄 Demo 4: Update existing documents")
	
	// First, insert a document without embedding
	_, err = db.Exec("INSERT INTO documents (content, embedding) VALUES (?, '')", 
		"This document needs an embedding")
	if err != nil {
		log.Fatal(err)
	}

	// Update it with computed embedding
	result, err := db.Exec(`
		UPDATE documents 
		SET embedding = get_embedding(content)
		WHERE embedding = '' OR embedding IS NULL
	`)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Updated %d documents with computed embeddings\n", rowsAffected)

	// Demo 5: Batch processing with subqueries
	fmt.Println("\n📊 Demo 5: Batch analysis")
	
	var totalDocs, docsWithEmbeddings int
	err = db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&totalDocs)
	if err != nil {
		log.Fatal(err)
	}
	
	err = db.QueryRow(`
		SELECT COUNT(*) FROM documents 
		WHERE embedding != '' AND embedding != '[]'
	`).Scan(&docsWithEmbeddings)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("📈 Total documents: %d\n", totalDocs)
	fmt.Printf("📈 Documents with embeddings: %d\n", docsWithEmbeddings)
	fmt.Printf("📈 Coverage: %.1f%%\n", float64(docsWithEmbeddings)/float64(totalDocs)*100)

	// Demo 6: Find documents most similar to the entire collection
	fmt.Println("\n🎯 Demo 6: Find documents most similar to collection average")
	
	// This is a complex query that finds documents similar to the "average" concept
	rows, err = db.Query(`
		WITH collection_query AS (
			SELECT get_embedding('technology database programming artificial intelligence') as avg_embedding
		)
		SELECT 
			content,
			cosine_similarity(
				embedding, 
				(SELECT avg_embedding FROM collection_query)
			) as centrality_score
		FROM documents
		ORDER BY centrality_score DESC
		LIMIT 3
	`)
	
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Most central documents:")
	for rows.Next() {
		var content string
		var score float64
		if err := rows.Scan(&content, &score); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %.3f | %s\n", score, content)
	}

	fmt.Println("\n✨ Demo completed! The get_embedding function enables:")
	fmt.Println("   • Real-time similarity search without pre-computation")
	fmt.Println("   • Dynamic text comparison in SQL")
	fmt.Println("   • Batch processing with computed embeddings")
	fmt.Println("   • Complex analytical queries")
	fmt.Println("   • Seamless integration of ML with traditional SQL")
}
