package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-labs/cmd/apps/embeddings/templates"

	"github.com/pkg/errors"
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

// EmbeddingsServer handles the embeddings API requests
type EmbeddingsServer struct {
	factory embeddings.ProviderFactory
	port    int
}

// NewEmbeddingsServer creates a new server with the given embeddings provider factory
func NewEmbeddingsServer(factory embeddings.ProviderFactory, port int) *EmbeddingsServer {
	return &EmbeddingsServer{
		factory: factory,
		port:    port,
	}
}

// ServerCommand defines the serve command using glazed
type ServerCommand struct {
	*cmds.CommandDescription
}

// ServerSettings contains configuration for the server
type ServerSettings struct {
	Port int `glazed.parameter:"port"`
}

func NewServerCommand() (*ServerCommand, error) {
	layers_, err := GetEmbeddingsLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings layers")
	}

	return &ServerCommand{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Start the embeddings server"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Port to run the server on"),
					parameters.WithDefault(8080),
				),
			),
			cmds.WithLayersList(layers_...),
		),
	}, nil
}

func (c *ServerCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Parse server settings
	s := &ServerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "could not initialize server settings")
	}

	// Create embeddings provider from parsed layers
	factory, err := embeddings.NewSettingsFactoryFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "could not create embeddings factory")
	}

	// Create server
	server := NewEmbeddingsServer(factory, s.Port)

	return server.Serve()
}

func init() {
	serverCmd, err := NewServerCommand()
	if err != nil {
		log.Fatalf("Error creating server command: %v", err)
	}

	// The rootCmd is now created in main.go, so we need to use a different approach
	// This init function will be called by main.go's init, so we need to export the command
	// for the parser to use.
	embeddings_commands = append(embeddings_commands, serverCmd)
}

func (s *EmbeddingsServer) Serve() error {
	// API endpoints
	http.HandleFunc("/compute-embeddings", s.handleComputeEmbeddings)
	http.HandleFunc("/compute-similarity", s.handleComputeSimilarity)

	// Web UI endpoints
	http.HandleFunc("/", s.handleHomePage)
	http.HandleFunc("/compare", s.handleCompare)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Web UI available at http://localhost:%d", s.port)
	return errors.Wrap(http.ListenAndServe(addr, nil), "failed to start server")
}

func (s *EmbeddingsServer) handleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	err := templates.ComparePage().Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *EmbeddingsServer) handleCompare(w http.ResponseWriter, r *http.Request) {
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
		simAB, err := s.computeSimilarityScore(r.Context(), textA, textB)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error computing similarity A-B: %v", err), http.StatusInternalServerError)
			return
		}
		similarityAB = formatSimilarity(simAB)
	}

	if textA != "" && textC != "" {
		simAC, err := s.computeSimilarityScore(r.Context(), textA, textC)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error computing similarity A-C: %v", err), http.StatusInternalServerError)
			return
		}
		similarityAC = formatSimilarity(simAC)
	}

	if textB != "" && textC != "" {
		simBC, err := s.computeSimilarityScore(r.Context(), textB, textC)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error computing similarity B-C: %v", err), http.StatusInternalServerError)
			return
		}
		similarityBC = formatSimilarity(simBC)
	}

	err := templates.SimilarityResults(similarityAB, similarityAC, similarityBC).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// computeSimilarityScore returns a similarity score between 0 and 1
func (s *EmbeddingsServer) computeSimilarityScore(ctx context.Context, text1 string, text2 string) (float64, error) {
	provider, err := s.factory.NewProvider()
	if err != nil {
		return 0, errors.Wrap(err, "failed to create embeddings provider")
	}

	// Generate embeddings for both texts
	embedding1, err := provider.GenerateEmbedding(ctx, text1)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate embeddings for text1")
	}

	embedding2, err := provider.GenerateEmbedding(ctx, text2)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate embeddings for text2")
	}

	// Calculate cosine similarity
	return computeCosineSimilarity(embedding1, embedding2), nil
}

// computeCosineSimilarity calculates the cosine similarity between two embedding vectors
func computeCosineSimilarity(vec1, vec2 []float32) float64 {
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

// formatSimilarity formats a similarity score as a string percentage
func formatSimilarity(similarity float64) string {
	percentage := similarity * 100
	return fmt.Sprintf("%.2f%%", percentage)
}

func (s *EmbeddingsServer) handleComputeEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComputeEmbeddingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider, err := s.factory.NewProvider()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create embeddings provider: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate embeddings
	embedding, err := provider.GenerateEmbedding(r.Context(), req.Text)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate embeddings: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert float32 embeddings to float64 for JSON response
	vector := make([]float64, len(embedding))
	for i, v := range embedding {
		vector[i] = float64(v)
	}

	resp := ComputeEmbeddingsResponse{
		Vector: vector,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *EmbeddingsServer) handleComputeSimilarity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComputeSimilarityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	similarity, err := s.computeSimilarityScore(r.Context(), req.Text1, req.Text2)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to compute similarity: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ComputeSimilarityResponse{
		Similarity: similarity,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
