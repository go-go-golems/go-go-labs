package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ocr-mistral-view",
		Short: "View OCR results from Mistral in a web browser",
		Long:  "A web application to view OCR results from Mistral in a user-friendly, browsable format",
		Run:   run,
	}

	rootCmd.Flags().StringP("input", "i", "", "Input JSON file path")
	rootCmd.Flags().StringP("port", "p", "8080", "Port to serve the web application")
	rootCmd.MarkFlagRequired("input")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	inputFile, _ := cmd.Flags().GetString("input")
	port, _ := cmd.Flags().GetString("port")

	// Initialize server
	server, err := NewServer(inputFile)
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	// Start server
	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
