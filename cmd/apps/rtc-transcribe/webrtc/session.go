package webrtc

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

const (
	sessionCleanupInterval = 5 * time.Minute
	sessionMaxAge          = 30 * time.Minute
	sessionMaxIdleTime     = 15 * time.Minute
)

// WebRTCSession represents a WebRTC session with its peer connection and related state
type WebRTCSession struct {
	ID             string
	PeerConnection *webrtc.PeerConnection
	WebSocket      *websocket.Conn
	wsMutex        sync.Mutex // Mutex to protect concurrent writes to the WebSocket
	CreatedAt      time.Time
	LastActivity   time.Time
}

// SessionManager handles tracking and cleanup of WebRTC sessions
type SessionManager struct {
	sessions map[string]*WebRTCSession
	mutex    sync.RWMutex
	stopChan chan struct{} // Channel to signal cleanup goroutine to stop
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*WebRTCSession),
		stopChan: make(chan struct{}),
	}

	// Start session cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

// GenerateSessionID creates a new unique session ID
func GenerateSessionID() string {
	return uuid.NewString()
}

// CreateSession creates a new WebRTC session
func (sm *SessionManager) CreateSession(id string, useIceServers bool) (*WebRTCSession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if session already exists
	if _, exists := sm.sessions[id]; exists {
		log.Warn().Str("sessionID", id).Msg("Attempted to create session that already exists")
		// Decide on behavior: return existing or error? Let's return existing for now.
		// If strict creation is needed, return an error here.
		return sm.sessions[id], nil // Or return nil, errors.New("session already exists")
	}

	// Create peer connection
	pc, err := CreatePeerConnection(useIceServers) // Pass useIceServers flag
	if err != nil {
		log.Error().Err(err).Str("sessionID", id).Msg("Failed to create peer connection for session")
		return nil, err
	}

	// Create and store session
	session := &WebRTCSession{
		ID:             id,
		PeerConnection: pc,
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
	}

	sm.sessions[id] = session

	log.Info().
		Str("sessionID", id).
		Msg("Created new WebRTC session")

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) (*WebRTCSession, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[id]
	if exists {
		// Optionally update last activity on access, depending on desired idle timeout logic
		// session.LastActivity = time.Now() // Uncomment if access should reset idle timer
	}
	return session, exists
}

// SetWebSocket associates a WebSocket connection with a session
func (sm *SessionManager) SetWebSocket(id string, ws *websocket.Conn) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[id]
	if !exists {
		log.Warn().Str("sessionID", id).Msg("Attempted to set WebSocket for non-existent session")
		return false
	}

	// Close existing WebSocket if any (shouldn't happen with current flow, but good practice)
	if session.WebSocket != nil {
		log.Warn().Str("sessionID", id).Msg("Replacing existing WebSocket in session")
		session.WebSocket.Close()
	}

	session.WebSocket = ws
	session.LastActivity = time.Now() // Update activity when WS connects

	log.Info().
		Str("sessionID", id).
		Msg("Associated WebSocket with session")

	return true
}

// RemoveSession removes a session and cleans up resources
func (sm *SessionManager) RemoveSession(id string) {
	sm.mutex.Lock()
	// Get session first while holding write lock
	session, exists := sm.sessions[id]
	if exists {
		delete(sm.sessions, id)
	}
	sm.mutex.Unlock() // Unlock before potentially long-running Close() calls

	if exists {
		log.Info().
			Str("sessionID", id).
			Msg("Removing WebRTC session")

		// Close peer connection if it exists
		if session.PeerConnection != nil {
			// Use a timeout for closing to prevent blocking indefinitely
			closeTimeout := 2 * time.Second
			closeErrChan := make(chan error, 1)
			go func() {
				closeErrChan <- session.PeerConnection.Close()
			}()

			select {
			case err := <-closeErrChan:
				if err != nil {
					log.Error().
						Err(err).
						Str("sessionID", id).
						Msg("Error closing peer connection")
				} else {
					log.Debug().Str("sessionID", id).Msg("Peer connection closed")
				}
			case <-time.After(closeTimeout):
				log.Warn().
					Str("sessionID", id).
					Dur("timeout", closeTimeout).
					Msg("Timed out closing peer connection")
			}
		}

		// Close WebSocket if it exists
		if session.WebSocket != nil {
			if err := session.WebSocket.Close(); err != nil {
				// Ignore "use of closed network connection" errors as they are expected if PC closed first
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) && err.Error() != "use of closed network connection" {
					log.Error().
						Err(err).
						Str("sessionID", id).
						Msg("Error closing WebSocket")
				} else {
					log.Debug().Str("sessionID", id).Msg("WebSocket closed")
				}
			}
		}

	} else {
		log.Warn().Str("sessionID", id).Msg("Attempted to remove non-existent session")
	}
}

// UpdateActivity updates the last activity timestamp for a session
func (sm *SessionManager) UpdateActivity(id string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if session, exists := sm.sessions[id]; exists {
		session.LastActivity = time.Now()
	}
}

// cleanupLoop periodically removes inactive sessions
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sessionCleanupInterval)
	defer ticker.Stop()

	log.Info().Dur("interval", sessionCleanupInterval).Msg("Session cleanup loop started")

	for {
		select {
		case <-ticker.C:
			sm.cleanupSessions()
		case <-sm.stopChan:
			log.Info().Msg("Session cleanup loop stopped")
			return
		}
	}
}

// cleanupSessions performs the actual cleanup logic
func (sm *SessionManager) cleanupSessions() {
	sm.mutex.RLock() // Use RLock first to identify candidates

	now := time.Now()
	var sessionsToRemove []string

	log.Debug().Int("active_sessions", len(sm.sessions)).Msg("Running session cleanup check")
	for id, session := range sm.sessions {
		isOld := now.Sub(session.CreatedAt) > sessionMaxAge
		isIdle := now.Sub(session.LastActivity) > sessionMaxIdleTime
		if isOld || isIdle {
			sessionsToRemove = append(sessionsToRemove, id)
			log.Info().
				Str("sessionID", id).
				Bool("isOld", isOld).
				Dur("age", now.Sub(session.CreatedAt)).
				Bool("isIdle", isIdle).
				Dur("idleTime", now.Sub(session.LastActivity)).
				Msg("Flagging inactive/old session for removal")
		}
	}
	sm.mutex.RUnlock()

	// Remove the flagged sessions
	if len(sessionsToRemove) > 0 {
		log.Info().Int("count", len(sessionsToRemove)).Msg("Removing inactive/old sessions")
		for _, id := range sessionsToRemove {
			sm.RemoveSession(id) // This acquires the write lock internally
		}
	} else {
		log.Debug().Msg("No inactive/old sessions found to remove")
	}
}

// StopCleanup gracefully stops the cleanup goroutine
func (sm *SessionManager) StopCleanup() {
	close(sm.stopChan)
}

// Initialize a global session manager
var GlobalSessionManager = NewSessionManager()
