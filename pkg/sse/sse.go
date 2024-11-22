package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Debug flag for logging
var debug = true

func debugLog(format string, v ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Event represents a server-sent event
type Event struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Client represents a connected SSE client
type Client struct {
	ID      string
	Events  chan Event
	Done    chan struct{}
	Request *http.Request
}

// EventBus manages all SSE clients and message distribution
type EventBus struct {
	clients map[string]*Client
	mu      sync.RWMutex

	Register   chan *Client
	Unregister chan *Client
}

// NewEventBus creates a new event bus instance
func NewEventBus() *EventBus {
	eb := &EventBus{
		clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
	go eb.Run()
	return eb
}

// Run starts the event bus loop
func (eb *EventBus) Run() {
	debugLog("EventBus started")
	for {
		select {
		case client := <-eb.Register:
			eb.mu.Lock()
			eb.clients[client.ID] = client
			eb.mu.Unlock()
			debugLog("Client registered: %s", client.ID)

		case client := <-eb.Unregister:
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

// GetClient returns a client by ID
func (eb *EventBus) GetClient(id string) (*Client, bool) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	client, exists := eb.clients[id]
	return client, exists
}

// writeSSE writes an event to the response writer
func writeSSE(w http.ResponseWriter, event Event) error {
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

// HandleSSE handles the SSE connection for a client
func HandleSSE(w http.ResponseWriter, r *http.Request, client *Client) error {
	debugLog("Starting SSE handler for client: %s", client.ID)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Keep connection open and stream events
	for {
		select {
		case event, ok := <-client.Events:
			if !ok {
				debugLog("Events channel closed for client: %s", client.ID)
				return nil
			}
			if err := writeSSE(w, event); err != nil {
				debugLog("Error writing SSE: %v", err)
				return err
			}
		case <-client.Done:
			debugLog("Client done signal received: %s", client.ID)
			return nil
		case <-r.Context().Done():
			debugLog("Request context done: %s", client.ID)
			return nil
		}
	}
}
