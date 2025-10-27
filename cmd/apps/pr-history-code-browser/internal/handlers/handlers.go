package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/go-go-labs/cmd/apps/pr-history-code-browser/internal/models"
	"github.com/rs/zerolog/log"
)

// Handler provides HTTP handlers for the API
type Handler struct {
	db *models.DB
}

// NewHandler creates a new handler instance
func NewHandler(db *models.DB) *Handler {
	return &Handler{db: db}
}

// HandleGetStats returns database statistics
func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetStats()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stats")
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	respondJSON(w, stats)
}

// HandleListCommits returns a paginated list of commits
func (h *Handler) HandleListCommits(w http.ResponseWriter, r *http.Request) {
	limit := getIntParam(r, "limit", 50)
	offset := getIntParam(r, "offset", 0)
	search := r.URL.Query().Get("search")

	commits, err := h.db.GetCommits(limit, offset, search)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get commits")
		http.Error(w, "Failed to get commits", http.StatusInternalServerError)
		return
	}

	respondJSON(w, commits)
}

// HandleGetCommit returns enriched details for a specific commit (with PR associations and notes)
func (h *Handler) HandleGetCommit(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		http.Error(w, "Commit hash required", http.StatusBadRequest)
		return
	}

	// Use enriched version with PR associations
	commitWithRefs, err := h.db.GetCommitWithPRAssociations(hash)
	if err != nil {
		log.Error().Err(err).Str("hash", hash).Msg("Failed to get commit with references")
		http.Error(w, "Commit not found", http.StatusNotFound)
		return
	}

	// Get symbols separately (not in CommitWithRefsAndPRs yet)
	symbols, err := h.db.GetCommitSymbols(commitWithRefs.Commit.ID)
	if err != nil {
		log.Error().Err(err).Int64("commitID", commitWithRefs.Commit.ID).Msg("Failed to get commit symbols")
		http.Error(w, "Failed to get commit symbols", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"commit":           commitWithRefs.Commit,
		"files":            commitWithRefs.Files,
		"symbols":          symbols,
		"pr_associations":  commitWithRefs.PRAssociations,
		"notes":            commitWithRefs.Notes,
	}

	respondJSON(w, response)
}

// HandleListPRs returns all PRs
func (h *Handler) HandleListPRs(w http.ResponseWriter, r *http.Request) {
	prs, err := h.db.GetPRs()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get PRs")
		http.Error(w, "Failed to get PRs", http.StatusInternalServerError)
		return
	}

	respondJSON(w, prs)
}

// HandleGetPR returns details for a specific PR
func (h *Handler) HandleGetPR(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid PR ID", http.StatusBadRequest)
		return
	}

	pr, err := h.db.GetPRByID(id)
	if err != nil {
		log.Error().Err(err).Int64("id", id).Msg("Failed to get PR")
		http.Error(w, "PR not found", http.StatusNotFound)
		return
	}

	respondJSON(w, pr)
}

// HandleListFiles returns a list of files
func (h *Handler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	limit := getIntParam(r, "limit", 100)
	offset := getIntParam(r, "offset", 0)
	pathPrefix := r.URL.Query().Get("prefix")

	files, err := h.db.GetFiles(pathPrefix, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get files")
		http.Error(w, "Failed to get files", http.StatusInternalServerError)
		return
	}

	respondJSON(w, files)
}

// HandleGetFileHistory returns enriched file details with history, related files, and notes
func (h *Handler) HandleGetFileHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	limit := getIntParam(r, "limit", 50)

	fileWithDetails, err := h.db.GetFileWithDetails(id, limit)
	if err != nil {
		log.Error().Err(err).Int64("fileID", id).Msg("Failed to get file details")
		http.Error(w, "Failed to get file details", http.StatusInternalServerError)
		return
	}

	respondJSON(w, fileWithDetails)
}

// HandleListAnalysisNotes returns analysis notes
func (h *Handler) HandleListAnalysisNotes(w http.ResponseWriter, r *http.Request) {
	limit := getIntParam(r, "limit", 50)
	offset := getIntParam(r, "offset", 0)
	noteType := r.URL.Query().Get("type")
	tags := r.URL.Query().Get("tags")

	notes, err := h.db.GetAnalysisNotes(noteType, tags, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get analysis notes")
		http.Error(w, "Failed to get analysis notes", http.StatusInternalServerError)
		return
	}

	respondJSON(w, notes)
}

// HandleGetSymbolHistory returns the history of a specific symbol
func (h *Handler) HandleGetSymbolHistory(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol name required", http.StatusBadRequest)
		return
	}

	limit := getIntParam(r, "limit", 50)

	history, err := h.db.GetSymbolHistory(symbol, limit)
	if err != nil {
		log.Error().Err(err).Str("symbol", symbol).Msg("Failed to get symbol history")
		http.Error(w, "Failed to get symbol history", http.StatusInternalServerError)
		return
	}

	respondJSON(w, history)
}

// HandleSearchSymbols searches for symbols by pattern
func (h *Handler) HandleSearchSymbols(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("q")
	if pattern == "" {
		http.Error(w, "Search pattern required", http.StatusBadRequest)
		return
	}

	limit := getIntParam(r, "limit", 100)

	symbols, err := h.db.SearchSymbols(pattern, limit)
	if err != nil {
		log.Error().Err(err).Str("pattern", pattern).Msg("Failed to search symbols")
		http.Error(w, "Failed to search symbols", http.StatusInternalServerError)
		return
	}

	respondJSON(w, symbols)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getIntParam(r *http.Request, param string, defaultValue int) int {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

