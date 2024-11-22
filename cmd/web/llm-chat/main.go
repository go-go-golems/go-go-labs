package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-labs/pkg/mockbot"
	"github.com/go-go-golems/go-go-labs/pkg/sse"
	"github.com/google/uuid"
)

var (
	debug    = true
	eventBus = sse.NewEventBus()
	bot      = mockbot.NewBotV2(eventBus)
)

func debugLog(format string, v ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	debugLog("Received %s request from %s", r.Method, r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		// Create new client for SSE connection
		client := &sse.Client{
			ID:      uuid.New().String(),
			Events:  make(chan sse.Event, 100),
			Done:    make(chan struct{}),
			Request: r,
		}
		debugLog("New SSE connection established for client: %s", client.ID)

		// Send client ID as first event
		initEvent := sse.Event{
			Type:    "connected",
			Content: client.ID,
		}

		// Register client and start SSE connection
		eventBus.Register <- client
		client.Events <- initEvent
		sse.HandleSSE(w, r, client)

	case http.MethodPost:
		var req mockbot.ChatRequest
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
		client, exists := eventBus.GetClient(clientID)
		if !exists {
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}

		debugLog("Received message from client %s: %+v", clientID, req)

		// Handle message in background
		go bot.HandleMessage(client, req)

		// Return success immediately
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// handleConversation handles conversation management operations
func handleConversation(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get conversation
		conv, exists := bot.GetConversation(clientID)
		if !exists {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}

		// Return conversation as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conv)

	case http.MethodPost:
		// Handle save/load operations
		operation := r.URL.Query().Get("op")
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Ensure filename is within the conversations directory
		filename = filepath.Join("conversations", filename)

		var err error
		switch operation {
		case "save":
			err = bot.SaveConversation(clientID, filename)
		case "load":
			err = bot.LoadConversation(clientID, filename)
		default:
			http.Error(w, "Invalid operation", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	// Enable line numbers in logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	debugLog("Server initializing...")

	// Create conversations directory if it doesn't exist
	if err := os.MkdirAll("conversations", 0755); err != nil {
		log.Fatal(err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/conversation", handleConversation)

	debugLog("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
