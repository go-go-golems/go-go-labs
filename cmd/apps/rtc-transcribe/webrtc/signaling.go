package webrtc

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SDPExchange represents the SDP offer or answer exchanged between peers
type SDPExchange struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer" or "answer"
	// SessionID is included by the client in the offer
	// SessionID string `json:"sessionId,omitempty"` // Keep commented if passed via query param
}

// Global variable to store the useIceServers flag
var UseIceServers bool

// HandleOffer handles incoming WebRTC offers from clients
func HandleOffer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	// Get SessionID from query parameter (preferred over body)
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		// Generate a new Session ID if none provided (optional, depends on client flow)
		// sessionID = GenerateSessionID() // Assuming GenerateSessionID exists in session.go
		// For now, require the client to send it.
		log.Error().Msg("Missing session ID in offer request query parameter")
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	logger := log.With().
		Str("component", "WebRTCSignaling").
		Str("sessionID", sessionID). // Include session ID in logs
		Str("remoteAddr", r.RemoteAddr).
		Str("userAgent", r.UserAgent()).
		Str("xForwardedFor", r.Header.Get("X-Forwarded-For")).
		Logger()

	logger.Info().Msg("Received WebRTC offer request")

	// Track the overall connection setup time
	defer func() {
		logger.Info().
			Dur("setupDuration", time.Since(startTime)).
			Msg("WebRTC offer processing completed") // Changed log message
	}()

	var sdp SDPExchange
	if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
		logger.Error().Err(err).Msg("Failed to decode SDP offer")
		http.Error(w, "Invalid SDP", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("sdpType", sdp.Type).
		Int("sdpLength", len(sdp.SDP)).
		Bool("useIceServers", UseIceServers).
		Msg("Decoded SDP offer")

	// --- Session Management Integration ---
	// Get or create the session associated with this ID
	// This creates the PeerConnection internally now.
	session, err := GlobalSessionManager.CreateSession(sessionID, UseIceServers)
	if err != nil {
		// CreateSession logs the specific error
		logger.Error().Err(err).Msg("Failed to get or create session")
		http.Error(w, "Failed to initialize session", http.StatusInternalServerError)
		return
	}
	peerConn := session.PeerConnection // Get the PC from the session
	// ------------------------------------

	// Set the remote description (client's offer)
	remoteDescStartTime := time.Now()
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}
	if err := peerConn.SetRemoteDescription(offer); err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to set remote description")).Msg("WebRTC error")
		http.Error(w, "Invalid remote description", http.StatusInternalServerError)
		GlobalSessionManager.RemoveSession(sessionID) // Clean up failed session
		return
	}
	logger.Debug().
		Dur("duration", time.Since(remoteDescStartTime)).
		Msg("Set remote description")

	// Set up audio track handler with enhanced logging
	handlerStartTime := time.Now()
	SetupAudioTrackHandler(peerConn)
	logger.Debug().
		Dur("duration", time.Since(handlerStartTime)).
		Msg("Audio track handler set up")

	// Create answer with enhanced logging
	answerStartTime := time.Now()
	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to create answer")).Msg("WebRTC error")
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		GlobalSessionManager.RemoveSession(sessionID) // Clean up failed session
		return
	}
	logger.Debug().
		Dur("duration", time.Since(answerStartTime)).
		Msg("Answer created")

	// Set local description (our answer) with enhanced logging
	localDescStartTime := time.Now()
	if err := peerConn.SetLocalDescription(answer); err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to set local description")).Msg("WebRTC error")
		http.Error(w, "Failed to set local description", http.StatusInternalServerError)
		GlobalSessionManager.RemoveSession(sessionID) // Clean up failed session
		return
	}
	logger.Debug().
		Dur("duration", time.Since(localDescStartTime)).
		Msg("Set local description")

	// Send answer back to client
	// The LocalDescription might be nil if SetLocalDescription failed, check?
	if peerConn.LocalDescription() == nil {
		logger.Error().Msg("Local description is nil after SetLocalDescription succeeded, unexpected state")
		http.Error(w, "Failed to get local description", http.StatusInternalServerError)
		GlobalSessionManager.RemoveSession(sessionID) // Clean up potentially broken session
		return
	}

	resp := SDPExchange{
		SDP:  peerConn.LocalDescription().SDP,
		Type: "answer",
	}

	w.Header().Set("Content-Type", "application/json")
	encodeStartTime := time.Now()
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Log error, but client might have disconnected. Session cleanup will handle it eventually.
		logger.Error().Err(errors.Wrap(err, "failed to encode SDP answer")).Msg("WebRTC error")
		// Don't return http.Error here as headers might already be sent.
		// GlobalSessionManager.RemoveSession(sessionID) // Let the cleanup goroutine handle this?
		return
	}
	logger.Debug().
		Dur("duration", time.Since(encodeStartTime)).
		Int("responseSize", len(resp.SDP)).
		Msg("Encoded and sent SDP answer")

	// Note: The connection setup is NOT complete here. ICE exchange happens via WebSocket.
	// The log message in defer is more accurate now.
}
