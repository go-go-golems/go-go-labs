package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// SSEClient represents a Server-Sent Events client
type SSEClient struct {
	ID       string
	Channel  chan []byte
	Request  *http.Request
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	lastPing time.Time
}

// SSEManager manages Server-Sent Events connections
type SSEManager struct {
	clients sync.Map
	mutex   sync.RWMutex
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	manager := &SSEManager{}
	
	// Start cleanup routine for dead connections
	go manager.cleanupRoutine()
	
	return manager
}

// cleanupRoutine periodically removes dead connections
func (m *SSEManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		m.clients.Range(func(key, value interface{}) bool {
			client := value.(*SSEClient)
			if now.Sub(client.lastPing) > 60*time.Second {
				m.removeClient(client.ID)
			}
			return true
		})
	}
}

// addClient adds a new SSE client
func (m *SSEManager) addClient(client *SSEClient) {
	m.clients.Store(client.ID, client)
	log.Info().Str("clientID", client.ID).Msg("SSE client connected")
}

// removeClient removes an SSE client
func (m *SSEManager) removeClient(clientID string) {
	if value, ok := m.clients.LoadAndDelete(clientID); ok {
		client := value.(*SSEClient)
		close(client.Channel)
		log.Info().Str("clientID", clientID).Msg("SSE client disconnected")
	}
}

// broadcast sends an event to all connected clients
func (m *SSEManager) broadcast(event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal SSE event")
		return
	}

	m.clients.Range(func(key, value interface{}) bool {
		client := value.(*SSEClient)
		select {
		case client.Channel <- data:
		default:
			// Channel is full, remove client
			m.removeClient(client.ID)
		}
		return true
	})
}

// Event broadcasting methods

func (m *SSEManager) BroadcastAgentStatusChanged(agentID, oldStatus, newStatus string, agent *models.Agent) {
	event := SSEEvent{
		Event: "agent_status_changed",
		Data: map[string]interface{}{
			"agent_id":   agentID,
			"old_status": oldStatus,
			"new_status": newStatus,
			"agent":      agent,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastAgentEventCreated(agentID string, agentEvent *models.Event) {
	event := SSEEvent{
		Event: "agent_event_created",
		Data: map[string]interface{}{
			"agent_id": agentID,
			"event":    agentEvent,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastAgentQuestionPosted(agentID, question string, agent *models.Agent) {
	event := SSEEvent{
		Event: "agent_question_posted",
		Data: map[string]interface{}{
			"agent_id": agentID,
			"question": question,
			"agent":    agent,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastAgentProgressUpdated(agentID string, progress, filesChanged, linesAdded, linesRemoved int) {
	event := SSEEvent{
		Event: "agent_progress_updated",
		Data: map[string]interface{}{
			"agent_id":      agentID,
			"progress":      progress,
			"files_changed": filesChanged,
			"lines_added":   linesAdded,
			"lines_removed": linesRemoved,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastTodoUpdated(agentID string, todo *models.TodoItem, action string) {
	event := SSEEvent{
		Event: "todo_updated",
		Data: map[string]interface{}{
			"agent_id": agentID,
			"todo":     todo,
			"action":   action,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastTaskAssigned(task *models.Task, agentID string) {
	event := SSEEvent{
		Event: "task_assigned",
		Data: map[string]interface{}{
			"task":     task,
			"agent_id": agentID,
		},
	}
	m.broadcast(event)
}

func (m *SSEManager) BroadcastCommandReceived(agentID string, command *models.Command) {
	event := SSEEvent{
		Event: "command_received",
		Data: map[string]interface{}{
			"agent_id": agentID,
			"command":  command,
		},
	}
	m.broadcast(event)
}

// SSEHandler handles Server-Sent Events connections
func (h *Handlers) SSEHandler(w http.ResponseWriter, r *http.Request) {
	// Check if connection supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErrorResponse(w, http.StatusInternalServerError, "SSE_NOT_SUPPORTED", "Server-Sent Events not supported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create client
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	client := &SSEClient{
		ID:       clientID,
		Channel:  make(chan []byte, 10),
		Request:  r,
		Writer:   w,
		Flusher:  flusher,
		lastPing: time.Now(),
	}

	// Add client to manager
	h.sse.addClient(client)
	defer h.sse.removeClient(clientID)

	// Send initial connection event
	initialData := SSEEvent{
		Event: "connected",
		Data: map[string]interface{}{
			"client_id": clientID,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	data, _ := json.Marshal(initialData)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	// Start ping routine
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Handle client connection
	for {
		select {
		case <-r.Context().Done():
			return
		case <-pingTicker.C:
			// Send ping
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
			client.lastPing = time.Now()
		case data := <-client.Channel:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
