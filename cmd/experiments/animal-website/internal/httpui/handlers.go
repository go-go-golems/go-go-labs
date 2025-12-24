package httpui

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/animal-website/internal/animals"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/animal-website/internal/ui"
	"github.com/rs/zerolog"
)

type Handlers struct {
	repo   *animals.Repository
	logger zerolog.Logger
}

func NewHandlers(repo *animals.Repository, logger zerolog.Logger) *Handlers {
	return &Handlers{
		repo:   repo,
		logger: logger,
	}
}

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleRoot)
	mux.HandleFunc("/animals", h.handleAnimals)
	mux.HandleFunc("/upload", h.handleUpload)
	mux.HandleFunc("/animals/clear", h.handleClear)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(ui.StaticFS))))
}

func (h *Handlers) handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/animals", http.StatusFound)
}

func (h *Handlers) handleAnimals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	animalsList, err := h.repo.List(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list animals")
		http.Error(w, "Failed to load animals", http.StatusInternalServerError)
		return
	}

	isHtmx := r.Header.Get("HX-Request") == "true"
	if isHtmx {
		// Return just the list fragment
		ui.AnimalsList(animalsList).Render(ctx, w)
		return
	}

	// Return full page
	ui.AnimalsPage(animalsList).Render(ctx, w)
}

func (h *Handlers) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ui.UploadPage().Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check content type
	if header.Header.Get("Content-Type") != "text/csv" && !strings.HasSuffix(header.Filename, ".csv") {
		h.logger.Warn().Str("content_type", header.Header.Get("Content-Type")).Msg("Unexpected content type")
		// Continue anyway - be forgiving
	}

	// Parse CSV
	names, err := animals.ParseCSV(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse CSV")
		http.Error(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusBadRequest)
		return
	}

	if len(names) == 0 {
		http.Error(w, "No animal names found in CSV", http.StatusBadRequest)
		return
	}

	// Determine mode
	mode := animals.InsertModeReplace
	if r.FormValue("mode") == "append" {
		mode = animals.InsertModeAppend
	}

	// Insert into database
	ctx := r.Context()
	inserted, err := h.repo.InsertMany(ctx, names, mode)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to insert animals")
		http.Error(w, "Failed to save animals", http.StatusInternalServerError)
		return
	}

	h.logger.Info().
		Int("total_names", len(names)).
		Int("inserted", inserted).
		Str("mode", string(mode)).
		Msg("Animals imported")

	isHtmx := r.Header.Get("HX-Request") == "true"
	if isHtmx {
		// Return updated list
		animalsList, err := h.repo.List(ctx)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to list animals after import")
			http.Error(w, "Failed to load animals", http.StatusInternalServerError)
			return
		}
		ui.AnimalsList(animalsList).Render(ctx, w)
		return
	}

	// Redirect to animals page
	http.Redirect(w, r, "/animals", http.StatusSeeOther)
}

func (h *Handlers) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := h.repo.Clear(ctx); err != nil {
		h.logger.Error().Err(err).Msg("Failed to clear animals")
		http.Error(w, "Failed to clear animals", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Msg("Animals cleared")

	isHtmx := r.Header.Get("HX-Request") == "true"
	if isHtmx {
		// Return empty list
		ui.AnimalsList([]animals.Animal{}).Render(ctx, w)
		return
	}

	// Redirect to animals page
	http.Redirect(w, r, "/animals", http.StatusSeeOther)
}

