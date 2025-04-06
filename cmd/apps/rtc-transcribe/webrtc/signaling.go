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
}

// Global variable to store the useIceServers flag
var UseIceServers bool

// HandleOffer handles incoming WebRTC offers from clients
func HandleOffer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = time.Now().Format("20060102-150405.000000")
	}

	logger := log.With().
		Str("component", "WebRTCSignaling").
		Str("requestID", requestID).
		Str("remoteAddr", r.RemoteAddr).
		Str("userAgent", r.UserAgent()).
		Str("xForwardedFor", r.Header.Get("X-Forwarded-For")).
		Logger()

	logger.Info().Msg("Received WebRTC offer request")

	// Track the overall connection setup time
	defer func() {
		logger.Info().
			Dur("setupDuration", time.Since(startTime)).
			Msg("WebRTC connection setup completed")
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

	// Create peer connection with enhanced logging
	peerConnStartTime := time.Now()
	peerConn, err := CreatePeerConnection(UseIceServers)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create peer connection")
		http.Error(w, "Failed to create peer connection", http.StatusInternalServerError)
		return
	}
	logger.Debug().
		Dur("duration", time.Since(peerConnStartTime)).
		Msg("Peer connection created")

	// Set the remote description (client's offer)
	remoteDescStartTime := time.Now()
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}
	if err := peerConn.SetRemoteDescription(offer); err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to set remote description")).Msg("WebRTC error")
		http.Error(w, "Invalid remote description", http.StatusInternalServerError)
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
		return
	}
	logger.Debug().
		Dur("duration", time.Since(localDescStartTime)).
		Msg("Set local description")

	// Send answer back to client
	resp := SDPExchange{
		SDP:  peerConn.LocalDescription().SDP,
		Type: "answer",
	}

	w.Header().Set("Content-Type", "application/json")
	encodeStartTime := time.Now()
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to encode SDP answer")).Msg("WebRTC error")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	logger.Debug().
		Dur("duration", time.Since(encodeStartTime)).
		Int("responseSize", len(resp.SDP)).
		Msg("Encoded and sent SDP answer")

	logger.Info().
		Dur("totalSetupTime", time.Since(startTime)).
		Msg("Successfully established WebRTC connection")
}
