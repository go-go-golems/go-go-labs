//go:build embed
// +build embed

package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed embed
var embeddedFiles embed.FS

// SetupStaticFiles configures the router to serve embedded static files
func SetupStaticFiles(r *mux.Router) {
	// Create a filesystem with just the embedded files
	fsys, err := fs.Sub(embeddedFiles, "embed")
	if err != nil {
		panic(err)
	}

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.FS(fsys)))
}
