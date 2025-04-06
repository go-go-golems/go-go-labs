package webrtc

import (
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// CreatePeerConnection creates and configures a new WebRTC peer connection
func CreatePeerConnection(useIceServers bool) (*webrtc.PeerConnection, error) {
	logger := log.With().
		Str("component", "WebRTCPeer").
		Str("connectionID", time.Now().Format("20060102-150405.000000")).
		Logger()

	startTime := time.Now()
	logger.Info().Msg("Creating new WebRTC peer connection")

	// Create a setting engine
	settingEngine := webrtc.SettingEngine{}

	// Configure to include loopback candidates
	// settingEngine.SetNAT1To1IPs([]string{"127.0.0.1"}, webrtc.ICECandidateTypeHost)

	// Include all network types including local ones
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeUDP6,
		webrtc.NetworkTypeTCP4,
		webrtc.NetworkTypeTCP6,
	})

	// Create the API with these settings
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	// Define configuration for WebRTC peer connection
	config := webrtc.Configuration{
		ICECandidatePoolSize: 5,
	}

	// Only add ICE servers if explicitly enabled
	if useIceServers {
		// Define ICE servers (STUN server for NAT traversal)
		stunServers := []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
			"stun:stun2.l.google.com:19302",
			"stun:stun3.l.google.com:19302",
			"stun:stun4.l.google.com:19302",
		}

		// Add ICE servers to config
		config.ICEServers = []webrtc.ICEServer{
			{
				URLs: stunServers,
			},
		}
		config.ICETransportPolicy = webrtc.ICETransportPolicyAll

		logger.Debug().
			Strs("stunServers", stunServers).
			Msg("Configured ICE servers")
	} else {
		logger.Debug().Msg("ICE servers disabled, using direct connections only")
	}

	// Create peer connection with your existing config
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create peer connection")
	}

	logger.Debug().
		Dur("duration", time.Since(startTime)).
		Msg("Peer connection created")

	// Set ICE connection state handler for debugging
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		stateLogger := logger.With().
			Str("state", connectionState.String()).
			Str("type", "ice").
			Logger()

		stateLogger.Info().Msg("ICE connection state changed")

		switch connectionState {
		case webrtc.ICEConnectionStateChecking:
			stateLogger.Debug().Msg("Checking ICE candidates")
		case webrtc.ICEConnectionStateConnected:
			stateLogger.Info().Msg("ICE connection established successfully")
		case webrtc.ICEConnectionStateFailed:
			stateLogger.Error().Msg("ICE connection failed - consider checking network/firewall")
		case webrtc.ICEConnectionStateDisconnected:
			stateLogger.Warn().Msg("ICE connection disconnected")
		}
	})

	// Add handler for ICE candidate events
	// REMOVED: This is now handled dynamically in websocket.go after WS connection
	/*
		peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate == nil {
				logger.Debug().Msg("ICE candidate gathering completed")
				return
			}

			candLogger := logger.With().
				Str("type", "iceCandidate").
				Logger()

			if candidate.Address != "" {
				candLogger.Info().
					Str("candidateType", string(candidate.Typ)).
					Str("candidateProtocol", string(candidate.Protocol)).
					Str("candidateAddress", candidate.Address).
					Int("candidatePort", int(candidate.Port)).
					Int("candidatePriority", int(candidate.Priority)).
					Int("candidateComponent", int(candidate.Component)).
					Msg("Gathered ICE candidate")
			}
		})
	*/

	// Set ICE gathering state handler
	peerConnection.OnICEGatheringStateChange(func(state webrtc.ICEGathererState) {
		logger.Debug().
			Str("state", state.String()).
			Str("type", "iceGathering").
			Msg("ICE gathering state changed")
	})

	// Set connection state handler for debugging
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		stateLogger := logger.With().
			Str("state", state.String()).
			Str("type", "connection").
			Logger()

		stateLogger.Info().Msg("Connection state changed")

		switch state {
		case webrtc.PeerConnectionStateConnected:
			stateLogger.Info().Msg("WebRTC connection established successfully")
		case webrtc.PeerConnectionStateFailed:
			stateLogger.Error().Msg("WebRTC connection failed")
		case webrtc.PeerConnectionStateClosed:
			stateLogger.Info().Msg("WebRTC connection closed")
		case webrtc.PeerConnectionStateDisconnected:
			stateLogger.Warn().Msg("WebRTC connection disconnected")
		}
	})

	// Set negotiation needed handler
	peerConnection.OnNegotiationNeeded(func() {
		logger.Debug().Msg("Negotiation needed for peer connection")
	})

	logger.Info().
		Dur("setupTime", time.Since(startTime)).
		Msg("WebRTC peer connection configured successfully")

	return peerConnection, nil
}
