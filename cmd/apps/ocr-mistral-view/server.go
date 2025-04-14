package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-go-golems/go-go-labs/cmd/apps/ocr-mistral-view/views"
	"github.com/pkg/errors"
)

// Server represents our web application server
type Server struct {
	router    chi.Router
	ocrData   views.OCRData
	tempDir   string
	hasImages bool
}

// NewServer creates a new server instance
func NewServer(inputFile string) (*Server, error) {
	// Read and parse JSON file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not read input file")
	}

	var ocrData views.OCRData
	if err := json.Unmarshal(data, &ocrData); err != nil {
		return nil, errors.Wrap(err, "could not parse JSON data")
	}

	// Create temp directory for images
	tempDir, err := os.MkdirTemp("", "ocr-mistral-images-*")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp directory")
	}

	// Check if there are any images
	hasImages := false
	for _, page := range ocrData.Pages {
		if len(page.Images) > 0 {
			hasImages = true
			break
		}
	}

	// Extract images if they exist
	if hasImages {
		for i, page := range ocrData.Pages {
			for _, img := range page.Images {
				imgData := img.ImageBase64
				// Strip the data:image/jpeg;base64, prefix if present
				if idx := strings.Index(imgData, ";base64,"); idx > 0 {
					imgData = imgData[idx+8:]
				}

				// Decode base64 to binary
				reader := strings.NewReader(imgData)
				decoder := base64.NewDecoder(base64.StdEncoding, reader)
				imgBytes, err := io.ReadAll(decoder)
				if err != nil {
					return nil, errors.Wrap(err, "could not decode image")
				}

				// Save image to temp directory
				imgPath := filepath.Join(tempDir, img.ID)
				if err := os.WriteFile(imgPath, imgBytes, 0644); err != nil {
					return nil, errors.Wrap(err, "could not write image file")
				}

				// Update the markdown to use the correct path
				ocrData.Pages[i].Markdown = strings.ReplaceAll(
					ocrData.Pages[i].Markdown,
					"]("+img.ID+")",
					"](/images/"+img.ID+")",
				)
			}
		}
	}

	// Initialize router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &Server{
		router:    r,
		ocrData:   ocrData,
		tempDir:   tempDir,
		hasImages: hasImages,
	}, nil
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	s.router.Get("/", s.handleIndex)
	s.router.Get("/page/{pageIndex}", s.handlePage)
	s.router.Get("/all", s.handleAllPages)
	s.router.Get("/images/{imageName}", s.handleImage)
	s.router.Get("/static/*", s.handleStatic)

	return s.router
}

// handleIndex displays the main page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/page/0", http.StatusTemporaryRedirect)
}

// handlePage displays a single page
func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	pageIndex, err := strconv.Atoi(chi.URLParam(r, "pageIndex"))
	if err != nil || pageIndex < 0 || pageIndex >= len(s.ocrData.Pages) {
		http.Error(w, "Invalid page index", http.StatusBadRequest)
		return
	}

	component := views.Page(s.ocrData, pageIndex)
	component.Render(r.Context(), w)
}

// handleAllPages displays all pages
func (s *Server) handleAllPages(w http.ResponseWriter, r *http.Request) {
	component := views.AllPages(s.ocrData)
	component.Render(r.Context(), w)
}

// handleImage serves image files
func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	imageName := chi.URLParam(r, "imageName")
	imagePath := filepath.Join(s.tempDir, imageName)

	// Simple security check to prevent directory traversal
	if !strings.HasPrefix(imagePath, s.tempDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, imagePath)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("static"))
	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}
