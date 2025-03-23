package main

import (
	"encoding/json"
	"fmt"
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
	http.HandleFunc("/compute-embeddings", handleComputeEmbeddings)
	http.HandleFunc("/compute-similarity", handleComputeSimilarity)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	return errors.Wrap(http.ListenAndServe(addr, nil), "failed to start server")
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

	// Mock implementation: generate a random similarity score between 0 and 1
	similarity := rand.Float64()

	resp := ComputeSimilarityResponse{
		Similarity: similarity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
