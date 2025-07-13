package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// TestFunctions provides a CLI command to test individual SQLite functions
func createTestCommand() *cobra.Command {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test individual SQLite functions",
	}

	// Test basic function registration
	testCmd.AddCommand(&cobra.Command{
		Use:   "basic",
		Short: "Test basic function registration",
		Run: func(cmd *cobra.Command, args []string) {
			testBasicFunction()
		},
	})

	// Test cosine similarity function
	testCmd.AddCommand(&cobra.Command{
		Use:   "cosine",
		Short: "Test cosine similarity function",
		Run: func(cmd *cobra.Command, args []string) {
			testCosineSimilarity()
		},
	})

	// Test JSON functions
	testCmd.AddCommand(&cobra.Command{
		Use:   "json",
		Short: "Test JSON handling functions",
		Run: func(cmd *cobra.Command, args []string) {
			testJSONFunctions()
		},
	})

	// Test vector operations
	testCmd.AddCommand(&cobra.Command{
		Use:   "vectors",
		Short: "Test vector operations",
		Run: func(cmd *cobra.Command, args []string) {
			testVectorOperations()
		},
	})

	// Test error handling
	testCmd.AddCommand(&cobra.Command{
		Use:   "errors",
		Short: "Test error handling in functions",
		Run: func(cmd *cobra.Command, args []string) {
			testErrorHandling()
		},
	})

	// Test embedding function
	testCmd.AddCommand(&cobra.Command{
		Use:   "embedding",
		Short: "Test get_embedding function (requires Ollama)",
		Run: func(cmd *cobra.Command, args []string) {
			testEmbeddingFunction()
		},
	})

	return testCmd
}

func testBasicFunction() {
	fmt.Println("Testing basic function registration...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register simple functions
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)

		// Register multiple test functions
		if err := sqliteConn.RegisterFunc("double", func(x float64) float64 {
			return x * 2
		}, true); err != nil {
			return err
		}

		if err := sqliteConn.RegisterFunc("concat_test", func(a, b string) string {
			return a + " " + b
		}, true); err != nil {
			return err
		}

		if err := sqliteConn.RegisterFunc("is_positive", func(x float64) bool {
			return x > 0
		}, true); err != nil {
			return err
		}

		fmt.Println("Functions registered successfully")
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to register functions: %v", err)
	}

	// Test the functions using the same connection
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"Double function", "SELECT double(5.0)", "10"},
		{"Concat function", "SELECT concat_test('hello', 'world')", "hello world"},
		{"Boolean function (true)", "SELECT is_positive(5.0)", "1"},
		{"Boolean function (false)", "SELECT is_positive(-5.0)", "0"},
	}

	for _, test := range tests {
		var result string
		err := conn.QueryRowContext(context.Background(), test.query).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", test.name, err)
			continue
		}

		if result == test.expected {
			fmt.Printf("✅ %s: %s = %s\n", test.name, test.query, result)
		} else {
			fmt.Printf("❌ %s: Expected %s, got %s\n", test.name, test.expected, result)
		}
	}
}

func testCosineSimilarity() {
	fmt.Println("Testing cosine similarity function...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a mock Ollama client for testing
	logger := zerolog.Nop() // Silent logger for tests
	mockOllama := NewOllamaClient("http://127.0.0.1:11434", "all-minilm:latest", logger)

	// Register the cosine similarity function
	if err := registerSQLiteFunctions(db, mockOllama); err != nil {
		log.Fatal(err)
	}

	// Test vectors with known similarities
	tests := []struct {
		name      string
		vector1   []float64
		vector2   []float64
		expected  float64
		tolerance float64
	}{
		{"Identical vectors", []float64{1, 0, 0}, []float64{1, 0, 0}, 1.0, 0.001},
		{"Orthogonal vectors", []float64{1, 0, 0}, []float64{0, 1, 0}, 0.0, 0.001},
		{"Opposite vectors", []float64{1, 0, 0}, []float64{-1, 0, 0}, -1.0, 0.001},
		{"45-degree vectors", []float64{1, 1, 0}, []float64{1, 0, 0}, 0.707, 0.01},
		{"Similar vectors", []float64{1, 2, 3}, []float64{1, 2, 3}, 1.0, 0.001},
	}

	for _, test := range tests {
		vec1JSON, _ := json.Marshal(test.vector1)
		vec2JSON, _ := json.Marshal(test.vector2)

		var result float64
		query := "SELECT cosine_similarity(?, ?)"
		err := db.QueryRow(query, string(vec1JSON), string(vec2JSON)).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", test.name, err)
			continue
		}

		if abs(result-test.expected) < test.tolerance {
			fmt.Printf("✅ %s: %.3f (expected %.3f)\n", test.name, result, test.expected)
		} else {
			fmt.Printf("❌ %s: %.3f (expected %.3f, diff %.3f)\n",
				test.name, result, test.expected, abs(result-test.expected))
		}
	}
}

func testJSONFunctions() {
	fmt.Println("Testing JSON handling...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register JSON test functions
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)

		// Function that returns vector length
		if err := sqliteConn.RegisterFunc("vector_length", func(vectorStr string) int {
			var vector []float64
			if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
				return -1
			}
			return len(vector)
		}, true); err != nil {
			return err
		}

		// Function that returns first element
		if err := sqliteConn.RegisterFunc("first_element", func(vectorStr string) float64 {
			var vector []float64
			if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
				return 0
			}
			if len(vector) == 0 {
				return 0
			}
			return vector[0]
		}, true); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test JSON handling
	tests := []struct {
		name     string
		query    string
		args     []interface{}
		expected string
	}{
		{"Vector length", "SELECT vector_length(?)", []interface{}{"[1,2,3,4,5]"}, "5"},
		{"Empty vector length", "SELECT vector_length(?)", []interface{}{"[]"}, "0"},
		{"Invalid JSON", "SELECT vector_length(?)", []interface{}{"invalid"}, "-1"},
		{"First element", "SELECT first_element(?)", []interface{}{"[42,1,2,3]"}, "42"},
		{"First element empty", "SELECT first_element(?)", []interface{}{"[]"}, "0"},
	}

	for _, test := range tests {
		var result string
		err := conn.QueryRowContext(context.Background(), test.query, test.args...).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", test.name, err)
			continue
		}

		if result == test.expected {
			fmt.Printf("✅ %s: %s\n", test.name, result)
		} else {
			fmt.Printf("❌ %s: Expected %s, got %s\n", test.name, test.expected, result)
		}
	}
}

func testVectorOperations() {
	fmt.Println("Testing vector operations...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register vector operation functions
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)

		// Dot product
		if err := sqliteConn.RegisterFunc("dot_product", func(aStr, bStr string) float64 {
			var a, b []float64
			if err := json.Unmarshal([]byte(aStr), &a); err != nil {
				return 0
			}
			if err := json.Unmarshal([]byte(bStr), &b); err != nil {
				return 0
			}
			if len(a) != len(b) {
				return 0
			}

			var sum float64
			for i := 0; i < len(a); i++ {
				sum += a[i] * b[i]
			}
			return sum
		}, true); err != nil {
			return err
		}

		// Vector magnitude
		if err := sqliteConn.RegisterFunc("magnitude", func(vectorStr string) float64 {
			var vector []float64
			if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
				return 0
			}

			var sum float64
			for _, val := range vector {
				sum += val * val
			}
			return sum // Return squared magnitude for simplicity
		}, true); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test vector operations
	tests := []struct {
		name      string
		query     string
		args      []interface{}
		expected  float64
		tolerance float64
	}{
		{"Dot product [1,2,3]·[4,5,6]", "SELECT dot_product(?, ?)",
			[]interface{}{"[1,2,3]", "[4,5,6]"}, 32.0, 0.001},
		{"Dot product orthogonal", "SELECT dot_product(?, ?)",
			[]interface{}{"[1,0,0]", "[0,1,0]"}, 0.0, 0.001},
		{"Magnitude [3,4]", "SELECT magnitude(?)",
			[]interface{}{"[3,4]"}, 25.0, 0.001}, // 3²+4² = 25
		{"Magnitude [1,1,1]", "SELECT magnitude(?)",
			[]interface{}{"[1,1,1]"}, 3.0, 0.001},
	}

	for _, test := range tests {
		var result float64
		err := conn.QueryRowContext(context.Background(), test.query, test.args...).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", test.name, err)
			continue
		}

		if abs(result-test.expected) < test.tolerance {
			fmt.Printf("✅ %s: %.3f\n", test.name, result)
		} else {
			fmt.Printf("❌ %s: %.3f (expected %.3f)\n", test.name, result, test.expected)
		}
	}
}

func testErrorHandling() {
	fmt.Println("Testing error handling...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a mock Ollama client for testing
	logger := zerolog.Nop() // Silent logger for tests
	mockOllama := NewOllamaClient("http://127.0.0.1:11434", "all-minilm:latest", logger)

	// Register the cosine similarity function (which handles errors)
	if err := registerSQLiteFunctions(db, mockOllama); err != nil {
		log.Fatal(err)
	}

	// Test error conditions
	tests := []struct {
		name        string
		query       string
		args        []interface{}
		expected    float64
		description string
	}{
		{"Invalid JSON first arg", "SELECT cosine_similarity(?, ?)",
			[]interface{}{"invalid", "[1,2,3]"}, 0.0, "Should return 0 for invalid JSON"},
		{"Invalid JSON second arg", "SELECT cosine_similarity(?, ?)",
			[]interface{}{"[1,2,3]", "invalid"}, 0.0, "Should return 0 for invalid JSON"},
		{"Mismatched dimensions", "SELECT cosine_similarity(?, ?)",
			[]interface{}{"[1,2,3]", "[1,2]"}, 0.0, "Should return 0 for mismatched dimensions"},
		{"Empty vectors", "SELECT cosine_similarity(?, ?)",
			[]interface{}{"[]", "[]"}, 0.0, "Should return 0 for empty vectors"},
		{"Zero vector", "SELECT cosine_similarity(?, ?)",
			[]interface{}{"[0,0,0]", "[1,2,3]"}, 0.0, "Should return 0 for zero vector"},
	}

	for _, test := range tests {
		var result float64
		err := db.QueryRow(test.query, test.args...).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Query failed - %v\n", test.name, err)
			continue
		}

		if result == test.expected {
			fmt.Printf("✅ %s: %.3f - %s\n", test.name, result, test.description)
		} else {
			fmt.Printf("❌ %s: %.3f (expected %.3f) - %s\n",
				test.name, result, test.expected, test.description)
		}
	}
}

func testEmbeddingFunction() {
	fmt.Println("Testing get_embedding function...")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create Ollama client for real testing
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	ollama := NewOllamaClient("http://127.0.0.1:11434", "all-minilm:latest", logger)

	// Register functions including get_embedding
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := registerSQLiteFunctions(db, ollama); err != nil {
		log.Fatal(err)
	}

	// Test embedding generation
	tests := []struct {
		name        string
		text        string
		expectEmpty bool
		description string
	}{
		{"Simple text", "hello world", false, "Should return valid embedding"},
		{"Empty text", "", true, "Should return empty array for empty text"},
		{"Complex text", "The quick brown fox jumps over the lazy dog", false, "Should handle longer text"},
		{"Technical text", "machine learning and artificial intelligence", false, "Should handle technical terms"},
	}

	for _, test := range tests {
		var result string
		err := conn.QueryRowContext(context.Background(), "SELECT get_embedding(?)", test.text).Scan(&result)
		if err != nil {
			fmt.Printf("❌ %s: Query failed - %v\n", test.name, err)
			continue
		}

		if test.expectEmpty {
			if result == "[]" {
				fmt.Printf("✅ %s: Empty array returned - %s\n", test.name, test.description)
			} else {
				fmt.Printf("❌ %s: Expected empty array, got %s - %s\n", test.name, result, test.description)
			}
		} else {
			// Parse the JSON to validate it's a proper embedding
			var embedding []float64
			if err := json.Unmarshal([]byte(result), &embedding); err != nil {
				fmt.Printf("❌ %s: Invalid JSON returned - %v\n", test.name, err)
				continue
			}

			if len(embedding) > 0 {
				fmt.Printf("✅ %s: Valid embedding with %d dimensions - %s\n",
					test.name, len(embedding), test.description)
			} else {
				fmt.Printf("❌ %s: Empty embedding returned - %s\n", test.name, test.description)
			}
		}
	}

	// Test SQL query integration
	fmt.Println("\nTesting SQL integration...")

	// Create a test table
	_, err = conn.ExecContext(context.Background(), `
		CREATE TABLE test_docs (
			id INTEGER PRIMARY KEY,
			content TEXT,
			embedding TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert document with computed embedding
	testText := "This is a test document"
	_, err = conn.ExecContext(context.Background(),
		"INSERT INTO test_docs (content, embedding) VALUES (?, get_embedding(?))",
		testText, testText)
	if err != nil {
		fmt.Printf("❌ SQL INSERT with get_embedding failed: %v\n", err)
		return
	}

	// Verify the insertion worked
	var content, embeddingStr string
	err = conn.QueryRowContext(context.Background(),
		"SELECT content, embedding FROM test_docs WHERE id = 1").Scan(&content, &embeddingStr)
	if err != nil {
		fmt.Printf("❌ Failed to read inserted data: %v\n", err)
		return
	}

	var embedding []float64
	if err := json.Unmarshal([]byte(embeddingStr), &embedding); err != nil {
		fmt.Printf("❌ Failed to parse stored embedding: %v\n", err)
		return
	}

	fmt.Printf("✅ SQL INSERT with get_embedding successful: %d dimensions stored\n", len(embedding))

	// Test similarity search with computed embedding
	queryText := "test document"
	rows, err := conn.QueryContext(context.Background(), `
		SELECT content, cosine_similarity(embedding, get_embedding(?)) as similarity
		FROM test_docs
		ORDER BY similarity DESC
	`, queryText)
	if err != nil {
		fmt.Printf("❌ Similarity search with get_embedding failed: %v\n", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var docContent string
		var similarity float64
		if err := rows.Scan(&docContent, &similarity); err != nil {
			fmt.Printf("❌ Failed to scan similarity result: %v\n", err)
			return
		}
		fmt.Printf("✅ Similarity search successful: '%s' similarity %.3f\n", docContent, similarity)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Add the test command to the main application
func addTestCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(createTestCommand())
}

// Standalone test runner (only used when run directly)
func testMain() {
	if len(os.Args) > 1 && os.Args[1] == "test-functions" {
		fmt.Println("Running SQLite function tests...\n")

		testBasicFunction()
		fmt.Println()

		testCosineSimilarity()
		fmt.Println()

		testJSONFunctions()
		fmt.Println()

		testVectorOperations()
		fmt.Println()

		testErrorHandling()

		fmt.Println("\nAll tests completed!")
		return
	}

	fmt.Println("This file contains test functions for SQLite custom functions.")
	fmt.Println("Run with: go run test_functions.go test-functions")
}
