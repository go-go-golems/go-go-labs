package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
)

type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type OllamaClient struct {
	BaseURL string
	Model   string
	logger  zerolog.Logger
}

func NewOllamaClient(baseURL, model string, logger zerolog.Logger) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		logger:  logger,
	}
}

func (c *OllamaClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	c.logger.Debug().Str("text", text).Msg("getting embedding")
	
	reqData := OllamaEmbeddingRequest{
		Model:  c.Model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	url := fmt.Sprintf("%s/api/embeddings", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	c.logger.Debug().Int("embedding_size", len(response.Embedding)).Msg("got embedding")
	return response.Embedding, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func registerSQLiteFunctions(db *sql.DB) error {
	sqliteConn, err := db.Conn(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to get connection")
	}
	defer sqliteConn.Close()

	return sqliteConn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*sqlite3.SQLiteConn)

		// Register cosine similarity function
		return conn.RegisterFunc("cosine_similarity", func(a, b string) float64 {
			var vecA, vecB []float64
			
			if err := json.Unmarshal([]byte(a), &vecA); err != nil {
				return 0
			}
			if err := json.Unmarshal([]byte(b), &vecB); err != nil {
				return 0
			}

			return cosineSimilarity(vecA, vecB)
		}, true)
	})
}

func setupDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		embedding TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_documents_content ON documents(content);
	`

	_, err := db.Exec(schema)
	return errors.Wrap(err, "failed to create schema")
}

func insertDocument(db *sql.DB, content string, embedding []float64) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return errors.Wrap(err, "failed to marshal embedding")
	}

	_, err = db.Exec("INSERT INTO documents (content, embedding) VALUES (?, ?)", content, string(embeddingJSON))
	return errors.Wrap(err, "failed to insert document")
}

func searchSimilarDocuments(db *sql.DB, queryEmbedding []float64, limit int) ([]struct {
	ID         int     `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}, error) {
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal query embedding")
	}

	query := `
	SELECT id, content, cosine_similarity(embedding, ?) as similarity
	FROM documents
	ORDER BY similarity DESC
	LIMIT ?
	`

	rows, err := db.Query(query, string(queryEmbeddingJSON), limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search query")
	}
	defer rows.Close()

	var results []struct {
		ID         int     `json:"id"`
		Content    string  `json:"content"`
		Similarity float64 `json:"similarity"`
	}

	for rows.Next() {
		var result struct {
			ID         int     `json:"id"`
			Content    string  `json:"content"`
			Similarity float64 `json:"similarity"`
		}
		if err := rows.Scan(&result.ID, &result.Content, &result.Similarity); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		results = append(results, result)
	}

	return results, nil
}

func main() {
	var logLevel string
	var ollamaURL string
	var ollamaModel string
	var dbPath string

	rootCmd := &cobra.Command{
		Use:   "sqlite-vector-search",
		Short: "SQLite vector search using Ollama embeddings",
		Run: func(cmd *cobra.Command, args []string) {
			// Setup logging
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Fatal(err)
			}
			logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)

			ctx := context.Background()

			// Initialize Ollama client
			ollama := NewOllamaClient(ollamaURL, ollamaModel, logger)

			// Open SQLite database
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			// Setup database schema
			if err := setupDatabase(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to setup database")
			}

			logger.Info().Msg("SQLite vector search initialized")

			// Sample documents
			sampleDocs := []string{
				"The quick brown fox jumps over the lazy dog",
				"Machine learning is a subset of artificial intelligence",
				"SQLite is a lightweight database engine",
				"Vector databases are useful for similarity search",
				"Natural language processing involves understanding text",
				"The cat sat on the mat",
				"Deep learning uses neural networks with multiple layers",
				"Embeddings capture semantic meaning of text",
			}

			// Insert sample documents
			logger.Info().Msg("inserting sample documents")
			for _, doc := range sampleDocs {
				embedding, err := ollama.GetEmbedding(ctx, doc)
				if err != nil {
					logger.Error().Err(err).Str("doc", doc).Msg("failed to get embedding")
					continue
				}

				if err := insertDocument(db, doc, embedding); err != nil {
					logger.Error().Err(err).Str("doc", doc).Msg("failed to insert document")
					continue
				}
				logger.Debug().Str("doc", doc).Msg("inserted document")
			}

			// Perform search
			query := "artificial intelligence and neural networks"
			logger.Info().Str("query", query).Msg("searching for similar documents")

			queryEmbedding, err := ollama.GetEmbedding(ctx, query)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to get query embedding")
			}

			results, err := searchSimilarDocuments(db, queryEmbedding, 5)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to search documents")
			}

			logger.Info().Int("count", len(results)).Msg("search completed")

			fmt.Printf("\nSearch results for: %s\n", query)
			fmt.Println(strings.Repeat("=", 60))
			for i, result := range results {
				fmt.Printf("%d. [%.4f] %s\n", i+1, result.Similarity, result.Content)
			}
		},
	}

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&ollamaURL, "ollama-url", "http://127.0.0.1:11434", "Ollama server URL")
	rootCmd.Flags().StringVar(&ollamaModel, "ollama-model", "all-minilm:latest", "Ollama model to use")
	rootCmd.Flags().StringVar(&dbPath, "db-path", "vector_search.db", "SQLite database path")

	// Add search command
	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for similar documents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Setup logging
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Fatal(err)
			}
			logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)

			ctx := context.Background()
			query := strings.Join(args, " ")

			// Initialize Ollama client
			ollama := NewOllamaClient(ollamaURL, ollamaModel, logger)

			// Open SQLite database
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			queryEmbedding, err := ollama.GetEmbedding(ctx, query)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to get query embedding")
			}

			limitStr := cmd.Flag("limit").Value.String()
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				limit = 5
			}

			results, err := searchSimilarDocuments(db, queryEmbedding, limit)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to search documents")
			}

			fmt.Printf("\nSearch results for: %s\n", query)
			fmt.Println(strings.Repeat("=", 60))
			for i, result := range results {
				fmt.Printf("%d. [%.4f] %s\n", i+1, result.Similarity, result.Content)
			}
		},
	}
	searchCmd.Flags().Int("limit", 5, "Number of results to return")

	// Add document insertion command
	addCmd := &cobra.Command{
		Use:   "add [text]",
		Short: "Add a document to the database",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Setup logging
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Fatal(err)
			}
			logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)

			ctx := context.Background()
			text := strings.Join(args, " ")

			// Initialize Ollama client
			ollama := NewOllamaClient(ollamaURL, ollamaModel, logger)

			// Open SQLite database
			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			// Setup database schema
			if err := setupDatabase(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to setup database")
			}

			embedding, err := ollama.GetEmbedding(ctx, text)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to get embedding")
			}

			if err := insertDocument(db, text, embedding); err != nil {
				logger.Fatal().Err(err).Msg("failed to insert document")
			}

			fmt.Printf("Added document: %s\n", text)
		},
	}

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(addCmd)
	addTestCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
