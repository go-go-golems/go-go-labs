Perfect — here's **Part 1: Project Description** and **Part 2: System Architecture** tailored for a **senior Go developer new to WebRTC**, focusing on clarity, background context, and practical implementation framing.

---

# **Part 1: Project Description**

### **Project Name:**  
**WhisperStream** — a real-time audio transcription server using WebRTC and Whisper.

### **Goal:**  
Build a Go server that receives audio streams from browser or native clients over **WebRTC**, transcribes them **in real-time using OpenAI’s Whisper** (via local model or API), and sends back the resulting text to the client **using Server-Sent Events (SSE)**.

### **Use Case Examples:**

- Live voice captioning in a browser
- Dictation interface for note-taking
- Real-time voice command processor
- Assistive transcription services

### **Why WebRTC?**  
WebRTC handles real-time audio streaming with excellent **low-latency**, **network traversal (ICE/STUN/TURN)**, **adaptive bitrate**, and **encryption (DTLS-SRTP)**. It's browser-native and perfect for high-quality, real-time audio delivery.

### **Why SSE for Text Output?**  
SSE is simple to implement with Go’s `net/http` standard library, works well for streaming text from server to client, and doesn’t require complex bidirectional protocols (unlike WebSockets or WebRTC DataChannel).

### **Tech Stack:**

| Component         | Technology        |
|------------------|-------------------|
| Audio Transport   | WebRTC (browser ↔ Go via pion/webrtc) |
| Audio Codec       | Opus (standard in WebRTC) |
| Audio Decoding    | `libopus` via Go bindings |
| STT Engine        | OpenAI Whisper (`whisper.cpp` or Whisper API) |
| Text Streaming    | Server-Sent Events (SSE) |
| Client            | Browser with JavaScript + microphone input |

---

# **Part 2: System Architecture**

The system has three key roles:

### 1. **The Client (Browser or Native App):**
- Captures microphone input
- Encodes it as Opus via WebRTC
- Connects to the server via ICE/STUN
- Sends SDP offer to the server via HTTP POST
- Starts audio stream
- Listens to transcription results via SSE

### 2. **The Server (in Go):**
- Accepts SDP offers via `/offer` endpoint
- Uses pion/webrtc to establish PeerConnection
- Receives Opus-encoded audio via WebRTC
- Decodes audio to raw PCM using `libopus`
- Streams PCM audio into Whisper (local or API)
- Sends transcription results via `/transcribe` SSE endpoint

### 3. **The Transcriber (Whisper):**
- Receives buffered PCM audio chunks (e.g., 1–5s)
- Returns transcribed text
- Works in either:
  - **Local mode**: Using `whisper.cpp` via Go bindings
  - **Remote mode**: Using OpenAI’s `/v1/audio/transcriptions` API

---

## **Flow Diagram**

```text
          [ Browser Client ]
         ┌──────────────────┐
         │  Mic Input (Opus)│
         │  WebRTC PeerConn │
         └────────┬─────────┘
                  │ SDP Offer
          HTTP POST /offer
                  ▼
         ┌────────────────────┐
         │   Go Server        │
         │ - pion/webrtc      │
         │ - /offer endpoint  │
         │ - OnTrack(audio)   │
         └───────┬────────────┘
                 │ RTP (Opus)
          Decoded via libopus
                 │ PCM
       Transcribed with Whisper
                 │ Text
         ┌───────▼────────────┐
         │ SSE: /transcribe   │
         └───────┬────────────┘
                 │
          [ EventSource in JS ]
         ┌──────────────────┐
         │ Real-Time Text UI│
         └──────────────────┘
```

---

## **Key Components and Protocols**

| Piece                  | Protocols / Tools Used                            |
|------------------------|--------------------------------------------------|
| Media negotiation      | SDP + ICE (via pion/webrtc)                      |
| NAT traversal          | STUN (e.g., Google's `stun.l.google.com:19302`)  |
| Audio streaming        | RTP over SRTP (Opus-encoded audio)               |
| Audio decoding         | Opus → PCM (`github.com/hraban/opus`)            |
| Speech recognition     | Whisper (via `whisper.cpp` or OpenAI API)       |
| Text delivery to client| HTTP Server-Sent Events                          |

---

## **Session Lifecycle Summary**

1. **Client captures audio**, creates a WebRTC offer, sends to `/offer`.
2. **Server accepts**, responds with answer, establishes secure peer connection.
3. **Client streams audio** via WebRTC to server.
4. **Server decodes Opus packets**, buffers audio in 1–5s chunks.
5. **Buffered audio is passed to Whisper**, transcribed.
6. **Transcribed text is streamed via `/transcribe` SSE** endpoint.
7. **Client receives and renders text updates** as they arrive.

---

Next up: in **Part 3**, we’ll walk through implementing this architecture in Go, step by step, including setting up pion, decoding audio, and integrating Whisper. Want me to continue with that now?
