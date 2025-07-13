package sse

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Subscriber represents a connected SSE client
type Subscriber struct {
	ID          string
	Writer      http.ResponseWriter
	Flusher     http.Flusher
	Closed      bool
	LastText    string
	ConnectedAt time.Time
	LastEventAt time.Time
	EventsSent  int
	RemoteAddr  string
	UserAgent   string
}

var (
	// subscribers is a map of all connected SSE clients
	subscribers     = make(map[string]*Subscriber)
	subscriberMutex sync.RWMutex
)

// NewSubscriber creates a new SSE subscriber
func NewSubscriber(w http.ResponseWriter, r *http.Request) (*Subscriber, error) {
	// Check if the response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	// Generate a random ID for this subscriber
	id := uuid.New().String()

	// Get client info
	remoteAddr := r.RemoteAddr
	userAgent := r.UserAgent()

	// Create new subscriber
	now := time.Now()
	return &Subscriber{
		ID:          id,
		Writer:      w,
		Flusher:     flusher,
		Closed:      false,
		ConnectedAt: now,
		LastEventAt: now,
		EventsSent:  0,
		RemoteAddr:  remoteAddr,
		UserAgent:   userAgent,
	}, nil
}

// SendEvent sends an SSE event to the subscriber
func (s *Subscriber) SendEvent(eventType, data string) error {
	if s.Closed {
		return errors.New("subscriber is closed")
	}

	logger := log.With().
		Str("component", "SSE").
		Str("action", "SendEvent").
		Str("subscriberID", s.ID).
		Str("eventType", eventType).
		Str("remoteAddr", s.RemoteAddr).
		Int("dataLength", len(data)).
		Logger()

	logger.Debug().Msg("Sending SSE event to subscriber")

	// Format the SSE event
	if eventType != "" {
		if _, err := fmt.Fprintf(s.Writer, "event: %s\n", eventType); err != nil {
			logger.Error().Err(err).Msg("Failed to write event type")
			return errors.Wrap(err, "failed to write event type")
		}
	}

	// Send the data
	if _, err := fmt.Fprintf(s.Writer, "data: %s\n\n", data); err != nil {
		logger.Error().Err(err).Msg("Failed to write event data")
		return errors.Wrap(err, "failed to write event data")
	}

	// Store the last sent text and update stats
	s.LastText = data
	s.LastEventAt = time.Now()
	s.EventsSent++

	connectionDuration := time.Since(s.ConnectedAt)
	logger.Debug().
		Int("totalEventsSent", s.EventsSent).
		Dur("connectionDuration", connectionDuration).
		Msg("SSE event sent successfully")

	// Flush to ensure it's sent immediately
	s.Flusher.Flush()
	return nil
}

// RegisterSubscriber adds a subscriber to the global subscribers map
func RegisterSubscriber(s *Subscriber) {
	logger := log.With().
		Str("component", "SSE").
		Str("action", "RegisterSubscriber").
		Str("subscriberID", s.ID).
		Str("remoteAddr", s.RemoteAddr).
		Str("userAgent", s.UserAgent).
		Logger()

	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	// Add to subscribers map
	subscribers[s.ID] = s

	// Log subscriber count
	subscriberCount := len(subscribers)
	logger.Info().
		Int("totalSubscribers", subscriberCount).
		Msg("New SSE subscriber registered")
}

// UnregisterSubscriber removes a subscriber from the global subscribers map
func UnregisterSubscriber(id string) {
	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	if subscriber, exists := subscribers[id]; exists {
		logger := log.With().
			Str("component", "SSE").
			Str("action", "UnregisterSubscriber").
			Str("subscriberID", id).
			Str("remoteAddr", subscriber.RemoteAddr).
			Int("eventsSent", subscriber.EventsSent).
			Dur("connectionDuration", time.Since(subscriber.ConnectedAt)).
			Logger()

		delete(subscribers, id)

		subscriberCount := len(subscribers)
		logger.Info().
			Int("totalSubscribers", subscriberCount).
			Msg("SSE subscriber unregistered")
	}
}

// BroadcastEvent sends an event to all subscribers
func BroadcastEvent(eventType, data string) {
	logger := log.With().
		Str("component", "SSE").
		Str("action", "BroadcastEvent").
		Str("eventType", eventType).
		Int("dataLength", len(data)).
		Logger()

	subscriberMutex.RLock()
	subscriberCount := len(subscribers)
	logger.Debug().
		Int("subscriberCount", subscriberCount).
		Msg("Broadcasting event to all subscribers")

	// If no subscribers, return early
	if subscriberCount == 0 {
		subscriberMutex.RUnlock()
		logger.Debug().Msg("No subscribers to broadcast to")
		return
	}

	var failedCount int
	var successCount int

	// Copy subscriber IDs to avoid holding the lock during sends
	subscriberIDs := make([]string, 0, subscriberCount)
	for id := range subscribers {
		subscriberIDs = append(subscriberIDs, id)
	}
	subscriberMutex.RUnlock()

	// Send to each subscriber without holding the lock
	for _, id := range subscriberIDs {
		// Re-acquire lock to get the subscriber
		subscriberMutex.RLock()
		subscriber, exists := subscribers[id]
		subscriberMutex.RUnlock()

		if !exists {
			continue
		}

		err := subscriber.SendEvent(eventType, data)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("subscriberID", id).
				Str("remoteAddr", subscriber.RemoteAddr).
				Msg("Failed to send event to subscriber")

			// Mark as closed for cleanup
			subscriber.Closed = true
			failedCount++
		} else {
			successCount++
		}
	}

	// Log broadcast results
	logger.Info().
		Int("totalSubscribers", subscriberCount).
		Int("succeededCount", successCount).
		Int("failedCount", failedCount).
		Msg("Broadcast complete")

	// Cleanup closed subscribers
	if failedCount > 0 {
		go cleanupClosedSubscribers()
	}
}

// cleanupClosedSubscribers removes subscribers that have been marked as closed
func cleanupClosedSubscribers() {
	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	closedCount := 0
	for id, subscriber := range subscribers {
		if subscriber.Closed {
			delete(subscribers, id)
			closedCount++
		}
	}

	if closedCount > 0 {
		log.Info().
			Str("component", "SSE").
			Int("closedCount", closedCount).
			Int("remainingCount", len(subscribers)).
			Msg("Cleaned up closed subscribers")
	}
}

// HandleSSE is the HTTP handler for the SSE endpoint
func HandleSSE(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("component", "SSE").
		Str("action", "HandleSSE").
		Str("remoteAddr", r.RemoteAddr).
		Str("userAgent", r.UserAgent()).
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()

	logger.Debug().Msg("Received SSE connection request")

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a new subscriber
	startTime := time.Now()
	subscriber, err := NewSubscriber(w, r)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("attemptDuration", time.Since(startTime)).
			Msg("Failed to create SSE subscriber")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	subscriberLogger := logger.With().
		Str("subscriberID", subscriber.ID).
		Logger()

	subscriberLogger.Info().Msg("Created new SSE subscriber")

	// Register the subscriber
	RegisterSubscriber(subscriber)

	// Send initial event to confirm connection
	if err := subscriber.SendEvent("connected", "Connected to transcription stream"); err != nil {
		subscriberLogger.Error().
			Err(err).
			Msg("Failed to send initial SSE event")
		UnregisterSubscriber(subscriber.ID)
		return
	}

	subscriberLogger.Info().Msg("SSE connection established successfully")

	// Wait for the client to disconnect
	notify := r.Context().Done()

	// Create a heartbeat ticker to keep the connection alive
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Handle disconnection or heartbeat in a separate goroutine
	disconnected := make(chan bool, 1)
	go func() {
		select {
		case <-notify:
			// Client disconnected
			subscriberLogger.Info().
				Dur("connectionDuration", time.Since(subscriber.ConnectedAt)).
				Int("eventsSent", subscriber.EventsSent).
				Msg("Client disconnected")
			disconnected <- true

		case <-heartbeatTicker.C:
			// Send a heartbeat to keep the connection alive
			if err := subscriber.SendEvent("heartbeat", fmt.Sprintf("%d", time.Now().UnixNano())); err != nil {
				subscriberLogger.Warn().
					Err(err).
					Dur("connectionDuration", time.Since(subscriber.ConnectedAt)).
					Msg("Failed to send heartbeat, closing connection")
				disconnected <- true
			}
		}
	}()

	// Wait for disconnection
	<-disconnected

	// Unregister subscriber when the connection is closed
	UnregisterSubscriber(subscriber.ID)
}

// SendTranscription broadcasts a transcription result to all subscribers
func SendTranscription(text string) {
	logger := log.With().
		Str("component", "SSE").
		Str("action", "SendTranscription").
		Int("textLength", len(text)).
		Logger()

	if text == "" {
		logger.Warn().Msg("Empty transcription text, not broadcasting")
		return
	}

	// Count words in the transcription
	wordCount := len(bytes.Fields([]byte(text)))

	logger.Info().
		Int("wordCount", wordCount).
		Str("text", text).
		Msg("Broadcasting transcription")

	BroadcastEvent("transcription", text)
}
