# Technical Deep Dive: Fixing WebRTC Connection Failures in `rtc-transcribe` via Trickle ICE

**Document Purpose:** This document explains why the WebRTC connection in the `rtc-transcribe` application often fails and provides a detailed technical plan for a developer to implement the necessary fix using the standard Trickle ICE mechanism with WebSockets.

**Target Audience:** A senior software developer who may be new to the intricacies of WebRTC signaling but is comfortable with Go and JavaScript development.

## 1. The Problem: Why Doesn't the Connection Work Reliably?

Users of the `rtc-transcribe` application experience frequent failures when attempting to establish the WebRTC connection needed to stream audio. While the initial handshake over HTTP seems to succeed, the audio stream itself often fails to start. This happens even when running the client and server on the same machine (`localhost`).

The root cause lies in **incomplete WebRTC signaling**, specifically the failure to properly exchange **ICE candidates** between the browser client and the Go server after the initial Session Description Protocol (SDP) offer/answer exchange.

## 2. WebRTC Signaling Fundamentals: Beyond the Initial Handshake

To understand the fix, we need to grasp the core components of establishing a WebRTC peer-to-peer connection:

*   **SDP (Session Description Protocol):** This protocol describes the *multimedia capabilities* of a peer.
    *   **Offer:** The initiating peer (usually the browser client in our case) creates an SDP offer detailing the codecs it supports (e.g., Opus audio), security parameters, etc.
    *   **Answer:** The receiving peer (our Go server) responds with an SDP answer, confirming the chosen codec and its own capabilities.
    *   `rtc-transcribe` currently handles this offer/answer exchange correctly using the `/offer` HTTP endpoint.

*   **ICE (Interactive Connectivity Establishment):** This framework finds the **best network path** for media to flow directly between peers. Modern networks are complex (NATs, firewalls), making direct connections tricky. ICE solves this by:
    *   **Gathering Candidates:** Each peer uses STUN and potentially TURN servers to discover potential IP address and port pairs (candidates) through which it might be reachable.
        *   **Host Candidates:** Direct IP/ports on the local machine (e.g., `192.168.1.x`, `127.0.0.1`).
        *   **Server Reflexive Candidates (srflx):** Public IP/port as seen by a STUN server (after NAT traversal).
        *   **Relay Candidates (relay):** IP/port on a TURN server, used as a last resort if direct connection fails (media is relayed through the TURN server).
    *   **Connectivity Checks:** Peers exchange their gathered candidates and perform checks to find a working candidate pair.

*   **The Crucial Link:** The SDP offer/answer defines *what* media to send, while ICE determines *how* (which network path) to send it. **Both peers need to know the other peer's viable ICE candidates to establish the media path.**

## 3. The Problem with Simple Offer/Answer: Introducing Trickle ICE

The initial SDP offer and answer often *don't* contain all the necessary ICE candidates. Why?

1.  **Asynchronous Gathering:** ICE candidate gathering can take time, especially involving STUN/TURN servers. Waiting for *all* candidates before sending the answer would introduce significant initial connection delay.
2.  **Efficiency:** Sending potentially dozens of candidates within the initial SDP bloats the message.

The standard solution is **Trickle ICE**:

*   Peers exchange the initial SDP offer/answer quickly (often containing *no* candidates or only preliminary ones).
*   **Asynchronously:** As each peer discovers its own ICE candidates, it "trickles" them (sends them immediately) to the other peer using a pre-established signaling channel.
*   The receiving peer adds these incoming candidates to its remote peer description using `addIceCandidate`.
*   This allows the connectivity checks to begin much earlier and proceed in parallel with candidate gathering.

**How is Trickle ICE usually implemented?**

While the initial SDP might be exchanged via HTTP POST (like in `rtc-transcribe`), the subsequent, asynchronous trickling of candidates typically requires a **persistent, bidirectional signaling channel**. **WebSockets** are the standard choice for this.

## 4. Analysis of `rtc-transcribe`: What's Missing?

Let's examine `rtc-transcribe` in light of Trickle ICE:

1.  **Initial SDP Exchange:** Uses HTTP `/offer`. This part is functional.
2.  **Server Candidate Gathering:** `webrtc/peer.go` sets `peerConnection.OnICECandidate`. This correctly *detects* candidates found by the Go server. **Problem:** It only *logs* these candidates; it **never sends them** to the browser client.
3.  **Client Candidate Gathering:** `static/index.html` sets `peerConnection.onicecandidate`. This correctly *detects* candidates found by the browser. **Problem:** It only *logs* these candidates; it **never sends them** to the Go server.
4.  **Signaling Channel for Trickling:** There is **no WebSocket or other persistent channel** established between the client and server for exchanging these candidates after the initial HTTP `/offer` response.
5.  **Adding Remote Candidates:** Consequently, neither the client (`addIceCandidate`) nor the server (`AddICECandidate`) has code to *receive* candidates from the peer and add them to the connection process.

**Conclusion:** `rtc-transcribe` performs only half of the required signaling (SDP offer/answer). It completely omits the crucial, standard Trickle ICE mechanism for exchanging candidates. Without this, the peers cannot determine a valid network path for the audio media, leading to connection failure.

## 5. Detailed Plan to Implement Trickle ICE in `rtc-transcribe`

This plan outlines the steps to add a WebSocket-based signaling channel for Trickle ICE.

### 5.1 Backend Modifications (Go)

**Dependencies:**

*   Add a WebSocket library: `go get github.com/gorilla/websocket`

**Data Structures (Consider placing in `webrtc` package or a new `signaling` package):**

```go
import (
	"sync"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// Represents a message exchanged over WebSocket for signaling
type WebSocketMessage struct {
	Type    string      `json:"type"`    // e.g., "candidate", "disconnect"
	Payload interface{} `json:"payload"` // Can be an ICE candidate, error message, etc.
}

// Thread-safe storage for active WebSocket connections, keyed by SessionID
var (
	wsConnections = make(map[string]*websocket.Conn)
	wsConnMutex   sync.RWMutex
)

// Thread-safe storage for active PeerConnections, keyed by SessionID
// Needed to find the PC when a candidate arrives via WebSocket
var (
	peerConnections = make(map[string]*webrtc.PeerConnection)
	pcMutex         sync.RWMutex
)

// Helper to safely add/remove/get WS connections
func AddWsConn(sessionID string, conn *websocket.Conn) {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()
	wsConnections[sessionID] = conn
}
// ... (Add GetWsConn, RemoveWsConn helpers)

// Helper to safely add/remove/get PeerConnections
func AddPeerConn(sessionID string, pc *webrtc.PeerConnection) {
	pcMutex.Lock()
	defer pcMutex.Unlock()
	peerConnections[sessionID] = pc
}
// ... (Add GetPeerConn, RemovePeerConn helpers)

// Function to send a message to a specific client
func SendWsMessage(sessionID string, message WebSocketMessage) error {
    wsConnMutex.RLock()
    conn, ok := wsConnections[sessionID]
    wsConnMutex.RUnlock()

    if !ok {
        return errors.New("WebSocket connection not found for sessionID: " + sessionID)
    }

    wsConnMutex.Lock() // Lock for writing
    defer wsConnMutex.Unlock()
    err := conn.WriteJSON(message)
    if err != nil {
        // Handle error, potentially remove connection
        log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to write WebSocket message")
        // Consider removing conn here if write fails permanently
        // delete(wsConnections, sessionID) 
        return err
    }
    return nil
}

// Cleanup function for when a session ends
func CleanupSession(sessionID string) {
    log.Info().Str("sessionID", sessionID).Msg("Cleaning up session")
    
    // Close and remove PeerConnection
    pcMutex.Lock()
    pc, okPC := peerConnections[sessionID]
    if okPC {
        delete(peerConnections, sessionID)
    }
    pcMutex.Unlock() // Unlock before potentially long Close()
    if okPC && pc != nil {
         if err := pc.Close(); err != nil {
            log.Error().Err(err).Str("sessionID", sessionID).Msg("Error closing PeerConnection")
        }
    }

    // Close and remove WebSocket connection
    wsConnMutex.Lock()
    conn, okWS := wsConnections[sessionID]
    if okWS {
        delete(wsConnections, sessionID)
    }
    wsConnMutex.Unlock() // Unlock before Close()
     if okWS && conn != nil {
        conn.Close()
    }
}
```

**Changes:**

1.  **Create WebSocket Upgrader:**
    ```go
    var upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            // Allow connections from any origin for now
            // TODO: Restrict this in production
            return true
        },
    }
    ```

2.  **Create WebSocket Handler (`/ws`):**
    *   Define a new HTTP handler function (e.g., `HandleWebSocket`).
    *   In the handler:
        *   Upgrade the HTTP connection: `conn, err := upgrader.Upgrade(w, r, nil)`
        *   Handle upgrade errors.
        *   Extract `sessionID` from the request URL query params (`r.URL.Query().Get("id")`). Validate it.
        *   Add the connection to the registry: `AddWsConn(sessionID, conn)`
        *   **Start a Read Loop (Goroutine):** Launch a goroutine to continuously read messages from this specific client's WebSocket connection.
            ```go
            go func(sid string, c *websocket.Conn) {
                defer CleanupSession(sid) // Ensure cleanup when loop exits
                for {
                    var msg WebSocketMessage
                    err := c.ReadJSON(&msg) // Use ReadJSON for simplicity
                    if err != nil {
                         if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                             log.Error().Err(err).Str("sessionID", sid).Msg("WebSocket read error")
                         } else {
                            log.Info().Err(err).Str("sessionID", sid).Msg("WebSocket closed")
                         }
                        break // Exit loop on error/close
                    }

                    switch msg.Type {
                    case "candidate":
                        HandleCandidateMessage(sid, msg.Payload)
                    // Add other message types if needed (e.g., "disconnect")
                    default:
                        log.Warn().Str("type", msg.Type).Str("sessionID", sid).Msg("Received unknown WebSocket message type")
                    }
                }
            }(sessionID, conn)
            ```
    *   Register this handler in `main.go`: `mux.HandleFunc("/ws", your_package.HandleWebSocket)`

3.  **Implement `HandleCandidateMessage`:**
    ```go
    func HandleCandidateMessage(sessionID string, payload interface{}) {
        pcMutex.RLock()
        pc, ok := peerConnections[sessionID]
        pcMutex.RUnlock() // Unlock before potentially long AddIceCandidate

        if !ok || pc == nil {
            log.Warn().Str("sessionID", sessionID).Msg("Received candidate for unknown/inactive PeerConnection")
            return
        }

        // Use map[string]interface{} for flexible JSON unmarshalling
        candidateMap, ok := payload.(map[string]interface{})
        if !ok {
             log.Error().Str("sessionID", sessionID).Interface("payload", payload).Msg("Invalid candidate payload format")
             return
        }

        // Marshal the map back to JSON bytes
        candidateBytes, err := json.Marshal(candidateMap)
         if err != nil {
             log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to marshal candidate payload")
             return
         }

        // Unmarshal into webrtc.ICECandidateInit
        var candidateInit webrtc.ICECandidateInit
        if err := json.Unmarshal(candidateBytes, &candidateInit); err != nil {
            log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to unmarshal candidate payload into ICECandidateInit")
            return
        }


        log.Debug().Str("sessionID", sessionID).Str("candidate", candidateInit.Candidate).Msg("Received candidate from client")

        // Add the candidate
        if err := pc.AddICECandidate(candidateInit); err != nil {
            // Important: Handle error if remote description isn't set yet
             if strings.Contains(err.Error(), "invalid state to add ICE candidate") && pc.RemoteDescription() == nil {
                 log.Warn().Str("sessionID", sessionID).Msg("Received candidate before remote description set. Buffering is needed (TODO).")
                 // TODO: Implement buffering for candidates received early
             } else {
                log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to add client ICE candidate")
             }
        }
    }
    ```

4.  **Modify `/offer` Handler (`webrtc/signaling.go`):**
    *   **Get `sessionID`:** The client needs to send its `sessionID` with the offer request. Modify the client JS and the `/offer` handler to expect/extract this ID (e.g., from a custom header or the JSON body).
    *   **Store PeerConnection:** After `peerConnection, err := CreatePeerConnection(...)` succeeds, store it: `AddPeerConn(sessionID, peerConnection)`.
    *   **Associate `sessionID` with PeerConnection:** The `OnICECandidate` handler needs the `sessionID` to send the candidate to the correct WebSocket. Pass the `sessionID` into `CreatePeerConnection` or set it on the PeerConnection object somehow (less ideal) or modify `OnICECandidate` to look it up. A simple approach is to pass it to `CreatePeerConnection` and capture it in the closure.

5.  **Modify `CreatePeerConnection` and `OnICECandidate` (`webrtc/peer.go`):**
    *   Change signature: `func CreatePeerConnection(useIceServers bool, sessionID string) (*webrtc.PeerConnection, error)`
    *   Modify the `peerConnection.OnICECandidate` handler:
        ```go
        peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
            if candidate == nil {
                log.Debug().Str("sessionID", sessionID).Msg("Server ICE candidate gathering completed")
                return
            }

            candJSON := candidate.ToJSON() // Get the candidate in JSON init format
            log.Info().Str("sessionID", sessionID).Str("candidate", candJSON.Candidate).Msg("Server gathered ICE candidate")

            // Prepare message to send via WebSocket
            message := WebSocketMessage{
                Type:    "candidate",
                Payload: candJSON, // Send the JSON representation
            }

            // Send the message to the specific client's WebSocket
            if err := SendWsMessage(sessionID, message); err != nil {
                log.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to send candidate over WebSocket")
                // Consider implications: if WS fails, PC might be useless
            }
        })
        ```

6.  **Handle Race Conditions (Buffering - TODO):** As noted in `HandleCandidateMessage`, it's possible to receive a candidate *before* the remote description (offer/answer) is set, causing `AddICECandidate` to fail. A robust solution involves temporarily buffering these early candidates and applying them immediately after `SetRemoteDescription` completes. This is left as a TODO for simplicity but is important for production.

7.  **Cleanup:** Ensure `CleanupSession(sessionID)` is called reliably when a client disconnects (WebSocket read loop exits) or the session ends for other reasons.

### 5.2 Frontend Modifications (JavaScript - `static/index.html`)

**Changes:**

1.  **Generate `sessionID` Earlier:** Ensure `sessionId` is generated right at the start before any network requests. (Already done).

2.  **Establish WebSocket Connection:**
    *   Add a state variable: `let ws = null;`
    *   Create a function to connect:
        ```javascript
        function connectWebSocket() {
            // Ensure sessionID is generated
            if (!sessionId) {
                console.error("Session ID not generated!");
                showError("Internal error: Missing session ID.");
                return;
            }

            // Construct WebSocket URL (use wss:// for https)
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${wsProtocol}//${window.location.host}/ws?id=${sessionId}`;
            console.log(`Connecting WebSocket to: ${wsUrl}`);

            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                console.log('WebSocket connection established');
                // You could update the UI status here if desired
            };

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log('WebSocket message received:', message);

                    switch (message.type) {
                        case 'candidate':
                            if (message.payload) {
                                handleRemoteCandidate(message.payload);
                            } else {
                                console.warn("Received candidate message with null payload");
                            }
                            break;
                        // Handle other message types from server if needed
                        default:
                            console.log("Received unhandled message type:", message.type);
                    }
                } catch (error) {
                    console.error('Failed to parse WebSocket message or handle it:', error);
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                showError('Signaling connection error. Please refresh.');
                // Consider attempting reconnection here
            };

            ws.onclose = (event) => {
                console.log('WebSocket connection closed:', event.code, event.reason);
                ws = null;
                // Update UI, potentially try reconnecting if closure was unexpected
                if (!stopBtn.disabled) { // Only show error if we weren't manually stopping
                   showError('Signaling connection lost. Please refresh.');
                }
            };
        }
        ```

3.  **Call `connectWebSocket`:** Call this function *after* the initial `/offer` HTTP request succeeds and *before* you expect to receive candidates from the server. A good place is right after `await peerConnection.setRemoteDescription(answer);` in `startTranscription`.

4.  **Modify `startTranscription` to Send `sessionID`:**
    *   When sending the POST request to `/offer`, include the `sessionId`:
        ```javascript
        // Send the offer to the server
        const response = await fetch('/offer', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sdp: peerConnection.localDescription.sdp,
            type: peerConnection.localDescription.type,
            sessionId: sessionId // Include session ID
          })
        });
        ```
    *   (Remember to update the Go `/offer` handler to read this `sessionId`).

5.  **Modify `peerConnection.onicecandidate`:**
    ```javascript
    peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
            console.log('Browser gathered ICE candidate:', event.candidate.candidate);
            if (ws && ws.readyState === WebSocket.OPEN) {
                const message = {
                    type: "candidate",
                    payload: event.candidate.toJSON() // Send the JSON representation
                };
                ws.send(JSON.stringify(message));
                console.log("Sent candidate to server via WebSocket");
            } else {
                console.warn('WebSocket not open. Cannot send candidate.');
                // TODO: Buffer candidate to send when WS opens?
            }
        } else {
            console.log('Browser ICE candidate gathering completed');
        }
    };
    ```

6.  **Implement `handleRemoteCandidate`:**
    ```javascript
     let candidateBuffer = []; // Buffer for early candidates
     let isRemoteDescriptionSet = false; // Flag

     async function handleRemoteCandidate(candidatePayload) {
         try {
            const candidate = new RTCIceCandidate(candidatePayload);
            console.log("Received remote candidate from server:", candidate.candidate);

             if (!peerConnection || !isRemoteDescriptionSet) {
                console.log("PeerConnection not ready or remote description not set, buffering candidate.");
                candidateBuffer.push(candidate);
                return;
            }

            await peerConnection.addIceCandidate(candidate);
            console.log("Successfully added remote ICE candidate.");
         } catch (error) {
             console.error('Error adding received ICE candidate:', error);
         }
     }

    // Add this function to process buffered candidates
     async function processCandidateBuffer() {
        while(candidateBuffer.length > 0) {
            const candidate = candidateBuffer.shift();
             console.log("Processing buffered remote candidate:", candidate.candidate);
            try {
                 await peerConnection.addIceCandidate(candidate);
                 console.log("Successfully added buffered remote ICE candidate.");
            } catch (error) {
                console.error('Error adding buffered ICE candidate:', error);
            }
        }
     }
    ```
    *   **Call `processCandidateBuffer`:** After `setRemoteDescription` succeeds, set `isRemoteDescriptionSet = true;` and immediately call `processCandidateBuffer();`.
        ```javascript
         // Set the remote description (answer)
         await peerConnection.setRemoteDescription(answer);
         console.log("Remote description set successfully.");
         isRemoteDescriptionSet = true; // Set the flag
         processCandidateBuffer(); // Process any buffered candidates
         
         // Connect WebSocket *after* setting remote description seems safer
         connectWebSocket(); 
        ```

7.  **Cleanup in `stopTranscription`:**
    *   Close the WebSocket connection: `if (ws) { ws.close(); ws = null; }`
    *   Reset the buffer/flag: `candidateBuffer = []; isRemoteDescriptionSet = false;`

## 6. Expected Outcome

By implementing these changes, `rtc-transcribe` will correctly use Trickle ICE via WebSockets.

*   The client and server will exchange ICE candidates asynchronously as they are discovered.
*   Connectivity checks will start earlier and have access to all discovered candidates.
*   WebRTC connections should establish much more reliably, including on `localhost` and across networks (assuming STUN/TURN servers are configured and reachable if needed).

This aligns the application with standard WebRTC signaling practices and resolves the core reason for the connection failures. Remember to handle errors gracefully throughout the implementation. 