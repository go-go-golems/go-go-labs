package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

// Example 1: Basic custom function registration
func ExampleBasicFunction() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register a simple custom function
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		// Register a function that squares a number
		return sqliteConn.RegisterFunc("square", func(x float64) float64 {
			return x * x
		}, true) // true means the function is deterministic
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test the function
	var result float64
	err = db.QueryRow("SELECT square(5)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("square(5) = %.1f\n", result) // Output: square(5) = 25.0
}

// Example 2: Working with JSON data in custom functions
func ExampleJSONFunction() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register a function that sums numbers in a JSON array
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		return sqliteConn.RegisterFunc("json_sum", func(jsonStr string) float64 {
			var numbers []float64
			if err := json.Unmarshal([]byte(jsonStr), &numbers); err != nil {
				return 0 // Return 0 for invalid JSON
			}
			
			sum := 0.0
			for _, num := range numbers {
				sum += num
			}
			return sum
		}, true)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test the function
	var result float64
	err = db.QueryRow("SELECT json_sum('[1, 2, 3, 4, 5]')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("json_sum([1,2,3,4,5]) = %.1f\n", result) // Output: json_sum([1,2,3,4,5]) = 15.0
}

// Example 3: Vector operations - dot product
func ExampleDotProduct() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register dot product function
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		return sqliteConn.RegisterFunc("dot_product", func(aStr, bStr string) float64 {
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
			
			var product float64
			for i := 0; i < len(a); i++ {
				product += a[i] * b[i]
			}
			return product
		}, true)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test the function
	var result float64
	err = db.QueryRow("SELECT dot_product('[1, 2, 3]', '[4, 5, 6]')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("dot_product([1,2,3], [4,5,6]) = %.1f\n", result) // Output: dot_product([1,2,3], [4,5,6]) = 32.0
}

// Example 4: Complete vector similarity with table
func ExampleVectorSimilarityTable() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE vectors (
			id INTEGER PRIMARY KEY,
			name TEXT,
			vector TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Register cosine similarity function
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		return sqliteConn.RegisterFunc("cosine_similarity", func(aStr, bStr string) float64 {
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
		}, true)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Insert sample vectors
	vectors := []struct {
		name   string
		vector []float64
	}{
		{"cat", []float64{1, 0, 0}},
		{"dog", []float64{0.8, 0.6, 0}},
		{"car", []float64{0, 0, 1}},
		{"kitten", []float64{0.9, 0.1, 0}},
	}

	for _, v := range vectors {
		vectorJSON, _ := json.Marshal(v.vector)
		_, err = db.Exec("INSERT INTO vectors (name, vector) VALUES (?, ?)", v.name, string(vectorJSON))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Search for similar vectors to "cat"
	query := `
		SELECT name, cosine_similarity(vector, ?) as similarity
		FROM vectors
		WHERE name != ?
		ORDER BY similarity DESC
		LIMIT 3
	`
	
	catVector, _ := json.Marshal([]float64{1, 0, 0})
	rows, err := db.Query(query, string(catVector), "cat")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Most similar to 'cat':")
	for rows.Next() {
		var name string
		var similarity float64
		if err := rows.Scan(&name, &similarity); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %s: %.3f\n", name, similarity)
	}
	// Output:
	// Most similar to 'cat':
	//   kitten: 0.995
	//   dog: 0.800
	//   car: 0.000
}

// Example 5: Error handling in custom functions
func ExampleErrorHandling() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register a function with proper error handling
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		// Function that validates and processes vector input
		return sqliteConn.RegisterFunc("safe_vector_length", func(vectorStr string) float64 {
			var vector []float64
			
			// Validate JSON
			if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
				// Log error in real application
				fmt.Printf("Invalid JSON: %s\n", vectorStr)
				return -1 // Return -1 to indicate error
			}
			
			// Validate vector is not empty
			if len(vector) == 0 {
				fmt.Printf("Empty vector: %s\n", vectorStr)
				return -1
			}
			
			// Calculate vector length (magnitude)
			var sum float64
			for _, val := range vector {
				sum += val * val
			}
			
			return math.Sqrt(sum)
		}, true)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test with valid input
	var result float64
	err = db.QueryRow("SELECT safe_vector_length('[3, 4]')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Length of [3,4]: %.1f\n", result) // Output: Length of [3,4]: 5.0

	// Test with invalid input
	err = db.QueryRow("SELECT safe_vector_length('invalid json')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Length of invalid JSON: %.1f\n", result) // Output: Length of invalid JSON: -1.0
}

// Example 6: Multiple parameter types
func ExampleMultipleParameters() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register a function with multiple parameter types
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		sqliteConn := driverConn.(*sqlite3.SQLiteConn)
		
		// Function that scales a vector by a scalar
		return sqliteConn.RegisterFunc("scale_vector", func(vectorStr string, scale float64) string {
			var vector []float64
			
			if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
				return "[]" // Return empty array for invalid input
			}
			
			// Scale each component
			for i := range vector {
				vector[i] *= scale
			}
			
			// Return as JSON string
			result, _ := json.Marshal(vector)
			return string(result)
		}, true)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Test the function
	var result string
	err = db.QueryRow("SELECT scale_vector('[1, 2, 3]', 2.5)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Scaled vector: %s\n", result) // Output: Scaled vector: [2.5,5,7.5]
}

// Run all examples
func RunAllExamples() {
	fmt.Println("=== Example 1: Basic Function ===")
	ExampleBasicFunction()
	
	fmt.Println("\n=== Example 2: JSON Function ===")
	ExampleJSONFunction()
	
	fmt.Println("\n=== Example 3: Dot Product ===")
	ExampleDotProduct()
	
	fmt.Println("\n=== Example 4: Vector Similarity Table ===")
	ExampleVectorSimilarityTable()
	
	fmt.Println("\n=== Example 5: Error Handling ===")
	ExampleErrorHandling()
	
	fmt.Println("\n=== Example 6: Multiple Parameters ===")
	ExampleMultipleParameters()
}
