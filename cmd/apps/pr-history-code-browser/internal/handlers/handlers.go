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

// HandleGetCommit returns details for a specific commit
func (h *Handler) HandleGetCommit(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		http.Error(w, "Commit hash required", http.StatusBadRequest)
		return
	}

	commit, err := h.db.GetCommitByHash(hash)
	if err != nil {
		log.Error().Err(err).Str("hash", hash).Msg("Failed to get commit")
		http.Error(w, "Commit not found", http.StatusNotFound)
		return
	}

	files, err := h.db.GetCommitFiles(commit.ID)
	if err != nil {
		log.Error().Err(err).Int64("commitID", commit.ID).Msg("Failed to get commit files")
		http.Error(w, "Failed to get commit files", http.StatusInternalServerError)
		return
	}

	symbols, err := h.db.GetCommitSymbols(commit.ID)
	if err != nil {
		log.Error().Err(err).Int64("commitID", commit.ID).Msg("Failed to get commit symbols")
		http.Error(w, "Failed to get commit symbols", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"commit":  commit,
		"files":   files,
		"symbols": symbols,
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

// HandleGetFileHistory returns commit history for a file
func (h *Handler) HandleGetFileHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	limit := getIntParam(r, "limit", 50)

	commits, err := h.db.GetFileHistory(id, limit)
	if err != nil {
		log.Error().Err(err).Int64("fileID", id).Msg("Failed to get file history")
		http.Error(w, "Failed to get file history", http.StatusInternalServerError)
		return
	}

	respondJSON(w, commits)
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

