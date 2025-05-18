package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed static
var staticFiles embed.FS

func spa() http.Handler {
	// Get a sub-filesystem rooted at the static directory
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("Failed to get sub filesystem:", err)
	}
	fileServer := http.FileServer(http.FS(static))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First, try to serve api requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			return
		}

		// For any other path, check if the file exists
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		_, err := static.Open(path)
		if err != nil {
			// File doesn't exist, serve index.html for client-side routing
			r.URL.Path = "/index.html"
		}

		fileServer.ServeHTTP(w, r)
	})
}

// Example API handler - replace with your actual API implementations
func apiRoutes() http.Handler {
	mux := http.NewServeMux()

	// Sample API endpoint
	mux.HandleFunc("/api/widgets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1,"name":"Widget 1"},{"id":2,"name":"Widget 2"}]`))
	})

	return mux
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/api/", apiRoutes())
	mux.Handle("/", spa())

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
