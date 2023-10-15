package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"
)

// Embed the file into the program. It requires Go 1.16 or later.
//
//go:embed index.html
var content embed.FS

// Streamer is responsible for sending data over a channel.
func Streamer(dataChan chan string) {
	// This is just an example; you might want to gather strings from a different source.
	strings := []string{"first", "second", "third", "fourth", "fifth"}
	for _, s := range strings {
		time.Sleep(2 * time.Second) // simulate delay
		dataChan <- s
	}
	close(dataChan) // No more data to send, close the channel.
}

// SSEHandler handles the Server-Sent Events.
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set the necessary headers to instruct the client to expect an SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel of strings.
	dataChan := make(chan string)

	log.Println("SSEHandler: Starting streamer")

	// Start a goroutine to stream data to the channel.
	go Streamer(dataChan)

	ctx := r.Context() // Get the context from the request.

	for {
		select {
		case <-ctx.Done():
			// The client closed the connection or the server is shutting down.
			fmt.Println("Client has closed the connection or server is shutting down")
			return
		case data, open := <-dataChan:
			if !open {
				// If our data channel was closed, finish the request
				fmt.Println("Channel has been closed")
				return
			}
			// Write to the ResponseWriter
			// SSE data needs to be formatted as "data: <payload>\n\n"
			_, err := fmt.Fprintf(w, "data: %s\n\n", data)
			if err != nil {
				return
			}

			// Flush the data immediately instead of buffering it for later.
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		}
	}
}

func main() {
	// Create a new serve mux (router) for handling our routes.
	mux := http.NewServeMux()

	// Handle the SSE endpoint
	mux.HandleFunc("/events", SSEHandler)

	fsys, err := fs.Sub(content, ".")
	if err != nil {
		log.Fatal(err)
		return
	}
	fileServer := http.FileServer(http.FS(fsys))
	mux.Handle("/", fileServer)

	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", mux))
}
