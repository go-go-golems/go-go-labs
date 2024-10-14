package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-go-golems/go-go-labs/pkg/snakemake"
	"github.com/spf13/cobra"
)

var (
	logFile string
	port    string
	host    string
)

var rootCmd = &cobra.Command{
	Use:   "snakelog",
	Short: "A web server to display Snakemake log information",
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&logFile, "log", "l", "logfiles/snakemake.log", "Path to the Snakemake log file")
	rootCmd.Flags().StringVarP(&port, "port", "p", "6060", "HTTP port to listen on")
	rootCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to bind the server to")
}

func run(cmd *cobra.Command, args []string) {
	logData, err := snakemake.ParseLog(logFile, true)
	if err != nil {
		fmt.Printf("Error parsing log file: %v\n", err)
		os.Exit(1)
	}

	tmpl := template.Must(template.ParseFiles("cmd/snakemake-viewer/templates/index.html", "cmd/snakemake-viewer/templates/job_details.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", logData)
	})

	http.HandleFunc("/job/", func(w http.ResponseWriter, r *http.Request) {
		jobID := strings.TrimPrefix(r.URL.Path, "/job/")
		for _, job := range logData.Jobs {
			if job.ID == jobID {
				tmpl.ExecuteTemplate(w, "job_details.html", job)
				return
			}
		}
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Server is running on http://%s\n", addr)
	http.ListenAndServe(addr, nil)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
