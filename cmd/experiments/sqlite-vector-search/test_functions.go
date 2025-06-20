package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
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
		name  string
		query string
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

	// Register the cosine similarity function
	if err := registerSQLiteFunctions(db); err != nil {
		log.Fatal(err)
	}

	// Test vectors with known similarities
	tests := []struct {
		name     string
		vector1  []float64
		vector2  []float64
		expected float64
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
		name     string
		query    string
		args     []interface{}
		expected float64
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

	// Register the cosine similarity function (which handles errors)
	if err := registerSQLiteFunctions(db); err != nil {
		log.Fatal(err)
	}

	// Test error conditions
	tests := []struct {
		name     string
		query    string
		args     []interface{}
		expected float64
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
