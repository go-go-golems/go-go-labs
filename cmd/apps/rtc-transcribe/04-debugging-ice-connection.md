# Debugging WebRTC Localhost Connection Failure (rtc-transcribe)

## 1. Purpose and Scope

This document summarizes the investigation into why the WebRTC connection between the browser client and the Go server fails, even when both are running on `localhost` and external STUN/TURN servers are disabled. The goal is to establish a stable local WebRTC audio stream for the `rtc-transcribe` application.

## 2. Situation Summary

The `rtc-transcribe` application aims to stream audio from a browser client to a Go server via WebRTC for real-time transcription. Despite disabling STUN/TURN servers, the connection starts failing.

The observed behavior is:
- The browser initiates the WebRTC connection.
- The Go server receives the SDP offer.
- Both client and server exchange ICE candidates (primarily `host` candidates representing local network interfaces).
- The ICE connection state eventually transitions to `failed` on both the client (visible in `about:webrtc` and JS console) and the server (visible in Go logs).
- Consequently, no audio data is successfully streamed from the client to the server.

## 3. Current Understanding of the Issue

- **ICE Failure:** The root cause is the failure of the ICE (Interactive Connectivity Establishment) process to find a working candidate pair for direct communication between the browser and the Go server, even though they reside on the same machine.
- **ICE is Still Necessary:** Even without external STUN/TURN servers, the ICE process is essential for WebRTC. It involves discovering local IP addresses/ports (`host` candidates) and performing connectivity checks between them.
- **Connectivity Checks Failing:** The logs indicate that while candidates are exchanged, the subsequent checks to confirm a direct path between the chosen candidate pair are failing.
- **Potential Causes:**
    - **Local Firewall:** The OS firewall might be blocking the specific UDP/TCP ports used by the ICE candidates, preventing the connectivity checks from succeeding.
    - **Network Interface Complexity:** The system has multiple network interfaces (localhost, physical, virtual/VPN/Docker). ICE might be selecting candidate pairs involving interfaces that aren't directly routable between the browser process and the Go server process.
    - **Browser/Library Issues:** Less likely, but potential subtle bugs or configuration mismatches in the browser's WebRTC implementation or the `pion/webrtc` library.

## 4. What We Tried

1.  **ICE Server Toggling:**
    - Introduced a `--use-ice-servers` command-line flag (defaulting to `false`) in `main.go`.
    - Modified `webrtc/peer.go` (`CreatePeerConnection`) to conditionally include `ICEServers` in the `webrtc.Configuration` based on the flag.
    - Updated `webrtc/signaling.go` (`HandleOffer`) to pass the flag value to `CreatePeerConnection`.
    - Modified the frontend Javascript (`static/index.html`) to check a `/ping` endpoint for the server's `useIceServers` setting and configure the `RTCPeerConnection` accordingly.
    - Updated the `/ping` endpoint in `main.go` to return the `useIceServers` status.
2.  **Enhanced Logging:**
    - Added detailed logging for ICE connection state changes, candidate gathering, and peer connection states in `webrtc/peer.go`.
    - Added logging for SDP decoding and peer connection creation steps in `webrtc/signaling.go`.
    - Added logging to confirm the setup of the audio track handler in `webrtc/audio.go` (`SetupAudioTrackHandler`).
    - **Crucially, added logging inside the `processOpusTrack` function in `webrtc/audio.go` to track:**
        - Entry into the function.
        - Entry into the RTP packet reading loop.
        - Reading of individual RTP packets.
        - Decoding steps.
        - Buffer filling and dispatching for transcription.

## 5. Key Findings / Current State

- The toggling mechanism for ICE servers works as intended; when the flag is false, STUN servers are not included in the configuration on either the client or server.
- ICE negotiation *begins* – `host` candidates are generated and exchanged via the SDP offer/answer.
- The Go logs show the server receiving the offer, creating the peer connection, setting remote/local descriptions, and sending the answer.
- The Go logs *also* show `Gathered ICE candidate` messages for its local interfaces.
- **Crucially, the connection consistently enters the `failed` state for both ICE and the overall PeerConnection.**
- The `about:webrtc` page confirms the `failed` state and shows no selected/active ICE candidate pair.
- **Based on the latest logs requested (but not yet provided in the conversation), we expect to see *no* log messages from within the `processOpusTrack`'s packet reading loop, confirming that the connection fails before any RTP data flows.**

## 6. Next Steps

1.  **Analyze Server Logs:** Carefully examine the *full* Go server logs generated after the most recent code changes (specifically the added logging in `processOpusTrack`). Confirm whether the `Remote track received` log appears, and if *any* `Read RTP packet from track` logs appear. This pinpoints *when* the failure occurs relative to track setup and data flow.
2.  **Check Local Firewall:** Temporarily disable the local OS firewall (e.g., `sudo ufw disable` on Ubuntu, or the equivalent for other OS/firewall software) and re-test the connection *without* ICE servers. If it connects, the firewall is the culprit, and specific rules will need to be added to allow the necessary local UDP/TCP traffic. Remember to re-enable the firewall afterwards.
3.  **Simplify Network Interfaces (If Possible):** If disabling the firewall doesn't help, investigate if the Go server or browser can be forced to bind to *only* the `127.0.0.1` interface to avoid complexities with other network cards. This might involve specific configurations in Pion or browser flags, requiring further research.
4.  **Review Pion ICE Configuration:** Check `pion/webrtc` documentation for specific settings related to localhost connections or ICE candidate filtering that might simplify the process when external servers are disabled.


## 5. Next Steps

1.  **Run with Latest Logging:** Execute the Go server again *without* the `--use-ice-servers` flag (ensuring it defaults to `false`).
    ```bash
    go run ./cmd/apps/rtc-transcribe/ --log-level debug
    ```
2.  **Attempt Connection:** Connect from the browser and click "Start Transcription". Wait for the connection to fail.
3.  **Analyze Go Server Logs:** Carefully examine the *full* output logs from the Go server. Specifically look for:
    - The `Remote track received` message in the `AudioTrackHandler` logs.
    - The `Starting to process Opus track` message from `OpusProcessor`.
    - Any `Read RTP packet from track` messages from `OpusProcessor`.
    - Any Opus decoding errors.
4.  **Interpret Results:**
    - **If `Remote track received` does NOT appear:** The ICE failure happens *before* the server even successfully registers the incoming audio track. The problem lies purely in the ICE negotiation phase. Focus should be on firewall rules and network interface selection/routing.
    - **If `Remote track received` appears, but `Read RTP packet from track` does NOT:** The track was signaled, but data packets are not arriving. This still points strongly to firewall/routing issues blocking the established data path.
    - **If `Read RTP packet from track` *does* appear:** Packets *are* arriving, but something else is wrong (e.g., Opus decoding issues, buffering problems, transcription service errors). This seems less likely given the current failure messages.
5.  **Based on the logs, investigate further:**
    - If firewall suspected: Temporarily disable the local firewall (e.g., `sudo ufw disable`) for testing ONLY and see if the connection succeeds. Re-enable immediately after testing. If it works, specific rules need to be added.
    - If network interfaces suspected: Analyze the candidate IPs in the logs and consider if net    - If network interfaces suspected: Analyze the candidate IPs in the logs and consider if network configuration might prevent communication between them. Try simplifying the network setup if possible (e.g., disconnecting VPNs, stopping Docker containers using host networking).