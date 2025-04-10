package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

const (
	serverPort = 8080
)

func main() {
	// Set up context with cancellation for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create embeddings factory with OpenAI text-embedding-3-small
	factory, err := createEmbeddingsFactory()
	if err != nil {
		log.Fatalf("Failed to create embeddings factory: %v", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: createHandler(factory),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on http://localhost:%d", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// createEmbeddingsFactory creates a factory for OpenAI text-embedding-3-small
func createEmbeddingsFactory() (embeddings.ProviderFactory, error) {
	// Create layers for configuration
	parsedLayers := layers.NewParsedLayers()

	// Set up OpenAI configuration
	err := parsedLayers.SetParameterValue(config.EmbeddingsSlug, "embeddings-type", "openai")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set embeddings type")
	}

	err = parsedLayers.SetParameterValue(config.EmbeddingsSlug, "embeddings-engine", "text-embedding-3-small")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set embeddings engine")
	}

	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable not set")
	}

	err = parsedLayers.SetParameterValue(config.EmbeddingsApiKeySlug, "openai-api-key", apiKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set OpenAI API key")
	}

	// Create factory from parsed layers
	factory, err := embeddings.NewSettingsFactoryFromParsedLayers(parsedLayers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create embeddings factory")
	}

	return factory, nil
}

// computeSimilarity calculates cosine similarity between two embedding vectors
func computeSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct float64
	var norm1 float64
	var norm2 float64

	for i := range vec1 {
		dotProduct += float64(vec1[i] * vec2[i])
		norm1 += float64(vec1[i] * vec1[i])
		norm2 += float64(vec2[i] * vec2[i])
	}

	// Avoid division by zero
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// createHandler sets up HTTP handlers for the application
func createHandler(factory embeddings.ProviderFactory) http.Handler {
	mux := http.NewServeMux()

	// Serve index page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		renderIndexPage(w, "", "", "", nil, nil)
	})

	// Handle form submission
	mux.HandleFunc("/compare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form values
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		textA := r.FormValue("text-a")
		textB := r.FormValue("text-b")
		textC := r.FormValue("text-c")

		// Create provider from factory
		provider, err := factory.NewProvider()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create embeddings provider: %v", err), http.StatusInternalServerError)
			return
		}

		// Calculate similarities
		similarities := make(map[string]float64)
		var errors []string

		// Only calculate similarities for non-empty text fields
		if textA != "" && textB != "" {
			sim, err := calculateSimilarity(r.Context(), provider, textA, textB)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to calculate A-B similarity: %v", err))
			} else {
				similarities["A-B"] = sim
			}
		}

		if textA != "" && textC != "" {
			sim, err := calculateSimilarity(r.Context(), provider, textA, textC)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to calculate A-C similarity: %v", err))
			} else {
				similarities["A-C"] = sim
			}
		}

		if textB != "" && textC != "" {
			sim, err := calculateSimilarity(r.Context(), provider, textB, textC)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to calculate B-C similarity: %v", err))
			} else {
				similarities["B-C"] = sim
			}
		}

		// Render page with results
		renderIndexPage(w, textA, textB, textC, similarities, errors)
	})

	return mux
}

// calculateSimilarity generates embeddings for two texts and returns their similarity
func calculateSimilarity(ctx context.Context, provider embeddings.Provider, text1, text2 string) (float64, error) {
	// Generate embeddings
	embedding1, err := provider.GenerateEmbedding(ctx, text1)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate embedding for text 1")
	}

	embedding2, err := provider.GenerateEmbedding(ctx, text2)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate embedding for text 2")
	}

	// Calculate similarity
	return computeSimilarity(embedding1, embedding2), nil
}

// renderIndexPage renders the HTML page with the form and results
func renderIndexPage(w http.ResponseWriter, textA, textB, textC string, similarities map[string]float64, errors []string) {
	w.Header().Set("Content-Type", "text/html")

	// Severance style CSS classes
	const pageTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Text Similarity Analyzer</title>
    <style>
        :root {
            --lumon-green: #71c1a1;
            --lumon-darkgreen: #2c594f;
            --lumon-blue: #143964;
            --lumon-light: #f5f5f5;
            --lumon-dark: #2b2b2b;
            --font-main: 'Helvetica Neue', Arial, sans-serif;
        }
        body {
            font-family: var(--font-main);
            background-color: var(--lumon-dark);
            color: var(--lumon-light);
            margin: 0;
            padding: 20px;
            display: flex;
            flex-direction: column;
            min-height: 100vh;
            box-sizing: border-box;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            padding: 30px;
            background-color: #1c1c1c;
            border-radius: 12px;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
        }
        h1 {
            color: var(--lumon-green);
            text-align: center;
            font-weight: 300;
            letter-spacing: 1px;
            margin-bottom: 30px;
            font-size: 28px;
        }
        form {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: var(--lumon-green);
            letter-spacing: 0.5px;
        }
        textarea {
            width: 100%;
            height: 100px;
            padding: 12px;
            border: 1px solid var(--lumon-darkgreen);
            background-color: #252525;
            color: var(--lumon-light);
            border-radius: 6px;
            resize: vertical;
            font-family: var(--font-main);
            font-size: 14px;
            box-sizing: border-box;
        }
        textarea:focus {
            border-color: var(--lumon-green);
            outline: none;
            box-shadow: 0 0 0 2px rgba(113, 193, 161, 0.2);
        }
        button {
            background-color: var(--lumon-darkgreen);
            color: white;
            border: none;
            padding: 12px 24px;
            font-size: 16px;
            cursor: pointer;
            border-radius: 6px;
            font-weight: 500;
            letter-spacing: 0.5px;
            margin-top: 10px;
            align-self: center;
            transition: background-color 0.2s;
        }
        button:hover {
            background-color: var(--lumon-green);
        }
        .results {
            margin-top: 30px;
            background-color: #252525;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid var(--lumon-green);
        }
        .result-item {
            margin-bottom: 10px;
            padding-bottom: 10px;
            border-bottom: 1px dashed #444;
        }
        .result-item:last-child {
            border-bottom: none;
            margin-bottom: 0;
            padding-bottom: 0;
        }
        .similarity-label {
            display: inline-block;
            width: 80px;
            color: var(--lumon-green);
            font-weight: bold;
        }
        .similarity-value {
            font-size: 18px;
            font-weight: 300;
        }
        .error {
            color: #e74c3c;
            background-color: rgba(231, 76, 60, 0.1);
            padding: 10px;
            border-radius: 5px;
            margin-bottom: 15px;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            color: #777;
            font-size: 12px;
        }
        .high-similarity {
            color: #2ecc71;
        }
        .medium-similarity {
            color: #f39c12;
        }
        .low-similarity {
            color: #e74c3c;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Text Similarity Analyzer</h1>
        
        {{if .Errors}}
        <div class="error">
            {{range .Errors}}
            <p>{{.}}</p>
            {{end}}
        </div>
        {{end}}
        
        <form action="/compare" method="post">
            <div>
                <label for="text-a">Text A</label>
                <textarea id="text-a" name="text-a" placeholder="Enter first text...">{{.TextA}}</textarea>
            </div>
            
            <div>
                <label for="text-b">Text B</label>
                <textarea id="text-b" name="text-b" placeholder="Enter second text...">{{.TextB}}</textarea>
            </div>
            
            <div>
                <label for="text-c">Text C</label>
                <textarea id="text-c" name="text-c" placeholder="Enter third text...">{{.TextC}}</textarea>
            </div>
            
            <button type="submit">Calculate Similarity</button>
        </form>
        
        {{if .Similarities}}
        <div class="results">
            {{if index .Similarities "A-B"}}
            <div class="result-item">
                <span class="similarity-label">A-B:</span>
                <span class="similarity-value {{getSimilarityClass (index .Similarities "A-B")}}">
                    {{formatSimilarity (index .Similarities "A-B")}}
                </span>
            </div>
            {{end}}
            
            {{if index .Similarities "A-C"}}
            <div class="result-item">
                <span class="similarity-label">A-C:</span>
                <span class="similarity-value {{getSimilarityClass (index .Similarities "A-C")}}">
                    {{formatSimilarity (index .Similarities "A-C")}}
                </span>
            </div>
            {{end}}
            
            {{if index .Similarities "B-C"}}
            <div class="result-item">
                <span class="similarity-label">B-C:</span>
                <span class="similarity-value {{getSimilarityClass (index .Similarities "B-C")}}">
                    {{formatSimilarity (index .Similarities "B-C")}}
                </span>
            </div>
            {{end}}
        </div>
        {{end}}
        
        <div class="footer">
            Powered by OpenAI text-embedding-3-small
        </div>
    </div>
</body>
</html>`

	// Format the template using simple string replacement
	html := pageTemplate

	// Replace placeholders with values
	html = strings.Replace(html, "{{.TextA}}", textA, -1)
	html = strings.Replace(html, "{{.TextB}}", textB, -1)
	html = strings.Replace(html, "{{.TextC}}", textC, -1)

	// Handle errors
	if len(errors) > 0 {
		errorHTML := `<div class="error">`
		for _, err := range errors {
			errorHTML += fmt.Sprintf("<p>%s</p>", err)
		}
		errorHTML += `</div>`
		html = strings.Replace(html, "{{if .Errors}}\n        <div class=\"error\">\n            {{range .Errors}}\n            <p>{{.}}</p>\n            {{end}}\n        </div>\n        {{end}}", errorHTML, -1)
	} else {
		html = strings.Replace(html, "{{if .Errors}}\n        <div class=\"error\">\n            {{range .Errors}}\n            <p>{{.}}</p>\n            {{end}}\n        </div>\n        {{end}}", "", -1)
	}

	// Handle similarities
	if len(similarities) > 0 {
		similarityHTML := `<div class="results">`

		if sim, ok := similarities["A-B"]; ok {
			class := getSimilarityClass(sim)
			formatted := formatSimilarity(sim)
			similarityHTML += fmt.Sprintf(`
            <div class="result-item">
                <span class="similarity-label">A-B:</span>
                <span class="similarity-value %s">
                    %s
                </span>
            </div>`, class, formatted)
		}

		if sim, ok := similarities["A-C"]; ok {
			class := getSimilarityClass(sim)
			formatted := formatSimilarity(sim)
			similarityHTML += fmt.Sprintf(`
            <div class="result-item">
                <span class="similarity-label">A-C:</span>
                <span class="similarity-value %s">
                    %s
                </span>
            </div>`, class, formatted)
		}

		if sim, ok := similarities["B-C"]; ok {
			class := getSimilarityClass(sim)
			formatted := formatSimilarity(sim)
			similarityHTML += fmt.Sprintf(`
            <div class="result-item">
                <span class="similarity-label">B-C:</span>
                <span class="similarity-value %s">
                    %s
                </span>
            </div>`, class, formatted)
		}

		similarityHTML += `</div>`

		html = strings.Replace(html, "{{if .Similarities}}\n        <div class=\"results\">\n            {{if index .Similarities \"A-B\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">A-B:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"A-B\")}}\">\n                    {{formatSimilarity (index .Similarities \"A-B\")}}\n                </span>\n            </div>\n            {{end}}\n            \n            {{if index .Similarities \"A-C\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">A-C:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"A-C\")}}\">\n                    {{formatSimilarity (index .Similarities \"A-C\")}}\n                </span>\n            </div>\n            {{end}}\n            \n            {{if index .Similarities \"B-C\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">B-C:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"B-C\")}}\">\n                    {{formatSimilarity (index .Similarities \"B-C\")}}\n                </span>\n            </div>\n            {{end}}\n        </div>\n        {{end}}", similarityHTML, -1)
	} else {
		html = strings.Replace(html, "{{if .Similarities}}\n        <div class=\"results\">\n            {{if index .Similarities \"A-B\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">A-B:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"A-B\")}}\">\n                    {{formatSimilarity (index .Similarities \"A-B\")}}\n                </span>\n            </div>\n            {{end}}\n            \n            {{if index .Similarities \"A-C\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">A-C:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"A-C\")}}\">\n                    {{formatSimilarity (index .Similarities \"A-C\")}}\n                </span>\n            </div>\n            {{end}}\n            \n            {{if index .Similarities \"B-C\"}}\n            <div class=\"result-item\">\n                <span class=\"similarity-label\">B-C:</span>\n                <span class=\"similarity-value {{getSimilarityClass (index .Similarities \"B-C\")}}\">\n                    {{formatSimilarity (index .Similarities \"B-C\")}}\n                </span>\n            </div>\n            {{end}}\n        </div>\n        {{end}}", "", -1)
	}

	// Write the HTML to the response
	_, err := w.Write([]byte(html))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// formatSimilarity formats a similarity score as a percentage
func formatSimilarity(similarity float64) string {
	return fmt.Sprintf("%.2f%%", similarity*100)
}

// getSimilarityClass returns a CSS class based on similarity score
func getSimilarityClass(similarity float64) string {
	switch {
	case similarity >= 0.7:
		return "high-similarity"
	case similarity >= 0.4:
		return "medium-similarity"
	default:
		return "low-similarity"
	}
}
