package webrtc

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"

	// "github.com/pkg/errors" // Not used currently, uncomment if needed
	"github.com/rs/zerolog/log"
)

// Message types for WebSocket communication
const (
	MessageTypeCandidate = "candidate"
	// Add other types like "offer", "answer", "error" if needed later
)

// Upgrader for HTTP to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		// For development, allow all origins.
		// Example: Check r.Header.Get("Origin") against allowed list
		return true
	},
}

// WebSocketMessage represents a generic message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`    // Correct tag format
	Payload json.RawMessage `json:"payload"` // Correct tag format
}

// CandidatePayload represents the structure expected for an ICE candidate payload
// Matches the structure sent by the browser's `event.candidate.toJSON()`
type CandidatePayload struct {
	Candidate        string `json:"candidate"`        // Correct tag format
	SDPMid           string `json:"sdpMid"`           // Correct tag format
	SDPMLineIndex    uint16 `json:"sdpMLineIndex"`    // Correct tag format
	UsernameFragment string `json:"usernameFragment"` // Correct tag format
}

// HandleWebSocket handles WebSocket connections for ICE candidate exchange
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		log.Warn().Msg("WebSocket connection attempt missing session ID")
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	wsLogger := log.With().Str("component", "WebSocket").Str("sessionID", sessionID).Logger()

	// Retrieve the session. It should have been created by HandleOffer beforehand.
	session, exists := GlobalSessionManager.GetSession(sessionID)
	if !exists {
		wsLogger.Error().Msg("WebSocket connection attempt for non-existent session")
		http.Error(w, "Invalid or expired session ID", http.StatusNotFound)
		return
	}

	wsLogger.Info().Msg("WebSocket connection request received")

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsLogger.Error().Err(err).Msg("Failed to upgrade to WebSocket")
		return
	}

	wsLogger.Info().Msg("WebSocket connection upgraded successfully")

	// Associate WebSocket with session
	associated := GlobalSessionManager.SetWebSocket(sessionID, ws)
	if !associated {
		wsLogger.Error().Msg("Failed to associate WebSocket with session (session likely removed)")
		_ = ws.Close()
		return
	}

	// --- Send Buffered Server Candidates --- //
	session.bufferMutex.Lock()
	bufferedCandidates := make([]webrtc.ICECandidateInit, len(session.candidateBuffer))
	copy(bufferedCandidates, session.candidateBuffer)
	session.candidateBuffer = session.candidateBuffer[:0] // Clear buffer after copying
	session.bufferMutex.Unlock()

	if len(bufferedCandidates) > 0 {
		wsLogger.Info().Int("count", len(bufferedCandidates)).Msg("Sending buffered server ICE candidates")
		for _, candidateInit := range bufferedCandidates {
			session.wsMutex.Lock() // Lock for writing to WS
			candidateJSON, err := json.Marshal(candidateInit)
			if err != nil {
				wsLogger.Error().Err(err).Msg("Failed to marshal buffered server ICE candidate")
				session.wsMutex.Unlock()
				continue
			}
			msg := WebSocketMessage{
				Type:    MessageTypeCandidate,
				Payload: candidateJSON,
			}
			if err := ws.WriteJSON(msg); err != nil {
				wsLogger.Error().Err(err).Msg("Failed to send buffered server ICE candidate via WebSocket")
				// Consider closing if write fails repeatedly
				session.wsMutex.Unlock()
				continue // Or break?
			}
			session.wsMutex.Unlock()
		}
	}
	// ---------------------------------- //

	// --- Setup Sending ICE Candidates ---
	// This must be set *after* the WS connection is associated with the session.
	session.PeerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			wsLogger.Debug().Msg("Server ICE candidate gathering complete")
			return
		}

		candidateJSON := candidate.ToJSON()
		payloadBytes, err := json.Marshal(candidateJSON)
		if err != nil {
			wsLogger.Error().Err(err).Msg("Failed to marshal server ICE candidate")
			return
		}

		msg := WebSocketMessage{
			Type:    MessageTypeCandidate,
			Payload: payloadBytes,
		}

		// Use the mutex from the session to protect concurrent writes
		session.wsMutex.Lock()
		wsToSend := session.WebSocket // Get the current WebSocket connection
		session.wsMutex.Unlock()

		if wsToSend == nil {
			wsLogger.Warn().Msg("WebSocket is nil when trying to send candidate, skipping")
			return
		}

		// Lock before writing
		session.wsMutex.Lock()
		err = wsToSend.WriteJSON(msg)
		session.wsMutex.Unlock()

		if err != nil {
			// Check if the error is due to a closed connection
			if errors.Is(err, net.ErrClosed) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				wsLogger.Warn().Err(err).Msg("Failed to send ICE candidate to client (connection closed)")
			} else {
				wsLogger.Error().Err(err).Msg("Failed to send ICE candidate to client")
			}
			// Consider closing the session or WS connection here if write errors persist
		} else {
			wsLogger.Debug().
				Str("candidate", candidateJSON.Candidate).
				// Str("type", string(candidate.Typ)). // Typ is deprecated, use ToJSON() fields
				// Str("address", candidate.Address). // Address is deprecated
				// Int("port", int(candidate.Port)). // Port is deprecated
				Msg("Sent server ICE candidate to client")
		}
	})

	wsLogger.Debug().Msg("OnICECandidate handler set for server-to-client candidates")

	// --- Start Reading Messages from Client ---
	go func() {
		defer func() {
			wsLogger.Info().Msg("WebSocket read loop exiting, cleaning up session")
			GlobalSessionManager.RemoveSession(sessionID) // Handles closing PC and WS
		}()

		wsLogger.Info().Msg("Starting WebSocket read loop")
		for {
			var msg WebSocketMessage
			err := ws.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					wsLogger.Info().Msgf("WebSocket closed by client or network: %v", err)
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					wsLogger.Warn().Err(err).Msg("WebSocket read timeout")
				} else {
					wsLogger.Error().Err(err).Msg("WebSocket read error")
				}
				break // Exit loop on any error
			}

			GlobalSessionManager.UpdateActivity(sessionID)
			wsLogger.Debug().Str("messageType", msg.Type).Msg("Received message from client")

			switch msg.Type {
			case MessageTypeCandidate:
				var payload CandidatePayload
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					wsLogger.Error().Err(err).Str("payload", string(msg.Payload)).Msg("Failed to unmarshal client ICE candidate payload")
					continue
				}

				wsLogger.Debug().
					Str("candidate", payload.Candidate).
					Str("sdpMid", payload.SDPMid).
					Uint16("sdpMLineIndex", payload.SDPMLineIndex).
					Msg("Received client ICE candidate")

				// Need pointer values for pion's ICECandidateInit
				sdpMidPtr := &payload.SDPMid
				// Handle empty sdpMid from browser (might be null)
				if payload.SDPMid == "" {
					sdpMidPtr = nil
				}
				sdpMLineIndexPtr := &payload.SDPMLineIndex

				candidateInit := webrtc.ICECandidateInit{
					Candidate:     payload.Candidate,
					SDPMid:        sdpMidPtr,
					SDPMLineIndex: sdpMLineIndexPtr,
				}

				if session.PeerConnection == nil {
					wsLogger.Error().Msg("PeerConnection is nil when trying to add client candidate")
					continue
				}

				connState := session.PeerConnection.ICEConnectionState()
				if connState == webrtc.ICEConnectionStateClosed || connState == webrtc.ICEConnectionStateFailed {
					wsLogger.Warn().Str("state", connState.String()).Msg("Skipping add candidate on closed/failed connection")
					continue
				}

				if err := session.PeerConnection.AddICECandidate(candidateInit); err != nil {
					if strings.Contains(err.Error(), "invalid state") && session.PeerConnection.RemoteDescription() == nil {
						wsLogger.Warn().Err(err).Msg("Failed to add client ICE candidate (likely due to unset remote description, pion should buffer)")
					} else {
						wsLogger.Error().Err(err).Msg("Failed to add client ICE candidate")
					}
					continue
				}

				wsLogger.Debug().
					Str("candidate", candidateInit.Candidate).
					Msg("Successfully added client ICE candidate to peer connection")

			default:
				wsLogger.Warn().Str("type", msg.Type).Msg("Received unhandled WebSocket message type")
			}
		}
	}()

	wsLogger.Info().Msg("WebSocket handler setup complete, read loop started.")
}

// Removed unnecessary global mutexes and misplaced import

// TODO: Consider adding mutexes for concurrent WebSocket writes if needed.
// var wsWriteMutex sync.Mutex // Example global mutex (better per-connection)

// Add any other necessary imports here

// Add any other necessary global variables here

// Add any other necessary functions here
