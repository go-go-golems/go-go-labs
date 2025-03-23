package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/embeddings/templates"
	"log"
	"math/rand"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ComputeEmbeddingsRequest struct {
	Text string `json:"text"`
}

type ComputeEmbeddingsResponse struct {
	Vector []float64 `json:"vector"`
}

type ComputeSimilarityRequest struct {
	Text1 string `json:"text1"`
	Text2 string `json:"text2"`
}

type ComputeSimilarityResponse struct {
	Similarity float64 `json:"similarity"`
}

func init() {
	var port int
	var serverCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the embeddings server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(port)
		},
	}

	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	rootCmd.AddCommand(serverCmd)
}

func runServer(port int) error {
	// API endpoints
	http.HandleFunc("/compute-embeddings", handleComputeEmbeddings)
	http.HandleFunc("/compute-similarity", handleComputeSimilarity)

	// Web UI endpoints
	http.HandleFunc("/", handleHomePage)
	http.HandleFunc("/compare", handleCompare)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Web UI available at http://localhost:%d", port)
	return errors.Wrap(http.ListenAndServe(addr, nil), "failed to start server")
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	err := templates.ComparePage().Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	textA := r.FormValue("textA")
	textB := r.FormValue("textB")
	textC := r.FormValue("textC")

	var similarityAB, similarityAC, similarityBC string

	// Only calculate similarities if both texts are provided
	if textA != "" && textB != "" {
		simAB := computeSimilarityScore(textA, textB)
		similarityAB = formatSimilarity(simAB)
	}

	if textA != "" && textC != "" {
		simAC := computeSimilarityScore(textA, textC)
		similarityAC = formatSimilarity(simAC)
	}

	if textB != "" && textC != "" {
		simBC := computeSimilarityScore(textB, textC)
		similarityBC = formatSimilarity(simBC)
	}

	err := templates.SimilarityResults(similarityAB, similarityAC, similarityBC).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// computeSimilarityScore returns a similarity score between 0 and 1
func computeSimilarityScore(text1 string, text2 string) float64 {
	// Mock implementation: random score between 0 and 1
	return rand.Float64()
}

// formatSimilarity formats a similarity score as a string percentage
func formatSimilarity(similarity float64) string {
	percentage := similarity * 100
	return fmt.Sprintf("%.2f%%", percentage)
}

func handleComputeEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComputeEmbeddingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mock implementation: generate a random 768-dimensional vector
	vector := make([]float64, 768)
	for i := range vector {
		vector[i] = rand.Float64()*2 - 1 // Random values between -1 and 1
	}

	resp := ComputeEmbeddingsResponse{
		Vector: vector,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleComputeSimilarity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComputeSimilarityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use the same function as for the web UI
	similarity := computeSimilarityScore(req.Text1, req.Text2)

	resp := ComputeSimilarityResponse{
		Similarity: similarity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
