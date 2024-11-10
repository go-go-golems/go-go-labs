package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

type AppState struct {
	CurrentStep        string
	JobDescription     string
	UploadedFile       *UploadedFile
	GeneratedJobConfig *JobConfig
	ChatHistory        []ChatMessage
	ShowDiagram        bool
	Error              *AppError
}

type UploadedFile struct {
	Name     string
	Size     int64
	Location string
}

type JobConfig struct {
	JobName     string `json:"jobName"`
	Queue       string `json:"queue"`
	Runtime     string `json:"runtime"`
	CPUs        int    `json:"cpus"`
	CoresPerCPU int    `json:"cores_per_cpu"`
	TotalCores  int    `json:"total_cores"`
	Memory      string `json:"memory"`
	Script      string `json:"script"`
	Priority    string `json:"priority"`
}

type ChatMessage struct {
	Role        string   `json:"role"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments,omitempty"`
}

type AppError struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func main() {
	cmd := &cobra.Command{
		Use:   "hpc-ui",
		Short: "Start the HPC job submission UI server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context())
		},
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx context.Context) error {
	// Parse templates in the correct order - base template first
	tmpl, err := template.New("base.html").ParseFS(templatesFS,
		"templates/base.html",
		"templates/*.html",
	)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/", handleIndex(tmpl))
	mux.HandleFunc("/submit-job", handleJobSubmission(tmpl)) // Pass template to handler
	mux.HandleFunc("/upload", handleFileUpload())
	mux.HandleFunc("/chat", handleChat())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Println("Starting server on :8080")
	return server.ListenAndServe()
}

func handleIndex(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := AppState{
			CurrentStep: "input",
			ShowDiagram: false,
		}
		if err := tmpl.ExecuteTemplate(w, "index.html", state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Handler implementations to be added:
func handleJobSubmission(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		description := r.FormValue("description")

		// Mock job config generation
		config := &JobConfig{
			JobName:     "test_job",
			Queue:       "default",
			Runtime:     "1h",
			CPUs:        2,
			CoresPerCPU: 4,
			TotalCores:  8,
			Memory:      "16GB",
			Priority:    "normal",
		}

		state := AppState{
			CurrentStep:        "review",
			JobDescription:     description,
			GeneratedJobConfig: config,
			ShowDiagram:        true,
		}

		if err := tmpl.ExecuteTemplate(w, "job_result", state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleFileUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Mock file processing - just return the name and size
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div class="text-sm">
			<span class="font-semibold">File:</span> %s
			<br>
			<span class="font-semibold">Size:</span> %.2f KB
		</div>`,
			header.Filename,
			float64(header.Size)/1024)
	}
}

func handleChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock chat response
		response := `
		<div class="bg-gray-100 p-4 rounded">
			<div class="mb-4">
				<strong>Assistant:</strong> I've analyzed your job configuration. 
				The current setup will use 8 cores (2 CPUs with 4 cores each) and 16GB memory 
				on the default queue. How can I help you refine this configuration?
			</div>
			<div class="flex">
				<input type="text" 
					class="flex-grow border rounded py-2 px-3 mr-2" 
					placeholder="Type your message...">
				<button class="bg-blue-500 text-white px-4 py-2 rounded">Send</button>
			</div>
		</div>`

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, response)
	}
}
