package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coder/hnsw"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
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

/*** Custom SQLite Driver with Auto-Function Registration *****************/

// Global references for auto-registration on each connection
var (
	hnswModule   *HNSWModule
	globalOllama *OllamaClient
)

// Custom SQLite driver that registers functions on every connection
type customSQLiteDriver struct {
	*sqlite3.SQLiteDriver
}

func (d *customSQLiteDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.SQLiteDriver.Open(name)
	if err != nil {
		return nil, err
	}

	// Register functions on this connection
	sqliteConn := conn.(*sqlite3.SQLiteConn)
	if err := registerFunctionsOnConnection(sqliteConn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func registerFunctionsOnConnection(conn *sqlite3.SQLiteConn) error {
	// Register HNSW virtual table module (only if not already registered)
	if hnswModule == nil {
		hnswModule = &HNSWModule{}
	}
	conn.CreateModule("hnsw", hnswModule) // Ignore error if already exists

	// Register HNSW helper functions
	conn.RegisterFunc("hnsw_add", func(id int64, emb string) int {
		if hnswModule != nil && hnswModule.table != nil {
			if emb == "" || emb == "[]" {
				return 1
			}
			hnswModule.Add(int(id), []byte(emb))
		}
		return 1
	}, false)

	conn.RegisterFunc("hnsw_del", func(id int64) int {
		if hnswModule != nil {
			hnswModule.Remove(int(id))
		}
		return 1
	}, false)

	conn.RegisterFunc("hnsw_save", func() int {
		if hnswModule != nil {
			hnswModule.Save()
		}
		return 1
	}, false)

	// Register cosine similarity function
	conn.RegisterFunc("cosine_similarity", func(a, b string) float64 {
		var vecA, vecB []float64

		if err := json.Unmarshal([]byte(a), &vecA); err != nil {
			return 0
		}
		if err := json.Unmarshal([]byte(b), &vecB); err != nil {
			return 0
		}

		return cosineSimilarity(vecA, vecB)
	}, true)

	// Register embedding generation function (only if Ollama client is available)
	if globalOllama != nil {
		conn.RegisterFunc("get_embedding", func(text string) string {
			if text == "" {
				return "[]"
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			embedding, err := globalOllama.GetEmbedding(ctx, text)
			if err != nil {
				globalOllama.logger.Error().Err(err).Str("text", text).Msg("failed to get embedding in SQL function")
				return "[]"
			}

			embeddingJSON, err := json.Marshal(embedding)
			if err != nil {
				globalOllama.logger.Error().Err(err).Msg("failed to marshal embedding")
				return "[]"
			}

			return string(embeddingJSON)
		}, false)
	}

	return nil
}

/*** HNSW Virtual Table Implementation ************************************/

func mustVec(blob []byte, dim int) []float32 {
	if len(blob) == 0 {
		// Return zero vector for empty input
		return make([]float32, dim)
	}

	var f64 []float64
	err := json.Unmarshal(blob, &f64)
	if err != nil {
		// Return zero vector for invalid JSON
		return make([]float32, dim)
	}

	if len(f64) != dim {
		// Return zero vector for dimension mismatch
		return make([]float32, dim)
	}

	v := make([]float32, dim)
	for i, x := range f64 {
		v[i] = float32(x)
	}
	return v
}

type HNSWModule struct {
	table *HNSWTable
}

func (m *HNSWModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	cfg := map[string]string{
		"dim":  "384",
		"m":    "16",
		"ef":   "200",
		"path": "",
	}
	for _, a := range args[3:] {
		k, v, _ := strings.Cut(a, "=")
		cfg[k] = v
	}

	// Declare the virtual table schema
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			rowid INT,
			distance REAL
		)`, args[0]))
	if err != nil {
		return nil, err
	}

	dim, _ := strconv.Atoi(cfg["dim"])
	mInt, _ := strconv.Atoi(cfg["m"])
	ef, _ := strconv.Atoi(cfg["ef"])
	idx := hnsw.NewGraph[int]()

	// Configure the graph with parameters
	idx.M = mInt
	idx.EfSearch = ef

	if p := cfg["path"]; p != "" {
		if f, err := os.Open(p); err == nil {
			_ = idx.Import(f)
			f.Close()
		}
	}

	tab := &HNSWTable{dim: dim, idx: idx, path: cfg["path"]}
	m.table = tab
	return tab, nil
}

func (m *HNSWModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *HNSWModule) DestroyModule() {}

// Module interface requires these methods
func (m *HNSWModule) SaveModule() error {
	m.Save()
	return nil
}

func (m *HNSWModule) Add(id int, blob []byte) {
	if m.table != nil {
		m.table.add(id, blob)
	}
}

func (m *HNSWModule) Remove(id int) {
	if m.table != nil {
		m.table.remove(id)
	}
}

func (m *HNSWModule) Save() {
	if m.table != nil {
		m.table.save()
	}
}

type HNSWTable struct {
	dim, nextRowid int
	idx            *hnsw.Graph[int]
	path           string
}

func (v *HNSWTable) BestIndex(csts []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	used := make([]bool, len(csts))
	return &sqlite3.IndexResult{
		IdxNum: 0,
		IdxStr: "default",
		Used:   used,
	}, nil
}

func (*HNSWTable) Destroy() error    { return nil }
func (*HNSWTable) Disconnect() error { return nil }

func (t *HNSWTable) Open() (sqlite3.VTabCursor, error) { return &HNSWCursor{t: t}, nil }

type HNSWCursor struct {
	t         *HNSWTable
	rowids    []int
	distances []float32
	pos       int
}

func (c *HNSWCursor) Filter(idxNum int, idxStr string, vals []any) error {
	vec := mustVec(vals[0].([]byte), c.t.dim)
	k := int(vals[1].(int64))

	results := c.t.idx.Search(vec, k)
	c.rowids = make([]int, len(results))
	c.distances = make([]float32, len(results))
	for i, result := range results {
		c.rowids[i] = result.Key
		// Calculate distance manually since it's not returned
		c.distances[i] = c.t.idx.Distance(vec, result.Value)
	}
	c.pos = 0
	return nil
}

func (c *HNSWCursor) Column(ctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		ctx.ResultInt(c.rowids[c.pos])
	case 1:
		ctx.ResultDouble(float64(c.distances[c.pos]))
	}
	return nil
}

func (c *HNSWCursor) Next() error           { c.pos++; return nil }
func (c *HNSWCursor) EOF() bool             { return c.pos >= len(c.rowids) }
func (c *HNSWCursor) Rowid() (int64, error) { return int64(c.rowids[c.pos]), nil }
func (c *HNSWCursor) Close() error          { return nil }

func (t *HNSWTable) add(id int, blob []byte) {
	node := hnsw.MakeNode(id, mustVec(blob, t.dim))
	t.idx.Add(node)
}

func (t *HNSWTable) remove(id int) {
	t.idx.Delete(id)
}

func (t *HNSWTable) save() {
	if t.path == "" {
		return
	}
	f, _ := os.Create(t.path)
	t.idx.Export(f)
	f.Close()
}

// Initialize the custom SQLite driver
func init() {
	sql.Register("sqlite3_with_functions", &customSQLiteDriver{
		SQLiteDriver: &sqlite3.SQLiteDriver{},
	})
}

// Simplified function registration - now just sets global Ollama client
func registerSQLiteFunctions(db *sql.DB, ollama *OllamaClient) error {
	// Set global Ollama client for auto-registration on each connection
	globalOllama = ollama
	return nil
}

func setupDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		embedding TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_documents_content ON documents(content);

	-- HNSW metadata table for sync state tracking
	CREATE TABLE IF NOT EXISTS hnsw_meta(
		id INTEGER PRIMARY KEY CHECK(id=1),
		max_rowid_indexed INTEGER NOT NULL
	);
	INSERT OR IGNORE INTO hnsw_meta(id, max_rowid_indexed) VALUES(1, 0);

	-- Optional BLOB storage for HNSW snapshots (alternative to file storage)
	CREATE TABLE IF NOT EXISTS hnsw_snapshot(
		id INTEGER PRIMARY KEY CHECK(id=1),
		data BLOB
	);

	-- HNSW virtual table for vector search
	CREATE VIRTUAL TABLE IF NOT EXISTS vss
	USING hnsw(dim=384,m=16,ef=200,path='hnsw.idx');

	-- Updated triggers to keep HNSW index in sync and maintain watermark
	CREATE TRIGGER IF NOT EXISTS docs_ai
	AFTER INSERT ON documents BEGIN
		SELECT hnsw_add(NEW.id, NEW.embedding);
		UPDATE hnsw_meta SET max_rowid_indexed = NEW.id WHERE id = 1;
	END;

	CREATE TRIGGER IF NOT EXISTS docs_ad
	AFTER DELETE ON documents BEGIN
		SELECT hnsw_del(OLD.id);
	END;
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
	ID       int     `json:"id"`
	Content  string  `json:"content"`
	Distance float64 `json:"distance"`
}, error) {
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal query embedding")
	}

	// For now, fall back to cosine similarity search since vss_knn syntax needs special handling
	query := `
	SELECT id, content, (1 - cosine_similarity(embedding, ?1)) as distance
	FROM documents
	ORDER BY distance
	LIMIT ?2
	`

	rows, err := db.Query(query, string(queryEmbeddingJSON), limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search query")
	}
	defer rows.Close()

	var results []struct {
		ID       int     `json:"id"`
		Content  string  `json:"content"`
		Distance float64 `json:"distance"`
	}

	for rows.Next() {
		var result struct {
			ID       int     `json:"id"`
			Content  string  `json:"content"`
			Distance float64 `json:"distance"`
		}
		if err := rows.Scan(&result.ID, &result.Content, &result.Distance); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		results = append(results, result)
	}

	return results, nil
}

// jsonToVec32 converts JSON string to []float32
func jsonToVec32(jsonStr string) []float32 {
	if jsonStr == "" || jsonStr == "[]" {
		return make([]float32, 384) // Return zero vector for empty input
	}

	var f64 []float64
	err := json.Unmarshal([]byte(jsonStr), &f64)
	if err != nil {
		return make([]float32, 384) // Return zero vector for invalid JSON
	}

	if len(f64) == 0 {
		return make([]float32, 384) // Return zero vector for empty array
	}

	v := make([]float32, len(f64))
	for i, x := range f64 {
		v[i] = float32(x)
	}
	return v
}

// bootstrapIndex loads persisted HNSW graph and performs catch-up scan
func bootstrapIndex(db *sql.DB, logger zerolog.Logger) error {
	logger.Info().Msg("bootstrapping HNSW index")

	// Force virtual table creation by accessing it once
	_, err := db.Query("SELECT rowid, distance FROM vss WHERE 0=1 LIMIT 1")
	if err != nil {
		logger.Debug().Err(err).Msg("virtual table access failed, will try direct initialization")
	}

	if hnswModule == nil || hnswModule.table == nil {
		logger.Warn().Msg("HNSW module not initialized, skipping bootstrap (index will sync via triggers)")
		return nil
	}

	// 1. Try to load persisted graph
	path := hnswModule.table.path
	if path != "" {
		if f, err := os.Open(path); err == nil {
			logger.Info().Str("path", path).Msg("loading HNSW index from file")
			err = hnswModule.table.idx.Import(f)
			f.Close()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to load HNSW index from file")
			}
		} else {
			logger.Debug().Str("path", path).Msg("no existing HNSW index file found")
		}
	} else {
		// Try BLOB storage
		var blob []byte
		err := db.QueryRow(`SELECT data FROM hnsw_snapshot WHERE id=1`).Scan(&blob)
		if err == nil && len(blob) > 0 {
			logger.Info().Int("blob_size", len(blob)).Msg("loading HNSW index from BLOB storage")
			err = hnswModule.table.idx.Import(strings.NewReader(string(blob)))
			if err != nil {
				logger.Warn().Err(err).Msg("failed to load HNSW index from BLOB")
			}
		} else {
			logger.Debug().Msg("no existing HNSW index BLOB found")
		}
	}

	// 2. Incremental catch-up scan
	var maxRowid int64
	err = db.QueryRow(`SELECT max_rowid_indexed FROM hnsw_meta WHERE id=1`).Scan(&maxRowid)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get max rowid, starting from 0")
		maxRowid = 0
	}

	logger.Info().Int64("max_rowid_indexed", maxRowid).Msg("starting catch-up scan")

	rows, err := db.Query(`
		SELECT id, embedding
		FROM documents
		WHERE id > ?
		ORDER BY id`, maxRowid)
	if err != nil {
		return errors.Wrap(err, "failed to query documents for catch-up")
	}
	defer rows.Close()

	catchupCount := 0
	for rows.Next() {
		var id int
		var embJSON string
		if err := rows.Scan(&id, &embJSON); err != nil {
			logger.Error().Err(err).Msg("failed to scan document row")
			continue
		}

		// Add to HNSW index
		node := hnsw.MakeNode(id, jsonToVec32(embJSON))
		hnswModule.table.idx.Add(node)
		maxRowid = int64(id)
		catchupCount++
	}

	// Update the watermark
	if catchupCount > 0 {
		_, err = db.Exec(`UPDATE hnsw_meta SET max_rowid_indexed = ? WHERE id=1`, maxRowid)
		if err != nil {
			logger.Error().Err(err).Msg("failed to update max_rowid_indexed")
		}
		logger.Info().Int("count", catchupCount).Int64("new_max_rowid", maxRowid).Msg("catch-up scan completed")
	} else {
		logger.Info().Msg("index was already current, no catch-up needed")
	}

	return nil
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
			db, err := sql.Open("sqlite3_with_functions", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db, ollama); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			// Setup database schema (this creates the virtual table and initializes the HNSW module)
			if err := setupDatabase(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to setup database")
			}

			// Bootstrap HNSW index with catch-up scan (after virtual table is created)
			if err := bootstrapIndex(db, logger); err != nil {
				logger.Fatal().Err(err).Msg("failed to bootstrap HNSW index")
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
				fmt.Printf("%d. [%.4f] %s\n", i+1, result.Distance, result.Content)
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
			db, err := sql.Open("sqlite3_with_functions", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db, ollama); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			// Bootstrap HNSW index with catch-up scan
			if err := bootstrapIndex(db, logger); err != nil {
				logger.Fatal().Err(err).Msg("failed to bootstrap HNSW index")
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
				fmt.Printf("%d. [%.4f] %s\n", i+1, result.Distance, result.Content)
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
			db, err := sql.Open("sqlite3_with_functions", dbPath)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to open database")
			}
			defer db.Close()

			// Register custom functions
			if err := registerSQLiteFunctions(db, ollama); err != nil {
				logger.Fatal().Err(err).Msg("failed to register functions")
			}

			// Setup database schema
			if err := setupDatabase(db); err != nil {
				logger.Fatal().Err(err).Msg("failed to setup database")
			}

			// Bootstrap HNSW index with catch-up scan
			if err := bootstrapIndex(db, logger); err != nil {
				logger.Fatal().Err(err).Msg("failed to bootstrap HNSW index")
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

	// Add demo command for embedding function
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run embedding function demo (requires Ollama)",
		Run: func(cmd *cobra.Command, args []string) {
			runEmbeddingDemo()
		},
	}

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(demoCmd)
	addTestCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
