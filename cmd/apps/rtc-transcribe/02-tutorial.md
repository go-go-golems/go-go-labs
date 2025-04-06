
# **Part 3: Implementation Tutorial**

---

## **0. Project Setup**

### **Folder Structure**

```
whisperstream/
├── main.go
├── webrtc/
│   ├── peer.go
│   ├── signaling.go
│   └── audio.go
├── transcribe/
│   ├── whisper_local.go  // or whisper_api.go
├── sse/
│   └── stream.go
├── static/
│   └── index.html
├── go.mod
└── go.sum
```

---

## **1. Install Required Dependencies**

```bash
go mod init github.com/yourname/whisperstream
go get github.com/pion/webrtc/v3
go get github.com/hraban/opus
```

If using Whisper locally:

- Install [whisper.cpp](https://github.com/ggerganov/whisper.cpp) and a Go binding (e.g., [whisper.cpp Go bindings](https://github.com/ggerganov/whisper.cpp/tree/master/bindings/go))

If using the OpenAI API:

```bash
go get github.com/go-resty/resty/v2
```

---

## **2. Signaling: Accepting WebRTC Offers**

Create `webrtc/signaling.go`:

```go
package webrtc

import (
	"encoding/json"
	"net/http"

	"github.com/pion/webrtc/v3"
)

type SDPExchange struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer"
}

func HandleOffer(w http.ResponseWriter, r *http.Request) {
	var sdp SDPExchange
	if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
		http.Error(w, "invalid SDP", 400)
		return
	}

	peerConn, err := CreatePeerConnection()
	if err != nil {
		http.Error(w, "failed to create peer", 500)
		return
	}

	// Set remote description
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}
	err = peerConn.SetRemoteDescription(offer)
	if err != nil {
		http.Error(w, "invalid remote description", 500)
		return
	}

	// Set OnTrack handler
	OnAudioTrack(peerConn)

	// Create and set answer
	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "failed to create answer", 500)
		return
	}
	err = peerConn.SetLocalDescription(answer)
	if err != nil {
		http.Error(w, "failed to set local desc", 500)
		return
	}

	// Send answer back
	resp := SDPExchange{
		SDP:  peerConn.LocalDescription().SDP,
		Type: "answer",
	}
	json.NewEncoder(w).Encode(resp)
}
```

---

## **3. WebRTC PeerConnection Setup**

In `webrtc/peer.go`:

```go
package webrtc

import "github.com/pion/webrtc/v3"

func CreatePeerConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	return webrtc.NewPeerConnection(config)
}
```

---

## **4. Handling the Incoming Audio Track**

In `webrtc/audio.go`:

```go
package webrtc

import (
	"github.com/hraban/opus"
	"github.com/pion/webrtc/v3"
	"io"
)

func OnAudioTrack(pc *webrtc.PeerConnection) {
	pc.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if track.Kind() != webrtc.RTPCodecTypeAudio {
			return
		}

		decoder, err := opus.NewDecoder(48000, 1)
		if err != nil {
			panic(err)
		}

		buf := make([]int16, 960*10) // 10 frames (200ms)
		var pcmChunk []int16

		for {
			packet, _, err := track.ReadRTP()
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			n, err := decoder.Decode(packet.Payload, buf)
			if err != nil {
				continue
			}

			pcmChunk = append(pcmChunk, buf[:n]...)

			if len(pcmChunk) > 48000*3 { // 3 seconds buffer
				go TranscribePCM(pcmChunk)
				pcmChunk = pcmChunk[:0]
			}
		}
	})
}
```

---

## **5. Transcribing Audio with Whisper**

### Option A: Local via `whisper.cpp`

In `transcribe/whisper_local.go`:

```go
package transcribe

import (
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

var model *whisper.Model

func InitModel(path string) error {
	var err error
	model, err = whisper.New(path)
	return err
}

func TranscribePCM(samples []int16) {
	ctx, _ := model.NewContext()
	float32Samples := make([]float32, len(samples))
	for i, v := range samples {
		float32Samples[i] = float32(v) / 32768.0
	}
	ctx.Process(float32Samples, nil, func(s whisper.Segment) {
		SendTextToClient(s.Text)
	}, nil)
}
```

### Option B: OpenAI Whisper API (simplified)

In `transcribe/whisper_api.go`:

```go
package transcribe

import (
	"bytes"
	"github.com/go-resty/resty/v2"
)

func TranscribePCM(samples []int16) {
	// Convert to .wav or .mp3 bytes (requires encoder)
	// then POST to OpenAI /v1/audio/transcriptions with multipart/form-data

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer YOUR_API_KEY").
		SetFileReader("file", "audio.wav", bytes.NewReader(audioBytes)).
		SetFormData(map[string]string{"model": "whisper-1"}).
		Post("https://api.openai.com/v1/audio/transcriptions")

	if err == nil {
		SendTextToClient(resp.String())
	}
}
```

---

## **6. SSE Streaming to Client**

In `sse/stream.go`:

```go
package sse

import (
	"fmt"
	"net/http"
)

var subscribers = make(map[string]http.ResponseWriter)

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // client session ID
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)
	subscribers[id] = w

	notify := r.Context().Done()
	<-notify
	delete(subscribers, id)
}

func SendTextToClient(text string) {
	for _, w := range subscribers {
		fmt.Fprintf(w, "data: %s\n\n", text)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
```

---

## **7. Minimal Frontend (Static HTML)**

In `static/index.html`:

```html
<!DOCTYPE html>
<html>
<head><title>WhisperStream</title></head>
<body>
  <button id="start">Start</button>
  <div id="out"></div>
  <script>
    const btn = document.getElementById("start");
    const out = document.getElementById("out");

    btn.onclick = async () => {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const pc = new RTCPeerConnection({ iceServers: [{ urls: "stun:stun.l.google.com:19302" }] });
      pc.addTrack(stream.getAudioTracks()[0], stream);

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      const res = await fetch("/offer", {
        method: "POST",
        body: JSON.stringify({ sdp: offer.sdp, type: offer.type }),
        headers: { "Content-Type": "application/json" }
      });
      const answer = await res.json();
      await pc.setRemoteDescription(answer);

      const es = new EventSource("/transcribe?id=browser");
      es.onmessage = e => out.innerHTML += `<p>${e.data}</p>`;
    };
  </script>
</body>
</html>
```

---

## **8. Main Entry Point**

In `main.go`:

```go
package main

import (
	"net/http"
	"your_project_path/webrtc"
	"your_project_path/sse"
	"your_project_path/transcribe"
)

func main() {
	transcribe.InitModel("ggml-base.en.bin")

	http.HandleFunc("/offer", webrtc.HandleOffer)
	http.HandleFunc("/transcribe", sse.SSEHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
```

