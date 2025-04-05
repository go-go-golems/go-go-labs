//go:build !embed
// +build !embed

package main

import (
	"github.com/gorilla/mux"
)

// SetupStaticFiles is a no-op in debug mode
func SetupStaticFiles(r *mux.Router) {
	// In debug mode, we don't serve static files from the backend
	// The frontend development server will handle this
}
