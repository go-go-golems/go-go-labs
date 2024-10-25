package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-go-golems/clay/pkg/autoreload"
	"github.com/go-go-golems/clay/pkg/watcher"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/go-go-golems/go-go-labs/pkg/svg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:   "svg",
	Short: "Render and serve SVG from YAML DSL",
	Long:  `A command-line tool to render and serve SVG files from YAML DSL input.`,
}

var renderCmd = &cobra.Command{
	Use:   "render <input-file>",
	Short: "Render SVG from YAML DSL",
	Run:   runRender,
}

var serveCmd = &cobra.Command{
	Use:   "serve <directory>",
	Short: "Serve and watch SVG files from YAML DSL",
	Run:   runServe,
}

var outputFile string
var port int

func init() {
	renderCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default is stdout)")
	_ = renderCmd.MarkFlagRequired("input")

	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")

	rootCmd.AddCommand(renderCmd, serveCmd)
}

func runRender(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Input file is required")
		_ = cmd.Usage()
		os.Exit(1)
	}

	inputFile := args[0]
	svgOutput := renderSVGFromYAML(inputFile)

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(svgOutput), 0644)
		if err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("SVG written to %s\n", outputFile)
	} else {
		fmt.Println(svgOutput)
	}
}

func runServe(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Directory to watch is required")
		_ = cmd.Usage()
		os.Exit(1)
	}

	directory := args[0]

	// Create a new WebSocket server instance
	wsServer := autoreload.NewWebSocketServer()

	// Set up the WebSocket handler
	http.HandleFunc("/ws", wsServer.WebSocketHandler())

	// Serve the JavaScript snippet at a specific endpoint
	http.HandleFunc("/autoreload.js", func(w http.ResponseWriter, r *http.Request) {
		js := wsServer.GetJavaScript("/ws")
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write([]byte(js))
	})

	// Serve the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderMainPage(w, directory)
	})

	// Set up the file watcher
	w := watcher.NewWatcher(
		watcher.WithPaths(directory),
		watcher.WithMask("**/*.yaml", "**/*.yml"),
		watcher.WithWriteCallback(func(path string) error {
			log.Printf("File changed: %s\n", path)
			wsServer.Broadcast("reload")
			return nil
		}),
	)

	// Start the watcher in a goroutine
	go func() {
		if err := w.Run(cmd.Context()); err != nil {
			log.Printf("Watcher error: %v\n", err)
		}
	}()

	// Start the HTTP server
	log.Printf("Server started on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func renderSVGFromYAML(inputFile string) string {
	// Read input YAML file
	yamlData, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Create an Emrichen interpreter
	interpreter, err := emrichen.NewInterpreter()
	if err != nil {
		fmt.Printf("Error creating Emrichen interpreter: %v\n", err)
		os.Exit(1)
	}

	// Process the YAML with Emrichen
	var processedYAML interface{}
	err = yaml.Unmarshal(yamlData, interpreter.CreateDecoder(&processedYAML))
	if err != nil {
		fmt.Printf("Error processing YAML with Emrichen: %v\n", err)
		os.Exit(1)
	}

	// Marshal the processed YAML back to bytes
	processedYAMLBytes, err := yaml.Marshal(processedYAML)
	if err != nil {
		fmt.Printf("Error marshaling processed YAML: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML into SVGDSL struct
	var dsl svg.SVGDSL
	err = yaml.Unmarshal(processedYAMLBytes, &dsl)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Render SVG
	svgOutput, err := svg.RenderSVG(&dsl.SVG)
	if err != nil {
		fmt.Printf("Error rendering SVG: %v\n", err)
		os.Exit(1)
	}

	return svgOutput
}

func renderMainPage(w http.ResponseWriter, directory string) {
	files, err := doublestar.FilepathGlob(filepath.Join(directory, "*.{yaml,yml}"))
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, "<html><head><title>SVG Renderer</title><script src='/autoreload.js'></script></head><body>")

	for _, file := range files {
		fileName := filepath.Base(file)
		sectionTitle := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		_, _ = fmt.Fprintf(w, "<h2>%s</h2>", sectionTitle)

		svgOutput, err := renderSVGFromYAMLWithErrorAndPanic(file)
		if err != nil {
			_, _ = fmt.Fprintf(w, "<div style='color: red; white-space: pre-wrap;'>Error rendering SVG: %+v</div>", err)
		} else {
			_, _ = fmt.Fprintf(w, "<div>%s</div>", svgOutput)
		}
	}

	_, _ = fmt.Fprintf(w, "</body></html>")
}

// Updated function to handle errors and panics
func renderSVGFromYAMLWithErrorAndPanic(inputFile string) (svgOutput string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v\n\nStack trace:\n%s", r, debug.Stack())
		}
	}()

	// Read input YAML file
	yamlData, err := os.ReadFile(inputFile)
	if err != nil {
		return "", fmt.Errorf("error reading input file: %w", err)
	}

	// Create an Emrichen interpreter
	interpreter, err := emrichen.NewInterpreter()
	if err != nil {
		return "", fmt.Errorf("error creating Emrichen interpreter: %w", err)
	}

	// Process the YAML with Emrichen
	var processedYAML interface{}
	err = yaml.Unmarshal(yamlData, interpreter.CreateDecoder(&processedYAML))
	if err != nil {
		return "", fmt.Errorf("error processing YAML with Emrichen: %w", err)
	}

	// Marshal the processed YAML back to bytes
	processedYAMLBytes, err := yaml.Marshal(processedYAML)
	if err != nil {
		return "", fmt.Errorf("error marshaling processed YAML: %w", err)
	}

	// Parse YAML into SVGDSL struct
	var dsl svg.SVGDSL
	err = yaml.Unmarshal(processedYAMLBytes, &dsl)
	if err != nil {
		return "", fmt.Errorf("error parsing YAML: %w", err)
	}

	// Render SVG
	svgOutput, err = svg.RenderSVG(&dsl.SVG)
	if err != nil {
		return "", fmt.Errorf("error rendering SVG: %w", err)
	}

	return svgOutput, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
