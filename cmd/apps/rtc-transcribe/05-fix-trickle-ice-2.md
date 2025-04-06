# Fixing WebRTC Connectivity in rtc-transcribe: Understanding and Implementing Trickle ICE

## 1. Introduction and Problem Overview

The `rtc-transcribe` application is designed to capture audio from a user's browser via WebRTC, stream it to a Go server, transcribe it with OpenAI's Whisper API, and return the transcription results via Server-Sent Events (SSE). However, the application is currently experiencing connection failures with the error:

```
Failed to start transcription: navigator.mediaDevices is undefined.
```

This occurs due to two separate issues:

1. **Browser API Availability**: Many browsers don't expose WebRTC APIs unless the page is served over HTTPS or from localhost.
2. **Missing Trickle ICE Implementation**: Even when the APIs are available, the WebRTC connection fails to establish because the current implementation lacks proper ICE candidate exchange (Trickle ICE).

We've addressed the first issue by adding feature detection and error handling in the browser. This document focuses on the second, more complex issue: **implementing Trickle ICE to make WebRTC connections work reliably**.

## 2. WebRTC and ICE: Technical Background

### 2.1 WebRTC Architecture Overview

WebRTC (Web Real-Time Communication) is a set of protocols and APIs that enable direct peer-to-peer communication between browsers or between browsers and servers. It handles media streams (audio, video) as well as data channels. A simplified WebRTC architecture consists of:

1. **Signaling**: The process of coordinating connection establishment between peers
2. **ICE Framework**: For NAT traversal and connectivity establishment
3. **DTLS**: For securing the connection
4. **SRTP/SRTCP**: For secure media transport
5. **SDP**: Session Description Protocol for negotiating media capabilities

### 2.2 The ICE Protocol

**ICE (Interactive Connectivity Establishment)** is a protocol that helps WebRTC peers find ways to communicate through NATs and firewalls. The ICE process involves:

1. **Gathering Candidates**: Each peer identifies possible address/port combinations (called "candidates") where it can receive data. These include:
   - **Host Candidates**: Local IP addresses/ports
   - **Server Reflexive Candidates**: Public-facing IP/port combinations obtained through STUN servers
   - **Relay Candidates**: TURN server addresses used as a last resort for particularly restrictive NATs

2. **Exchanging Candidates**: Each peer must inform the other about all of its candidates.

3. **Connectivity Checks**: Peers test each possible candidate pair to find working combinations.

4. **Selection and Nomination**: The best working pair is selected for media transmission.

### 2.3 The SDP Exchange Process

WebRTC uses SDP (Session Description Protocol) for describing the session parameters. The basic SDP exchange follows an offer/answer model:

1. The initiating peer (offerer) creates an SDP "offer" containing its media capabilities.
2. The receiving peer (answerer) responds with an SDP "answer" containing its compatible capabilities.

### 2.4 Trickle ICE vs. Vanilla ICE

There are two approaches to handling ICE candidates in WebRTC:

#### Vanilla ICE (Traditional Method)
1. Gather **all** ICE candidates first
2. Include them in the SDP offer/answer
3. Perform a single exchange of complete SDPs with all candidates embedded

This approach is simple but introduces latency, as peers must wait for all candidates to be gathered before establishing a connection.

#### Trickle ICE (Enhanced Method)
1. Exchange initial SDP offer/answer **without** waiting for all candidates
2. Send candidates individually as they're discovered using a secondary signaling channel
3. Add received candidates to the peer connection as they arrive

Trickle ICE dramatically improves connection setup time and reliability, particularly in complex network environments.

## 3. Current Implementation Analysis

### 3.1 How rtc-transcribe Currently Handles WebRTC

Our `rtc-transcribe` application currently implements a partial WebRTC setup:

#### 3.1.1 Client-Side (static/index.html)
```javascript
// Create WebRTC peer connection
peerConnection = new RTCPeerConnection(rtcConfig);

// Log ICE candidate events, but doesn't send them to the server
peerConnection.onicecandidate = (event) => {
  if (event.candidate) {
    console.log('New ICE candidate gathered:', {
      candidate: event.candidate.candidate,
      sdpMid: event.candidate.sdpMid,
      sdpMLineIndex: event.candidate.sdpMLineIndex
    });
  } else {
    console.log('ICE candidate gathering completed');
  }
};

// Add the audio track to the peer connection
mediaStream.getAudioTracks().forEach(track => {
  peerConnection.addTrack(track, mediaStream);
});

// Create and set the local description (offer)
const offer = await peerConnection.createOffer();
await peerConnection.setLocalDescription(offer);

// Send the offer to the server via HTTP
const response = await fetch('/offer', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    sdp: peerConnection.localDescription.sdp,
    type: peerConnection.localDescription.type
  })
});

// Get the answer from the server
const answerData = await response.json();
await peerConnection.setRemoteDescription(new RTCSessionDescription({
  type: 'answer',
  sdp: answerData.sdp
}));
```

#### 3.1.2 Server-Side (webrtc/signaling.go and webrtc/peer.go)
```go
// HandleOffer receives the client's SDP offer via HTTP
func HandleOffer(w http.ResponseWriter, r *http.Request) {
  // Decode the client's SDP offer
  var sdp SDPExchange
  if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
    // Error handling...
  }
  
  // Create a new peer connection
  peerConn, err := CreatePeerConnection(UseIceServers)
  
  // Set the remote description (client's offer)
  offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp.SDP}
  if err := peerConn.SetRemoteDescription(offer); err != nil {
    // Error handling...
  }
  
  // Set up audio track handler
  SetupAudioTrackHandler(peerConn)
  
  // Create and set the local description (answer)
  answer, err := peerConn.CreateAnswer(nil)
  if err := peerConn.SetLocalDescription(answer); err != nil {
    // Error handling...
  }
  
  // Send the answer back to the client
  resp := SDPExchange{
    SDP:  peerConn.LocalDescription().SDP,
    Type: "answer",
  }
  json.NewEncoder(w).Encode(resp)
}

// In peer.go
peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
  if candidate == nil {
    logger.Debug().Msg("ICE candidate gathering completed")
    return
  }

  // Log the candidate but don't send it anywhere
  candLogger.Info().
    Str("candidateType", string(candidate.Typ)).
    Str("candidateProtocol", string(candidate.Protocol)).
    Str("candidateAddress", candidate.Address).
    // Additional properties logged...
    Msg("Gathered ICE candidate")
});
```

### 3.2 What's Missing: The Trickle ICE Implementation

The current implementation has these critical gaps:

1. **No Transport for ICE Candidates**: While candidates are *detected* on both sides (client and server), there's no mechanism to *exchange* these candidates after the initial SDP offer/answer exchange.

2. **No Session Tracking**: The server has no way to associate incoming ICE candidates with the correct peer connection after the initial HTTP exchange.

3. **Incomplete ICE Configuration**: The ICE configuration lacks proper handling for the various ICE gathering and connection states.

### 3.3 Why It Fails

WebRTC connection establishment typically requires multiple ICE candidates to be exchanged. When using HTTP for the initial offer/answer, some critical ICE candidates might be discovered *after* this initial exchange. Without a secondary channel to send these late candidates, the WebRTC connection often fails to establish, especially:

1. On localhost, where the "best" candidates might be discovered after the initial SDP exchange
2. Through certain NATs or firewalls, where STUN-discovered candidates arrive after the initial exchange
3. In complex network environments, where multiple connection paths need to be tested

## 4. Working Examples Analysis

To understand how to fix the issue, we analyzed two working examples:

### 4.1 save-to-disk Example (Manual Base64 Exchange)

The `save-to-disk` example uses a synchronous ICE approach:

1. It waits for ICE gathering to complete **before** creating the final SDP answer
2. This means all candidates are embedded in a single SDP message
3. The SDP is manually exchanged via base64-encoded strings
4. This approach works but introduces latency and a poor user experience

```go
// Signal exchange using manual copy-paste of base64 (save-to-disk example)
// Create a channel to signal when gathering is complete
gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

// Set the remote description
if err = peerConnection.SetRemoteDescription(offer); err != nil {
  // Error handling...
}

// Create answer
answer, err := peerConnection.CreateAnswer(nil)
if err != nil {
  // Error handling...
}

// Sets the LocalDescription and starts gathering ICE candidates
if err = peerConnection.SetLocalDescription(answer); err != nil {
  // Error handling...
}

// Block until ICE gathering is complete (this is the key difference)
<-gatherComplete

// Encode the local description to base64 for manual exchange
localDescriptionEncoded := base64.StdEncoding.EncodeToString([]byte(peerConnection.LocalDescription().SDP))
fmt.Println(localDescriptionEncoded)
```

### 4.2 trickle-ice Example (WebSocket-based Candidate Exchange)

The `trickle-ice` example uses the proper Trickle ICE approach:

1. It exchanges the initial SDP offer/answer over a WebSocket connection
2. It then sends individual ICE candidates as they're discovered over the same WebSocket
3. It adds received candidates to the peer connection as they arrive

```go
// Server side (trickle-ice example)
peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
  // If candidate is nil, gathering is complete
  if candidate == nil {
    return
  }
  
  // Encode the candidate to JSON
  candidateJSON, err := json.Marshal(candidate.ToJSON())
  if err != nil {
    // Error handling...
  }
  
  // Send the candidate over WebSocket
  if err = c.WriteJSON(map[string]interface{}{
    "type": "candidate",
    "payload": string(candidateJSON),
  }); err != nil {
    // Error handling...
  }
})

// Client side (trickle-ice example)
pc.onicecandidate = (event) => {
  // If candidate is null, gathering is complete
  if (!event.candidate) {
    return;
  }
  
  // Send the candidate over WebSocket
  ws.send(JSON.stringify({
    type: 'candidate',
    payload: event.candidate
  }));
};

// Handle candidates received from WebSocket
ws.onmessage = (e) => {
  const message = JSON.parse(e.data);
  
  if (message.type === 'candidate') {
    // Add the received candidate to the peer connection
    pc.addIceCandidate(new RTCIceCandidate(JSON.parse(message.payload)));
  }
};
```

## 5. Detailed Implementation Plan for rtc-transcribe

To fix the WebRTC connectivity issue in `rtc-transcribe`, we need to implement a proper Trickle ICE mechanism. Here's a comprehensive implementation plan:

### 5.1 Create a Session Management System

First, we need a way to track WebRTC sessions and associate WebSocket connections with the correct peer connections.

Create a new file: `webrtc/session.go`:

```go
package webrtc

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

// WebRTCSession represents a WebRTC session with its peer connection and related state
type WebRTCSession struct {
	ID             string
	PeerConnection *webrtc.PeerConnection
	WebSocket      *websocket.Conn
	CreatedAt      time.Time
	LastActivity   time.Time
}

// SessionManager handles tracking and cleanup of WebRTC sessions
type SessionManager struct {
	sessions map[string]*WebRTCSession
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*WebRTCSession),
	}
	
	// Start session cleanup goroutine
	go sm.cleanupSessions()
	
	return sm
}

// CreateSession creates a new WebRTC session
func (sm *SessionManager) CreateSession(id string, useIceServers bool) (*WebRTCSession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	// Create peer connection
	pc, err := CreatePeerConnection(useIceServers)
	if err != nil {
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
	return session, exists
}

// SetWebSocket associates a WebSocket connection with a session
func (sm *SessionManager) SetWebSocket(id string, ws *websocket.Conn) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	session, exists := sm.sessions[id]
	if !exists {
		return false
	}
	
	session.WebSocket = ws
	session.LastActivity = time.Now()
	
	log.Info().
		Str("sessionID", id).
		Msg("Associated WebSocket with session")
		
	return true
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(id string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if session, exists := sm.sessions[id]; exists {
		// Close peer connection if it exists
		if session.PeerConnection != nil {
			if err := session.PeerConnection.Close(); err != nil {
				log.Error().
					Err(err).
					Str("sessionID", id).
					Msg("Error closing peer connection")
			}
		}
		
		// Close WebSocket if it exists
		if session.WebSocket != nil {
			if err := session.WebSocket.Close(); err != nil {
				log.Error().
					Err(err).
					Str("sessionID", id).
					Msg("Error closing WebSocket")
			}
		}
		
		delete(sm.sessions, id)
		
		log.Info().
			Str("sessionID", id).
			Msg("Removed WebRTC session")
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

// cleanupSessions periodically removes inactive sessions
func (sm *SessionManager) cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.mutex.Lock()
		
		now := time.Now()
		var sessionsToRemove []string
		
		for id, session := range sm.sessions {
			// If session is more than 30 minutes old or inactive for 15 minutes
			if now.Sub(session.CreatedAt) > 30*time.Minute || 
				now.Sub(session.LastActivity) > 15*time.Minute {
				sessionsToRemove = append(sessionsToRemove, id)
			}
		}
		
		sm.mutex.Unlock()
		
		// Remove the flagged sessions
		for _, id := range sessionsToRemove {
			log.Info().
				Str("sessionID", id).
				Msg("Cleaning up inactive session")
			sm.RemoveSession(id)
		}
	}
}

// Initialize a global session manager
var GlobalSessionManager = NewSessionManager()
```

### 5.2 Add WebSocket Handler for ICE Candidate Exchange

Create a new file: `webrtc/websocket.go`:

```go
package webrtc

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Message types for WebSocket communication
const (
	MessageTypeCandidate = "candidate"
)

// Upgrader for HTTP to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// CandidateMessage represents an ICE candidate message
type CandidateMessage struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex uint16 `json:"sdpMLineIndex"`
}

// HandleWebSocket handles WebSocket connections for ICE candidate exchange
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}
	
	// Retrieve the session
	session, exists := GlobalSessionManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}
	
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Msg("Failed to upgrade to WebSocket")
		return
	}
	
	// Associate WebSocket with session
	if !GlobalSessionManager.SetWebSocket(sessionID, ws) {
		ws.Close()
		return
	}
	
	// Setup ICE candidate handler to send candidates to client
	session.PeerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		
		// Convert ICE candidate to JSON
		candidateJSON, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Error().
				Err(err).
				Str("sessionID", sessionID).
				Msg("Failed to marshal ICE candidate")
			return
		}
		
		// Create the WebSocket message
		msg := WebSocketMessage{
			Type:    MessageTypeCandidate,
			Payload: candidateJSON,
		}
		
		// Send the candidate to the client
		if err := ws.WriteJSON(msg); err != nil {
			log.Error().
				Err(err).
				Str("sessionID", sessionID).
				Msg("Failed to send ICE candidate to client")
			return
		}
		
		log.Debug().
			Str("sessionID", sessionID).
			Str("candidateType", string(candidate.Typ)).
			Str("candidateAddress", candidate.Address).
			Int("candidatePort", int(candidate.Port)).
			Msg("Sent ICE candidate to client")
	})
	
	// Main WebSocket read loop
	go func() {
		defer func() {
			ws.Close()
			GlobalSessionManager.RemoveSession(sessionID)
		}()
		
		for {
			// Read message from WebSocket
			var msg WebSocketMessage
			if err := ws.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, 
					websocket.CloseGoingAway, 
					websocket.CloseAbnormalClosure) {
					log.Error().
						Err(err).
						Str("sessionID", sessionID).
						Msg("WebSocket read error")
				}
				break
			}
			
			// Update session activity timestamp
			GlobalSessionManager.UpdateActivity(sessionID)
			
			// Handle different message types
			switch msg.Type {
			case MessageTypeCandidate:
				var candidate CandidateMessage
				if err := json.Unmarshal(msg.Payload, &candidate); err != nil {
					log.Error().
						Err(err).
						Str("sessionID", sessionID).
						Msg("Failed to unmarshal ICE candidate")
					continue
				}
				
				// Add ICE candidate to peer connection
				if err := session.PeerConnection.AddICECandidate(webrtc.ICECandidateInit{
					Candidate:     candidate.Candidate,
					SDPMid:        &candidate.SDPMid,
					SDPMLineIndex: &candidate.SDPMLineIndex,
				}); err != nil {
					log.Error().
						Err(err).
						Str("sessionID", sessionID).
						Msg("Failed to add ICE candidate")
					continue
				}
				
				log.Debug().
					Str("sessionID", sessionID).
					Str("candidate", candidate.Candidate).
					Msg("Added ICE candidate from client")
			}
		}
	}()
}
```

### 5.3 Modify the Offer Handler to Use Session Management

Update `webrtc/signaling.go`:

```go
// HandleOffer handles incoming WebRTC offers from clients
func HandleOffer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		sessionID = time.Now().Format("20060102-150405.000000")
	}
	
	logger := log.With().
		Str("component", "WebRTCSignaling").
		Str("sessionID", sessionID).
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

	// Create a new session with peer connection
	session, err := GlobalSessionManager.CreateSession(sessionID, UseIceServers)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	
	peerConn := session.PeerConnection

	// Set the remote description (client's offer)
	remoteDescStartTime := time.Now()
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}
	if err := peerConn.SetRemoteDescription(offer); err != nil {
		logger.Error().Err(errors.Wrap(err, "failed to set remote description")).Msg("WebRTC error")
		http.Error(w, "Invalid remote description", http.StatusInternalServerError)
		GlobalSessionManager.RemoveSession(sessionID)
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
		GlobalSessionManager.RemoveSession(sessionID)
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
		GlobalSessionManager.RemoveSession(sessionID)
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
		GlobalSessionManager.RemoveSession(sessionID)
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
```

### 5.4 Update the Server Entry Point

Update `main.go` to register the WebSocket handler:

```go
// In the main.go file, update the route registration in the run function
mux.HandleFunc("/offer", webrtc_handlers.HandleOffer)
mux.HandleFunc("/ws", webrtc_handlers.HandleWebSocket)  // Add this line
mux.HandleFunc("/transcribe", sse.HandleSSE)
```

### 5.5 Update the Client-Side JavaScript

Update `static/index.html` to implement Trickle ICE on the client side:

```javascript
// Add these variables to the existing variables section
let wsConnection = null;

// Connect to WebSocket for ICE candidate exchange
function connectWebSocket() {
  // Close any existing connection
  if (wsConnection) {
    wsConnection.close();
  }
  
  // Connect to WebSocket endpoint
  wsConnection = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?id=${sessionId}`);
  
  // Handle WebSocket open event
  wsConnection.onopen = function() {
    console.log('WebSocket connection established for ICE candidates');
  };
  
  // Handle WebSocket messages (ICE candidates from server)
  wsConnection.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    // Handle ICE candidates from server
    if (message.type === 'candidate' && peerConnection) {
      const candidate = JSON.parse(message.payload);
      console.log('Received ICE candidate from server:', candidate);
      
      try {
        peerConnection.addIceCandidate(new RTCIceCandidate({
          candidate: candidate.candidate,
          sdpMid: candidate.sdpMid,
          sdpMLineIndex: candidate.sdpMLineIndex
        })).catch(err => {
          console.error('Error adding received ICE candidate:', err);
        });
      } catch (error) {
        console.error('Error parsing ICE candidate:', error);
      }
    }
  };
  
  // Handle WebSocket errors
  wsConnection.onerror = function(error) {
    console.error('WebSocket error:', error);
    connectionStatus.className = 'status disconnected';
    connectionStatus.textContent = 'WebSocket connection error';
  };
  
  // Handle WebSocket close
  wsConnection.onclose = function() {
    console.log('WebSocket connection closed');
    wsConnection = null;
  };
}

// Update the existing onicecandidate handler in startTranscription function
peerConnection.onicecandidate = (event) => {
  if (event.candidate) {
    console.log('New ICE candidate gathered:', {
      candidate: event.candidate.candidate,
      sdpMid: event.candidate.sdpMid,
      sdpMLineIndex: event.candidate.sdpMLineIndex
    });
    
    // Send the ICE candidate to the server via WebSocket
    if (wsConnection && wsConnection.readyState === WebSocket.OPEN) {
      wsConnection.send(JSON.stringify({
        type: 'candidate',
        payload: {
          candidate: event.candidate.candidate,
          sdpMid: event.candidate.sdpMid,
          sdpMLineIndex: event.candidate.sdpMLineIndex
        }
      }));
    } else {
      console.warn('WebSocket not connected, unable to send ICE candidate');
    }
  } else {
    console.log('ICE candidate gathering completed');
  }
};

// Update the startTranscription function to pass sessionId in the offer request
// Modify the fetch call when sending the offer:
const response = await fetch(`/offer?id=${sessionId}`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    sdp: peerConnection.localDescription.sdp,
    type: peerConnection.localDescription.type
  })
});

// Add a call to connectWebSocket before creating the peer connection
connectWebSocket();

// Update the stopTranscription function to close the WebSocket
function stopTranscription() {
  // Existing cleanup code...
  
  // Close the WebSocket connection
  if (wsConnection) {
    wsConnection.close();
    wsConnection = null;
  }
  
  // Existing UI update code...
}
```

### 5.6 Add Dependencies

Update the project dependencies to include the gorilla/websocket package:

```
go get github.com/gorilla/websocket
```

## 6. Testing and Verification

To verify the implementation works correctly, follow these testing steps:

### 6.1 Local Testing

1. Build and run the server:
   ```bash
   cd go-go-labs
   go run cmd/apps/rtc-transcribe/main.go
   ```

2. Open the application in a browser: `http://localhost:8080`

3. Check browser console for WebRTC and WebSocket connection logs

4. Check server logs for ICE candidate exchange logs

### 6.2 Connection State Verification

1. Watch the ICE connection state changes:
   - Browser: `peerConnection.iceConnectionState` should progress from `new` to `checking` to `connected`
   - Server: ICE connection state logs should show the same progression

2. Verify audio track reception:
   - Server logs should show "Remote track received" and "Starting Opus audio processing"
   - Transcriptions should begin to appear via the SSE connection

### 6.3 Common Issues and Troubleshooting

1. **WebSocket Connection Issues**:
   - Check for CORS errors in browser console
   - Verify the WebSocket endpoint is accessible
   - Ensure proper websocket URL construction (ws:// vs wss://)

2. **ICE Candidate Exchange Issues**:
   - Check browser console for ICE candidate gathering
   - Verify candidates are being sent and received via WebSocket
   - Look for "Added ICE candidate" logs on the server

3. **Session Management Issues**:
   - Ensure sessionId is consistent between offer and WebSocket connections
   - Look for "Created new WebRTC session" and "Associated WebSocket with session" logs

4. **Peer Connection Issues**:
   - Check for ICE connection state "failed" logs
   - Verify STUN servers are accessible if using ICE servers
   - Try both with and without ICE servers flag

## 7. Conclusion and Future Enhancements

By implementing the Trickle ICE mechanism with WebSockets, we've addressed the core issue preventing WebRTC connections from establishing in the `rtc-transcribe` application. This approach:

1. **Improves Connection Success Rate**: By exchanging ICE candidates as they're discovered
2. **Reduces Connection Setup Time**: By starting connectivity checks earlier
3. **Works Reliably on Localhost**: By properly exchanging host candidates
4. **Handles NAT Traversal**: By supporting the exchange of server-reflexive and relay candidates

### Potential Future Enhancements

1. **Connection Monitoring**: Add heartbeats to detect and recover from silent connection failures
2. **STUN/TURN Server Configuration**: Allow configurable ICE server settings
3. **Reconnection Logic**: Add automatic reconnection when WebSocket or WebRTC connections fail
4. **Secure Signaling**: Add authentication to the WebSocket and offer endpoints
5. **Network Diagnostics**: Enhance the diagnostics tool to provide more information about ICE candidates and connectivity

This implementation adheres to the WebRTC specification and follows industry best practices for Trickle ICE, ensuring robust real-time communication across different network environments. 