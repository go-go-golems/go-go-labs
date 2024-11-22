package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Client represents a connected SSE client
type Client struct {
	ID      string
	Events  chan StreamEvent
	Done    chan struct{}
	Request *http.Request
}

// EventBus manages all SSE clients and message distribution
type EventBus struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func NewEventBus() *EventBus {
	return &EventBus{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

var (
	debug    = true
	eventBus = NewEventBus()
)

func debugLog(format string, v ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func (eb *EventBus) Run() {
	debugLog("EventBus started")
	for {
		select {
		case client := <-eb.register:
			eb.mu.Lock()
			eb.clients[client.ID] = client
			eb.mu.Unlock()
			debugLog("Client registered: %s", client.ID)

		case client := <-eb.unregister:
			eb.mu.Lock()
			if _, ok := eb.clients[client.ID]; ok {
				delete(eb.clients, client.ID)
				close(client.Events)
			}
			eb.mu.Unlock()
			debugLog("Client unregistered: %s", client.ID)
		}
	}
}

func writeSSE(w http.ResponseWriter, event StreamEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}
	debugLog("Sending SSE event: %s", string(data))

	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		return fmt.Errorf("failed to write SSE event: %v", err)
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	} else {
		return fmt.Errorf("response writer does not support flushing")
	}
	return nil
}

func handleSSE(w http.ResponseWriter, r *http.Request, client *Client) {
	debugLog("Starting SSE handler for client: %s", client.ID)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Register client
	eventBus.register <- client

	// Ensure cleanup
	defer func() {
		eventBus.unregister <- client
		// Don't close the Done channel here - it's closed in streamResponse
	}()

	// Keep connection open and stream events
	for {
		select {
		case event, ok := <-client.Events:
			if !ok {
				debugLog("Events channel closed for client: %s", client.ID)
				return
			}
			if err := writeSSE(w, event); err != nil {
				debugLog("Error writing SSE: %v", err)
				return
			}
		case <-client.Done:
			debugLog("Client done signal received: %s", client.ID)
			return
		case <-r.Context().Done():
			debugLog("Request context done: %s", client.ID)
			return
		}
	}
}

func streamResponse(client *Client, response string) {
	// Only close the Done channel here
	defer func() {
		debugLog("Closing done channel for client: %s", client.ID)
		// close(client.Done)
	}()

	debugLog("Starting response streaming for client: %s", client.ID)

	// Send "thinking" event
	client.Events <- StreamEvent{Type: "thinking", Content: ""}
	time.Sleep(1 * time.Second)

	// Split response into words and stream them
	words := strings.Fields(response)
	for i, word := range words {
		// Check if client is still connected
		select {
		case <-client.Done:
			debugLog("Client disconnected, stopping stream: %s", client.ID)
			return
		default:
			client.Events <- StreamEvent{Type: "token", Content: word}
			if i < len(words)-1 {
				client.Events <- StreamEvent{Type: "token", Content: " "}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	client.Events <- StreamEvent{Type: "done", Content: ""}
	debugLog("Finished streaming response for client: %s", client.ID)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	debugLog("Received %s request from %s", r.Method, r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		// Create new client for SSE connection
		client := &Client{
			ID:      uuid.New().String(),
			Events:  make(chan StreamEvent, 100),
			Done:    make(chan struct{}),
			Request: r,
		}
		debugLog("New SSE connection established for client: %s", client.ID)

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Send client ID as first event
		initEvent := StreamEvent{
			Type:    "connected",
			Content: client.ID,
		}
		if err := writeSSE(w, initEvent); err != nil {
			http.Error(w, "Failed to send init event", http.StatusInternalServerError)
			return
		}

		// Start SSE connection
		handleSSE(w, r, client)

	case http.MethodPost:
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get client ID from header
		clientID := r.Header.Get("X-Client-ID")
		if clientID == "" {
			http.Error(w, "Client ID required", http.StatusBadRequest)
			return
		}

		// Find existing client
		eventBus.mu.RLock()
		client, exists := eventBus.clients[clientID]
		eventBus.mu.RUnlock()

		if !exists {
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}

		debugLog("Received message from client %s: %+v", clientID, req)

		// Get the last message and start streaming response
		lastMessage := req.Messages[len(req.Messages)-1]
		response := reverseString(lastMessage.Content)
		go streamResponse(client, response)

		// Return success immediately
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	debugLog("Server initializing...")

	// Start event bus
	go eventBus.Run()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve index.html at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// Chat endpoint
	http.HandleFunc("/api/chat", handleChat)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
